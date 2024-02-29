package capacity_planning

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
)

func ListServerCapacity(request *Request) ([]*ResponseCapClassification, error) {
	// 缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	// 查询云产品规划表
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
	// 查询云产品基线表
	var cloudProductBaselineList []*entity.CloudProductBaseline
	if err := db.Where("id IN (?)", cloudProductIdList).Find(&cloudProductBaselineList).Error; err != nil {
		return nil, err
	}
	var cloudProductCodeList []string
	var cloudProductIdCodeMap = make(map[int64]string)
	for _, v := range cloudProductBaselineList {
		cloudProductCodeList = append(cloudProductCodeList, v.ProductCode)
		cloudProductIdCodeMap[v.Id] = v.ProductCode
	}
	// 根据产品编码-售卖规格、产品编码-增值服务筛选容量输入列表
	var screenCloudProductSellSpecMap = make(map[string]interface{})
	var screenCloudProductValueAddedServiceMap = make(map[string]interface{})
	for _, v := range cloudProductPlanningList {
		// 根据产品编码-售卖规格
		if util.IsNotBlank(v.SellSpec) {
			screenCloudProductSellSpecMap[fmt.Sprintf("%s-%s", cloudProductIdCodeMap[v.ProductId], v.SellSpec)] = nil
		}
		// 产品编码-增值服务
		if util.IsNotBlank(v.ValueAddedService) {
			for _, valueAddedService := range strings.Split(v.ValueAddedService, ",") {
				screenCloudProductValueAddedServiceMap[fmt.Sprintf("%s-%s", cloudProductIdCodeMap[v.ProductId], valueAddedService)] = nil
			}
		}
	}
	// 查询容量换算表
	var capConvertBaselineList []*entity.CapConvertBaseline
	if err := db.Where("product_code IN (?) AND version_id = ?", cloudProductCodeList, cloudProductPlanningList[0].VersionId).Find(&capConvertBaselineList).Error; err != nil {
		return nil, err
	}
	// 查询是否已有保存容量规划
	var serverCapPlanningList []*entity.ServerCapPlanning
	if err := db.Where("plan_id = ?", request.PlanId).Find(&serverCapPlanningList).Error; err != nil {
		return nil, err
	}
	var serverCapPlanningMap = make(map[int64]*entity.ServerCapPlanning)
	// 单独处理ecs产品指标
	var ecsCapacity *EcsCapacity
	for _, v := range serverCapPlanningList {
		if v.Type == 2 && util.IsNotBlank(v.Special) {
			util.ToObject(v.Special, &ecsCapacity)
			continue
		}
		serverCapPlanningMap[v.CapacityBaselineId] = v
	}
	var capConvertBaselineMap = make(map[string][]*ResponseFeatures)
	var responseCapConvertList []*ResponseCapConvert
	for _, v := range capConvertBaselineList {
		// 根据产品编码-售卖规格、产品编码-增值服务筛选容量输入列表
		if util.IsNotBlank(v.SellSpecs) {
			if _, ok := screenCloudProductSellSpecMap[fmt.Sprintf("%s-%s", v.ProductCode, v.SellSpecs)]; !ok {
				continue
			}
		}
		if util.IsNotBlank(v.ValueAddedService) {
			if _, ok := screenCloudProductValueAddedServiceMap[fmt.Sprintf("%s-%s", v.ProductCode, v.ValueAddedService)]; !ok {
				continue
			}
		}
		key := fmt.Sprintf("%v-%v", v.ProductCode, v.CapPlanningInput)
		if _, ok := capConvertBaselineMap[key]; !ok {
			responseCapConvert := &ResponseCapConvert{
				VersionId:        v.VersionId,
				ProductName:      v.ProductName,
				ProductCode:      v.ProductCode,
				SellSpecs:        v.SellSpecs,
				CapPlanningInput: v.CapPlanningInput,
				Unit:             v.Unit,
				FeatureId:        v.Id,
				FeatureMode:      v.FeaturesMode,
				Description:      v.Description,
			}
			if serverCapPlanningMap[v.Id] != nil {
				responseCapConvert.Number = serverCapPlanningMap[v.Id].Number
			}
			responseCapConvertList = append(responseCapConvertList, responseCapConvert)
		}
		if util.IsNotBlank(v.Features) {
			capConvertBaselineMap[key] = append(capConvertBaselineMap[key], &ResponseFeatures{Id: v.Id, Name: v.Features})
		}
	}
	// 整理容量指标的特性
	var productCapConvertMap = make(map[string][]*ResponseCapConvert)
	var specialMap = make(map[string]*EcsCapacity)
	for i, v := range responseCapConvertList {
		key := fmt.Sprintf("%v-%v", v.ProductCode, v.CapPlanningInput)
		responseCapConvertList[i].Features = capConvertBaselineMap[key]
		// 回显容量规划数据
		for _, feature := range capConvertBaselineMap[key] {
			if serverCapPlanningMap[feature.Id] != nil {
				responseCapConvertList[i].FeatureId = feature.Id
				// 处理特性的输入值
				responseCapConvertList[i].Number = serverCapPlanningMap[feature.Id].Number
				responseCapConvertList[i].FeatureNumber = serverCapPlanningMap[feature.Id].FeatureNumber
			}
		}
		productCapConvertMap[v.ProductCode] = append(productCapConvertMap[v.ProductCode], v)
		// 单独处理ecs产品指标
		if v.ProductCode == constant.ProductCodeECS {
			specialMap[constant.ProductCodeECS] = ecsCapacity
		}
	}
	// 按产品分类
	var response []*ResponseCapClassification
	for k, v := range productCapConvertMap {
		classification := fmt.Sprintf("%v-%v", v[0].ProductName, v[0].SellSpecs)
		response = append(response, &ResponseCapClassification{
			Classification: classification,
			ProductName:    v[0].ProductName,
			ProductCode:    v[0].ProductCode,
			ProductType:    v[0].ProductCode,
			CapConvert:     v,
			Special:        specialMap[k],
		})
	}
	return response, nil
}

func SaveServerCapacity(request *Request) error {
	// 缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	// 保存服务器规划
	serverPlanningList, err := createServerPlanning(db, request)
	if err != nil {
		return err
	}
	serverPlanningMap, nodeRoleBaselineMap, serverBaselineMap, err := handleServerPlanning(db, serverPlanningList)
	if err != nil {
		return err
	}
	// 获取入参中所有容量输入指标id List
	var serverCapacityIdList []int64
	for _, v := range request.ServerCapacityList {
		serverCapacityIdList = append(serverCapacityIdList, v.Id)
	}
	// 单独处理ecs容量规划-按规格数量计算，将计算节点相关容量id写入List
	if request.EcsCapacity != nil {
		serverCapacityIdList = append(serverCapacityIdList, request.EcsCapacity.CapacityIdList...)
	}
	// 查询容量指标基线数据map
	capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, err := getCapBaseline(db, serverCapacityIdList)
	if err != nil {
		return err
	}
	var newServerPlanningList []*entity.ServerPlanning
	var serverCapPlanningList []*entity.ServerCapPlanning
	var productCodeNodeRoleIdMap = make(map[string][]int64)
	// 保存容量规划数据
	for _, v := range request.ServerCapacityList {
		// 查询容量换算表
		capConvertBaseline := capConvertBaselineMap[v.Id]
		// 构建服务器容量规划
		serverCapPlanning := &entity.ServerCapPlanning{
			PlanId:             request.PlanId,
			Type:               1,
			CapacityBaselineId: v.Id,
			Number:             v.Number,
			FeatureNumber:      v.FeatureNumber,
			VersionId:          capConvertBaseline.VersionId,
			ProductName:        capConvertBaseline.ProductName,
			ProductCode:        capConvertBaseline.ProductCode,
			SellSpecs:          capConvertBaseline.SellSpecs,
			CapPlanningInput:   capConvertBaseline.CapPlanningInput,
			Unit:               capConvertBaseline.Unit,
			FeaturesMode:       capConvertBaseline.FeaturesMode,
			Features:           capConvertBaseline.Features,
			ValueAddedService:  capConvertBaseline.ValueAddedService,
			Special:            "{}",
		}
		// 查询容量实际资源消耗表
		capConvertBaselineKey := fmt.Sprintf("%v-%v-%v-%v-%v", capConvertBaseline.ProductCode, capConvertBaseline.SellSpecs, capConvertBaseline.ValueAddedService, capConvertBaseline.CapPlanningInput, capConvertBaseline.Features)
		capActualResBaselineList := capActualResBaselineMap[capConvertBaselineKey]
		// 查不到容量规划基线表1和表2对不上的数据，例如CNBH
		if len(capActualResBaselineList) == 0 && capConvertBaseline.ProductCode == constant.ProductCodeCNBH {
			for key, capActualResBaselines := range capActualResBaselineMap {
				if strings.Contains(key, constant.ProductCodeCNBH) {
					capActualResBaselineList = capActualResBaselines
				}
			}
		}
		var nodeRoleBaseline *entity.NodeRoleBaseline
		for _, capActualResBaseline := range capActualResBaselineList {
			// 查询容量服务器数量计算
			capServerCalcBaseline := capServerCalcBaselineMap[capActualResBaseline.ExpendResCode]
			// 查询角色节点
			nodeRoleBaseline = nodeRoleBaselineMap[capServerCalcBaseline.ExpendNodeRoleCode]
			if nodeRoleBaseline != nil {
				productCodeNodeRoleIdMap[capActualResBaseline.ProductCode] = append(productCodeNodeRoleIdMap[capActualResBaseline.ProductCode], nodeRoleBaseline.Id)
			}
		}
		// 为了计算CKE的容器集群数的容量规划输入
		if capConvertBaseline.ProductCode == constant.ProductCodeCKE {
			nodeRoleBaseline = nodeRoleBaselineMap[constant.NodeRoleCodeCompute]
		}
		serverCapPlanningList = append(serverCapPlanningList, serverCapPlanning)
	}
	// for i, v := range serverCapPlanningList {
	// 	serverCapPlanningList[i].NodeRoleId = productCodeNodeRoleIdMap[v.ProductCode]
	// }
	// 计算个节点角色的服务器数量
	nodeRoleCapNumberMap, ecsServerPlanning, ecsServerCapPlanning := computing(db, request, capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, serverPlanningMap, nodeRoleBaselineMap, serverBaselineMap)
	if ecsServerPlanning != nil {
		newServerPlanningList = append(newServerPlanningList, ecsServerPlanning)
	}
	if ecsServerCapPlanning != nil {
		serverCapPlanningList = append(serverCapPlanningList, ecsServerCapPlanning)
	}
	// 将各角色的服务器数量写入服务器规划数据
	for k, v := range nodeRoleCapNumberMap {
		serverPlanningMap[k].Number = v
		newServerPlanningList = append(newServerPlanningList, serverPlanningMap[k])
	}
	if err = db.Transaction(func(tx *gorm.DB) error {
		// 保存服务器规划
		if err = tx.Save(&newServerPlanningList).Error; err != nil {
			return err
		}
		// 保存服务器容量规划
		if err = tx.Where("plan_id = ?", request.PlanId).Delete(&entity.ServerCapPlanning{}).Error; err != nil {
			return err
		}
		if err = tx.Create(&serverCapPlanningList).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func SingleComputing(request *RequestServerCapacityCount) (*ResponseCapCount, error) {
	// 缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	// 查询容量规划
	var serverCapPlanningList []*entity.ServerCapPlanning
	if err := db.Where("plan_id = ?", request.PlanId).Find(&serverCapPlanningList).Error; err != nil {
		return nil, err
	}
	var nodeRoleBaseline = &entity.NodeRoleBaseline{}
	if err := db.Where("id = ?", request.NodeRoleId).Find(&nodeRoleBaseline).Error; err != nil {
		return nil, err
	}
	var capServerCalcBaselines []*entity.CapServerCalcBaseline
	if err := db.Where("version_id = ? and expend_node_role_code = ?", nodeRoleBaseline.VersionId, nodeRoleBaseline.NodeRoleCode).Find(&capServerCalcBaselines).Error; err != nil {
		return nil, err
	}
	var expendResCodes []string
	var calcComputeNode bool
	if nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeCompute {
		calcComputeNode = true
	}
	for _, capServerCalcBaseline := range capServerCalcBaselines {
		expendResCodes = append(expendResCodes, capServerCalcBaseline.ExpendResCode)
	}
	var computingRequest = &Request{PlanId: request.PlanId}
	var capConvertBaselines []*entity.CapConvertBaseline
	if err := db.Table(entity.CapConvertBaselineTable+" ccb").Select("DISTINCT ccb.*").
		Joins("LEFT JOIN cap_actual_res_baseline carb ON ccb.product_code = carb.product_code and ccb.sell_specs = carb.sell_specs and ccb.value_added_service = carb.value_added_service and ccb.cap_planning_input = carb.sell_unit").
		Where("ccb.version_id= ? and carb.expend_res_code in (?) ", nodeRoleBaseline.VersionId, expendResCodes).
		Find(&capConvertBaselines).Error; err != nil {
		return nil, err
	}
	if calcComputeNode {
		// 添加容量规划基线的表1和表2对不上的数据(CNBH)和CKE的容器集群数
		var extraCapConvertBaselines []*entity.CapConvertBaseline
		var extraProductCodes = []string{constant.ProductCodeCKE, constant.ProductCodeCNBH}
		if err := db.Table(entity.CapConvertBaselineTable).Where("version_id = ? and product_code in (?)", nodeRoleBaseline.VersionId, extraProductCodes).
			Find(&extraCapConvertBaselines).Error; err != nil {
			return nil, err
		}
		for _, extraCapConvertBaseline := range extraCapConvertBaselines {
			if extraCapConvertBaseline.ProductCode == constant.ProductCodeCKE && (extraCapConvertBaseline.CapPlanningInput == constant.CapPlanningInputVCpu || extraCapConvertBaseline.CapPlanningInput == constant.CapPlanningInputMemory) {
				continue
			}
			capConvertBaselines = append(capConvertBaselines, extraCapConvertBaseline)
		}
	}
	capConvertBaselineIdMap := make(map[int64]*entity.CapConvertBaseline)
	for _, capConvertBaseline := range capConvertBaselines {
		capConvertBaselineIdMap[capConvertBaseline.Id] = capConvertBaseline
	}

	var capacityBaselineIdList []int64
	for _, v := range serverCapPlanningList {
		// 判断指标是否有ecs按规则计算，单独处理ecs容量规划-按规格数量计算
		if v.Type == 2 && util.IsNotBlank(v.Special) {
			util.ToObject(v.Special, &computingRequest.EcsCapacity)
			for _, capacityId := range computingRequest.EcsCapacity.CapacityIdList {
				if _, ok := capConvertBaselineIdMap[capacityId]; ok {
					capacityBaselineIdList = append(capacityBaselineIdList, capacityId)
				}
			}
			continue
		}
		if _, ok := capConvertBaselineIdMap[v.CapacityBaselineId]; !ok {
			continue
		}
		capacityBaselineIdList = append(capacityBaselineIdList, v.CapacityBaselineId)
		// 构建计算参数
		computingRequest.ServerCapacityList = append(computingRequest.ServerCapacityList, &RequestServerCapacity{Id: v.CapacityBaselineId, Number: v.Number, FeatureNumber: v.FeatureNumber})
	}
	if len(capacityBaselineIdList) == 0 {
		return nil, nil
	}
	// 查询容量指标基线
	capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, err := getCapBaseline(db, capacityBaselineIdList)
	if err != nil {
		return nil, err
	}
	var serverPlanning = &entity.ServerPlanning{}
	if err = db.Where("plan_id = ? AND node_role_id = ?", request.PlanId, request.NodeRoleId).Find(&serverPlanning).Error; err != nil {
		return nil, err
	}
	serverPlanning.ServerBaselineId = request.ServerBaselineId
	var serverBaseline = &entity.ServerBaseline{}
	if err = db.Where("id = ?", request.ServerBaselineId).Find(&serverBaseline).Error; err != nil {
		return nil, err
	}
	// 计算服务器数量
	var serverPlanningMap = map[int64]*entity.ServerPlanning{serverPlanning.NodeRoleId: serverPlanning}
	var nodeRoleBaselineMap = map[string]*entity.NodeRoleBaseline{nodeRoleBaseline.NodeRoleCode: nodeRoleBaseline}
	var serverBaselineMap = map[int64]*entity.ServerBaseline{serverBaseline.Id: serverBaseline}
	nodeRoleCapNumberMap, _, _ := computing(db, computingRequest, capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, serverPlanningMap, nodeRoleBaselineMap, serverBaselineMap)
	return &ResponseCapCount{Number: nodeRoleCapNumberMap[nodeRoleBaseline.Id]}, nil
}

func GetNodeRoleCapMap(db *gorm.DB, request *Request, serverPlanningMap map[int64]*entity.ServerPlanning, nodeRoleBaselineMap map[string]*entity.NodeRoleBaseline, serverBaselineMap map[int64]*entity.ServerBaseline) (map[int64]int, error) {
	var serverCapPlanningList []*entity.ServerCapPlanning
	if err := db.Where("plan_id = ?", request.PlanId).Find(&serverCapPlanningList).Error; err != nil {
		return nil, err
	}
	if serverCapPlanningList == nil || len(serverCapPlanningList) == 0 {
		return nil, nil
	}
	var nodeRoleCapNumberMap = make(map[int64]int)
	if len(serverCapPlanningList) == 0 {
		return nodeRoleCapNumberMap, nil
	}
	var productCodeList []string
	var capacityBaselineIdList []int64
	var computingRequest = &Request{PlanId: request.PlanId}
	for _, v := range serverCapPlanningList {
		productCodeList = append(productCodeList, v.ProductCode)
		capacityBaselineIdList = append(capacityBaselineIdList, v.CapacityBaselineId)
		// 判断指标是否有ecs按规则计算，单独处理ecs容量规划-按规格数量计算
		if v.Type == 2 && util.IsNotBlank(v.Special) {
			util.ToObject(v.Special, &computingRequest.EcsCapacity)
			continue
		}
		// 构建计算参数
		computingRequest.ServerCapacityList = append(computingRequest.ServerCapacityList, &RequestServerCapacity{Id: v.CapacityBaselineId, Number: v.Number, FeatureNumber: v.FeatureNumber})
	}
	// 查询容量指标基线数据map
	capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, err := getCapBaseline(db, capacityBaselineIdList)
	if err != nil {
		return nil, err
	}
	// 计算各角色节点服务器数量
	nodeRoleCapNumberMap, _, _ = computing(db, computingRequest, capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, serverPlanningMap, nodeRoleBaselineMap, serverBaselineMap)
	return nodeRoleCapNumberMap, nil
}

func computing(db *gorm.DB, request *Request, capConvertBaselineMap map[int64]*entity.CapConvertBaseline, capActualResBaselineMap map[string][]*entity.CapActualResBaseline, capServerCalcBaselineMap map[string]*entity.CapServerCalcBaseline,
	serverPlanningMap map[int64]*entity.ServerPlanning, nodeRoleBaselineMap map[string]*entity.NodeRoleBaseline, serverBaselineMap map[int64]*entity.ServerBaseline) (map[int64]int, *entity.ServerPlanning, *entity.ServerCapPlanning) {

	// 角色节点id为key，服务器数量为value
	var nodeRoleCapNumberMap = make(map[int64]int)
	// 部分产品特殊处理
	var serverCapacityMap = make(map[int64]float64)
	for _, v := range request.ServerCapacityList {
		serverCapacityMap[v.Id] = float64(v.Number)
	}
	specialCapActualResMap := SpecialCapacityComputing(serverCapacityMap, capConvertBaselineMap)
	// 计算各角色节点的总消耗map，消耗资源编码为key，总消耗为value
	var expendResCodeMap = make(map[string]float64)

	// 处理没有输入，按照固定数量计算的数据，DSP、CWP、DES，查询云产品规划表，看看是否包含了这3个云产品
	extraCalcProductCodes := []string{constant.ProductCodeDSP, constant.ProductCodeCWP, constant.ProductCodeDES}
	var extraProductCodes []string
	if err := db.Table(entity.CloudProductPlanningTable+" cpp").Select("DISTINCT cpb.product_code").
		Joins("LEFT JOIN cloud_product_baseline cpb on cpp.product_id = cpb.id").
		Where("cpb.product_code in (?) and cpp.plan_id = ?", extraCalcProductCodes, request.PlanId).
		Find(&extraProductCodes).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, nil, nil
	}
	if len(extraProductCodes) > 0 {
		var versionId int64
		for _, capConvertBaseline := range capConvertBaselineMap {
			versionId = capConvertBaseline.VersionId
			break
		}
		// 查询容量实际资源消耗表
		var capActualResBaselineList []*entity.CapActualResBaseline
		if err := db.Where("version_id = ? AND product_code IN (?)", versionId, extraProductCodes).Find(&capActualResBaselineList).Error; err != nil {
			return nil, nil, nil
		}
		extraCapActualResBaselineMap := make(map[string][]*entity.CapActualResBaseline)
		for _, capActualResBaseline := range capActualResBaselineList {
			extraCapActualResBaselineMap[capActualResBaseline.ProductCode] = append(extraCapActualResBaselineMap[capActualResBaseline.ProductCode], capActualResBaseline)
		}
		for _, extraProductCode := range extraProductCodes {
			capActualResBaselines := extraCapActualResBaselineMap[extraProductCode]
			for _, capActualResBaseline := range capActualResBaselines {
				// 总消耗
				capActualResNumber := capActualRes(1, 0, extraProductCode, capActualResBaseline)
				expendResCodeMap[capActualResBaseline.ExpendResCode] += capActualResNumber
			}
		}
	}

	// 产品编码为key，容量输入列表为value
	var ecsResourceProductMap = make(map[string][]*RequestServerCapacity)
	// ecs超分比
	var ecsOverallocation float64
	// 计算所有消耗资源的总量
	for _, v := range request.ServerCapacityList {
		// 查询容量换算表
		capConvertBaseline := capConvertBaselineMap[v.Id]
		// 查询容量实际资源消耗表
		capConvertBaselineKey := fmt.Sprintf("%v-%v-%v-%v-%v", capConvertBaseline.ProductCode, capConvertBaseline.SellSpecs, capConvertBaseline.ValueAddedService, capConvertBaseline.CapPlanningInput, capConvertBaseline.Features)
		capActualResBaselineList := capActualResBaselineMap[capConvertBaselineKey]
		if len(capActualResBaselineList) == 0 {
			if capConvertBaseline.ProductCode == constant.ProductCodeCKE {
				// 放入CKE容量规划表1有的但是表2没有的数据
				ecsResourceProductMap[capConvertBaseline.ProductCode] = append(ecsResourceProductMap[capConvertBaseline.ProductCode], v)
			}
			// 判断是否是CNBH产品
			if capConvertBaseline.ProductCode == constant.ProductCodeCNBH {
				switch capConvertBaseline.CapPlanningInput {
				case constant.CapPlanningInputOneHundred:
					v.Number = int(serverCapacityMap[capConvertBaseline.Id] * constant.CapPlanningInputOneHundredInt)
				case constant.CapPlanningInputFiveHundred:
					v.Number = int(serverCapacityMap[capConvertBaseline.Id] * constant.CapPlanningInputFiveHundredInt)
				case constant.CapPlanningInputOneThousand:
					v.Number = int(serverCapacityMap[capConvertBaseline.Id] * constant.CapPlanningInputOneThousandInt)
				}
				for key, capActualResBaselines := range capActualResBaselineMap {
					if strings.Contains(key, constant.ProductCodeCNBH) {
						capActualResBaselineList = capActualResBaselines
					}
				}
			}
		}
		for _, capActualResBaseline := range capActualResBaselineList {
			// 如果ecs容量规划-按规格数量计算，则将CKE、ECS_VCPU的容量输入信息放入，不判断ECS_MEM的目的是为了不造成数据重复
			if request.EcsCapacity != nil && (capConvertBaseline.ProductCode == constant.ProductCodeCKE || capActualResBaseline.ExpendResCode == constant.ExpendResCodeECSVCpu) {
				ecsResourceProductMap[capConvertBaseline.ProductCode] = append(ecsResourceProductMap[capConvertBaseline.ProductCode], v)
				continue
			}
			// 过滤特殊产品特殊计算
			if _, ok := SpecialProduct[capConvertBaseline.ProductCode]; ok {
				continue
			}
			// 总消耗
			capActualResNumber := capActualRes(v.Number, v.FeatureNumber, capConvertBaseline.ProductCode, capActualResBaseline)
			expendResCodeMap[capActualResBaseline.ExpendResCode] += capActualResNumber
		}
		if capConvertBaseline.ProductCode == constant.ProductCodeECS && capConvertBaseline.Features == constant.FeaturesNameOverallocation {
			ecsOverallocation = float64(v.FeatureNumber)
		}
	}
	// 计算每个角色节点的服务器数量
	for k, capActualResNumber := range expendResCodeMap {
		// ecs特殊处理容量超分特性，所有产品的vCpu总和/超分，加上CKE消耗
		if k == constant.ExpendResCodeECSVCpu {
			capActualResNumber = (capActualResNumber + specialCapActualResMap[k]) / ecsOverallocation
		} else {
			// 加上特殊产品计算的总消耗
			capActualResNumber += specialCapActualResMap[k]
		}
		nodeRoleCode := capServerCalcBaselineMap[k].ExpendNodeRoleCode
		nodeRoleBaseline := nodeRoleBaselineMap[nodeRoleCode]
		if nodeRoleBaseline == nil {
			continue
		}
		serverPlanning := serverPlanningMap[nodeRoleBaseline.Id]
		// 计算各角色节点单个服务器消耗
		capServerCalcNumber := capServerCalc(k, capServerCalcBaselineMap[k], serverBaselineMap[serverPlanning.ServerBaselineId])
		// 总消耗除以单个服务器消耗，等于服务器数量
		serverNumber := math.Ceil(capActualResNumber / capServerCalcNumber)
		if nodeRoleCapNumberMap[nodeRoleBaseline.Id] < int(serverNumber) {
			nodeRoleCapNumberMap[nodeRoleBaseline.Id] = int(serverNumber)
		}
	}
	var ecsServerPlanning *entity.ServerPlanning
	var ecsServerCapPlanning *entity.ServerCapPlanning
	// 单独处理ecs容量规划-按规格数量计算
	if request.EcsCapacity != nil {
		// 根据计算节点角色id查询服务器规划数据
		for _, v := range nodeRoleBaselineMap {
			if v.NodeRoleCode == constant.NodeRoleCodeCompute {
				ecsServerPlanning = serverPlanningMap[v.Id]
			}
		}
		// 查询计算节点的服务器基线配置
		serverBaseline := serverBaselineMap[ecsServerPlanning.ServerBaselineId]
		number := handleEcsData(request.EcsCapacity, serverBaseline, ecsResourceProductMap, capConvertBaselineMap, capServerCalcBaselineMap)
		ecsServerPlanning.Number = number
		ecsServerCapPlanning = &entity.ServerCapPlanning{
			PlanId:        request.PlanId,
			ProductCode:   constant.ProductCodeECS,
			Type:          2,
			FeatureNumber: request.EcsCapacity.FeatureNumber,
			VersionId:     serverBaselineMap[ecsServerPlanning.ServerBaselineId].VersionId,
			Special:       util.ToString(request.EcsCapacity)}
	}
	return nodeRoleCapNumberMap, ecsServerPlanning, ecsServerCapPlanning
}

func getCapBaseline(db *gorm.DB, serverCapacityIdList []int64) (map[int64]*entity.CapConvertBaseline, map[string][]*entity.CapActualResBaseline, map[string]*entity.CapServerCalcBaseline, error) {
	var capConvertBaselineList []*entity.CapConvertBaseline
	if err := db.Where("id IN (?)", serverCapacityIdList).Find(&capConvertBaselineList).Error; err != nil {
		return nil, nil, nil, err
	}
	if len(capConvertBaselineList) == 0 {
		return nil, nil, nil, errors.New("服务器容量规划指标不存在")
	}
	// 查询容量输入表，id为key
	var capConvertBaselineMap = make(map[int64]*entity.CapConvertBaseline)
	var productCoedList []string
	for _, v := range capConvertBaselineList {
		capConvertBaselineMap[v.Id] = v
		productCoedList = append(productCoedList, v.ProductCode)
	}
	// 查询容量实际资源消耗表
	var capActualResBaselineList []*entity.CapActualResBaseline
	if err := db.Where("version_id = ? AND product_code IN (?)", capConvertBaselineList[0].VersionId, productCoedList).Find(&capActualResBaselineList).Error; err != nil {
		return nil, nil, nil, err
	}
	// 产品编码-售卖规格-消耗资源-特性为key
	var capActualResBaselineListMap = make(map[string][]*entity.CapActualResBaseline)
	var expendResCodeList []string
	for _, v := range capActualResBaselineList {
		key := fmt.Sprintf("%v-%v-%v-%v-%v", v.ProductCode, v.SellSpecs, v.ValueAddedService, v.SellUnit, v.Features)
		capActualResBaselineListMap[key] = append(capActualResBaselineListMap[key], v)
		expendResCodeList = append(expendResCodeList, v.ExpendResCode)
	}
	// 查询容量服务器数量计算
	var capServerCalcBaselineList []*entity.CapServerCalcBaseline
	if err := db.Where("version_id = ? AND expend_res_code IN (?)", capConvertBaselineList[0].VersionId, expendResCodeList).Find(&capServerCalcBaselineList).Error; err != nil {
		return nil, nil, nil, err
	}
	// 消耗资源编码为key
	var capServerCalcBaselineMap = make(map[string]*entity.CapServerCalcBaseline)
	for _, v := range capServerCalcBaselineList {
		capServerCalcBaselineMap[v.ExpendResCode] = v
	}
	return capConvertBaselineMap, capActualResBaselineListMap, capServerCalcBaselineMap, nil
}

func handleEcsData(ecsCapacity *EcsCapacity, serverBaseline *entity.ServerBaseline, ecsResourceProductMap map[string][]*RequestServerCapacity, capConvertBaselineMap map[int64]*entity.CapConvertBaseline, capServerCalcBaselineMap map[string]*entity.CapServerCalcBaseline) int {
	var items []util.Item
	// 计算ecs小箱子
	for _, v := range ecsCapacity.List {
		width := float64(v.CpuNumber)
		// 每个实例额外消耗内存（单位：M）=138M(libvirt)+8M(IO)+16M(GPU)+8M*vCPU+内存(M)/512，ARM额外128M
		extraMemorySpent := float64(138 + 8 + 16 + 8*v.CpuNumber + v.MemoryNumber*1024/512)
		if serverBaseline.Arch == constant.CpuArchARM {
			extraMemorySpent += 128
		}
		height := float64(v.MemoryNumber) + extraMemorySpent/1024
		items = append(items, util.Item{Size: util.Rectangle{Width: width, Height: height}, Number: v.Count})
	}
	// 计算其它计算节点相关的产品小箱子
	for k, v := range ecsResourceProductMap {
		switch k {
		case constant.ProductCodeCKE:
			var vCpu, cluster float64
			for _, requestCapacity := range v {
				switch capConvertBaselineMap[requestCapacity.Id].CapPlanningInput {
				case constant.CapPlanningInputVCpu:
					vCpu = float64(requestCapacity.Number)
				case constant.CapPlanningInputContainerCluster:
					// TODO 没有放入cluster数量
					cluster = float64(requestCapacity.Number)
				}
			}
			waterLevel, _ := strconv.ParseFloat(capServerCalcBaselineMap[constant.ExpendResCodeECSVCpu].WaterLevel, 64)
			items = append(items, util.Item{Size: util.Rectangle{Width: 16, Height: 32}, Number: int(math.Ceil(vCpu/waterLevel/14.6 + cluster*3))})
		case constant.ProductCodeCNBH:
			// 前面已经算好了CNBH每个资产类型的数量，这里不需要计算了
			var assetsNumber float64
			for _, requestCapacity := range v {
				assetsNumber += float64(requestCapacity.Number)
			}
			items = append(items, util.Item{Size: util.Rectangle{Width: 2, Height: 4}, Number: int(math.Ceil(assetsNumber / 50))})
		case constant.ProductCodeCNFW:
			for _, requestCapacity := range v {
				if capConvertBaselineMap[requestCapacity.Id].SellSpecs == constant.SellSpecsStandardEdition {
					items = append(items, util.Item{Size: util.Rectangle{Width: 4, Height: 8}, Number: requestCapacity.Number})
				}
				if capConvertBaselineMap[requestCapacity.Id].SellSpecs == constant.SellSpecsUltimateEdition {
					items = append(items, util.Item{Size: util.Rectangle{Width: 8, Height: 16}, Number: requestCapacity.Number})
				}
			}
		case constant.ProductCodeCWAF:
			var firewallNumber float64
			for _, requestCapacity := range v {
				firewallNumber += float64(requestCapacity.Number)
			}
			items = append(items, util.Item{Size: util.Rectangle{Width: 1, Height: 2}, Number: int(firewallNumber)})
		}
	}
	// 节点固定开销5C8G，则单节点可用vCPU=(节点总vCPU*-5）*70%*超分系数N；单节点可用内存=(节点总内存*-8）*70%，为大箱子的长宽
	var boxSize = util.Rectangle{Width: float64((serverBaseline.Cpu-5)*ecsCapacity.FeatureNumber) * 0.7, Height: float64(serverBaseline.Memory-8) * 0.7}
	boxes := util.Pack(items, boxSize)
	return len(boxes)
}

func createServerPlanning(db *gorm.DB, request *Request) ([]*entity.ServerPlanning, error) {
	if err := db.Where("plan_id = ?", request.PlanId).Delete(&entity.ServerPlanning{}).Error; err != nil {
		return nil, err
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
		})
	}
	if err := db.Create(&serverPlanningEntityList).Error; err != nil {
		return nil, err
	}
	return serverPlanningEntityList, nil
}

func handleServerPlanning(db *gorm.DB, serverPlanningList []*entity.ServerPlanning) (map[int64]*entity.ServerPlanning, map[string]*entity.NodeRoleBaseline, map[int64]*entity.ServerBaseline, error) {
	// 将服务器规划数据封装成map，角色节点id为key
	var serverPlanningMap = make(map[int64]*entity.ServerPlanning)
	var nodeRoleIdIdList []int64
	var serverBaselineIdList []int64
	for _, v := range serverPlanningList {
		serverPlanningMap[v.NodeRoleId] = v
		nodeRoleIdIdList = append(nodeRoleIdIdList, v.NodeRoleId)
		serverBaselineIdList = append(serverBaselineIdList, v.ServerBaselineId)
	}
	// 查询节点角色，为了计算容量计算表3的数据
	var nodeRoleBaselineList []*entity.NodeRoleBaseline
	if err := db.Where("id IN (?)", nodeRoleIdIdList).Find(&nodeRoleBaselineList).Error; err != nil {
		return nil, nil, nil, err
	}
	var nodeRoleBaselineMap = make(map[string]*entity.NodeRoleBaseline)
	for _, v := range nodeRoleBaselineList {
		nodeRoleBaselineMap[v.NodeRoleCode] = v
	}
	// 查询服务器基线数据，为了计算容量计算表3的数据
	var serverBaselineList []*entity.ServerBaseline
	if err := db.Where("id IN (?)", serverBaselineIdList).Find(&serverBaselineList).Error; err != nil {
		return nil, nil, nil, err
	}
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	for _, v := range serverBaselineList {
		serverBaselineMap[v.Id] = v
	}
	return serverPlanningMap, nodeRoleBaselineMap, serverBaselineMap, nil
}

var SpecialProduct = map[string]interface{}{constant.ProductCodeCKE: nil}

func SpecialCapacityComputing(serverCapacityMap map[int64]float64, capConvertBaselineMap map[int64]*entity.CapConvertBaseline) map[string]float64 {
	// 按产品将容量输入参数分类
	var productCapMap = make(map[string][]*entity.CapConvertBaseline)
	for _, v := range capConvertBaselineMap {
		productCapMap[v.ProductCode] = append(productCapMap[v.ProductCode], v)
	}
	var capActualResMap = make(map[string]float64)
	for k, v := range productCapMap {
		if k == constant.ProductCodeCKE {
			var vCpu, memory, cluster float64
			for _, capConvertBaseline := range v {
				switch capConvertBaseline.CapPlanningInput {
				case constant.CapPlanningInputVCpu:
					vCpu = serverCapacityMap[capConvertBaseline.Id]
				case constant.CapPlanningInputMemory:
					memory = serverCapacityMap[capConvertBaseline.Id]
				case constant.CapPlanningInputContainerCluster:
					cluster = serverCapacityMap[capConvertBaseline.Id]
				}
			}
			// TODO 0.7的水位从哪个配置来？还是自己写死的？
			cpuCapActualRes := 48*cluster + 16*vCpu/0.7/14.6
			memoryCapActualRes := 96*cluster + 32*memory/0.7/29.4
			capActualResMap[constant.ExpendResCodeECSVCpu] = cpuCapActualRes
			capActualResMap[constant.ExpendResCodeECSMemory] = memoryCapActualRes
		}
	}
	return capActualResMap
}

func capActualRes(number, featureNumber int, productCode string, capActualResBaseline *entity.CapActualResBaseline) float64 {
	if productCode == constant.ProductCodeECS {
		return float64(number)
	}
	if featureNumber <= 0 {
		featureNumber = 1
	}
	numerator, _ := strconv.ParseFloat(capActualResBaseline.OccRatioNumerator, 64)
	if numerator == 0 {
		numerator = float64(featureNumber)
	}
	denominator, _ := strconv.ParseFloat(capActualResBaseline.OccRatioDenominator, 64)
	if denominator == 0 {
		denominator = float64(featureNumber)
	}
	// 总消耗
	return float64(number) / numerator * denominator
}

func capServerCalc(expendResCode string, capServerCalcBaseline *entity.CapServerCalcBaseline, serverBaseline *entity.ServerBaseline) float64 {
	// 判断用哪个容量参数
	var singleCapacity int
	if strings.Contains(expendResCode, "_VCPU") {
		singleCapacity = serverBaseline.Cpu
	}
	if strings.Contains(expendResCode, "_MEM") {
		singleCapacity = serverBaseline.Memory
	}
	if strings.Contains(expendResCode, "_DISK") {
		singleCapacity = serverBaseline.StorageDiskNum * serverBaseline.StorageDiskCapacity
	}

	nodeWastage, _ := strconv.ParseFloat(capServerCalcBaseline.NodeWastage, 64)
	waterLevel, _ := strconv.ParseFloat(capServerCalcBaseline.WaterLevel, 64)
	// 单个服务器消耗
	if capServerCalcBaseline.NodeWastageCalcType == 1 {
		return (float64(singleCapacity) - nodeWastage) * waterLevel
	} else {
		return (float64(singleCapacity) * (1 - nodeWastage)) * waterLevel
	}
}
