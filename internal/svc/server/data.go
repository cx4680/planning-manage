package server

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math"
	"strconv"
	"time"
)

func ListServer(request *Request) ([]*Server, error) {
	//缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	//查询云产品规划表
	var productIdList []int64
	if err := db.Model(&entity.CloudProductPlanning{}).Select("product_id").Where("plan_id = ?", request.PlanId).Find(&productIdList).Error; err != nil {
		return nil, err
	}
	if len(productIdList) == 0 {
		return nil, errors.New("该方案未找到关联产品")
	}
	//查询云产品和角色关联表
	var nodeRoleIdList []int64
	if err := db.Model(&entity.CloudProductNodeRoleRel{}).Select("node_role_id").Where("product_id IN (?)", productIdList).Find(&nodeRoleIdList).Error; err != nil {
		return nil, err
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
	//查询混合部署方式map
	mixedNodeRoleMap, err := getMixedNodeRoleMap(db, nodeRoleIdList)
	if err != nil {
		return nil, err
	}
	//查询已保存的服务器规划表
	var serverPlanningList []*Server
	if err = db.Where("plan_id = ?", request.PlanId).Find(&serverPlanningList).Error; err != nil {
		return nil, err
	}
	var nodeRoleServerPlanningMap = make(map[int64]*Server)
	for _, v := range serverPlanningList {
		nodeRoleServerPlanningMap[v.NodeRoleId] = v
	}
	// 计算已保存的容量规划指标
	nodeRoleCapMap, err := getNodeRoleCapMap(db, request, nodeRoleServerBaselineListMap)
	if err != nil {
		return nil, err
	}
	//构建返回体
	var list []*Server
	for _, v := range nodeRoleBaselineList {
		serverPlanning := &Server{}
		//若服务器规划有保存过，则加载已保存的数据
		if nodeRoleServerPlanningMap[v.Id] != nil && util.IsBlank(request.NetworkInterface) && util.IsBlank(request.CpuType) {
			serverPlanning = nodeRoleServerPlanningMap[v.Id]
			serverPlanning.ServerBomCode = serverBaselineMap[serverPlanning.ServerBaselineId].BomCode
			serverPlanning.ServerArch = serverBaselineMap[serverPlanning.ServerBaselineId].Arch
		} else {
			serverPlanning.PlanId = request.PlanId
			serverPlanning.NodeRoleId = v.Id
			serverPlanning.Number = v.MinimumNum
			if nodeRoleCapMap[v.Id] != 0 {
				serverPlanning.Number = nodeRoleCapMap[v.Id]
			}
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
				NetworkInterface: request.NetworkInterface,
				CpuType:          request.CpuType,
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
			BusinessPlanStage: constant.BusinessPlanningNetworkDevice,
			UpdateUserId:      request.UserId,
			UpdateTime:        time.Now(),
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
	//缓存预编译 会话模式
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
	//缓存预编译 会话模式
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
	//查询云产品规划表
	var cloudProductPlanning = &entity.CloudProductPlanning{}
	if err := data.DB.Where("plan_id = ?", request.PlanId).Find(&cloudProductPlanning).Error; err != nil {
		return nil, nil, err
	}
	if cloudProductPlanning.Id == 0 {
		return nil, nil, errors.New("云产品规划不存在")
	}
	//查询云产品基线表
	var cloudProductBaseline = &entity.CloudProductBaseline{}
	if err := data.DB.Where("id = ?", cloudProductPlanning.ProductId).Find(&cloudProductBaseline).Error; err != nil {
		return nil, nil, err
	}
	if cloudProductBaseline.Id == 0 {
		return nil, nil, errors.New("云产品基线不存在")
	}
	//查询服务器基线表
	var serverBaselineList []*entity.ServerBaseline
	if err := data.DB.Where("version_id = ?", cloudProductBaseline.VersionId).Find(&serverBaselineList).Error; err != nil {
		return nil, nil, err
	}
	//查询是否有已保存的方案
	var serverPlanning = &entity.ServerPlanning{}
	if err := db.Where("plan_id = ?", request.PlanId).Find(&serverPlanning).Error; err != nil {
		return nil, nil, err
	}
	return serverBaselineList, serverPlanning, nil
}

func ListServerCapacity(request *Request) ([]*ResponseCapClassification, error) {
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
	var cloudProductCodeList []string
	if err := db.Model(&entity.CloudProductBaseline{}).Select("product_code").Where("id IN (?)", cloudProductIdList).Find(&cloudProductCodeList).Error; err != nil {
		return nil, err
	}
	//查询容量换算表
	var capConvertBaselineList []*entity.CapConvertBaseline
	if err := db.Where("product_code IN (?) AND version_id = ?", cloudProductCodeList, cloudProductPlanningList[0].VersionId).Find(&capConvertBaselineList).Error; err != nil {
		return nil, err
	}
	//查询是否已有保存容量规划
	var serverCapPlanningList []*entity.ServerCapPlanning
	if err := db.Where("plan_id = ?", request.PlanId).Find(&serverCapPlanningList).Error; err != nil {
		return nil, err
	}
	var serverCapPlanningMap = make(map[int64]*entity.ServerCapPlanning)
	for _, v := range serverCapPlanningList {
		serverCapPlanningMap[v.CapacityBaselineId] = v
	}
	var capConvertBaselineMap = make(map[string][]*ResponseFeatures)
	var responseCapConvertList []*ResponseCapConvert
	for _, v := range capConvertBaselineList {
		key := v.ProductCode + v.SellSpecs + v.CapPlanningInput
		if _, ok := capConvertBaselineMap[key]; !ok {
			responseCapConvert := &ResponseCapConvert{
				VersionId:        v.VersionId,
				ProductName:      v.ProductName,
				ProductCode:      v.ProductCode,
				SellSpecs:        v.SellSpecs,
				CapPlanningInput: v.CapPlanningInput,
				Unit:             v.Unit,
				FeatureType:      FeatureMap[v.Features],
				Description:      v.Description,
			}
			responseCapConvert.FeatureId = v.Id
			responseCapConvertList = append(responseCapConvertList, responseCapConvert)
		}
		responseFeatures := &ResponseFeatures{
			Id:   v.Id,
			Name: v.Features,
		}
		capConvertBaselineMap[key] = append(capConvertBaselineMap[key], responseFeatures)
	}
	//整理容量指标的特性
	var classificationMap = make(map[string][]*ResponseCapConvert)
	for i, v := range responseCapConvertList {
		key := v.ProductCode + v.SellSpecs + v.CapPlanningInput
		responseCapConvertList[i].Features = capConvertBaselineMap[key]
		//回显容量规划数据
		for _, feature := range capConvertBaselineMap[key] {
			if serverCapPlanningMap[feature.Id] != nil {
				responseCapConvertList[i].FeatureId = feature.Id
				responseCapConvertList[i].Number = serverCapPlanningMap[feature.Id].Number
				responseCapConvertList[i].FeatureNumber = serverCapPlanningMap[feature.Id].FeatureNumber
			}
		}

		classification := fmt.Sprintf("%s-%s", v.ProductName, v.SellSpecs)
		classificationMap[classification] = append(classificationMap[classification], v)
	}
	//按产品分类
	var response []*ResponseCapClassification
	for k, v := range classificationMap {
		response = append(response, &ResponseCapClassification{
			Classification: k,
			CapConvert:     v,
		})
	}
	return response, nil
}

func SaveServerCapacity(request *Request) error {
	//缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	var serverCapPlanningList []*entity.ServerCapPlanning
	var nodeRoleBaselineMap = make(map[int64]float64)
	//查询服务器容量数据map
	var serverCapacityIdList []int64
	for _, v := range request.ServerCapacityList {
		serverCapacityIdList = append(serverCapacityIdList, v.Id)
	}
	var capConvertBaselineList []*entity.CapConvertBaseline
	if err := db.Where("id IN ?", serverCapacityIdList).Find(&capConvertBaselineList).Error; err != nil {
		return err
	}
	if len(capConvertBaselineList) == 0 {
		return errors.New("服务器容量规划指标不存在")
	}
	var capConvertBaselineMap = make(map[int64]*entity.CapConvertBaseline)
	for _, v := range capConvertBaselineList {
		capConvertBaselineMap[v.Id] = v
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, v := range request.ServerCapacityList {
			capConvertBaseline := capConvertBaselineMap[v.Id]
			//查询容量实际资源消耗表
			var capActualResBaseline = &entity.CapActualResBaseline{}
			if err := tx.Where("version_id = ? AND product_code = ? AND sell_specs = ? AND sell_unit = ? AND features = ?",
				capConvertBaseline.VersionId, capConvertBaseline.ProductCode, capConvertBaseline.SellSpecs, capConvertBaseline.CapPlanningInput, capConvertBaseline.Features).
				Find(&capActualResBaseline).Error; err != nil {
				return err
			}
			if capActualResBaseline.Id == 0 {
				return errors.New("服务器容量规划特性不存在")
			}
			//查询容量服务器数量计算
			var capServerCalcBaseline = &entity.CapServerCalcBaseline{}
			if err := tx.Where("version_id = ? AND expend_res_code = ?", capConvertBaseline.VersionId, capActualResBaseline.ExpendResCode).Find(&capServerCalcBaseline).Error; err != nil {
				return err
			}
			if capServerCalcBaseline.Id == 0 {
				return errors.New("服务器数量计算数据不存在")
			}
			//查询服务器规划是否有已保存数据
			var nodeRoleBaseline = &entity.NodeRoleBaseline{}
			if err := tx.Where("version_id = ? AND node_role_code = ?", capConvertBaseline.VersionId, capServerCalcBaseline.ExpendNodeRoleCode).Find(&nodeRoleBaseline).Error; err != nil {
				return err
			}
			if nodeRoleBaseline.Id == 0 {
				return errors.New("节点角色数据不存在")
			}
			var serverPlanning = &entity.ServerPlanning{}
			if err := tx.Where("plan_id = ? AND node_role_id = ?", request.PlanId, nodeRoleBaseline.Id).Find(&serverPlanning).Error; err != nil {
				return err
			}
			//查询服务器基线数据
			var serverBaseline = &entity.ServerBaseline{}
			if serverPlanning.Id != 0 {
				if err := tx.Where("id = ?", serverPlanning.ServerBaselineId).Find(&serverBaseline).Error; err != nil {
					return err
				}
			} else {
				var serverNodeRoleRelList []*entity.ServerNodeRoleRel
				if err := tx.Where("node_role_id = ?", nodeRoleBaseline.Id).Find(&serverNodeRoleRelList).Error; err != nil {
					return err
				}
				if len(serverNodeRoleRelList) == 0 {
					return errors.New("服务器和节点角色关联数据不存在")
				}
				var serverBaselineIdList []int64
				for _, serverNodeRoleRel := range serverNodeRoleRelList {
					serverBaselineIdList = append(serverBaselineIdList, serverNodeRoleRel.ServerId)
				}
				if err := tx.Where("id IN (?) AND network_interface = ? AND cpu_type = ?", serverBaselineIdList, request.NetworkInterface, request.CpuType).Find(&serverBaseline).Error; err != nil {
					return err
				}
			}
			if serverBaseline.Id == 0 {
				return errors.New("服务器基线数据不存在")
			}
			//计算服务器规划数据
			numerator, _ := strconv.ParseFloat(capActualResBaseline.OccRatioNumerator, 64)
			if numerator == 0 {
				numerator = float64(v.FeatureNumber)
			}
			denominator, _ := strconv.ParseFloat(capActualResBaseline.OccRatioDenominator, 64)
			if numerator == 0 || denominator == 0 {
				numerator = 1
				denominator = 1
			}
			capacityNumber := float64(v.Number) / numerator * denominator

			nodeWastage, _ := strconv.ParseFloat(capServerCalcBaseline.NodeWastage, 64)
			waterLevel, _ := strconv.ParseFloat(capServerCalcBaseline.WaterLevel, 64)
			var consumeNumber float64
			if capServerCalcBaseline.NodeWastageCalcType == 1 {
				consumeNumber = (float64(serverBaseline.Cpu) - nodeWastage) * waterLevel
			} else {
				consumeNumber = (float64(serverBaseline.Cpu) * (1 - nodeWastage)) * waterLevel
			}
			serverNumber := math.Ceil(capacityNumber / consumeNumber)
			//保存服务器规划数据
			if nodeRoleBaselineMap[nodeRoleBaseline.Id] < serverNumber {
				nodeRoleBaselineMap[nodeRoleBaseline.Id] = serverNumber
			}
			if serverPlanning.Id == 0 {
				now := datetime.GetNow()
				serverPlanning = &entity.ServerPlanning{
					PlanId:           request.PlanId,
					NodeRoleId:       nodeRoleBaseline.Id,
					MixedNodeRoleId:  nodeRoleBaseline.Id,
					ServerBaselineId: serverBaseline.Id,
					NetworkInterface: request.NetworkInterface,
					CpuType:          request.CpuType,
					CreateUserId:     request.UserId,
					CreateTime:       now,
					UpdateUserId:     request.UserId,
					UpdateTime:       now,
					DeleteState:      0,
				}
			}
			serverPlanning.Number = int(nodeRoleBaselineMap[nodeRoleBaseline.Id])
			if err := tx.Save(&serverPlanning).Error; err != nil {
				return err
			}
			//构建服务器容量规划
			serverCapPlanningList = append(serverCapPlanningList, &entity.ServerCapPlanning{
				PlanId:             request.PlanId,
				NodeRoleId:         nodeRoleBaseline.Id,
				CapacityBaselineId: v.Id,
				Number:             v.Number,
				FeatureNumber:      v.FeatureNumber,
				CapacityNumber:     strconv.FormatFloat(capacityNumber, 'f', 3, 64),
				ExpendResCode:      capActualResBaseline.ExpendResCode,
			})
		}
		//保存服务器容量规划
		if err := tx.Where("plan_id = ?", request.PlanId).Delete(&entity.ServerCapPlanning{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&serverCapPlanningList).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func DownloadServer(planId int64) ([]ResponseDownloadServer, string, error) {
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
	//构建文件名称
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

func getNodeRoleServerBaselineMap(db *gorm.DB, nodeRoleIdList []int64, request *Request) (map[int64]*entity.ServerBaseline, map[int64][]*Baseline, map[int64][]*entity.ServerBaseline, error) {
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
	var nodeRoleServerBaselineListMap = make(map[int64][]*Baseline)
	var screenNodeRoleServerBaselineListMap = make(map[int64][]*entity.ServerBaseline)
	for k, serverIdList := range nodeRoleServerRelMap {
		for _, serverId := range serverIdList {
			serverBaseline := serverBaselineMap[serverId]
			if serverBaseline != nil {
				nodeRoleServerBaselineListMap[k] = append(nodeRoleServerBaselineListMap[k], &Baseline{
					Id:                serverBaseline.Id,
					BomCode:           serverBaseline.BomCode,
					NetworkInterface:  serverBaseline.NetworkInterface,
					Cpu:               serverBaseline.Cpu,
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

func getNodeRoleCapMap(db *gorm.DB, request *Request, nodeRoleServerBaselineListMap map[int64][]*Baseline) (map[int64]int, error) {
	var serverCapPlanningList []*entity.ServerCapPlanning
	if err := db.Where("plan_id = ?", request.PlanId).Find(&serverCapPlanningList).Error; err != nil {
		return nil, err
	}
	var nodeRoleCapMap = make(map[int64]int)
	if len(serverCapPlanningList) != 0 {
		var capPlanningNodeRoleIdList []int64
		var expendResCodeList []string
		for _, v := range serverCapPlanningList {
			capPlanningNodeRoleIdList = append(capPlanningNodeRoleIdList, v.NodeRoleId)
			expendResCodeList = append(expendResCodeList, v.ExpendResCode)
		}
		//查询服务器和角色关联表
		var capPlanningNodeRoleRelList []*entity.ServerNodeRoleRel
		if err := db.Where("node_role_id IN (?)", capPlanningNodeRoleIdList).Find(&capPlanningNodeRoleRelList).Error; err != nil {
			return nil, err
		}
		//查询容量计算表
		var capServerCalcBaselineList []*entity.CapServerCalcBaseline
		if err := db.Where("expend_res_code IN (?)", expendResCodeList).Find(&capServerCalcBaselineList).Error; err != nil {
			return nil, err
		}
		var capServerCalcBaselineMap = make(map[string]*entity.CapServerCalcBaseline)
		for _, v := range capServerCalcBaselineList {
			capServerCalcBaselineMap[v.ExpendResCode] = v
		}
		for _, v := range serverCapPlanningList {
			for _, serverBaseline := range nodeRoleServerBaselineListMap[v.NodeRoleId] {
				if serverBaseline.CpuType == request.CpuType {
					capServerCalcBaseline := capServerCalcBaselineMap[v.ExpendResCode]
					capacityNumber, _ := strconv.ParseFloat(v.CapacityNumber, 64)
					nodeWastage, _ := strconv.ParseFloat(capServerCalcBaseline.NodeWastage, 64)
					waterLevel, _ := strconv.ParseFloat(capServerCalcBaseline.WaterLevel, 64)
					var consumeNumber float64
					if capServerCalcBaseline.NodeWastageCalcType == 1 {
						consumeNumber = (float64(serverBaseline.Cpu) - nodeWastage) * waterLevel
					} else {
						consumeNumber = (float64(serverBaseline.Cpu) * (1 - nodeWastage)) * waterLevel
					}
					serverNumber := math.Ceil(capacityNumber / consumeNumber)
					nodeRoleCapMap[v.NodeRoleId] = int(serverNumber)
				}
			}
		}
	}
	return nodeRoleCapMap, nil
}

func getServerShelveList(planId int64) ([]*ResponseServerShelve, error) {
	//查询服务器规划表
	var serverPlanning []*entity.ServerPlanning
	if err := data.DB.Where("plan_id = ?", planId).Find(&serverPlanning).Error; err != nil {
		return nil, err
	}
	var NodeRoleIdList []int64
	var ServerBaselineIdList []int64
	for _, v := range serverPlanning {
		NodeRoleIdList = append(NodeRoleIdList, v.NodeRoleId)
		ServerBaselineIdList = append(ServerBaselineIdList, v.ServerBaselineId)
	}
	//查询节点角色表
	var nodeRoleList []*entity.NodeRoleBaseline
	if err := data.DB.Where("id IN (?)", NodeRoleIdList).Find(&nodeRoleList).Error; err != nil {
		return nil, err
	}
	var nodeRoleNameMap = make(map[int64]string)
	for _, v := range nodeRoleList {
		nodeRoleNameMap[v.Id] = v.NodeRoleName
	}
	//查询服务器基线表
	var serverBaseline []*entity.ServerBaseline
	if err := data.DB.Where("id IN (?)", ServerBaselineIdList).Find(&serverBaseline).Error; err != nil {
		return nil, err
	}
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	for _, v := range serverBaseline {
		serverBaselineMap[v.Id] = v
	}
	var savedServerShelve []*entity.ServerShelve
	if err := data.DB.Where("plan_id = ?", planId).Find(&savedServerShelve).Error; err != nil {
		return nil, err
	}
	var savedServerShelveMap = make(map[int64]*entity.ServerShelve)
	for _, v := range savedServerShelve {
		savedServerShelveMap[v.NodeRoleId] = v
	}
	var serverShelve []*ResponseServerShelve
	for _, v := range serverPlanning {
		var responseServerShelve = &ResponseServerShelve{}
		if savedServerShelveMap[v.NodeRoleId] != nil {
			responseServerShelve.ServerShelve = savedServerShelveMap[v.NodeRoleId]
		} else {
			responseServerShelve.ServerShelve = &entity.ServerShelve{
				PlanId:                planId,
				NodeRoleId:            v.NodeRoleId,
				Model:                 serverBaselineMap[v.ServerBaselineId].BomCode,
				MachineRoomAbbr:       "",
				MachineRoomNumber:     "",
				ColumnNumber:          "",
				CabinetAsw:            "",
				CabinetNumber:         "",
				CabinetOriginalNumber: "",
				CabinetLocation:       "",
				SlotPosition:          "",
				NetworkInterface:      "",
				BmcUserName:           "",
				BmcPassword:           "",
				BmcIp:                 "",
				BmcMac:                "",
				Mask:                  "",
				Gateway:               "",
			}
		}
		responseServerShelve.NodeRoleName = nodeRoleNameMap[v.NodeRoleId]
		serverShelve = append(serverShelve, responseServerShelve)
	}
	return serverShelve, nil
}
