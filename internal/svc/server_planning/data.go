package server_planning

import (
	"errors"
	"fmt"
	"math"
	"sort"
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
	for _, cloudProductPlanning := range cloudProductPlanningList {
		productIdList = append(productIdList, cloudProductPlanning.ProductId)
	}
	// 查询云产品和角色关联表
	var nodeRoleIdList []int64
	var cloudProductNodeRoleRelList []*entity.CloudProductNodeRoleRel
	if err := db.Model(&entity.CloudProductNodeRoleRel{}).Where("product_id IN (?)", productIdList).Find(&cloudProductNodeRoleRelList).Error; err != nil {
		return nil, err
	}
	for _, cloudProductNodeRoleRel := range cloudProductNodeRoleRelList {
		nodeRoleIdList = append(nodeRoleIdList, cloudProductNodeRoleRel.NodeRoleId)
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
	mixedNodeRoleMap, err := getMixedNodeRoleMap(db, nodeRoleBaselineList, request.PlanId)
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
			serverBaseline := screenNodeRoleServerBaselineMap[nodeRoleBaseline.Id]
			resourcePoolMap := make(map[int64]*entity.ResourcePool)
			for _, resourcePool := range nodeRoleResourcePoolList {
				resourcePoolMap[resourcePool.Id] = resourcePool
			}
			for _, originServerPlanning := range nodeRoleServerPlannings {
				// 如果在服务器规划页面删除了资源池，则跳过
				resourcePool, ok := resourcePoolMap[originServerPlanning.ServerPlanning.ResourcePoolId]
				if !ok {
					continue
				}
				serverPlanning := &Server{}
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
				serverPlanning.SupportMultiResourcePool = nodeRoleBaseline.SupportMultiResourcePool
				serverPlanning.ServerBaselineList = nodeRoleServerBaselineListMap[nodeRoleBaseline.Id]
				serverPlanning.MixedNodeRoleList = mixedNodeRoleMap[nodeRoleBaseline.Id]
				serverPlanning.ResourcePoolName = resourcePool.ResourcePoolName
				serverPlanning.DefaultResourcePool = resourcePool.DefaultResourcePool
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
				delete(resourcePoolMap, resourcePool.Id)
			}
			if len(resourcePoolMap) > 0 {
				resourcePoolIds := make([]int64, 0, len(resourcePoolMap))
				for resourcePoolId := range resourcePoolMap {
					resourcePoolIds = append(resourcePoolIds, resourcePoolId)
				}
				// 对keys进行排序
				sort.SliceStable(resourcePoolIds, func(i, j int) bool {
					return resourcePoolIds[i] < resourcePoolIds[j]
				})
				for _, resourcePoolId := range resourcePoolIds {
					resourcePool := resourcePoolMap[resourcePoolId]
					// 表示在服务器规划页面新增了资源池
					serverPlanning := &Server{}
					serverPlanning.PlanId = request.PlanId
					serverPlanning.NodeRoleId = nodeRoleBaseline.Id
					serverPlanning.Number = nodeRoleBaseline.MinimumNum
					// 列表加载机型
					serverPlanning.ServerBaselineId = serverBaseline.Id
					serverPlanning.ServerBomCode = serverBaseline.BomCode
					serverPlanning.ServerArch = serverBaseline.Arch
					serverPlanning.MixedNodeRoleId = nodeRoleBaseline.Id
					serverPlanning.NodeRoleName = nodeRoleBaseline.NodeRoleName
					serverPlanning.NodeRoleClassify = nodeRoleBaseline.Classify
					serverPlanning.NodeRoleAnnotation = nodeRoleBaseline.Annotation
					serverPlanning.SupportDpdk = nodeRoleBaseline.SupportDPDK
					serverPlanning.SupportMultiResourcePool = nodeRoleBaseline.SupportMultiResourcePool
					serverPlanning.ServerBaselineList = nodeRoleServerBaselineListMap[nodeRoleBaseline.Id]
					serverPlanning.MixedNodeRoleList = mixedNodeRoleMap[nodeRoleBaseline.Id]
					serverPlanning.NetworkInterface = serverBaseline.NetworkInterface
					serverPlanning.CpuType = serverBaseline.CpuType
					serverPlanning.ResourcePoolName = resourcePool.ResourcePoolName
					serverPlanning.ResourcePoolId = resourcePool.Id
					serverPlanning.DefaultResourcePool = resourcePool.DefaultResourcePool
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
				}
			}
		} else {
			return nil, errors.New("服务器规划无数据，请重新规划云产品配置")
		}
	}
	// 计算已保存的容量规划指标
	resourcePoolCapMap, err := capacity_planning.GetResourcePoolCapMap(db, &capacity_planning.Request{PlanId: request.PlanId}, resourcePoolIdServerPlanningMap, nodeRoleCodeBaselineMap, serverBaselineMap)
	if err != nil {
		return nil, err
	}
	var resourceIdList []int64
	var bmsServerNumber int
	var serverNumber int
	bmsNodeRoleBaseline := nodeRoleCodeBaselineMap[constant.NodeRoleCodeBMS]
	bmsGWNodeRoleBaseline := nodeRoleCodeBaselineMap[constant.NodeRoleCodeBMSGW]
	masterNodeRoleBaseline := nodeRoleCodeBaselineMap[constant.NodeRoleCodeMaster]
	bmsGWServerPlanningIndex := -1
	masterServerPlanningIndex := -1
	for i, server := range list {
		resourceIdList = append(resourceIdList, server.ResourcePoolId)
		if serverPlanning, ok := resourcePoolIdServerPlanningMap[server.ResourcePoolId]; ok && util.IsBlank(request.NetworkInterface) && util.IsBlank(request.CpuType) {
			list[i].Number = serverPlanning.Number
			continue
		}
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
		if bmsNodeRoleBaseline != nil && bmsGWNodeRoleBaseline != nil {
			if server.NodeRoleId == bmsNodeRoleBaseline.Id {
				bmsServerNumber += server.Number
			}
			if server.NodeRoleId == bmsGWNodeRoleBaseline.Id {
				bmsGWServerPlanningIndex = i
			}
		}
		if masterNodeRoleBaseline != nil && server.NodeRoleId == masterNodeRoleBaseline.Id {
			masterServerPlanningIndex = i
		} else {
			if bmsGWNodeRoleBaseline == nil || server.NodeRoleId != bmsGWNodeRoleBaseline.Id {
				serverNumber += server.Number
			}
		}
	}
	if bmsGWServerPlanningIndex != -1 {
		bmsGWServerNumber := int(math.Ceil(float64(bmsServerNumber)/30)) * 2
		bmsGWServerPlanning := list[bmsGWServerPlanningIndex]
		if bmsGWServerPlanning.Number < bmsGWServerNumber {
			bmsGWServerPlanning.Number = bmsGWServerNumber
			list[bmsGWServerPlanningIndex].Number = bmsGWServerNumber
		}
		serverNumber += bmsGWServerPlanning.Number
	}
	if masterServerPlanningIndex != -1 {
		var cloudProductBaselineList []*entity.CloudProductBaseline
		if err = data.DB.Where("id in (?)", productIdList).Find(&cloudProductBaselineList).Error; err != nil {
			return nil, err
		}
		pureIaaS := true
		for _, cloudProductBaseline := range cloudProductBaselineList {
			if cloudProductBaseline.ProductType != constant.ProductTypeCompute && cloudProductBaseline.ProductType != constant.ProductTypeNetwork && cloudProductBaseline.ProductType != constant.ProductTypeStorage {
				pureIaaS = false
				break
			}
		}
		var masterNumber int
		if pureIaaS {
			var projectManage *entity.ProjectManage
			if err = data.DB.Where("id = (select project_id from plan_manage where id = ?)", request.PlanId).Find(&projectManage).Error; err != nil {
				return nil, err
			}
			var azManageList []*entity.AzManage
			if err = data.DB.Where("region_id = ?", projectManage.RegionId).Find(&azManageList).Error; err != nil {
				return nil, err
			}
			var cellManage *entity.CellManage
			if err = data.DB.Where("id = ?", projectManage.CellId).Find(&cellManage).Error; err != nil {
				return nil, err
			}
			if len(azManageList) > 1 {
				if cellManage.Type == constant.CellTypeControl {
					if serverNumber <= 495 {
						masterNumber = 5
					} else if serverNumber <= 1991 {
						masterNumber = 9
					} else {
						masterNumber = 15
					}
				} else {
					if serverNumber <= 197 {
						masterNumber = 3
					} else if serverNumber <= 495 {
						masterNumber = 5
					} else if serverNumber <= 1991 {
						masterNumber = 9
					} else {
						masterNumber = 15
					}
				}
			} else {
				if serverNumber <= 197 {
					masterNumber = 3
				} else if serverNumber <= 495 {
					masterNumber = 5
				} else if serverNumber <= 1991 {
					masterNumber = 9
				} else {
					masterNumber = 15
				}
			}
		} else {
			if serverNumber <= 195 {
				masterNumber = 5
			} else if serverNumber <= 493 {
				masterNumber = 7
			} else if serverNumber <= 1991 {
				masterNumber = 9
			} else {
				masterNumber = 15
			}
		}
		// 是否和原始服务器数量比较，如果比较且之前的数据大于现有的数据，则不修改
		if list[masterServerPlanningIndex].Number < masterNumber {
			list[masterServerPlanningIndex].Number = masterNumber
		}
	}
	return list, nil
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

func getMixedNodeRoleMap(db *gorm.DB, nodeRoleBaselineList []*entity.NodeRoleBaseline, planId int64) (map[int64][]*MixedNodeRole, error) {
	var nodeRoleIdList []int64
	var nodeRoleBaselineMap = make(map[int64]*entity.NodeRoleBaseline)
	var mixedNodeRoleMap = make(map[int64][]*MixedNodeRole)
	for _, nodeRoleBaseline := range nodeRoleBaselineList {
		nodeRoleId := nodeRoleBaseline.Id
		nodeRoleIdList = append(nodeRoleIdList, nodeRoleId)
		nodeRoleBaselineMap[nodeRoleId] = nodeRoleBaseline
		mixedNodeRoleMap[nodeRoleId] = append(mixedNodeRoleMap[nodeRoleId], &MixedNodeRole{
			Id:   nodeRoleId,
			Name: "独立部署",
		})
	}
	var nodeRoleMixedDeployList []*entity.NodeRoleMixedDeploy
	if err := db.Where("node_role_id IN (?)", nodeRoleIdList).Find(&nodeRoleMixedDeployList).Error; err != nil {
		return nil, err
	}
	var mixedNodeRoleIdList []int64
	for _, nodeRoleMixedDeploy := range nodeRoleMixedDeployList {
		mixedNodeRoleIdList = append(mixedNodeRoleIdList, nodeRoleMixedDeploy.MixedNodeRoleId)
	}
	var resourcePoolList []*entity.ResourcePool
	if err := db.Where("plan_id = ? and node_role_id IN (?)", planId, mixedNodeRoleIdList).Find(&resourcePoolList).Error; err != nil {
		return nil, err
	}
	nodeRoleResourcePoolsMap := make(map[int64][]*entity.ResourcePool)
	for _, resourcePool := range resourcePoolList {
		nodeRoleResourcePoolsMap[resourcePool.NodeRoleId] = append(nodeRoleResourcePoolsMap[resourcePool.NodeRoleId], resourcePool)
	}
	for _, nodeRoleMixedDeploy := range nodeRoleMixedDeployList {
		mixNodeRole := &MixedNodeRole{
			Id:   nodeRoleBaselineMap[nodeRoleMixedDeploy.MixedNodeRoleId].Id,
			Name: "混合部署：" + nodeRoleBaselineMap[nodeRoleMixedDeploy.MixedNodeRoleId].NodeRoleName,
		}
		for _, resourcePool := range nodeRoleResourcePoolsMap[nodeRoleMixedDeploy.MixedNodeRoleId] {
			mixNodeRole.MixResourcePoolList = append(mixNodeRole.MixResourcePoolList, &MixedResourcePool{
				ResourcePoolId:   resourcePool.Id,
				ResourcePoolName: resourcePool.ResourcePoolName,
			})
		}
		mixedNodeRoleMap[nodeRoleMixedDeploy.NodeRoleId] = append(mixedNodeRoleMap[nodeRoleMixedDeploy.NodeRoleId], mixNodeRole)
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
