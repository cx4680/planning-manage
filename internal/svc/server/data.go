package server

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"errors"
	"gorm.io/gorm"
)

type ResponseCapacity struct {
	Id              int64  `json:"id"`
	ProductId       int64  `json:"productId"`
	ProductName     string `json:"productName"`
	CapacitySpecs   string `json:"capacitySpecs"`
	SalesSpecs      string `json:"salesSpecs"`
	OverbookingRate string `json:"overbookingRate"`
	Number          string `json:"number"`
	Unit            string `json:"unit"`
}

func ListServer(request *Request) ([]*entity.ServerPlanning, error) {
	//查询云产品规划表
	var cloudProductPlanningList []*entity.CloudProductPlanning
	if err := data.DB.Where("plan_id = ?", request.PlanId).Find(&cloudProductPlanningList).Error; err != nil {
		return nil, err
	}
	if len(cloudProductPlanningList) == 0 {
		return nil, errors.New("该方案未找到关联产品")
	}
	var productIdList []int64
	for _, v := range cloudProductPlanningList {
		productIdList = append(productIdList, v.ProductId)
	}
	//查询云产品和角色关联表
	var cloudProductNodeRoleRelList []*entity.CloudProductNodeRoleRel
	if err := data.DB.Where("product_id IN (?)", productIdList).Find(&cloudProductNodeRoleRelList).Error; err != nil {
		return nil, err
	}
	var nodeRoleIdList []int64
	for _, v := range cloudProductNodeRoleRelList {
		nodeRoleIdList = append(nodeRoleIdList, v.NodeRoleId)
	}
	//查询角色表
	var nodeRoleBaselineList []*entity.NodeRoleBaseline
	if err := data.DB.Where("id IN (?)", nodeRoleIdList).Find(&nodeRoleBaselineList).Error; err != nil {
		return nil, err
	}
	//查询服务器规划表
	var serverPlanningList []*entity.ServerPlanning
	if err := data.DB.Where("plan_id = ?", request.PlanId).Find(&serverPlanningList).Error; err != nil {
		return nil, err
	}
	var serverPlanningMap = make(map[int64]*entity.ServerPlanning)
	for _, v := range serverPlanningList {
		serverPlanningMap[v.NodeRoleId] = v
	}
	//查询机型
	serverModelMap, err := getServerModel(nodeRoleBaselineList)
	if err != nil {
		return nil, err
	}
	//查询cpuType
	serverBaselineIdMap, serverBaselineCpuTypeMap, err := getServerBaselineMap(productIdList[0])
	if err != nil {
		return nil, err
	}
	//查询部署方式map
	mixedNodeRoleMap, err := getMixedNodeRoleMap(nodeRoleIdList)
	if err != nil {
		return nil, err
	}
	//构建返回体
	var list []*entity.ServerPlanning
	for _, v := range nodeRoleBaselineList {
		serverPlanning := &entity.ServerPlanning{}
		_, ok := serverPlanningMap[v.Id]
		if ok {
			serverPlanning = serverPlanningMap[v.Id]
			serverPlanning.ServerBaselineId = serverBaselineIdMap[serverPlanning.ServerBaselineId].Id
			serverPlanning.ServerModel = serverBaselineIdMap[serverPlanning.ServerBaselineId].BomCode
			serverPlanning.ServerArch = serverBaselineIdMap[serverPlanning.ServerBaselineId].Arch
		} else {
			serverPlanning.PlanId = request.PlanId
			serverPlanning.NodeRoleId = v.Id
			serverPlanning.Number = v.MinimumNum
		}
		if util.IsNotBlank(request.CpuType) {
			serverPlanning.ServerBaselineId = serverBaselineCpuTypeMap[request.CpuType].Id
			serverPlanning.ServerModel = serverBaselineCpuTypeMap[request.CpuType].BomCode
			serverPlanning.ServerArch = serverBaselineCpuTypeMap[request.CpuType].Arch
		}
		serverPlanning.NodeRoleName = v.NodeRoleName
		serverPlanning.NodeRoleClassify = v.Classify
		serverPlanning.NodeRoleAnnotation = v.Annotation
		serverPlanning.SupportDpdk = v.SupportDPDK
		serverPlanning.MixedNodeRoleList = mixedNodeRoleMap[v.Id]
		serverPlanning.ServerModelList = serverModelMap[v.Id]
		list = append(list, serverPlanning)
	}
	return list, nil
}

func SaveServer(request *Request) error {
	if err := checkBusiness(request); err != nil {
		return err
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("planId = ?", request.PlanId).Delete(&entity.ServerPlanning{})
		now := datetime.GetNow()
		var serverPlanningEntityList []*entity.ServerPlanning
		for _, v := range request.serverList {
			serverPlanningEntityList = append(serverPlanningEntityList, &entity.ServerPlanning{
				PlanId:           request.PlanId,
				NodeRoleId:       v.NodeRoleId,
				MixedNodeRoleId:  v.MixedNodeRoleId,
				ServerBaselineId: v.ServerBaselineId,
				Number:           v.Number,
				OpenDpdk:         v.OpenDpdk,
				CreateUserId:     request.UserId,
				CreateTime:       now,
				UpdateUserId:     request.UserId,
				UpdateTime:       now,
				DeleteState:      0,
			})
		}
		if err := data.DB.Create(serverPlanningEntityList).Error; err != nil {
			return err
		}
		if err := data.DB.Model(&entity.PlanManage{}).Where("id = ?", request.PlanId).Update("BusinessPanStage", 2).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func ListServerCpuType(request *Request) ([]string, error) {
	//查询云产品规划表
	var cloudProductPlanning = &entity.CloudProductPlanning{}
	if err := data.DB.Where("plan_id = ?", request.PlanId).First(&cloudProductPlanning).Error; err != nil {
		return nil, err
	}
	//查询云产品基线表
	var cloudProductBaseline = &entity.CloudProductBaseline{}
	if err := data.DB.Where("id = ?", cloudProductPlanning.ProductId).First(&cloudProductBaseline).Error; err != nil {
		return nil, err
	}
	//查询服务器基线表
	var serverBaselineList []*entity.ServerBaseline
	if err := data.DB.Where("version_id = ?", cloudProductBaseline.VersionId).Find(&serverBaselineList).Error; err != nil {
		return nil, err
	}
	var cpuTypeList []string
	for _, v := range serverBaselineList {
		cpuTypeList = append(cpuTypeList, v.CpuType)
	}
	return cpuTypeList, nil
}

func ListServerCapacity(request *Request) ([]*ResponseCapacity, error) {
	//查询云产品规划表
	var cloudProductPlanning = &entity.CloudProductPlanning{}
	if err := data.DB.Where("plan_id = ?", request.PlanId).First(&cloudProductPlanning).Error; err != nil {
		return nil, err
	}
	//查询云产品基线表
	var cloudProductBaseline = &entity.CloudProductBaseline{}
	if err := data.DB.Where("id = ?", cloudProductPlanning.ProductId).First(&cloudProductBaseline).Error; err != nil {
		return nil, err
	}
	//todo 缺少产品容量指标
	var capacityList []*ResponseCapacity

	return capacityList, nil
}

//func ListServerModel(request *Request) ([]*entity.ServerBaseline, error) {
//	//查询云产品规划表
//	var cloudProductPlanning = &entity.CloudProductPlanning{}
//	if err := data.DB.Where("plan_id = ?", request.PlanId).First(&cloudProductPlanning).Error; err != nil {
//		return nil, err
//	}
//	//查询服务器和角色关联表
//	var serverNodeRoleRelList []*entity.ServerNodeRoleRel
//	if err := data.DB.Where("node_role_id = ?", request.NodeRoleId).Find(&serverNodeRoleRelList).Error; err != nil {
//		return nil, err
//	}
//	var serverIdList []int64
//	for _, v := range serverNodeRoleRelList {
//		serverIdList = append(serverIdList, v.ServerId)
//	}
//	//查询服务器基线表
//	var serverBaselineList []*entity.ServerBaseline
//	if err := data.DB.Where("id IN (?)", serverIdList).Find(&serverBaselineList).Error; err != nil {
//		return nil, err
//	}
//	return serverBaselineList, nil
//}

func QueryServerPlanningListByPlanId(planId int64) ([]entity.ServerPlanning, error) {
	var serverPlanningList []entity.ServerPlanning
	if err := data.DB.Table(entity.ServerPlanningTable).Where("plan_id = ? and delete_state = 0", planId).Find(&serverPlanningList).Error; err != nil {
		return serverPlanningList, err
	}
	return serverPlanningList, nil
}

func checkBusiness(request *Request) error {
	//校验planId
	var planCount int64
	if err := data.DB.Model(&entity.PlanManage{}).Where("id = ? AND delete_state = ?", request.Id, 0).Count(&planCount).Error; err != nil {
		return err
	}
	if planCount == 0 {
		return errors.New("方案不存在")
	}
	return nil
}

func getMixedNodeRoleMap(nodeRoleIdList []int64) (map[int64][]*entity.MixedNodeRole, error) {
	var nodeRoleMixedDeployList []*entity.NodeRoleMixedDeploy
	if err := data.DB.Where("node_role_id IN (?)", nodeRoleIdList).Find(&nodeRoleMixedDeployList).Error; err != nil {
		return nil, err
	}
	var mixedNodeRoleIdList []int64
	for _, v := range nodeRoleMixedDeployList {
		mixedNodeRoleIdList = append(mixedNodeRoleIdList, v.MixedNodeRoleId)
	}
	var mixedNodeRoleBaselineList []*entity.NodeRoleBaseline
	if err := data.DB.Where("id IN (?)", nodeRoleIdList).Find(&mixedNodeRoleBaselineList).Error; err != nil {
		return nil, err
	}
	var nodeRoleBaselineMap = make(map[int64]*entity.NodeRoleBaseline)
	for _, v := range mixedNodeRoleBaselineList {
		nodeRoleBaselineMap[v.Id] = v
	}
	var mixedNodeRoleMap = make(map[int64][]*entity.MixedNodeRole)
	for _, v := range nodeRoleMixedDeployList {
		mixedNodeRoleMap[v.NodeRoleId] = append(mixedNodeRoleMap[v.NodeRoleId], &entity.MixedNodeRole{
			Id:   nodeRoleBaselineMap[v.MixedNodeRoleId].Id,
			Name: "混合部署：" + nodeRoleBaselineMap[v.MixedNodeRoleId].NodeRoleName,
		})
	}
	for k := range mixedNodeRoleMap {
		mixedNodeRoleMap[k] = append(mixedNodeRoleMap[k], &entity.MixedNodeRole{
			Id:   k,
			Name: "独立部署",
		})
	}
	return mixedNodeRoleMap, nil
}

func getServerModel(nodeRoleBaselineList []*entity.NodeRoleBaseline) (map[int64][]*entity.ServerModel, error) {
	var nodeRoleIdList []int64
	for _, v := range nodeRoleBaselineList {
		nodeRoleIdList = append(nodeRoleIdList, v.Id)
	}
	//查询服务器和角色关联表
	var serverNodeRoleRelList []*entity.ServerNodeRoleRel
	if err := data.DB.Where("node_role_id IN (?)", nodeRoleIdList).Find(&serverNodeRoleRelList).Error; err != nil {
		return nil, err
	}
	var serverNodeRoleRelMap = make(map[int64][]int64)
	for _, v := range serverNodeRoleRelList {
		serverNodeRoleRelMap[v.NodeRoleId] = append(serverNodeRoleRelMap[v.NodeRoleId], v.ServerId)
	}
	//查询服务器基线表
	var serverModelMap = make(map[int64][]*entity.ServerModel)
	for k, v := range serverNodeRoleRelMap {
		var serverBaselineList []*entity.ServerBaseline
		if err := data.DB.Where("id IN (?)", v).Find(&serverBaselineList).Error; err != nil {
			return nil, err
		}
		for _, serverBaseline := range serverBaselineList {
			serverModelMap[k] = append(serverModelMap[k], &entity.ServerModel{
				Id:                serverBaseline.Id,
				BomCode:           serverBaseline.BomCode,
				Arch:              serverBaseline.Arch,
				ConfigurationInfo: serverBaseline.ConfigurationInfo,
			})
		}
	}
	return serverModelMap, nil
}

func getServerBaselineMap(productId int64) (map[int64]*entity.ServerBaseline, map[string]*entity.ServerBaseline, error) {
	var serverBaselineIdMap = make(map[int64]*entity.ServerBaseline)
	var serverBaselineCpuTypeMap = make(map[string]*entity.ServerBaseline)
	//查询云产品基线表
	var cloudProductBaseline = &entity.CloudProductBaseline{}
	if err := data.DB.Where("id = ?", productId).First(&cloudProductBaseline).Error; err != nil {
		return nil, nil, err
	}
	//查询服务器基线表
	var serverBaselineList []*entity.ServerBaseline
	if err := data.DB.Where("version_id = ?", cloudProductBaseline.VersionId).Find(&serverBaselineList).Error; err != nil {
		return nil, nil, err
	}
	for _, v := range serverBaselineList {
		serverBaselineIdMap[v.Id] = v
		serverBaselineCpuTypeMap[v.CpuType] = v
	}
	return serverBaselineIdMap, serverBaselineCpuTypeMap, nil
}
