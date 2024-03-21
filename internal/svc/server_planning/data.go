package server_planning

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/capacity_planning"
)

func ListServer(request *Request) ([]*Server, error) {
	// 缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	// 查询云产品规划表
	var cloudProductPlanningList []*entity.CloudProductPlanning
	if err := db.Model(&entity.CloudProductPlanning{}).Where("plan_id = ?", request.PlanId).Find(&cloudProductPlanningList).Error; err != nil {
		return nil, err
	}
	if len(cloudProductPlanningList) == 0 {
		return nil, errors.New("该方案未找到关联产品")
	}
	var productIdList []int64
	dpdkCloudProductMap := make(map[int64]*entity.CloudProductPlanning)
	for _, cloudProductPlanning := range cloudProductPlanningList {
		productIdList = append(productIdList, cloudProductPlanning.ProductId)
		if strings.Contains(cloudProductPlanning.SellSpec, constant.SellSpecDPDK) {
			dpdkCloudProductMap[cloudProductPlanning.ProductId] = cloudProductPlanning
		}
	}
	// 查询云产品和角色关联表
	var nodeRoleIdList []int64
	var cloudProductNodeRoleRelList []*entity.CloudProductNodeRoleRel
	if err := db.Model(&entity.CloudProductNodeRoleRel{}).Where("product_id IN (?)", productIdList).Find(&cloudProductNodeRoleRelList).Error; err != nil {
		return nil, err
	}
	dpdkNodeRoleMap := make(map[int64][]*entity.CloudProductNodeRoleRel)
	for _, cloudProductNodeRoleRel := range cloudProductNodeRoleRelList {
		nodeRoleIdList = append(nodeRoleIdList, cloudProductNodeRoleRel.NodeRoleId)
		if _, ok := dpdkCloudProductMap[cloudProductNodeRoleRel.ProductId]; ok {
			dpdkNodeRoleMap[cloudProductNodeRoleRel.NodeRoleId] = append(dpdkNodeRoleMap[cloudProductNodeRoleRel.NodeRoleId], cloudProductNodeRoleRel)
		}
	}
	// 查询角色表
	var nodeRoleBaselineList []*entity.NodeRoleBaseline
	if err := db.Where("id IN (?)", nodeRoleIdList).Find(&nodeRoleBaselineList).Error; err != nil {
		return nil, err
	}
	// 查询角色服务器基线map
	serverBaselineMap, nodeRoleServerBaselineListMap, screenNodeRoleServerBaselineMap, nodeRoleCodeBaselineMap, err := getNodeRoleServerBaselineMap(db, nodeRoleIdList, nodeRoleBaselineList, request)
	if err != nil {
		return nil, err
	}
	// 查询混合部署方式map
	mixedNodeRoleMap, err := getMixedNodeRoleMap(db, nodeRoleIdList)
	if err != nil {
		return nil, err
	}
	// 查询已保存的服务器规划表
	var serverPlanningList []*Server
	if err = db.Model(&entity.ServerPlanning{}).Where("plan_id = ? AND node_role_id IN (?)", request.PlanId, nodeRoleIdList).Find(&serverPlanningList).Error; err != nil {
		return nil, err
	}
	var nodeRoleServerPlanningsMap = make(map[int64][]*Server)
	for _, v := range serverPlanningList {
		nodeRoleServerPlanningsMap[v.NodeRoleId] = append(nodeRoleServerPlanningsMap[v.NodeRoleId], v)
	}
	var resourcePoolList []*entity.ResourcePool
	if err = db.Model(&entity.ResourcePool{}).Where("plan_id = ? AND node_role_id IN (?)", request.PlanId, nodeRoleIdList).Find(&resourcePoolList).Error; err != nil {
		return nil, err
	}
	resourcePoolNodeRoleIdMap := make(map[int64][]*entity.ResourcePool)
	for _, resourcePool := range resourcePoolList {
		resourcePoolNodeRoleIdMap[resourcePool.NodeRoleId] = append(resourcePoolNodeRoleIdMap[resourcePool.NodeRoleId], resourcePool)
	}
	resourcePoolIdServerPlanningMap := make(map[int64]*entity.ServerPlanning)
	// 构建返回体
	var list []*Server
	nodeRoleIdNodeRoleMap := make(map[int64]*entity.NodeRoleBaseline)
	for _, nodeRoleBaseline := range nodeRoleBaselineList {
		nodeRoleIdNodeRoleMap[nodeRoleBaseline.Id] = nodeRoleBaseline
		nodeRoleServerPlannings := nodeRoleServerPlanningsMap[nodeRoleBaseline.Id]
		// 若服务器规划有保存过，则加载已保存的数据
		if len(nodeRoleServerPlannings) > 0 {
			nodeRoleResourcePoolList := resourcePoolNodeRoleIdMap[nodeRoleBaseline.Id]
			resourcePoolMap := make(map[int64]*entity.ResourcePool)
			for _, resourcePool := range nodeRoleResourcePoolList {
				resourcePoolMap[resourcePool.Id] = resourcePool
			}
			serverBaseline := screenNodeRoleServerBaselineMap[nodeRoleBaseline.Id]
			/**
			1、 如果修改了云产品规划的售卖规格，（1）之前带DPDK，现在不带，dpdkNodeRoleMap就是空的，需要去掉依赖的DPDK资源池（2）之前不带DPDK，现在带，dpdkNodeRoleMap不为空，则需要添加新的DPDK资源池
			*/
			for i, originServerPlanning := range nodeRoleServerPlannings {
				serverPlanning := &Server{}
				if originServerPlanning.ServerPlanning.OpenDpdk == constant.OpenDpdk && len(dpdkNodeRoleMap[nodeRoleBaseline.Id]) == 0 {
					continue
				}
				if util.IsBlank(request.NetworkInterface) && util.IsBlank(request.CpuType) {
					serverPlanning = originServerPlanning
					serverPlanning.ServerBomCode = serverBaselineMap[originServerPlanning.ServerPlanning.ServerBaselineId].BomCode
					serverPlanning.ServerArch = serverBaselineMap[originServerPlanning.ServerPlanning.ServerBaselineId].Arch
				} else {
					serverPlanning.PlanId = request.PlanId
					serverPlanning.NodeRoleId = nodeRoleBaseline.Id
					serverPlanning.Number = nodeRoleBaseline.MinimumNum
					// 列表加载机型
					serverPlanning.ServerBaselineId = serverBaseline.Id
					serverPlanning.ServerBomCode = serverBaseline.BomCode
					serverPlanning.ServerArch = serverBaseline.Arch
					serverPlanning.MixedNodeRoleId = nodeRoleBaseline.Id
					serverPlanning.ResourcePoolId = originServerPlanning.ServerPlanning.ResourcePoolId
					serverPlanning.OpenDpdk = originServerPlanning.ServerPlanning.OpenDpdk
				}
				serverPlanning.NodeRoleName = nodeRoleBaseline.NodeRoleName
				serverPlanning.NodeRoleClassify = nodeRoleBaseline.Classify
				serverPlanning.NodeRoleAnnotation = nodeRoleBaseline.Annotation
				serverPlanning.SupportDpdk = nodeRoleBaseline.SupportDPDK
				serverPlanning.ServerBaselineList = nodeRoleServerBaselineListMap[nodeRoleBaseline.Id]
				serverPlanning.MixedNodeRoleList = mixedNodeRoleMap[nodeRoleBaseline.Id]
				resourcePool := resourcePoolMap[originServerPlanning.ServerPlanning.ResourcePoolId]
				if resourcePool != nil {
					serverPlanning.ResourcePoolName = resourcePool.ResourcePoolName
				} else {
					openDpdk := constant.CloseDpdk
					resourcePoolName := fmt.Sprintf("%s-%s-%d", nodeRoleBaseline.NodeRoleName, constant.ResourcePoolDefaultName, i+1)
					if nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeNFV {
						resourcePoolName = constant.NFVResourcePoolNameKernel
					}
					if originServerPlanning.ServerPlanning.OpenDpdk == constant.OpenDpdk {
						openDpdk = constant.OpenDpdk
						if nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeNFV {
							resourcePoolName = constant.NFVResourcePoolNameDpdk
						}
					}
					resourcePool = &entity.ResourcePool{
						PlanId:           request.PlanId,
						NodeRoleId:       nodeRoleBaseline.Id,
						ResourcePoolName: resourcePoolName,
						OpenDpdk:         openDpdk,
					}
					if err = db.Table(entity.ResourcePoolTable).Save(&resourcePool).Error; err != nil {
						log.Errorf("save resource pool error: %v", err)
						return nil, err
					}
					serverPlanning.ResourcePoolId = resourcePool.Id
					serverPlanning.ResourcePoolName = resourcePool.ResourcePoolName
				}
				list = append(list, serverPlanning)
				resourcePoolIdServerPlanningMap[serverPlanning.ResourcePoolId] = &entity.ServerPlanning{
					PlanId:           serverPlanning.PlanId,
					NodeRoleId:       serverPlanning.NodeRoleId,
					ServerBaselineId: serverPlanning.ServerBaselineId,
					MixedNodeRoleId:  serverPlanning.MixedNodeRoleId,
					Number:           serverPlanning.Number,
					OpenDpdk:         serverPlanning.OpenDpdk,
					ResourcePoolId:   serverPlanning.ResourcePoolId,
				}
			}
		} else {
			serverPlanning := &Server{}
			serverPlanning.PlanId = request.PlanId
			serverPlanning.NodeRoleId = nodeRoleBaseline.Id
			serverPlanning.Number = nodeRoleBaseline.MinimumNum
			// 列表加载机型
			serverBaseline := screenNodeRoleServerBaselineMap[nodeRoleBaseline.Id]
			serverPlanning.ServerBaselineId = serverBaseline.Id
			serverPlanning.ServerBomCode = serverBaseline.BomCode
			serverPlanning.ServerArch = serverBaseline.Arch
			serverPlanning.MixedNodeRoleId = nodeRoleBaseline.Id
			serverPlanning.NodeRoleName = nodeRoleBaseline.NodeRoleName
			serverPlanning.NodeRoleClassify = nodeRoleBaseline.Classify
			serverPlanning.NodeRoleAnnotation = nodeRoleBaseline.Annotation
			serverPlanning.SupportDpdk = nodeRoleBaseline.SupportDPDK
			serverPlanning.ServerBaselineList = nodeRoleServerBaselineListMap[nodeRoleBaseline.Id]
			serverPlanning.MixedNodeRoleList = mixedNodeRoleMap[nodeRoleBaseline.Id]
			resourcePoolList = resourcePoolNodeRoleIdMap[nodeRoleBaseline.Id]
			var resourcePool *entity.ResourcePool
			if len(resourcePoolList) > 0 {
				resourcePool = resourcePoolList[0]
			} else {
				resourcePoolName := fmt.Sprintf("%s-%s-%d", nodeRoleBaseline.NodeRoleName, constant.ResourcePoolDefaultName, 1)
				if nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeNFV {
					resourcePoolName = constant.NFVResourcePoolNameKernel
				}
				resourcePool = &entity.ResourcePool{
					PlanId:           request.PlanId,
					NodeRoleId:       nodeRoleBaseline.Id,
					ResourcePoolName: resourcePoolName,
					OpenDpdk:         constant.CloseDpdk,
				}
				if err = db.Table(entity.ResourcePoolTable).Save(&resourcePool).Error; err != nil {
					log.Errorf("save resource pool error: %v", err)
					return nil, err
				}
			}
			serverPlanning.ResourcePoolName = resourcePool.ResourcePoolName
			serverPlanning.ResourcePoolId = resourcePool.Id
			list = append(list, serverPlanning)
			resourcePoolIdServerPlanningMap[serverPlanning.ResourcePoolId] = &entity.ServerPlanning{
				PlanId:           serverPlanning.PlanId,
				NodeRoleId:       serverPlanning.NodeRoleId,
				ServerBaselineId: serverPlanning.ServerBaselineId,
				MixedNodeRoleId:  serverPlanning.MixedNodeRoleId,
				Number:           serverPlanning.Number,
				OpenDpdk:         serverPlanning.OpenDpdk,
				ResourcePoolId:   resourcePool.Id,
			}
			if len(dpdkNodeRoleMap[nodeRoleBaseline.Id]) > 0 {
				dpdkServerPlanning, err := addDpdkServerPlanning(db, request.PlanId, nodeRoleBaseline, serverBaseline, nodeRoleServerBaselineListMap, mixedNodeRoleMap, resourcePoolList, resourcePoolIdServerPlanningMap)
				if err != nil {
					return nil, err
				}
				list = append(list, dpdkServerPlanning)
			}
		}
	}
	// 计算已保存的容量规划指标
	resourcePoolCapMap, err := capacity_planning.GetResourcePoolCapMap(db, &capacity_planning.Request{PlanId: request.PlanId}, resourcePoolIdServerPlanningMap, nodeRoleCodeBaselineMap, serverBaselineMap)
	if err != nil {
		return nil, err
	}
	var resourceIdList []int64
	for i, server := range list {
		// 资源池变了，就去节点角色最小数
		number, ok := resourcePoolCapMap[server.ResourcePoolId]
		if ok {
			// 处理节点最小数
			if number < nodeRoleIdNodeRoleMap[server.NodeRoleId].MinimumNum {
				number = nodeRoleIdNodeRoleMap[server.NodeRoleId].MinimumNum
			}
			if number > server.Number {
				list[i].Number = number
			}
		} else {
			list[i].Number = nodeRoleIdNodeRoleMap[server.NodeRoleId].MinimumNum
		}
		resourceIdList = append(resourceIdList, server.ResourcePoolId)
	}
	if err = db.Table(entity.ResourcePoolTable).Where("plan_id = ? and id not in (?)", request.PlanId, resourceIdList).Delete(&entity.ResourcePool{}).Error; err != nil {
		log.Errorf("delete resource pool error: %v", err)
		return nil, err
	}
	return list, nil
}

func addDpdkServerPlanning(db *gorm.DB, planId int64, nodeRoleBaseline *entity.NodeRoleBaseline, serverBaseline *entity.ServerBaseline, nodeRoleServerBaselineListMap map[int64][]*Baseline, mixedNodeRoleMap map[int64][]*MixedNodeRole, resourcePoolList []*entity.ResourcePool, resourcePoolServerPlanningMap map[int64]*entity.ServerPlanning) (*Server, error) {
	dpdkServerPlanning := &Server{}
	var resourcePool *entity.ResourcePool
	dpdkServerPlanning.PlanId = planId
	dpdkServerPlanning.NodeRoleId = nodeRoleBaseline.Id
	dpdkServerPlanning.Number = nodeRoleBaseline.MinimumNum
	// 列表加载机型
	dpdkServerPlanning.ServerBaselineId = serverBaseline.Id
	dpdkServerPlanning.ServerBomCode = serverBaseline.BomCode
	dpdkServerPlanning.ServerArch = serverBaseline.Arch
	dpdkServerPlanning.MixedNodeRoleId = nodeRoleBaseline.Id
	dpdkServerPlanning.NodeRoleName = nodeRoleBaseline.NodeRoleName
	dpdkServerPlanning.NodeRoleClassify = nodeRoleBaseline.Classify
	dpdkServerPlanning.NodeRoleAnnotation = nodeRoleBaseline.Annotation
	dpdkServerPlanning.SupportDpdk = nodeRoleBaseline.SupportDPDK
	dpdkServerPlanning.ServerBaselineList = nodeRoleServerBaselineListMap[nodeRoleBaseline.Id]
	dpdkServerPlanning.MixedNodeRoleList = mixedNodeRoleMap[nodeRoleBaseline.Id]
	dpdkServerPlanning.OpenDpdk = constant.OpenDpdk
	if len(resourcePoolList) > 1 {
		resourcePool = resourcePoolList[len(resourcePoolList)-1]
	} else {
		resourcePoolName := fmt.Sprintf("%s-%s-%d", nodeRoleBaseline.NodeRoleName, constant.ResourcePoolDefaultName, 2)
		if nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeNFV {
			resourcePoolName = constant.NFVResourcePoolNameDpdk
		}
		resourcePool = &entity.ResourcePool{
			PlanId:           planId,
			NodeRoleId:       nodeRoleBaseline.Id,
			ResourcePoolName: resourcePoolName,
			OpenDpdk:         constant.OpenDpdk,
		}
		if err := db.Table(entity.ResourcePoolTable).Save(&resourcePool).Error; err != nil {
			log.Errorf("save resource pool error: %v", err)
			return nil, err
		}
	}
	dpdkServerPlanning.ResourcePoolName = resourcePool.ResourcePoolName
	dpdkServerPlanning.ResourcePoolId = resourcePool.Id
	resourcePoolServerPlanningMap[nodeRoleBaseline.Id] = &entity.ServerPlanning{
		PlanId:           dpdkServerPlanning.PlanId,
		NodeRoleId:       dpdkServerPlanning.NodeRoleId,
		ServerBaselineId: dpdkServerPlanning.ServerBaselineId,
		MixedNodeRoleId:  dpdkServerPlanning.MixedNodeRoleId,
		Number:           dpdkServerPlanning.Number,
		OpenDpdk:         dpdkServerPlanning.OpenDpdk,
		ResourcePoolId:   resourcePool.Id,
	}
	return dpdkServerPlanning, nil
}

func SaveServer(request *Request) error {
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		if err := CreateServerPlanning(tx, request); err != nil {
			return err
		}
		if err := tx.Model(entity.PlanManage{}).Where("id = ?", request.PlanId).Updates(&entity.PlanManage{
			BusinessPlanStage: constant.BusinessPlanningNetworkDevice,
			UpdateUserId:      request.UserId,
			UpdateTime:        time.Now(),
		}).Error; err != nil {
			return err
		}
		for _, server := range request.ServerList {
			if err := tx.Table(entity.ResourcePoolTable).Where("id = ?", server.ResourcePoolId).Update("open_dpdk", server.OpenDpdk).Error; err != nil {
				log.Errorf("update resourcePool error: %v", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func CreateServerPlanning(db *gorm.DB, request *Request) error {
	if err := checkBusiness(db, request); err != nil {
		return err
	}
	if err := db.Where("plan_id = ?", request.PlanId).Delete(&entity.ServerPlanning{}).Error; err != nil {
		return err
	}
	now := datetime.GetNow()
	var serverPlanningEntityList []*entity.ServerPlanning
	for _, v := range request.ServerList {
		serverPlanningEntityList = append(serverPlanningEntityList, &entity.ServerPlanning{
			PlanId:           request.PlanId,
			NodeRoleId:       v.NodeRoleId,
			MixedNodeRoleId:  v.MixedNodeRoleId,
			ServerBaselineId: v.ServerBaselineId,
			Number:           v.Number,
			OpenDpdk:         v.OpenDpdk,
			NetworkInterface: request.NetworkInterface,
			CpuType:          request.CpuType,
			CreateUserId:     request.UserId,
			CreateTime:       now,
			UpdateUserId:     request.UserId,
			UpdateTime:       now,
			DeleteState:      0,
			ResourcePoolId:   v.ResourcePoolId,
		})
	}
	if err := db.Create(&serverPlanningEntityList).Error; err != nil {
		return err
	}
	return nil
}

func ListServerNetworkType(request *Request) ([]string, error) {
	// 缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	serverBaselineList, serverPlanning, err := getServerType(db, request)
	if err != nil {
		return nil, err
	}
	var networkTypeMap = make(map[string]interface{})
	var networkTypeList []string
	if serverPlanning.Id != 0 {
		networkTypeMap[serverPlanning.NetworkInterface] = struct{}{}
		networkTypeList = append(networkTypeList, serverPlanning.NetworkInterface)
	}
	for _, v := range serverBaselineList {
		if _, ok := networkTypeMap[v.NetworkInterface]; !ok {
			networkTypeMap[v.NetworkInterface] = struct{}{}
			networkTypeList = append(networkTypeList, v.NetworkInterface)
		}
	}
	return networkTypeList, nil
}

func ListServerCpuType(request *Request) ([]string, error) {
	// 缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	serverBaselineList, serverPlanning, err := getServerType(db, request)
	if err != nil {
		return nil, err
	}
	var cpuTypeMap = make(map[string]interface{})
	var cpuTypeList []string
	if serverPlanning.Id != 0 {
		cpuTypeMap[serverPlanning.CpuType] = struct{}{}
		cpuTypeList = append(cpuTypeList, serverPlanning.CpuType)
	}
	for _, v := range serverBaselineList {
		if _, ok := cpuTypeMap[v.CpuType]; !ok {
			cpuTypeMap[v.CpuType] = struct{}{}
			cpuTypeList = append(cpuTypeList, v.CpuType)
		}
	}
	return cpuTypeList, nil
}

func getServerType(db *gorm.DB, request *Request) ([]*entity.ServerBaseline, *entity.ServerPlanning, error) {
	// 查询云产品规划表
	var cloudProductPlanning = &entity.CloudProductPlanning{}
	if err := data.DB.Where("plan_id = ?", request.PlanId).Find(&cloudProductPlanning).Error; err != nil {
		return nil, nil, err
	}
	if cloudProductPlanning.Id == 0 {
		return nil, nil, errors.New("云产品规划不存在")
	}
	// 查询云产品基线表
	var cloudProductBaseline = &entity.CloudProductBaseline{}
	if err := data.DB.Where("id = ?", cloudProductPlanning.ProductId).Find(&cloudProductBaseline).Error; err != nil {
		return nil, nil, err
	}
	if cloudProductBaseline.Id == 0 {
		return nil, nil, errors.New("云产品基线不存在")
	}
	// 查询服务器基线表
	var serverBaselineList []*entity.ServerBaseline
	if err := data.DB.Where("version_id = ?", cloudProductBaseline.VersionId).Find(&serverBaselineList).Error; err != nil {
		return nil, nil, err
	}
	// 查询是否有已保存的方案
	var serverPlanning = &entity.ServerPlanning{}
	if err := db.Where("plan_id = ?", request.PlanId).Find(&serverPlanning).Error; err != nil {
		return nil, nil, err
	}
	return serverBaselineList, serverPlanning, nil
}

func DownloadServer(planId int64) ([]ResponseDownloadServer, string, error) {
	// 查询服务器规划列表
	var serverList []*entity.ServerPlanning
	if err := data.DB.Where("plan_id = ?", planId).Find(&serverList).Error; err != nil {
		return nil, "", err
	}
	// 查询关联的角色和设备，封装成map
	var nodeRoleIdList, serverBaselineIdList []int64
	for _, v := range serverList {
		nodeRoleIdList = append(nodeRoleIdList, v.NodeRoleId)
		serverBaselineIdList = append(serverBaselineIdList, v.ServerBaselineId)
	}
	var nodeRoleList []*entity.NodeRoleBaseline
	if err := data.DB.Where("id IN (?)", nodeRoleIdList).Find(&nodeRoleList).Error; err != nil {
		return nil, "", err
	}
	var nodeRoleMap = make(map[int64]*entity.NodeRoleBaseline)
	for _, v := range nodeRoleList {
		nodeRoleMap[v.Id] = v
	}
	var serverBaselineList []*entity.ServerBaseline
	if err := data.DB.Where("id IN (?)", serverBaselineIdList).Find(&serverBaselineList).Error; err != nil {
		return nil, "", err
	}
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	for _, v := range serverBaselineList {
		serverBaselineMap[v.Id] = v
	}
	// 构建返回体
	var response []ResponseDownloadServer
	var total int
	for _, v := range serverList {
		response = append(response, ResponseDownloadServer{
			NodeRole:   nodeRoleMap[v.NodeRoleId].NodeRoleName,
			ServerType: serverBaselineMap[v.ServerBaselineId].Arch,
			BomCode:    serverBaselineMap[v.ServerBaselineId].BomCode,
			Spec:       serverBaselineMap[v.ServerBaselineId].ConfigurationInfo,
			Number:     strconv.Itoa(v.Number),
		})
		total += v.Number
	}
	response = append(response, ResponseDownloadServer{
		Number: "总计：" + strconv.Itoa(total) + "台",
	})
	// 构建文件名称
	var planManage = &entity.PlanManage{}
	if err := data.DB.Where("id = ? AND delete_state = ?", planId, 0).Find(&planManage).Error; err != nil {
		return nil, "", err
	}
	if planManage.Id == 0 {
		return nil, "", errors.New("方案不存在")
	}
	var projectManage = &entity.ProjectManage{}
	if err := data.DB.Where("id = ? AND delete_state = ?", planManage.ProjectId, 0).First(&projectManage).Error; err != nil {
		return nil, "", err
	}
	fileName := projectManage.Name + "-" + planManage.Name + "-" + "服务器规划清单"
	return response, fileName, nil
}

func getMixedNodeRoleMap(db *gorm.DB, nodeRoleIdList []int64) (map[int64][]*MixedNodeRole, error) {
	var nodeRoleIdMap = make(map[int64]interface{})
	var newNodeRoleId []int64
	for _, v := range nodeRoleIdList {
		if _, ok := nodeRoleIdMap[v]; !ok {
			nodeRoleIdMap[v] = struct{}{}
			newNodeRoleId = append(newNodeRoleId, v)
		}
	}
	var nodeRoleMixedDeployList []*entity.NodeRoleMixedDeploy
	if err := db.Where("node_role_id IN (?)", newNodeRoleId).Find(&nodeRoleMixedDeployList).Error; err != nil {
		return nil, err
	}
	var mixedNodeRoleIdList []int64
	for _, v := range nodeRoleMixedDeployList {
		mixedNodeRoleIdList = append(mixedNodeRoleIdList, v.MixedNodeRoleId)
	}
	var mixedNodeRoleBaselineList []*entity.NodeRoleBaseline
	if err := db.Where("id IN (?)", newNodeRoleId).Find(&mixedNodeRoleBaselineList).Error; err != nil {
		return nil, err
	}
	var nodeRoleBaselineMap = make(map[int64]*entity.NodeRoleBaseline)
	for _, v := range mixedNodeRoleBaselineList {
		nodeRoleBaselineMap[v.Id] = v
	}
	var mixedNodeRoleMap = make(map[int64][]*MixedNodeRole)
	for _, v := range newNodeRoleId {
		mixedNodeRoleMap[v] = append(mixedNodeRoleMap[v], &MixedNodeRole{
			Id:   v,
			Name: "独立部署",
		})
	}
	for _, v := range nodeRoleMixedDeployList {
		mixedNodeRoleMap[v.NodeRoleId] = append(mixedNodeRoleMap[v.NodeRoleId], &MixedNodeRole{
			Id:   nodeRoleBaselineMap[v.MixedNodeRoleId].Id,
			Name: "混合部署：" + nodeRoleBaselineMap[v.MixedNodeRoleId].NodeRoleName,
		})
	}
	return mixedNodeRoleMap, nil
}

func getNodeRoleServerBaselineMap(db *gorm.DB, nodeRoleIdList []int64, nodeRoleBaselineList []*entity.NodeRoleBaseline, request *Request) (map[int64]*entity.ServerBaseline, map[int64][]*Baseline, map[int64]*entity.ServerBaseline, map[string]*entity.NodeRoleBaseline, error) {
	var nodeRoleBaselineMap = make(map[int64]*entity.NodeRoleBaseline)
	for _, v := range nodeRoleBaselineList {
		nodeRoleBaselineMap[v.Id] = v
	}
	// 查询服务器和角色关联表
	var serverNodeRoleRelList []*entity.ServerNodeRoleRel
	if err := db.Where("node_role_id IN (?)", nodeRoleIdList).Find(&serverNodeRoleRelList).Error; err != nil {
		return nil, nil, nil, nil, err
	}
	var nodeRoleServerRelMap = make(map[int64][]int64)
	var serverBaselineIdList []int64
	for _, v := range serverNodeRoleRelList {
		nodeRoleServerRelMap[v.NodeRoleId] = append(nodeRoleServerRelMap[v.NodeRoleId], v.ServerId)
		serverBaselineIdList = append(serverBaselineIdList, v.ServerId)
	}
	// 查询服务器基线表
	var serverBaselineList []*entity.ServerBaseline
	if err := db.Where("id IN (?)", serverBaselineIdList).Find(&serverBaselineList).Error; err != nil {
		return nil, nil, nil, nil, err
	}
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	for _, v := range serverBaselineList {
		serverBaselineMap[v.Id] = v
	}
	// 查询服务器基线表
	var nodeRoleServerBaselineListMap = make(map[int64][]*Baseline)
	var screenNodeRoleServerBaselineMap = make(map[int64]*entity.ServerBaseline)
	var nodeRoleCodeBaselineMap = make(map[string]*entity.NodeRoleBaseline)
	for k, serverIdList := range nodeRoleServerRelMap {
		for _, serverId := range serverIdList {
			serverBaseline := serverBaselineMap[serverId]
			if serverBaseline == nil {
				continue
			}
			nodeRoleServerBaselineListMap[k] = append(nodeRoleServerBaselineListMap[k], &Baseline{
				Id:                  serverBaseline.Id,
				BomCode:             serverBaseline.BomCode,
				NetworkInterface:    serverBaseline.NetworkInterface,
				CpuType:             serverBaseline.CpuType,
				Cpu:                 serverBaseline.Cpu,
				Memory:              serverBaseline.Memory,
				StorageDiskNum:      serverBaseline.StorageDiskNum,
				StorageDiskCapacity: serverBaseline.StorageDiskCapacity,
				Arch:                serverBaseline.Arch,
				ConfigurationInfo:   serverBaseline.ConfigurationInfo,
			})
			if request.NetworkInterface == serverBaseline.NetworkInterface && serverBaseline.CpuType == request.CpuType {
				screenNodeRoleServerBaselineMap[k] = serverBaseline
				nodeRoleCodeBaselineMap[nodeRoleBaselineMap[k].NodeRoleCode] = nodeRoleBaselineMap[k]
			}
			if screenNodeRoleServerBaselineMap[k] == nil {
				screenNodeRoleServerBaselineMap[k] = serverBaseline
				nodeRoleCodeBaselineMap[nodeRoleBaselineMap[k].NodeRoleCode] = nodeRoleBaselineMap[k]
			}
		}
	}
	return serverBaselineMap, nodeRoleServerBaselineListMap, screenNodeRoleServerBaselineMap, nodeRoleCodeBaselineMap, nil
}

func getServerShelveDownloadTemplate(planId int64) ([]ShelveDownload, string, error) {
	// 查询服务器规划表
	serverPlanningList, err := getServerShelvePlanningList(planId)
	if err != nil {
		return nil, "", err
	}
	if len(serverPlanningList) == 0 {
		return nil, "", errors.New("服务器未规划")
	}
	// 查询机柜
	cabinetIdleSlotList, err := getCabinetInfo(planId)
	if err != nil {
		return nil, "", err
	}
	// 构建返回体
	var response []ShelveDownload
	var sortNumber = 0
	for _, v := range serverPlanningList {
		for i := 1; i <= v.Number; i++ {
			if sortNumber >= len(cabinetIdleSlotList) {
				return nil, "", errors.New("槽位数量不足，请修改机房勘察表")
			}
			response = append(response, ShelveDownload{
				SortNumber:            sortNumber + 1,
				NodeRoleName:          v.NodeRoleName,
				Model:                 v.ServerBomCode,
				MachineRoomAbbr:       cabinetIdleSlotList[sortNumber].MachineRoomAbbr,
				MachineRoomNumber:     cabinetIdleSlotList[sortNumber].MachineRoomNum,
				ColumnNumber:          cabinetIdleSlotList[sortNumber].ColumnNum,
				CabinetAsw:            cabinetIdleSlotList[sortNumber].CabinetAsw,
				CabinetNumber:         cabinetIdleSlotList[sortNumber].CabinetNum,
				CabinetOriginalNumber: cabinetIdleSlotList[sortNumber].OriginalNum,
				CabinetLocation:       cabinetIdleSlotList[sortNumber].CabinetLocation,
				SlotPosition:          cabinetIdleSlotList[sortNumber].IdleSlot,
				NetworkInterface:      v.NetworkInterface,
			})
			sortNumber++
		}
	}
	// 构建文件名称
	var planManage = &entity.PlanManage{}
	if err = data.DB.Where("id = ? AND delete_state = ?", planId, 0).Find(&planManage).Error; err != nil {
		return nil, "", err
	}
	if planManage.Id == 0 {
		return nil, "", errors.New("方案不存在")
	}
	var projectManage = &entity.ProjectManage{}
	if err = data.DB.Where("id = ? AND delete_state = ?", planManage.ProjectId, 0).First(&projectManage).Error; err != nil {
		return nil, "", err
	}
	fileName := projectManage.Name + "-" + planManage.Name + "-" + "服务器上架模板"
	return response, fileName, nil
}

func getServerShelvePlanningList(planId int64) ([]*Server, error) {
	// 查询服务器规划表
	var serverPlanning []*Server
	if err := data.DB.Model(&entity.ServerPlanning{}).Where("plan_id = ?", planId).Order("shelve_priority ASC").Find(&serverPlanning).Error; err != nil {
		return nil, err
	}
	var nodeRoleIdList []int64
	var serverBaselineIdList []int64
	for _, v := range serverPlanning {
		nodeRoleIdList = append(nodeRoleIdList, v.NodeRoleId)
		serverBaselineIdList = append(serverBaselineIdList, v.ServerBaselineId)
	}
	// 查询节点角色表
	var nodeRoleList []*entity.NodeRoleBaseline
	if err := data.DB.Where("id IN (?)", nodeRoleIdList).Find(&nodeRoleList).Error; err != nil {
		return nil, err
	}
	var nodeRoleNameMap = make(map[int64]string)
	for _, v := range nodeRoleList {
		nodeRoleNameMap[v.Id] = v.NodeRoleName
	}
	// 查询服务器基线表
	var serverBaseline []*entity.ServerBaseline
	if err := data.DB.Where("id IN (?)", serverBaselineIdList).Find(&serverBaseline).Error; err != nil {
		return nil, err
	}
	var serverBaselineMap = make(map[int64]string)
	for _, v := range serverBaseline {
		serverBaselineMap[v.Id] = v.BomCode
	}
	// 查询服务器上架表
	var serverShelveCount int64
	if err := data.DB.Model(&entity.ServerShelve{}).Where("plan_id = ?", planId).Count(&serverShelveCount).Error; err != nil {
		return nil, err
	}
	var upload int
	if serverShelveCount > 0 {
		upload = 1
	}
	for i, v := range serverPlanning {
		serverPlanning[i].NodeRoleName = nodeRoleNameMap[v.NodeRoleId]
		serverPlanning[i].ServerBomCode = serverBaselineMap[v.ServerBaselineId]
		serverPlanning[i].Upload = upload
	}
	return serverPlanning, nil
}

func getCabinetInfo(planId int64) ([]*Cabinet, error) {
	var cabinetInfoList []*entity.CabinetInfo
	if err := data.DB.Where("plan_id = ? AND cabinet_type = ?", planId, 2).Order("id ASC").Find(&cabinetInfoList).Error; err != nil {
		return nil, err
	}
	var cabinetMap = make(map[int64]*entity.CabinetInfo)
	var cabinetIdList []int64
	var cabinetResidualRackServerNumMap = make(map[int64]int)    // 剩余上架服务器数
	var networkDeviceShelveCabinetIdMap = make(map[string]int64) // 网络设备与机柜的关联信息
	for _, v := range cabinetInfoList {
		cabinetMap[v.Id] = v
		cabinetIdList = append(cabinetIdList, v.Id)
		cabinetResidualRackServerNumMap[v.Id] = v.ResidualRackServerNum
		networkDeviceShelveCabinetIdMap[fmt.Sprintf("%v-%v-%v-%v", v.CabinetAsw, v.MachineRoomAbbr, v.MachineRoomNum, v.CabinetNum)] = v.Id
	}
	// 查询机柜槽位
	var cabinetIdleSlotRelList []*entity.CabinetIdleSlotRel
	if err := data.DB.Where("cabinet_id IN (?)", cabinetIdList).Order("cabinet_id ASC, idle_slot_number ASC").Find(&cabinetIdleSlotRelList).Error; err != nil {
		return nil, err
	}
	if len(cabinetIdleSlotRelList) == 0 {
		return nil, errors.New("机柜槽位为空，请检查机房勘察表是否填写错误")
	}
	// 查询网络设备占用槽位
	var networkDeviceShelveList []*entity.NetworkDeviceShelve
	if err := data.DB.Where("plan_id = ?", planId).Find(&networkDeviceShelveList).Error; err != nil {
		return nil, err
	}
	var networkDeviceShelveSlotPositionMap = make(map[int64]map[int]interface{})
	for _, v := range networkDeviceShelveList {
		slotPositionSplit := strings.Split(v.SlotPosition, "-")
		for _, lotPositionString := range slotPositionSplit {
			lotPosition, _ := strconv.Atoi(lotPositionString)
			cabinetId := networkDeviceShelveCabinetIdMap[fmt.Sprintf("%v-%v-%v-%v", v.DeviceLogicalId, v.MachineRoomAbbr, v.MachineRoomNumber, v.CabinetNumber)]
			if networkDeviceShelveSlotPositionMap[cabinetId] == nil {
				networkDeviceShelveSlotPositionMap[cabinetId] = make(map[int]interface{})
			}
			networkDeviceShelveSlotPositionMap[cabinetId][lotPosition] = struct{}{}
		}
	}
	log.Infof("网络设备占用槽位: %+v", networkDeviceShelveSlotPositionMap)
	var cabinetIdleSlotNumberMap = make(map[int64]int)
	var cabinetIdleSlotListMap = make(map[int64][]*Cabinet)
	for _, v := range cabinetIdleSlotRelList {
		// 1号槽位不上架
		if v.IdleSlotNumber == 1 {
			continue
		}
		// 过滤网络设备占用的槽位
		if networkDeviceShelveSlotPositionMap[v.CabinetId] != nil {
			if _, ok := networkDeviceShelveSlotPositionMap[v.CabinetId][v.IdleSlotNumber]; ok {
				continue
			}
		}
		if cabinetIdleSlotNumberMap[v.CabinetId] == 0 {
			cabinetIdleSlotNumberMap[v.CabinetId] = v.IdleSlotNumber
		} else {
			// 两个槽位相邻，且不超过机柜的剩余上架服务器数
			if v.IdleSlotNumber-cabinetIdleSlotNumberMap[v.CabinetId] == 1 && len(cabinetIdleSlotListMap) <= cabinetResidualRackServerNumMap[v.CabinetId] {
				cabinetIdleSlotListMap[v.CabinetId] = append(cabinetIdleSlotListMap[v.CabinetId],
					&Cabinet{
						CabinetInfo:     cabinetMap[v.CabinetId],
						CabinetLocation: fmt.Sprintf("%v-%v", cabinetMap[v.CabinetId].MachineRoomAbbr, cabinetMap[v.CabinetId].CabinetNum),
						IdleSlot:        fmt.Sprintf("%d-%d", cabinetIdleSlotNumberMap[v.CabinetId], v.IdleSlotNumber),
					})
				cabinetIdleSlotNumberMap[v.CabinetId] = 0
			} else {
				cabinetIdleSlotNumberMap[v.CabinetId] = v.IdleSlotNumber
			}
		}
	}
	// 跨机柜上架，将所有机柜槽位列表纵向排布，然后从低到高横向上架
	var cabinetIdleSlotList []*Cabinet
	var maxLength = 1 // 槽位最多的机柜的槽位数量
	var index = 0     // 下标
	for index < maxLength {
		for _, v := range cabinetInfoList {
			idleSlotList := cabinetIdleSlotListMap[v.Id]
			if maxLength < len(idleSlotList) {
				maxLength = len(idleSlotList)
			}
			if index < len(idleSlotList) {
				cabinetIdleSlotList = append(cabinetIdleSlotList, idleSlotList[index])
			}
		}
		index++
	}
	return cabinetIdleSlotList, nil
}

func UploadServerShelve(planId int64, serverShelveDownload []ShelveDownload, userId string) error {
	if len(serverShelveDownload) == 0 {
		return errors.New("数据为空")
	}
	now := datetime.GetNow()
	// 查询服务器规划表
	var serverPlanning []*entity.ServerPlanning
	if err := data.DB.Where("plan_id = ?", planId).Find(&serverPlanning).Error; err != nil {
		return err
	}
	var NodeRoleIdList []int64
	for _, v := range serverPlanning {
		NodeRoleIdList = append(NodeRoleIdList, v.NodeRoleId)
	}
	// 查询节点角色表
	var nodeRoleList []*entity.NodeRoleBaseline
	if err := data.DB.Where("id IN (?)", NodeRoleIdList).Find(&nodeRoleList).Error; err != nil {
		return err
	}
	var nodeRoleNameMap = make(map[string]int64)
	for _, v := range nodeRoleList {
		nodeRoleNameMap[v.NodeRoleName] = v.Id
	}
	// 查询机柜信息
	var cabinetInfoList []*entity.CabinetInfo
	if err := data.DB.Where("plan_id = ?", planId).Find(&cabinetInfoList).Error; err != nil {
		return err
	}
	var cabinetInfoMap = make(map[string]*entity.CabinetInfo)
	for _, v := range cabinetInfoList {
		cabinetInfoMap[fmt.Sprintf("%v-%v-%v-%v-%v-%v", v.MachineRoomAbbr, v.MachineRoomNum, v.ColumnNum, v.CabinetAsw, v.CabinetNum, v.OriginalNum)] = v
	}
	var serverShelveList []*entity.ServerShelve
	for _, v := range serverShelveDownload {
		if util.IsBlank(v.Sn) {
			return errors.New("表单所有参数不能为空")
		}
		key := fmt.Sprintf("%v-%v-%v-%v-%v-%v", v.MachineRoomAbbr, v.MachineRoomNumber, v.ColumnNumber, v.CabinetAsw, v.CabinetNumber, v.CabinetOriginalNumber)
		cabinetInfo := cabinetInfoMap[key]
		if cabinetInfo == nil {
			return errors.New("机柜信息错误：" + key)
		}
		serverShelveList = append(serverShelveList, &entity.ServerShelve{
			SortNumber:            v.SortNumber,
			PlanId:                planId,
			NodeRoleId:            nodeRoleNameMap[v.NodeRoleName],
			Sn:                    v.Sn,
			Model:                 v.Model,
			CabinetId:             cabinetInfo.Id,
			MachineRoomAbbr:       v.MachineRoomAbbr,
			MachineRoomNumber:     v.MachineRoomNumber,
			ColumnNumber:          v.ColumnNumber,
			CabinetAsw:            v.CabinetAsw,
			CabinetNumber:         v.CabinetNumber,
			CabinetOriginalNumber: v.CabinetOriginalNumber,
			CabinetLocation:       v.CabinetLocation,
			SlotPosition:          v.SlotPosition,
			NetworkInterface:      v.NetworkInterface,
			BmcUserName:           v.BmcUserName,
			BmcPassword:           v.BmcPassword,
			BmcIp:                 v.BmcIp,
			BmcMac:                v.BmcMac,
			Mask:                  v.Mask,
			Gateway:               v.Gateway,
			CreateUserId:          userId,
			CreateTime:            now,
		})
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&entity.ServerShelve{}, "plan_id = ?", planId).Error; err != nil {
			return err
		}
		if err := tx.CreateInBatches(&serverShelveList, 10).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func saveServerPlanning(request *Request) error {
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		for _, v := range request.ServerList {
			if err := tx.Model(&entity.ServerPlanning{}).Where("plan_id = ? AND node_role_id = ?", request.PlanId, v.NodeRoleId).Updates(map[string]interface{}{"business_attributes": v.BusinessAttributes, "shelve_mode": v.ShelveMode, "shelve_priority": v.ShelvePriority}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func saveServerShelve(request *Request) error {
	// 查询服务器上架表
	var cabinetIdList []int64
	if err := data.DB.Model(&entity.ServerShelve{}).Select("cabinet_id").Where("plan_id = ?", request.PlanId).Group("cabinet_id").Find(&cabinetIdList).Error; err != nil {
		return err
	}
	if len(cabinetIdList) == 0 {
		return errors.New("服务器未上架")
	}
	// 查询机柜信息
	var cabinetCount int64
	if err := data.DB.Model(&entity.CabinetInfo{}).Where("id IN (?)", cabinetIdList).Count(&cabinetCount).Error; err != nil {
		return err
	}
	if int64(len(cabinetIdList)) != cabinetCount {
		return errors.New("机房信息已修改，请重新下载服务器上架模板并上传")
	}
	if err := data.DB.Updates(&entity.PlanManage{Id: request.PlanId, DeliverPlanStage: constant.DeliverPlanningIp}).Error; err != nil {
		return err
	}
	return nil
}

func getServerShelveDownload(planId int64) ([]ShelveDownload, string, error) {
	var serverShelveList []*entity.ServerShelve
	if err := data.DB.Where("plan_id = ?", planId).Find(&serverShelveList).Error; err != nil {
		return nil, "", err
	}
	// 查询节点角色表
	var nodeRoleIdList []int64
	for _, v := range serverShelveList {
		nodeRoleIdList = append(nodeRoleIdList, v.NodeRoleId)
	}
	var nodeRoleList []*entity.NodeRoleBaseline
	if err := data.DB.Where("id IN (?)", nodeRoleIdList).Find(&nodeRoleList).Error; err != nil {
		return nil, "", err
	}
	var nodeRoleNameMap = make(map[int64]string)
	for _, v := range nodeRoleList {
		nodeRoleNameMap[v.Id] = v.NodeRoleName
	}
	var response []ShelveDownload
	for _, v := range serverShelveList {
		response = append(response, ShelveDownload{
			SortNumber:            v.SortNumber,
			NodeRoleName:          nodeRoleNameMap[v.NodeRoleId],
			Sn:                    v.Sn,
			Model:                 v.Model,
			MachineRoomAbbr:       v.MachineRoomAbbr,
			MachineRoomNumber:     v.MachineRoomNumber,
			ColumnNumber:          v.ColumnNumber,
			CabinetAsw:            v.CabinetAsw,
			CabinetNumber:         v.CabinetNumber,
			CabinetOriginalNumber: v.CabinetOriginalNumber,
			CabinetLocation:       v.CabinetLocation,
			SlotPosition:          v.SlotPosition,
			NetworkInterface:      v.NetworkInterface,
			BmcUserName:           v.BmcUserName,
			BmcPassword:           v.BmcPassword,
			BmcIp:                 v.BmcIp,
			BmcMac:                v.BmcMac,
			Mask:                  v.Mask,
			Gateway:               v.Gateway,
		})
	}
	// 构建文件名称
	var planManage = &entity.PlanManage{}
	if err := data.DB.Where("id = ? AND delete_state = ?", planId, 0).Find(&planManage).Error; err != nil {
		return nil, "", err
	}
	if planManage.Id == 0 {
		return nil, "", errors.New("方案不存在")
	}
	var projectManage = &entity.ProjectManage{}
	if err := data.DB.Where("id = ? AND delete_state = ?", planManage.ProjectId, 0).First(&projectManage).Error; err != nil {
		return nil, "", err
	}
	fileName := projectManage.Name + "-" + planManage.Name + "-" + "服务器上架清单"
	return response, fileName, nil
}

func checkBusiness(db *gorm.DB, request *Request) error {
	// 校验planId
	var planCount int64
	if err := db.Model(&entity.PlanManage{}).Where("id = ? AND delete_state = ?", request.PlanId, 0).Count(&planCount).Error; err != nil {
		return err
	}
	if planCount == 0 {
		return errors.New("方案不存在")
	}
	return nil
}
