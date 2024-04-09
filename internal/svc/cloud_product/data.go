package cloud_product

import (
	"fmt"
	"strings"
	"time"

	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/capacity_planning"
)

func getVersionListByProjectId(projectId int64) ([]entity.SoftwareVersion, error) {
	var project entity.ProjectManage
	if err := data.DB.Where("id = ? AND delete_state = 0", projectId).Find(&project).Error; err != nil {
		return nil, err
	}
	var cloudPlatform = &entity.CloudPlatformManage{}
	if err := data.DB.Where("id = ? AND delete_state = 0", project.CloudPlatformId).Find(&cloudPlatform).Error; err != nil {
		return nil, err
	}
	var versionList []entity.SoftwareVersion
	if err := data.DB.Where("cloud_platform_type = ?", cloudPlatform.Type).Find(&versionList).Error; err != nil {
		log.Errorf("[getVersionListByProjectId] get version by platformType error,%v", err)
		return nil, err
	}
	return versionList, nil
}

func getCloudProductBaseListByVersionId(versionId int64) ([]CloudProductBaselineResponse, error) {
	var baselineList []entity.CloudProductBaseline
	if err := data.DB.Where("version_id = ? ", versionId).Find(&baselineList).Error; err != nil {
		return nil, err
	}
	// 查询依赖的云产品
	cloudProductDependList, err := getDependProductIds()
	if err != nil {
		return nil, err
	}
	var responseList []CloudProductBaselineResponse
	for _, baseline := range baselineList {
		sellSpecs := make([]string, 1)
		if strings.Index(baseline.SellSpecs, constant.Comma) > 0 {
			sellSpecs = strings.Split(baseline.SellSpecs, constant.Comma)
		} else {
			sellSpecs[0] = baseline.SellSpecs
		}
		var valueAddedServices []string
		if baseline.ValueAddedService != "" {
			valueAddedServices = strings.Split(baseline.ValueAddedService, constant.Comma)
		}
		var dependProductId int64
		var dependProductName string
		for _, depend := range cloudProductDependList {
			if depend.ID == baseline.Id {
				dependProductId = depend.DependId
				dependProductName = depend.DependProductName
			}
		}
		responseData := CloudProductBaselineResponse{
			ID:                 baseline.Id,
			VersionId:          baseline.VersionId,
			ProductType:        baseline.ProductType,
			ProductName:        baseline.ProductName,
			ProductCode:        baseline.ProductCode,
			SellSpecs:          sellSpecs,
			ValueAddedServices: valueAddedServices,
			AuthorizedUnit:     baseline.AuthorizedUnit,
			WhetherRequired:    baseline.WhetherRequired,
			Instructions:       baseline.Instructions,
			DependProductId:    dependProductId,
			DependProductName:  dependProductName,
		}
		responseList = append(responseList, responseData)
	}
	return responseList, nil
}

func saveCloudProductPlanning(request CloudProductPlanningRequest, currentUserId string) error {
	var cloudProductPlanningList []*entity.CloudProductPlanning
	for _, cloudProduct := range request.ProductList {
		var valueAddedService string
		if len(cloudProduct.ValueAddedServices) > 0 {
			valueAddedService = strings.Join(cloudProduct.ValueAddedServices, constant.Comma)
		}
		cloudProductPlanning := &entity.CloudProductPlanning{
			PlanId:            request.PlanId,
			ProductId:         cloudProduct.ProductId,
			VersionId:         request.VersionId,
			SellSpec:          cloudProduct.SellSpec,
			ValueAddedService: valueAddedService,
			ServiceYear:       request.ServiceYear,
			CreateTime:        time.Now(),
			UpdateTime:        time.Now(),
		}
		cloudProductPlanningList = append(cloudProductPlanningList, cloudProductPlanning)
	}

	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		// 删除原有的清单
		if err := tx.Delete(&entity.CloudProductPlanning{}, "plan_id=?", request.PlanId).Error; err != nil {
			log.Errorf("[saveCloudProductPlanning] delete cloudProductPlanning by planId error,%v", err)
			return err
		}
		// 保存云服务规划清单
		if err := tx.Table(entity.CloudProductPlanningTable).CreateInBatches(&cloudProductPlanningList, len(cloudProductPlanningList)).Error; err != nil {
			log.Errorf("[saveCloudProductPlanning] batch create cloudProductPlanning error, %v", err)
			return err
		}
		// 更新方案业务规划阶段
		if err := tx.Table(entity.PlanManageTable).Where("id = ?", request.PlanId).
			Update("business_plan_stage", constant.BusinessPlanningServer).
			Update("stage", constant.PlanStagePlanning).
			Update("update_user_id", currentUserId).
			Update("update_time", time.Now()).Error; err != nil {
			log.Errorf("[saveCloudProductPlanning] update plan business stage error, %v", err)
			return err
		}
		if err := HandleResourcePoolAndServerPlanning(tx, request.PlanId, cloudProductPlanningList, currentUserId); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Errorf("[saveCloudProductPlanning] error, %v", err)
		return err
	}
	return nil
}

func HandleResourcePoolAndServerPlanning(db *gorm.DB, planId int64, cloudProductPlanningList []*entity.CloudProductPlanning, currentUserId string) error {
	var productIdList []int64
	dpdkCloudProductMap := make(map[int64]*entity.CloudProductPlanning)
	for _, cloudProductPlanning := range cloudProductPlanningList {
		productIdList = append(productIdList, cloudProductPlanning.ProductId)
		if cloudProductPlanning.SellSpec == constant.SellSpecHighPerformanceType {
			dpdkCloudProductMap[cloudProductPlanning.ProductId] = cloudProductPlanning
		}
	}
	if len(dpdkCloudProductMap) > 0 {
		var cloudProductBaselineList []*entity.CloudProductBaseline
		if err := db.Where("id IN (?)", productIdList).Find(&cloudProductBaselineList).Error; err != nil {
			return err
		}
		for _, cloudProductBaseline := range cloudProductBaselineList {
			if _, ok := dpdkCloudProductMap[cloudProductBaseline.Id]; ok && cloudProductBaseline.ProductType != constant.ProductTypeNetwork {
				delete(dpdkCloudProductMap, cloudProductBaseline.Id)
			}
		}
	}
	// 查询云产品和角色关联表
	var nodeRoleIdList []int64
	var cloudProductNodeRoleRelList []*entity.CloudProductNodeRoleRel
	if err := db.Model(&entity.CloudProductNodeRoleRel{}).Where("product_id IN (?)", productIdList).Find(&cloudProductNodeRoleRelList).Error; err != nil {
		return err
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
		return err
	}
	// 查询角色服务器基线map
	serverBaselineMap, screenNodeRoleServerBaselineMap, nodeRoleCodeBaselineMap, err := getNodeRoleServerBaselineMap(db, nodeRoleIdList, nodeRoleBaselineList)
	if err != nil {
		return err
	}
	// 查询已保存的服务器规划表
	var serverPlanningList []*entity.ServerPlanning
	if err = db.Where("plan_id = ? AND node_role_id IN (?)", planId, nodeRoleIdList).Find(&serverPlanningList).Error; err != nil {
		return err
	}
	var nodeRoleServerPlanningsMap = make(map[int64][]*entity.ServerPlanning)
	for _, serverPlanning := range serverPlanningList {
		nodeRoleServerPlanningsMap[serverPlanning.NodeRoleId] = append(nodeRoleServerPlanningsMap[serverPlanning.NodeRoleId], serverPlanning)
	}
	var resourcePoolList []*entity.ResourcePool
	if err = db.Where("plan_id = ? AND node_role_id IN (?)", planId, nodeRoleIdList).Find(&resourcePoolList).Error; err != nil {
		return err
	}
	resourcePoolNodeRoleIdMap := make(map[int64][]*entity.ResourcePool)
	for _, resourcePool := range resourcePoolList {
		resourcePoolNodeRoleIdMap[resourcePool.NodeRoleId] = append(resourcePoolNodeRoleIdMap[resourcePool.NodeRoleId], resourcePool)
	}
	resourcePoolIdServerPlanningMap := make(map[int64]*entity.ServerPlanning)
	// 构建返回体
	var list []*entity.ServerPlanning
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
			var addDpdkServerPlanningFlag bool
			for i, serverPlanning := range nodeRoleServerPlannings {
				if serverPlanning.OpenDpdk == constant.OpenDpdk && len(dpdkNodeRoleMap[nodeRoleBaseline.Id]) == 0 {
					continue
				}
				resourcePool := resourcePoolMap[serverPlanning.ResourcePoolId]
				if resourcePool != nil {
					if serverPlanning.OpenDpdk == constant.CloseDpdk && len(dpdkNodeRoleMap[nodeRoleBaseline.Id]) > 0 && len(nodeRoleServerPlannings) == 1 {
						addDpdkServerPlanningFlag = true
					}
				} else {
					openDpdk := constant.CloseDpdk
					resourcePoolName := fmt.Sprintf("%s-%s-%d", nodeRoleBaseline.NodeRoleName, constant.ResourcePoolDefaultName, i+1)
					if nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeNFV {
						resourcePoolName = constant.NFVResourcePoolNameKernel
						if serverPlanning.OpenDpdk == constant.OpenDpdk {
							openDpdk = constant.OpenDpdk
							resourcePoolName = constant.NFVResourcePoolNameDpdk
						}
					}
					resourcePool = &entity.ResourcePool{
						PlanId:              planId,
						NodeRoleId:          nodeRoleBaseline.Id,
						ResourcePoolName:    resourcePoolName,
						OpenDpdk:            openDpdk,
						DefaultResourcePool: constant.Yes,
					}
					if err = db.Table(entity.ResourcePoolTable).Save(&resourcePool).Error; err != nil {
						log.Errorf("save resource pool error: %v", err)
						return err
					}
					serverPlanning.ResourcePoolId = resourcePool.Id
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
			if addDpdkServerPlanningFlag && nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeNFV {
				dpdkServerPlanning, err := addDpdkServerPlanning(db, planId, nodeRoleBaseline, serverBaseline, resourcePoolIdServerPlanningMap)
				if err != nil {
					return err
				}
				list = append(list, dpdkServerPlanning)
			}
		} else {
			serverBaseline := screenNodeRoleServerBaselineMap[nodeRoleBaseline.Id]
			serverPlanning := &entity.ServerPlanning{
				PlanId:           planId,
				NodeRoleId:       nodeRoleBaseline.Id,
				Number:           nodeRoleBaseline.MinimumNum,
				ServerBaselineId: serverBaseline.Id,
				MixedNodeRoleId:  nodeRoleBaseline.Id,
				NetworkInterface: serverBaseline.NetworkInterface,
				CpuType:          serverBaseline.CpuType,
			}
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
					PlanId:              planId,
					NodeRoleId:          nodeRoleBaseline.Id,
					ResourcePoolName:    resourcePoolName,
					OpenDpdk:            constant.CloseDpdk,
					DefaultResourcePool: constant.Yes,
				}
				if err = db.Table(entity.ResourcePoolTable).Save(&resourcePool).Error; err != nil {
					log.Errorf("save resource pool error: %v", err)
					return err
				}
			}
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
			if len(dpdkNodeRoleMap[nodeRoleBaseline.Id]) > 0 && nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeNFV {
				dpdkServerPlanning, err := addDpdkServerPlanning(db, planId, nodeRoleBaseline, serverBaseline, resourcePoolIdServerPlanningMap)
				if err != nil {
					return err
				}
				list = append(list, dpdkServerPlanning)
			}
		}
	}
	// 计算已保存的容量规划指标
	resourcePoolCapMap, err := capacity_planning.GetResourcePoolCapMap(db, &capacity_planning.Request{PlanId: planId}, resourcePoolIdServerPlanningMap, nodeRoleCodeBaselineMap, serverBaselineMap)
	if err != nil {
		return err
	}
	var resourceIdList []int64
	for i, server := range list {
		resourceIdList = append(resourceIdList, server.ResourcePoolId)
		if serverPlanning, ok := resourcePoolIdServerPlanningMap[server.ResourcePoolId]; ok {
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
	}
	if _, err = capacity_planning.HandleBmsGWAndMasterServerNum(list, nodeRoleCodeBaselineMap, true); err != nil {
		return err
	}
	if err = db.Table(entity.ResourcePoolTable).Where("plan_id = ? and id not in (?)", planId, resourceIdList).Delete(&entity.ResourcePool{}).Error; err != nil {
		log.Errorf("delete resource pool error: %v", err)
		return err
	}
	if err = db.Where("plan_id = ?", planId).Delete(&entity.ServerPlanning{}).Error; err != nil {
		log.Errorf("delete server planning error: %v", err)
		return err
	}
	if err = db.Table(entity.ServerCapPlanningTable).Where("plan_id = ? and resource_pool_id not in (?)", planId, resourceIdList).Delete(&entity.ServerCapPlanning{}).Error; err != nil {
		log.Errorf("delete server cap planning error: %v", err)
		return err
	}
	now := datetime.GetNow()
	var serverPlanningEntityList []*entity.ServerPlanning
	for _, serverPlanning := range list {
		serverPlanningEntityList = append(serverPlanningEntityList, &entity.ServerPlanning{
			PlanId:           planId,
			NodeRoleId:       serverPlanning.NodeRoleId,
			MixedNodeRoleId:  serverPlanning.MixedNodeRoleId,
			ServerBaselineId: serverPlanning.ServerBaselineId,
			Number:           serverPlanning.Number,
			OpenDpdk:         serverPlanning.OpenDpdk,
			NetworkInterface: serverPlanning.NetworkInterface,
			CpuType:          serverPlanning.CpuType,
			CreateUserId:     currentUserId,
			CreateTime:       now,
			UpdateUserId:     currentUserId,
			UpdateTime:       now,
			DeleteState:      0,
			ResourcePoolId:   serverPlanning.ResourcePoolId,
		})
	}
	if err = db.Create(&serverPlanningEntityList).Error; err != nil {
		return err
	}
	return nil
}

func getNodeRoleServerBaselineMap(db *gorm.DB, nodeRoleIdList []int64, nodeRoleBaselineList []*entity.NodeRoleBaseline) (map[int64]*entity.ServerBaseline, map[int64]*entity.ServerBaseline, map[string]*entity.NodeRoleBaseline, error) {
	var nodeRoleBaselineMap = make(map[int64]*entity.NodeRoleBaseline)
	for _, v := range nodeRoleBaselineList {
		nodeRoleBaselineMap[v.Id] = v
	}
	// 查询服务器和角色关联表
	var serverNodeRoleRelList []*entity.ServerNodeRoleRel
	if err := db.Where("node_role_id IN (?)", nodeRoleIdList).Find(&serverNodeRoleRelList).Error; err != nil {
		return nil, nil, nil, err
	}
	var nodeRoleServerRelMap = make(map[int64][]int64)
	var serverBaselineIdList []int64
	for _, serverNodeRoleRel := range serverNodeRoleRelList {
		nodeRoleServerRelMap[serverNodeRoleRel.NodeRoleId] = append(nodeRoleServerRelMap[serverNodeRoleRel.NodeRoleId], serverNodeRoleRel.ServerId)
		serverBaselineIdList = append(serverBaselineIdList, serverNodeRoleRel.ServerId)
	}
	// 查询服务器基线表
	var serverBaselineList []*entity.ServerBaseline
	if err := db.Where("id IN (?)", serverBaselineIdList).Find(&serverBaselineList).Error; err != nil {
		return nil, nil, nil, err
	}
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	for _, serverBaseline := range serverBaselineList {
		serverBaselineMap[serverBaseline.Id] = serverBaseline
	}
	// 查询服务器基线表
	var screenNodeRoleServerBaselineMap = make(map[int64]*entity.ServerBaseline)
	var nodeRoleCodeBaselineMap = make(map[string]*entity.NodeRoleBaseline)
	for nodeRoleId, serverIdList := range nodeRoleServerRelMap {
		for _, serverId := range serverIdList {
			serverBaseline := serverBaselineMap[serverId]
			if serverBaseline == nil {
				continue
			}
			if screenNodeRoleServerBaselineMap[nodeRoleId] == nil {
				screenNodeRoleServerBaselineMap[nodeRoleId] = serverBaseline
				nodeRoleCodeBaselineMap[nodeRoleBaselineMap[nodeRoleId].NodeRoleCode] = nodeRoleBaselineMap[nodeRoleId]
			}
		}
	}
	return serverBaselineMap, screenNodeRoleServerBaselineMap, nodeRoleCodeBaselineMap, nil
}

func addDpdkServerPlanning(db *gorm.DB, planId int64, nodeRoleBaseline *entity.NodeRoleBaseline, serverBaseline *entity.ServerBaseline, resourcePoolServerPlanningMap map[int64]*entity.ServerPlanning) (*entity.ServerPlanning, error) {
	dpdkServerPlanning := &entity.ServerPlanning{
		PlanId:           planId,
		NodeRoleId:       nodeRoleBaseline.Id,
		Number:           nodeRoleBaseline.MinimumNum,
		ServerBaselineId: serverBaseline.Id,
		MixedNodeRoleId:  nodeRoleBaseline.Id,
		NetworkInterface: serverBaseline.NetworkInterface,
		CpuType:          serverBaseline.CpuType,
		OpenDpdk:         constant.OpenDpdk,
	}
	resourcePool := &entity.ResourcePool{
		PlanId:              planId,
		NodeRoleId:          nodeRoleBaseline.Id,
		ResourcePoolName:    constant.NFVResourcePoolNameDpdk,
		OpenDpdk:            constant.OpenDpdk,
		DefaultResourcePool: constant.Yes,
	}
	if err := db.Table(entity.ResourcePoolTable).Save(&resourcePool).Error; err != nil {
		log.Errorf("save resource pool error: %v", err)
		return nil, err
	}
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

func ListCloudProductPlanningByPlanId(planId int64) ([]entity.CloudProductPlanning, error) {
	var cloudProductPlanningList []entity.CloudProductPlanning
	if err := data.DB.Table(entity.CloudProductPlanningTable).Where("plan_id=?", planId).Scan(&cloudProductPlanningList).Error; err != nil {
		log.Errorf("[ListCloudProductPlanningByPlanId] error, %v", err)
		return nil, err
	}
	return cloudProductPlanningList, nil
}

func exportCloudProductPlanningByPlanId(planId int64) (string, []CloudProductPlanningExportResponse, error) {
	var planManage entity.PlanManage
	if err := data.DB.Table(entity.PlanManageTable).Where("id=?", planId).Scan(&planManage).Error; err != nil {
		log.Errorf("[exportCloudProductPlanningByPlanId] get planManage by id err, %v", err)
		return "", nil, err
	}

	var projectManage entity.ProjectManage
	if err := data.DB.Table(entity.ProjectManageTable).Where("id=?", planManage.ProjectId).Scan(&projectManage).Error; err != nil {
		log.Errorf("[exportCloudProductPlanningByPlanId] get projectManage by id err, %v", err)
		return "", nil, err
	}

	var cloudProductPlanningList []entity.CloudProductPlanning
	if err := data.DB.Table(entity.CloudProductPlanningTable).Where("plan_id=?", planId).Scan(&cloudProductPlanningList).Error; err != nil {
		log.Errorf("[exportCloudProductPlanningByPlanId] error, %v", err)
		return "", nil, err
	}

	var response []CloudProductPlanningExportResponse
	if err := data.DB.Table("cloud_product_planning cpp").Select("cpb.product_type,cpb.product_name, cpp.sell_spec, cpp.value_added_service, cpb.instructions").
		Joins("LEFT JOIN cloud_product_baseline cpb ON cpb.id = cpp.product_id").
		Where("cpp.plan_id=?", planId).
		Find(&response).Error; err != nil {
		log.Errorf("[exportCloudProductPlanningByPlanId] query db error")
		return "", nil, err
	}
	return projectManage.Name + "-" + planManage.Name + "-" + "云产品清单", response, nil
}

func getDependProductIds() ([]CloudProductBaselineDependResponse, error) {
	var cloudProductDependList []CloudProductBaselineDependResponse
	if err := data.DB.Table("cloud_product_depend_rel cpdr").
		Select("cpdr.product_id id, cpb.id dependId, cpb.product_name dependProductName, cpb.product_code dependProductCode").
		Joins("LEFT JOIN cloud_product_baseline cpb ON cpb.id = cpdr.depend_product_id").
		Find(&cloudProductDependList).Error; err != nil {

		log.Errorf("[getDependProductIds] query db err, %v", err)
		return nil, err
	}
	return cloudProductDependList, nil
}
