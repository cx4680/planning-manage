package server

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"errors"
	"gorm.io/gorm"
	"strconv"
	"time"
)

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
	//查询角色服务器基线map
	serverBaselineMap, nodeRoleServerBaselineListMap, err := getNodeRoleServerBaselineMap(nodeRoleIdList, request)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	//查询混合部署方式map
	mixedNodeRoleMap, err := getMixedNodeRoleMap(nodeRoleIdList)
	if err != nil {
		return nil, err
	}
	//查询已保存的服务器规划表
	var serverPlanningList []*entity.ServerPlanning
	if err = data.DB.Where("plan_id = ?", request.PlanId).Find(&serverPlanningList).Error; err != nil {
		return nil, err
	}
	var serverPlanningMap = make(map[int64]*entity.ServerPlanning)
	for _, v := range serverPlanningList {
		serverPlanningMap[v.NodeRoleId] = v
	}
	//构建返回体
	var list []*entity.ServerPlanning
	for _, v := range nodeRoleBaselineList {
		serverPlanning := &entity.ServerPlanning{}
		_, ok := serverPlanningMap[v.Id]
		if ok {
			serverPlanning = serverPlanningMap[v.Id]
			serverPlanning.ServerBomCode = serverBaselineMap[serverPlanning.ServerBaselineId].BomCode
			serverPlanning.ServerArch = serverBaselineMap[serverPlanning.ServerBaselineId].Arch
		} else {
			serverPlanning.PlanId = request.PlanId
			serverPlanning.NodeRoleId = v.Id
			serverPlanning.Number = v.MinimumNum
			if len(nodeRoleServerBaselineListMap[v.Id]) != 0 {
				serverPlanning.ServerBaselineId = nodeRoleServerBaselineListMap[v.Id][0].Id
				serverPlanning.ServerBomCode = nodeRoleServerBaselineListMap[v.Id][0].BomCode
				serverPlanning.ServerArch = nodeRoleServerBaselineListMap[v.Id][0].Arch
			}
			serverPlanning.MixedNodeRoleId = v.Id
		}
		serverPlanning.NodeRoleName = v.NodeRoleName
		serverPlanning.NodeRoleClassify = v.Classify
		serverPlanning.NodeRoleAnnotation = v.Annotation
		serverPlanning.SupportDpdk = v.SupportDPDK
		serverPlanning.ServerBaselineList = nodeRoleServerBaselineListMap[v.Id]
		serverPlanning.MixedNodeRoleList = mixedNodeRoleMap[v.Id]
		list = append(list, serverPlanning)
	}
	return list, nil
}

func SaveServer(request *Request) error {
	if err := checkBusiness(request); err != nil {
		return err
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("plan_id = ?", request.PlanId).Delete(&entity.ServerPlanning{})
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
				CreateUserId:     request.UserId,
				CreateTime:       now,
				UpdateUserId:     request.UserId,
				UpdateTime:       now,
				DeleteState:      0,
			})
		}
		if err := tx.Create(&serverPlanningEntityList).Error; err != nil {
			return err
		}
		if err := tx.Model(entity.PlanManage{}).Where("id = ?", request.PlanId).Updates(&entity.PlanManage{
			BusinessPanStage: constant.NETWORK_DEVICE_PLAN,
			UpdateUserId:     request.UserId,
			UpdateTime:       time.Now(),
		}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func ListServerNetworkType(request *Request) ([]string, error) {
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
	var networkTypeMap = make(map[string]interface{})
	var networkTypeList []string
	for _, v := range serverBaselineList {
		_, ok := networkTypeMap[v.NetworkInterface]
		if !ok {
			networkTypeMap[v.NetworkInterface] = struct{}{}
			networkTypeList = append(networkTypeList, v.NetworkInterface)
		}
	}
	return networkTypeList, nil
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
	var cpuTypeMap = make(map[string]interface{})
	var cpuTypeList []string
	for _, v := range serverBaselineList {
		_, ok := cpuTypeMap[v.CpuType]
		if !ok {
			cpuTypeMap[v.CpuType] = struct{}{}
			cpuTypeList = append(cpuTypeList, v.CpuType)
		}
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

func DownloadServer(planId int64) ([]*ResponseDownloadServer, string, error) {
	//查询服务器规划列表
	var serverList []*entity.ServerPlanning
	if err := data.DB.Where("plan_id = ?", planId).Find(&serverList).Error; err != nil {
		return nil, "", err
	}
	//查询关联的角色和设备，封装成map
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
	//构建返回体
	var response []*ResponseDownloadServer
	var total int
	for _, v := range serverList {
		response = append(response, &ResponseDownloadServer{
			NodeRole:   nodeRoleMap[v.NodeRoleId].NodeRoleName,
			ServerType: serverBaselineMap[v.ServerBaselineId].Arch,
			BomCode:    serverBaselineMap[v.ServerBaselineId].BomCode,
			Spec:       serverBaselineMap[v.ServerBaselineId].Spec,
			Number:     strconv.Itoa(v.Number),
		})
		total += v.Number
	}
	response = append(response, &ResponseDownloadServer{
		Number: "总计：" + strconv.Itoa(total) + "台",
	})
	//构建文件名称
	var planManage = &entity.PlanManage{}
	if err := data.DB.Where("id = ? AND delete_state = ?", planId, 0).First(&planManage).Error; err != nil {
		return nil, "", err
	}
	var projectManage = &entity.ProjectManage{}
	if err := data.DB.Where("id = ? AND delete_state = ?", planManage.ProjectId, 0).First(&projectManage).Error; err != nil {
		return nil, "", err
	}
	fileName := projectManage.Name + "-" + planManage.Name + "-" + "服务器规划清单"
	return response, fileName, nil
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
	if err := data.DB.Model(&entity.PlanManage{}).Where("id = ? AND delete_state = ?", request.PlanId, 0).Count(&planCount).Error; err != nil {
		return err
	}
	if planCount == 0 {
		return errors.New("方案不存在")
	}
	return nil
}

func getMixedNodeRoleMap(nodeRoleIdList []int64) (map[int64][]*entity.MixedNodeRole, error) {
	var nodeRoleIdMap = make(map[int64]interface{})
	var newNodeRoleId []int64
	for _, v := range nodeRoleIdList {
		_, ok := nodeRoleIdMap[v]
		if !ok {
			nodeRoleIdMap[v] = struct{}{}
			newNodeRoleId = append(newNodeRoleId, v)
		}
	}
	var nodeRoleMixedDeployList []*entity.NodeRoleMixedDeploy
	if err := data.DB.Where("node_role_id IN (?)", newNodeRoleId).Find(&nodeRoleMixedDeployList).Error; err != nil {
		return nil, err
	}
	var mixedNodeRoleIdList []int64
	for _, v := range nodeRoleMixedDeployList {
		mixedNodeRoleIdList = append(mixedNodeRoleIdList, v.MixedNodeRoleId)
	}
	var mixedNodeRoleBaselineList []*entity.NodeRoleBaseline
	if err := data.DB.Where("id IN (?)", newNodeRoleId).Find(&mixedNodeRoleBaselineList).Error; err != nil {
		return nil, err
	}
	var nodeRoleBaselineMap = make(map[int64]*entity.NodeRoleBaseline)
	for _, v := range mixedNodeRoleBaselineList {
		nodeRoleBaselineMap[v.Id] = v
	}
	var mixedNodeRoleMap = make(map[int64][]*entity.MixedNodeRole)
	for _, v := range newNodeRoleId {
		mixedNodeRoleMap[v] = append(mixedNodeRoleMap[v], &entity.MixedNodeRole{
			Id:   v,
			Name: "独立部署",
		})
	}
	for _, v := range nodeRoleMixedDeployList {
		mixedNodeRoleMap[v.NodeRoleId] = append(mixedNodeRoleMap[v.NodeRoleId], &entity.MixedNodeRole{
			Id:   nodeRoleBaselineMap[v.MixedNodeRoleId].Id,
			Name: "混合部署：" + nodeRoleBaselineMap[v.MixedNodeRoleId].NodeRoleName,
		})
	}
	return mixedNodeRoleMap, nil
}

func getNodeRoleServerBaselineMap(nodeRoleIdList []int64, request *Request) (map[int64]*entity.ServerBaseline, map[int64][]*entity.ServerModel, error) {
	//查询服务器和角色关联表
	var serverNodeRoleRelList []*entity.ServerNodeRoleRel
	if err := data.DB.Where("node_role_id IN (?)", nodeRoleIdList).Find(&serverNodeRoleRelList).Error; err != nil {
		return nil, nil, err
	}
	var nodeRoleServerRelMap = make(map[int64][]int64)
	var serverBaselineIdList []int64
	for _, v := range serverNodeRoleRelList {
		nodeRoleServerRelMap[v.NodeRoleId] = append(nodeRoleServerRelMap[v.NodeRoleId], v.ServerId)
		serverBaselineIdList = append(serverBaselineIdList, v.ServerId)
	}
	//查询服务器基线表
	var screenSql, screenParams = " id IN (?) ", []interface{}{serverBaselineIdList}
	if util.IsNotBlank(request.NetworkInterface) {
		screenSql += " AND network_interface = ? "
		screenParams = append(screenParams, request.NetworkInterface)
	}
	if util.IsNotBlank(request.CpuType) {
		screenSql += " AND cpu_type = ? "
		screenParams = append(screenParams, request.CpuType)
	}
	var serverBaselineList []*entity.ServerBaseline
	if err := data.DB.Where(screenSql, screenParams...).Find(&serverBaselineList).Error; err != nil {
		return nil, nil, err
	}
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	for _, v := range serverBaselineList {
		serverBaselineMap[v.Id] = v
	}

	//查询服务器基线表
	var nodeRoleServerBaselineMap = make(map[int64][]*entity.ServerModel)
	for k, serverIdList := range nodeRoleServerRelMap {
		for _, serverId := range serverIdList {
			nodeRoleServerBaselineMap[k] = append(nodeRoleServerBaselineMap[k], &entity.ServerModel{
				Id:                serverBaselineMap[serverId].Id,
				BomCode:           serverBaselineMap[serverId].BomCode,
				NetworkInterface:  serverBaselineMap[serverId].NetworkInterface,
				CpuType:           serverBaselineMap[serverId].CpuType,
				Arch:              serverBaselineMap[serverId].Arch,
				ConfigurationInfo: serverBaselineMap[serverId].ConfigurationInfo,
			})
		}
	}
	return serverBaselineMap, nodeRoleServerBaselineMap, nil
}

func getServerBaselineMap(productId int64, request *Request) (map[int64]*entity.ServerBaseline, error) {
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	//查询云产品基线表
	var cloudProductBaseline = &entity.CloudProductBaseline{}
	if err := data.DB.Where("id = ?", productId).First(&cloudProductBaseline).Error; err != nil {
		return nil, err
	}
	//查询服务器基线表
	screenSql, screenParams := " version_id = ? ", []interface{}{cloudProductBaseline.VersionId}
	if util.IsNotBlank(request.NetworkInterface) {
		screenSql += " AND network_interface = ? "
		screenParams = append(screenParams, request.NetworkInterface)
	}
	if util.IsNotBlank(request.CpuType) {
		screenSql += " AND cpu_type = ? "
		screenParams = append(screenParams, request.CpuType)
	}
	var serverBaselineList []*entity.ServerBaseline
	if err := data.DB.Where(screenSql, screenParams...).Find(&serverBaselineList).Error; err != nil {
		return nil, err
	}
	for _, v := range serverBaselineList {
		serverBaselineMap[v.Id] = v
	}
	return serverBaselineMap, nil
}
