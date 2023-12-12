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
	//缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	//查询云产品规划表
	var cloudProductPlanningList []*entity.CloudProductPlanning
	if err := db.Where("plan_id = ?", request.PlanId).Find(&cloudProductPlanningList).Error; err != nil {
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
	if err := db.Where("product_id IN (?)", productIdList).Find(&cloudProductNodeRoleRelList).Error; err != nil {
		return nil, err
	}
	var nodeRoleIdList []int64
	for _, v := range cloudProductNodeRoleRelList {
		nodeRoleIdList = append(nodeRoleIdList, v.NodeRoleId)
	}
	//查询角色表
	var nodeRoleBaselineList []*entity.NodeRoleBaseline
	if err := db.Where("id IN (?)", nodeRoleIdList).Find(&nodeRoleBaselineList).Error; err != nil {
		return nil, err
	}
	//查询角色服务器基线map
	serverBaselineMap, nodeRoleServerBaselineListMap, screenNodeRoleServerBaselineListMap, err := getNodeRoleServerBaselineMap(db, nodeRoleIdList, request)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	//查询混合部署方式map
	mixedNodeRoleMap, err := getMixedNodeRoleMap(db, nodeRoleIdList)
	if err != nil {
		return nil, err
	}
	//查询已保存的服务器规划表
	var serverPlanningList []*entity.ServerPlanning
	if err = db.Where("plan_id = ?", request.PlanId).Find(&serverPlanningList).Error; err != nil {
		return nil, err
	}
	var nodeRoleServerPlanningMap = make(map[int64]*entity.ServerPlanning)
	for _, v := range serverPlanningList {
		nodeRoleServerPlanningMap[v.NodeRoleId] = v
	}
	//构建返回体
	var list []*entity.ServerPlanning
	for _, v := range nodeRoleBaselineList {
		serverPlanning := &entity.ServerPlanning{}
		//若服务器规划有保存过，则加载已保存的数据
		if _, ok := nodeRoleServerPlanningMap[v.Id]; ok {
			serverPlanning = nodeRoleServerPlanningMap[v.Id]
			serverPlanning.ServerBomCode = serverBaselineMap[serverPlanning.ServerBaselineId].BomCode
			serverPlanning.ServerArch = serverBaselineMap[serverPlanning.ServerBaselineId].Arch
		} else {
			serverPlanning.PlanId = request.PlanId
			serverPlanning.NodeRoleId = v.Id
			serverPlanning.Number = v.MinimumNum
			//列表加载查询的可用机型
			serverBaselineList := screenNodeRoleServerBaselineListMap[v.Id]
			if len(serverBaselineList) != 0 {
				serverPlanning.ServerBaselineId = serverBaselineList[0].Id
				serverPlanning.ServerBomCode = serverBaselineList[0].BomCode
				serverPlanning.ServerArch = serverBaselineList[0].Arch
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
	var cloudProductPlanningList []*entity.CloudProductPlanning
	if err := data.DB.Where("plan_id = ?", request.PlanId).Find(&cloudProductPlanningList).Error; err != nil {
		return nil, err
	}
	if len(cloudProductPlanningList) == 0 {
		return nil, errors.New("云产品规划不存在")
	}
	//查询云产品基线表
	var cloudProductBaselineList []*entity.CloudProductBaseline
	if err := data.DB.Where("id = ?", cloudProductPlanningList[0].ProductId).Find(&cloudProductBaselineList).Error; err != nil {
		return nil, err
	}
	if len(cloudProductPlanningList) == 0 {
		return nil, errors.New("云产品基线不存在")
	}
	//查询服务器基线表
	var serverBaselineList []*entity.ServerBaseline
	if err := data.DB.Where("version_id = ?", cloudProductBaselineList[0].VersionId).Find(&serverBaselineList).Error; err != nil {
		return nil, err
	}
	var networkTypeMap = make(map[string]interface{})
	var networkTypeList []string
	for _, v := range serverBaselineList {
		if _, ok := networkTypeMap[v.NetworkInterface]; !ok {
			networkTypeMap[v.NetworkInterface] = struct{}{}
			networkTypeList = append(networkTypeList, v.NetworkInterface)
		}
	}
	return networkTypeList, nil
}

func ListServerCpuType(request *Request) ([]string, error) {
	//查询云产品规划表
	var cloudProductPlanningList []*entity.CloudProductPlanning
	if err := data.DB.Where("plan_id = ?", request.PlanId).Find(&cloudProductPlanningList).Error; err != nil {
		return nil, err
	}
	if len(cloudProductPlanningList) == 0 {
		return nil, errors.New("云产品规划不存在")
	}
	//查询云产品基线表
	var cloudProductBaselineList []*entity.CloudProductBaseline
	if err := data.DB.Where("id = ?", cloudProductPlanningList[0].ProductId).Find(&cloudProductBaselineList).Error; err != nil {
		return nil, err
	}
	if len(cloudProductPlanningList) == 0 {
		return nil, errors.New("云产品基线不存在")
	}
	//查询服务器基线表
	var serverBaselineList []*entity.ServerBaseline
	if err := data.DB.Where("version_id = ?", cloudProductBaselineList[0].VersionId).Find(&serverBaselineList).Error; err != nil {
		return nil, err
	}
	var cpuTypeMap = make(map[string]interface{})
	var cpuTypeList []string
	for _, v := range serverBaselineList {
		if _, ok := cpuTypeMap[v.CpuType]; !ok {
			cpuTypeMap[v.CpuType] = struct{}{}
			cpuTypeList = append(cpuTypeList, v.CpuType)
		}
	}
	return cpuTypeList, nil
}

func ListServerCapacity(request *Request) ([]*ResponseCapConvert, error) {
	//缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	//查询云产品规划表
	var cloudProductPlanningList []*entity.CloudProductPlanning
	if err := db.Where("plan_id = ?", request.PlanId).Find(&cloudProductPlanningList).Error; err != nil {
		return nil, err
	}
	if len(cloudProductPlanningList) == 0 {
		return nil, errors.New("云产品规划错误")
	}
	var cloudProductIdList []int64
	for _, v := range cloudProductPlanningList {
		cloudProductIdList = append(cloudProductIdList, v.ProductId)
	}
	//查询云产品基线表
	var cloudProductBaselineList []*entity.CloudProductBaseline
	if err := db.Where("id IN (?) AND version_id = ?", cloudProductIdList, cloudProductPlanningList[0].VersionId).Find(&cloudProductBaselineList).Error; err != nil {
		return nil, err
	}
	var cloudProductCodeList []string
	for _, v := range cloudProductBaselineList {
		cloudProductCodeList = append(cloudProductCodeList, v.ProductCode)
	}
	//查询容量换算表
	var capConvertBaselineList []*entity.CapConvertBaseline
	if err := db.Where("product_code IN (?)", cloudProductCodeList).Find(&capConvertBaselineList).Error; err != nil {
		return nil, err
	}
	var capConvertBaselineMap = make(map[string][]*ResponseFeatures)
	var ServerPlanningBaseline []*ResponseCapConvert
	for _, v := range capConvertBaselineList {
		key := v.ProductCode + v.SellSpecs + v.CapPlanningInput
		if _, ok := capConvertBaselineMap[key]; !ok {
			ServerPlanningBaseline = append(ServerPlanningBaseline, &ResponseCapConvert{
				VersionId:        v.VersionId,
				ProductName:      v.ProductName,
				ProductCode:      v.ProductCode,
				SellSpecs:        v.SellSpecs,
				CapPlanningInput: v.CapPlanningInput,
				Unit:             v.Unit,
				Description:      v.Description,
			})
		}
		capConvertBaselineMap[key] = append(capConvertBaselineMap[key], &ResponseFeatures{
			Id:   v.Id,
			Name: v.Features,
		})
	}
	for i, v := range ServerPlanningBaseline {
		key := v.ProductCode + v.SellSpecs + v.CapPlanningInput
		ServerPlanningBaseline[i].Features = capConvertBaselineMap[key]
	}
	return ServerPlanningBaseline, nil
}

func SaveServerCapacity(request *Request) error {
	var serverCapPlanning = &entity.ServerCapPlanning{
		PlanId:             request.PlanId,
		CapacityBaselineId: request.CapacityBaselineId,
		Number:             request.Number,
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		//查询服务器容量数据
		var capConvertBaselineList []*entity.CapConvertBaseline
		if err := tx.Where("id = ?", request.CapacityBaselineId).Find(&capConvertBaselineList).Error; err != nil {
			return err
		}
		if len(capConvertBaselineList) == 0 {
			return errors.New("服务器容量规划指标不存在")
		}
		capConvertBaseline := capConvertBaselineList[0]
		var capActualResBaselineList []*entity.CapActualResBaseline
		if err := tx.Where("version_id = ? AND product_code = ? AND sell_specs = ? AND sell_unit = ? AND features = ?",
			capConvertBaseline.VersionId, capConvertBaseline.ProductCode, capConvertBaseline.SellSpecs, capConvertBaseline.CapPlanningInput, capConvertBaseline.Features).
			Find(&capActualResBaselineList).Error; err != nil {
			return err
		}
		if len(capActualResBaselineList) == 0 {
			return errors.New("服务器容量规划特性不存在")
		}
		capActualResBaseline := capActualResBaselineList[0]
		var capServerCalcBaselineList []*entity.CapServerCalcBaseline
		if err := tx.Where("expend_res_code = ?", capActualResBaseline.ExpendResCode).Find(&capServerCalcBaselineList).Error; err != nil {
			return err
		}
		if len(capServerCalcBaselineList) == 0 {
			return errors.New("服务器数量计算数据不存在")
		}
		//计算服务器规划数据

		//保存服务器规划数据

		//保存服务器容量规划
		if err := tx.Create(&serverCapPlanning).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
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

func getMixedNodeRoleMap(db *gorm.DB, nodeRoleIdList []int64) (map[int64][]*entity.MixedNodeRole, error) {
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

func getNodeRoleServerBaselineMap(db *gorm.DB, nodeRoleIdList []int64, request *Request) (map[int64]*entity.ServerBaseline, map[int64][]*entity.ServerPlanningBaseline, map[int64][]*entity.ServerBaseline, error) {
	//查询服务器和角色关联表
	var serverNodeRoleRelList []*entity.ServerNodeRoleRel
	if err := db.Where("node_role_id IN (?)", nodeRoleIdList).Find(&serverNodeRoleRelList).Error; err != nil {
		return nil, nil, nil, err
	}
	var nodeRoleServerRelMap = make(map[int64][]int64)
	var serverBaselineIdList []int64
	for _, v := range serverNodeRoleRelList {
		nodeRoleServerRelMap[v.NodeRoleId] = append(nodeRoleServerRelMap[v.NodeRoleId], v.ServerId)
		serverBaselineIdList = append(serverBaselineIdList, v.ServerId)
	}
	//查询服务器基线表
	var serverBaselineList []*entity.ServerBaseline
	if err := db.Where("id IN (?)", serverBaselineIdList).Find(&serverBaselineList).Error; err != nil {
		return nil, nil, nil, err
	}
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	for _, v := range serverBaselineList {
		serverBaselineMap[v.Id] = v
	}
	//查询服务器基线表
	var nodeRoleServerBaselineListMap = make(map[int64][]*entity.ServerPlanningBaseline)
	var screenNodeRoleServerBaselineListMap = make(map[int64][]*entity.ServerBaseline)
	for k, serverIdList := range nodeRoleServerRelMap {
		for _, serverId := range serverIdList {
			serverBaseline := serverBaselineMap[serverId]
			if serverBaseline != nil {
				nodeRoleServerBaselineListMap[k] = append(nodeRoleServerBaselineListMap[k], &entity.ServerPlanningBaseline{
					Id:                serverBaseline.Id,
					BomCode:           serverBaseline.BomCode,
					NetworkInterface:  serverBaseline.NetworkInterface,
					CpuType:           serverBaseline.CpuType,
					Arch:              serverBaseline.Arch,
					ConfigurationInfo: serverBaseline.ConfigurationInfo,
				})
				if (util.IsBlank(request.NetworkInterface) || serverBaseline.NetworkInterface == request.NetworkInterface) && (util.IsBlank(request.CpuType) || serverBaseline.CpuType == request.CpuType) {
					screenNodeRoleServerBaselineListMap[k] = append(screenNodeRoleServerBaselineListMap[k], serverBaseline)
				}
			}
		}
	}
	return serverBaselineMap, nodeRoleServerBaselineListMap, screenNodeRoleServerBaselineListMap, nil
}
