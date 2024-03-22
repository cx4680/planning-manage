package capacity_planning

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/opentrx/seata-golang/v2/pkg/util/log"
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
	dpdkCloudProductIdMap := make(map[int64]*entity.CloudProductPlanning)
	for _, cloudProductPlanning := range cloudProductPlanningList {
		cloudProductIdList = append(cloudProductIdList, cloudProductPlanning.ProductId)
		if strings.Contains(cloudProductPlanning.SellSpec, constant.SellSpecDPDK) {
			dpdkCloudProductIdMap[cloudProductPlanning.ProductId] = cloudProductPlanning
		}
	}
	// 查询云产品基线表
	var cloudProductBaselineList []*entity.CloudProductBaseline
	if err := db.Where("id IN (?)", cloudProductIdList).Find(&cloudProductBaselineList).Error; err != nil {
		return nil, err
	}
	var cloudProductCodeList []string
	var cloudProductIdBaselineMap = make(map[int64]*entity.CloudProductBaseline)
	var cloudProductCodeBaselineMap = make(map[string]*entity.CloudProductBaseline)
	dpdkCloudProductCodeMap := make(map[string]*entity.CloudProductBaseline)
	for _, cloudProductBaseline := range cloudProductBaselineList {
		cloudProductCodeList = append(cloudProductCodeList, cloudProductBaseline.ProductCode)
		cloudProductIdBaselineMap[cloudProductBaseline.Id] = cloudProductBaseline
		cloudProductCodeBaselineMap[cloudProductBaseline.ProductCode] = cloudProductBaseline
		if _, ok := dpdkCloudProductIdMap[cloudProductBaseline.Id]; ok {
			dpdkCloudProductCodeMap[cloudProductBaseline.ProductCode] = cloudProductBaseline
		}
	}
	// 根据产品编码-售卖规格、产品编码-增值服务筛选容量输入列表
	var screenCloudProductSellSpecMap = make(map[string]interface{})
	var screenCloudProductValueAddedServiceMap = make(map[string]interface{})
	for _, v := range cloudProductPlanningList {
		// 根据产品编码-售卖规格
		if util.IsNotBlank(v.SellSpec) {
			screenCloudProductSellSpecMap[fmt.Sprintf("%s-%s", cloudProductIdBaselineMap[v.ProductId].ProductCode, v.SellSpec)] = nil
		}
		// 产品编码-增值服务
		if util.IsNotBlank(v.ValueAddedService) {
			for _, valueAddedService := range strings.Split(v.ValueAddedService, ",") {
				screenCloudProductValueAddedServiceMap[fmt.Sprintf("%s-%s", cloudProductIdBaselineMap[v.ProductId].ProductCode, valueAddedService)] = nil
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
	resourcePoolIdEcsCapacityMap := make(map[int64]*EcsCapacity)
	for _, serverCapPlanning := range serverCapPlanningList {
		if serverCapPlanning.Type == 2 && util.IsNotBlank(serverCapPlanning.Special) {
			var ecsCapacity *EcsCapacity
			util.ToObject(serverCapPlanning.Special, &ecsCapacity)
			resourcePoolIdEcsCapacityMap[serverCapPlanning.ResourcePoolId] = ecsCapacity
			continue
		}
		serverCapPlanningMap[serverCapPlanning.CapacityBaselineId] = serverCapPlanning
	}

	// 处理按照云产品编码与服务器规划的服务器关联关系
	var resourcePoolList []*entity.ResourcePool
	if err := db.Where("plan_id = ?", request.PlanId).Find(&resourcePoolList).Error; err != nil {
		return nil, err
	}
	nodeRoleIdResourcePoolMap := make(map[int64][]*entity.ResourcePool)
	for _, resourcePool := range resourcePoolList {
		nodeRoleIdResourcePoolMap[resourcePool.NodeRoleId] = append(nodeRoleIdResourcePoolMap[resourcePool.NodeRoleId], resourcePool)
	}
	var cloudProductNodeRoleRelList []*entity.CloudProductNodeRoleRel
	// 只查询资源节点角色数据，去掉管控资源节点角色
	if err := db.Where("product_id IN (?) and node_role_type = 0", cloudProductIdList).Find(&cloudProductNodeRoleRelList).Error; err != nil {
		return nil, err
	}
	productCodeResourcePoolMap := make(map[string][]*entity.ResourcePool)
	for _, cloudProductNodeRoleRel := range cloudProductNodeRoleRelList {
		productCode := cloudProductIdBaselineMap[cloudProductNodeRoleRel.ProductId].ProductCode
		productCodeResourcePoolMap[productCode] = append(productCodeResourcePoolMap[productCode], nodeRoleIdResourcePoolMap[cloudProductNodeRoleRel.NodeRoleId]...)
	}

	capConvertBaselineMap := make(map[string][]*ResponseFeatures)
	resourcePoolCapConvertMap := make(map[int64][]*ResponseCapConvert)
	extraProductCodeResponseCapConvertMap := make(map[string][]*ResponseCapConvert)
	for _, capConvertBaseline := range capConvertBaselineList {
		if capConvertBaseline.CapPlanningInput == "" {
			continue
		}
		// 根据产品编码-售卖规格、产品编码-增值服务筛选容量输入列表
		if util.IsNotBlank(capConvertBaseline.SellSpecs) {
			if _, ok := screenCloudProductSellSpecMap[fmt.Sprintf("%s-%s", capConvertBaseline.ProductCode, capConvertBaseline.SellSpecs)]; !ok {
				continue
			}
		}
		if util.IsNotBlank(capConvertBaseline.ValueAddedService) {
			if _, ok := screenCloudProductValueAddedServiceMap[fmt.Sprintf("%s-%s", capConvertBaseline.ProductCode, capConvertBaseline.ValueAddedService)]; !ok {
				continue
			}
		}
		key := fmt.Sprintf("%v-%v", capConvertBaseline.ProductCode, capConvertBaseline.CapPlanningInput)
		if _, ok := capConvertBaselineMap[key]; !ok {
			productCodeResourcePoolList := productCodeResourcePoolMap[capConvertBaseline.ProductCode]
			if len(productCodeResourcePoolList) == 0 {
				// 处理只用算BOM的容量规划输入数据，因为云产品没有关联节点角色导致需要额外处理
				responseCapConvert := &ResponseCapConvert{
					VersionId:        capConvertBaseline.VersionId,
					ProductName:      capConvertBaseline.ProductName,
					ProductCode:      capConvertBaseline.ProductCode,
					ProductType:      cloudProductCodeBaselineMap[capConvertBaseline.ProductCode].ProductType,
					SellSpecs:        capConvertBaseline.SellSpecs,
					CapPlanningInput: capConvertBaseline.CapPlanningInput,
					Unit:             capConvertBaseline.Unit,
					FeatureId:        capConvertBaseline.Id,
					FeatureMode:      capConvertBaseline.FeaturesMode,
					Description:      capConvertBaseline.Description,
				}
				if serverCapPlanningMap[capConvertBaseline.Id] != nil {
					responseCapConvert.Number = serverCapPlanningMap[capConvertBaseline.Id].Number
				}
				extraProductCodeResponseCapConvertMap[capConvertBaseline.ProductCode] = append(extraProductCodeResponseCapConvertMap[capConvertBaseline.ProductCode], responseCapConvert)
			}
			for _, productCodeResourcePool := range productCodeResourcePoolList {
				responseCapConvert := &ResponseCapConvert{
					VersionId:        capConvertBaseline.VersionId,
					ProductName:      capConvertBaseline.ProductName,
					ProductCode:      capConvertBaseline.ProductCode,
					ProductType:      cloudProductCodeBaselineMap[capConvertBaseline.ProductCode].ProductType,
					SellSpecs:        capConvertBaseline.SellSpecs,
					CapPlanningInput: capConvertBaseline.CapPlanningInput,
					Unit:             capConvertBaseline.Unit,
					FeatureId:        capConvertBaseline.Id,
					FeatureMode:      capConvertBaseline.FeaturesMode,
					Description:      capConvertBaseline.Description,
					ResourcePoolId:   productCodeResourcePool.Id,
					ResourcePoolName: productCodeResourcePool.ResourcePoolName,
				}
				if serverCapPlanningMap[capConvertBaseline.Id] != nil {
					responseCapConvert.Number = serverCapPlanningMap[capConvertBaseline.Id].Number
				}
				resourcePoolCapConvertMap[productCodeResourcePool.Id] = append(resourcePoolCapConvertMap[productCodeResourcePool.Id], responseCapConvert)
			}
		}
		if util.IsNotBlank(capConvertBaseline.Features) {
			capConvertBaselineMap[key] = append(capConvertBaselineMap[key], &ResponseFeatures{Id: capConvertBaseline.Id, Name: capConvertBaseline.Features})
		}
	}
	// 整理容量指标的特性
	for _, responseCapConverts := range resourcePoolCapConvertMap {
		for i, responseCapConvert := range responseCapConverts {
			key := fmt.Sprintf("%v-%v", responseCapConvert.ProductCode, responseCapConvert.CapPlanningInput)
			responseCapConverts[i].Features = capConvertBaselineMap[key]
			// 回显容量规划数据
			for _, feature := range capConvertBaselineMap[key] {
				if serverCapPlanningMap[feature.Id] != nil {
					responseCapConverts[i].FeatureId = feature.Id
					// 处理特性的输入值
					responseCapConverts[i].Number = serverCapPlanningMap[feature.Id].Number
					responseCapConverts[i].FeatureNumber = serverCapPlanningMap[feature.Id].FeatureNumber
				}
			}
		}
	}

	// 按产品分类
	var response []*ResponseCapClassification
	for productCode, resourcePools := range productCodeResourcePoolMap {
		responseCapClassification := &ResponseCapClassification{}
		for _, resourcePool := range resourcePools {
			responseCapConverts := resourcePoolCapConvertMap[resourcePool.Id]
			var resourcePoolCapConverts []*ResponseCapConvert
			for _, responseCapConvert := range responseCapConverts {
				if responseCapConvert.ProductCode == productCode {
					// （1）云产品没有选择DPDK规格，如果资源池开启了DPDK，则跳过 （2）云产品选择了DPDK规格，如果资源池没有开启DPDK，则跳过
					_, ok := dpdkCloudProductCodeMap[productCode]
					if (!ok && resourcePool.OpenDpdk == constant.OpenDpdk) || (ok && resourcePool.OpenDpdk == constant.CloseDpdk) {
						continue
					}
					resourcePoolCapConverts = append(resourcePoolCapConverts, responseCapConvert)
				}
			}
			if len(resourcePoolCapConverts) > 0 {
				responseCapClassification.Classification = fmt.Sprintf("%v-%v", resourcePoolCapConverts[0].ProductName, resourcePoolCapConverts[0].SellSpecs)
				responseCapClassification.ProductCode = productCode
				responseCapClassification.ProductName = resourcePoolCapConverts[0].ProductName
				responseCapClassification.ProductType = resourcePoolCapConverts[0].ProductType
				resourcePoolCapConvert := &ResourcePoolCapConvert{
					ResourcePoolId:      resourcePool.Id,
					ResourcePoolName:    resourcePool.ResourcePoolName,
					ResponseCapConverts: resourcePoolCapConverts,
					Specials:            resourcePoolIdEcsCapacityMap[resourcePool.Id],
				}
				if productCode == constant.ProductCodeECS {
					resourcePoolCapConvert.Specials = resourcePoolIdEcsCapacityMap[resourcePool.Id]
				}
				responseCapClassification.ResourcePoolCapConverts = append(responseCapClassification.ResourcePoolCapConverts, resourcePoolCapConvert)
			}
		}
		if responseCapClassification.Classification != "" {
			response = append(response, responseCapClassification)
		}
	}
	// 处理只用算BOM的容量规划输入数据，因为云产品没有关联节点角色导致需要额外处理
	for productCode, responseCapConverts := range extraProductCodeResponseCapConvertMap {
		responseCapClassification := &ResponseCapClassification{}
		for i, responseCapConvert := range responseCapConverts {
			key := fmt.Sprintf("%v-%v", responseCapConvert.ProductCode, responseCapConvert.CapPlanningInput)
			responseCapConverts[i].Features = capConvertBaselineMap[key]
			// 回显容量规划数据
			for _, feature := range capConvertBaselineMap[key] {
				if serverCapPlanningMap[feature.Id] != nil {
					responseCapConverts[i].FeatureId = feature.Id
					// 处理特性的输入值
					responseCapConverts[i].Number = serverCapPlanningMap[feature.Id].Number
					responseCapConverts[i].FeatureNumber = serverCapPlanningMap[feature.Id].FeatureNumber
				}
			}
		}
		responseCapClassification.Classification = fmt.Sprintf("%v-%v", responseCapConverts[0].ProductName, responseCapConverts[0].SellSpecs)
		responseCapClassification.ProductCode = productCode
		responseCapClassification.ProductName = responseCapConverts[0].ProductName
		responseCapClassification.ProductType = responseCapConverts[0].ProductType
		responseCapClassification.ResourcePoolCapConverts = append(responseCapClassification.ResourcePoolCapConverts, &ResourcePoolCapConvert{
			ResponseCapConverts: responseCapConverts,
		})
		response = append(response, responseCapClassification)
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
	resourcePoolServerPlanningMap, nodeRoleCodeBaselineMap, nodeRoleIdBaselineMap, serverBaselineMap, err := handleServerPlanning(db, serverPlanningList)
	if err != nil {
		return err
	}
	// 获取入参中所有容量输入指标id List
	var serverCapacityIdList []int64
	for _, serverCapacity := range request.ServerCapacityList {
		for _, requestServerCapacity := range serverCapacity.CommonServerCapacityList {
			serverCapacityIdList = append(serverCapacityIdList, requestServerCapacity.Id)
		}
		if serverCapacity.EcsCapacity != nil {
			serverCapacityIdList = append(serverCapacityIdList, serverCapacity.EcsCapacity.CapacityIdList...)
		}
	}
	// 查询容量指标基线数据map
	capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, err := getCapBaseline(db, serverCapacityIdList)
	if err != nil {
		return err
	}
	var newServerPlanningList []*entity.ServerPlanning
	var serverCapPlanningList []*entity.ServerCapPlanning
	// 保存容量规划数据
	resourcePoolIdServerCapacityMap := make(map[int64]*ResourcePoolServerCapacity)
	for _, resourcePoolServerCapacity := range request.ServerCapacityList {
		resourcePoolId := resourcePoolServerCapacity.ResourcePoolId
		if _, ok := resourcePoolIdServerCapacityMap[resourcePoolId]; ok {
			resourcePoolIdServerCapacityMap[resourcePoolId].CommonServerCapacityList = append(resourcePoolIdServerCapacityMap[resourcePoolId].CommonServerCapacityList, resourcePoolServerCapacity.CommonServerCapacityList...)
			if resourcePoolServerCapacity.EcsCapacity != nil {
				resourcePoolIdServerCapacityMap[resourcePoolId].EcsCapacity = resourcePoolServerCapacity.EcsCapacity
			}
		} else {
			// 资源池id是0表示不用计算容量，用在后面的bom
			if resourcePoolId != 0 {
				resourcePoolIdServerCapacityMap[resourcePoolId] = resourcePoolServerCapacity
			}
		}
		for _, requestServerCapacity := range resourcePoolServerCapacity.CommonServerCapacityList {
			// 查询容量换算表
			capConvertBaseline := capConvertBaselineMap[requestServerCapacity.Id]
			// 构建服务器容量规划
			serverCapPlanning := &entity.ServerCapPlanning{
				PlanId:             request.PlanId,
				Type:               1,
				CapacityBaselineId: requestServerCapacity.Id,
				Number:             requestServerCapacity.Number,
				FeatureNumber:      requestServerCapacity.FeatureNumber,
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
				ResourcePoolId:     resourcePoolId,
			}
			serverCapPlanningList = append(serverCapPlanningList, serverCapPlanning)
		}
	}
	// 计算个节点角色的服务器数量
	for resourcePoolId, resourcePoolServerCapacity := range resourcePoolIdServerCapacityMap {
		serverPlanning := resourcePoolServerPlanningMap[resourcePoolId]
		resourcePoolCapNum, ecsServerPlanning, ecsServerCapPlanning, err := computing(db, resourcePoolServerCapacity, capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, serverPlanning, nodeRoleCodeBaselineMap, serverBaselineMap)
		if err != nil {
			return err
		}
		if ecsServerPlanning != nil {
			newServerPlanningList = append(newServerPlanningList, ecsServerPlanning)
		} else {
			minimumNum := nodeRoleIdBaselineMap[serverPlanning.NodeRoleId].MinimumNum
			if resourcePoolCapNum < minimumNum {
				resourcePoolCapNum = minimumNum
			}
			serverPlanning.Number = resourcePoolCapNum
			newServerPlanningList = append(newServerPlanningList, serverPlanning)
		}
		if ecsServerCapPlanning != nil {
			serverCapPlanningList = append(serverCapPlanningList, ecsServerCapPlanning)
		}
	}
	if err = db.Transaction(func(tx *gorm.DB) error {
		// 保存服务器规划
		if err = tx.Save(&newServerPlanningList).Error; err != nil {
			return err
		}
		// 修改资源池表的dpdk状态
		for _, server := range request.ServerList {
			if err = tx.Table(entity.ResourcePoolTable).Where("id = ?", server.ResourcePoolId).Update("open_dpdk", server.OpenDpdk).Error; err != nil {
				log.Errorf("update resourcePool error: %v", err)
				return err
			}
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
	if err := db.Where("plan_id = ? and resource_pool_id = ?", request.PlanId, request.ResourcePoolId).Find(&serverCapPlanningList).Error; err != nil {
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
	for _, capServerCalcBaseline := range capServerCalcBaselines {
		expendResCodes = append(expendResCodes, capServerCalcBaseline.ExpendResCode)
	}
	var resourcePoolServerCapacity = &ResourcePoolServerCapacity{ResourcePoolId: request.ResourcePoolId}
	var capConvertBaselines []*entity.CapConvertBaseline
	if err := db.Table(entity.CapConvertBaselineTable+" ccb").Select("DISTINCT ccb.*").
		Joins("LEFT JOIN cap_actual_res_baseline carb ON ccb.product_code = carb.product_code and ccb.sell_specs = carb.sell_specs and ccb.value_added_service = carb.value_added_service and ccb.cap_planning_input = carb.sell_unit").
		Where("ccb.version_id= ? and carb.version_id= ? and carb.expend_res_code in (?) ", nodeRoleBaseline.VersionId, nodeRoleBaseline.VersionId, expendResCodes).
		Find(&capConvertBaselines).Error; err != nil {
		return nil, err
	}
	var extraCapConvertBaselines []*entity.CapConvertBaseline
	if nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeCompute {
		// 添加容量规划基线的表1和表2对不上的数据(CNBH)和CKE的容器集群数
		var extraProductCodes = []string{constant.ProductCodeCKE, constant.ProductCodeCNBH}
		if err := db.Table(entity.CapConvertBaselineTable).Where("version_id = ? and product_code in (?)", nodeRoleBaseline.VersionId, extraProductCodes).
			Find(&extraCapConvertBaselines).Error; err != nil {
			return nil, err
		}
	}
	if nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodePAASCompute || nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodePAASData {
		// 添加容量规划基线的表1和表2对不上的PAAS数据
		var extraProductCodes = []string{constant.ProductCodeKAFKA, constant.ProductCodeROCKETMQ, constant.ProductCodeRABBITMQ, constant.ProductCodeAPIM, constant.ProductCodeCONNECT, constant.ProductCodeCLS, constant.ProductCodeCLCP}
		if err := db.Table(entity.CapConvertBaselineTable).Where("version_id = ? and product_code in (?)", nodeRoleBaseline.VersionId, extraProductCodes).
			Find(&extraCapConvertBaselines).Error; err != nil {
			return nil, err
		}
	}
	if nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeNFV {
		// 添加容量规划基线的表1和表2对不上的NFV数据
		var extraProductCodes = []string{constant.ProductCodeNLB, constant.ProductCodeSLB}
		if err := db.Table(entity.CapConvertBaselineTable).Where("version_id = ? and product_code in (?)", nodeRoleBaseline.VersionId, extraProductCodes).
			Find(&extraCapConvertBaselines).Error; err != nil {
			return nil, err
		}
	}
	if nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeNETWORK {
		// 添加容量规划基线的表1和表2对不上的NETWORK数据
		var extraProductCodes = []string{constant.ProductCodeBGW}
		if err := db.Table(entity.CapConvertBaselineTable).Where("version_id = ? and product_code in (?)", nodeRoleBaseline.VersionId, extraProductCodes).
			Find(&extraCapConvertBaselines).Error; err != nil {
			return nil, err
		}
	}
	if nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeBMS {
		// 添加容量规划基线的表1和表2对不上的BMS数据
		var extraProductCodes = []string{constant.ProductCodeBMS}
		if err := db.Table(entity.CapConvertBaselineTable).Where("version_id = ? and product_code in (?)", nodeRoleBaseline.VersionId, extraProductCodes).
			Find(&extraCapConvertBaselines).Error; err != nil {
			return nil, err
		}
	}
	if nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeBIGDATA {
		// 添加容量规划基线的表1和表2对不上的BIG-DATA数据
		var extraProductCodes = []string{constant.ProductCodeCIK}
		if err := db.Table(entity.CapConvertBaselineTable).Where("version_id = ? and product_code in (?)", nodeRoleBaseline.VersionId, extraProductCodes).
			Find(&extraCapConvertBaselines).Error; err != nil {
			return nil, err
		}
	}
	if nodeRoleBaseline.NodeRoleCode == constant.NodeRoleCodeDATABASE {
		// 添加容量规划基线的表1和表2对不上的DATABASE数据
		var extraProductCodes = []string{constant.ProductCodeCEASQLDW, constant.ProductCodeCEASQLCK, constant.ProductCodeMYSQL, constant.ProductCodeCEASQLTX, constant.ProductCodePOSTGRESQL, constant.ProductCodeMONGODB, constant.ProductCodeINFLUXDB, constant.ProductCodeES, constant.ProductCodeREDIS, constant.ProductCodeDTS}
		if err := db.Table(entity.CapConvertBaselineTable).Where("version_id = ? and product_code in (?)", nodeRoleBaseline.VersionId, extraProductCodes).
			Find(&extraCapConvertBaselines).Error; err != nil {
			return nil, err
		}
	}
	for _, extraCapConvertBaseline := range extraCapConvertBaselines {
		if extraCapConvertBaseline.ProductCode == constant.ProductCodeCKE && (extraCapConvertBaseline.CapPlanningInput == constant.CapPlanningInputVCpu || extraCapConvertBaseline.CapPlanningInput == constant.CapPlanningInputMemory) {
			continue
		}
		capConvertBaselines = append(capConvertBaselines, extraCapConvertBaseline)
	}
	capConvertBaselineIdMap := make(map[int64]*entity.CapConvertBaseline)
	for _, capConvertBaseline := range capConvertBaselines {
		capConvertBaselineIdMap[capConvertBaseline.Id] = capConvertBaseline
	}

	var capacityBaselineIdList []int64
	for _, serverCapPlanning := range serverCapPlanningList {
		// 判断指标是否有ecs按规则计算，单独处理ecs容量规划-按规格数量计算
		if serverCapPlanning.Type == 2 && util.IsNotBlank(serverCapPlanning.Special) {
			util.ToObject(serverCapPlanning.Special, &resourcePoolServerCapacity.EcsCapacity)
			for _, capacityId := range resourcePoolServerCapacity.EcsCapacity.CapacityIdList {
				if _, ok := capConvertBaselineIdMap[capacityId]; ok {
					capacityBaselineIdList = append(capacityBaselineIdList, capacityId)
				}
			}
			continue
		}
		if _, ok := capConvertBaselineIdMap[serverCapPlanning.CapacityBaselineId]; !ok {
			continue
		}
		capacityBaselineIdList = append(capacityBaselineIdList, serverCapPlanning.CapacityBaselineId)
		// 构建计算参数
		resourcePoolServerCapacity.CommonServerCapacityList = append(resourcePoolServerCapacity.CommonServerCapacityList, &RequestServerCapacity{
			Id:            serverCapPlanning.CapacityBaselineId,
			Number:        serverCapPlanning.Number,
			FeatureNumber: serverCapPlanning.FeatureNumber,
		})
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
	if err = db.Where("plan_id = ? AND node_role_id = ? AND resource_pool_id = ?", request.PlanId, request.NodeRoleId, request.ResourcePoolId).Find(&serverPlanning).Error; err != nil {
		return nil, err
	}
	serverPlanning.ServerBaselineId = request.ServerBaselineId
	var serverBaseline = &entity.ServerBaseline{}
	if err = db.Where("id = ?", request.ServerBaselineId).Find(&serverBaseline).Error; err != nil {
		return nil, err
	}
	// 计算服务器数量
	var nodeRoleBaselineMap = map[string]*entity.NodeRoleBaseline{nodeRoleBaseline.NodeRoleCode: nodeRoleBaseline}
	var serverBaselineMap = map[int64]*entity.ServerBaseline{serverBaseline.Id: serverBaseline}
	resourcePoolCapNumber, ecsServerPlanning, _, err := computing(db, resourcePoolServerCapacity, capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, serverPlanning, nodeRoleBaselineMap, serverBaselineMap)
	if err != nil {
		return nil, err
	}
	if ecsServerPlanning != nil {
		resourcePoolCapNumber += ecsServerPlanning.Number
	}
	if resourcePoolCapNumber < nodeRoleBaseline.MinimumNum {
		resourcePoolCapNumber = nodeRoleBaseline.MinimumNum
	}
	return &ResponseCapCount{Number: resourcePoolCapNumber}, nil
}

func GetResourcePoolCapMap(db *gorm.DB, request *Request, resourcePoolServerPlanningMap map[int64]*entity.ServerPlanning, nodeRoleBaselineMap map[string]*entity.NodeRoleBaseline, serverBaselineMap map[int64]*entity.ServerBaseline) (map[int64]int, error) {
	var serverCapPlanningList []*entity.ServerCapPlanning
	if err := db.Where("plan_id = ?", request.PlanId).Find(&serverCapPlanningList).Error; err != nil {
		return nil, err
	}
	if serverCapPlanningList == nil || len(serverCapPlanningList) == 0 {
		return nil, nil
	}
	var resourcePoolCapNumberMap = make(map[int64]int)
	if len(serverCapPlanningList) == 0 {
		return resourcePoolCapNumberMap, nil
	}
	var productCodeList []string
	var capacityBaselineIdList []int64
	resourcePoolServerCapacityMap := make(map[int64]*ResourcePoolServerCapacity)
	for _, serverCapPlanning := range serverCapPlanningList {
		// 资源池id为0表示不用计算容量规划，后面计算BOM用
		if serverCapPlanning.ResourcePoolId == 0 {
			continue
		}
		productCodeList = append(productCodeList, serverCapPlanning.ProductCode)
		capacityBaselineIdList = append(capacityBaselineIdList, serverCapPlanning.CapacityBaselineId)
		resourcePoolServerCapacity, ok := resourcePoolServerCapacityMap[serverCapPlanning.ResourcePoolId]
		if !ok {
			resourcePoolServerCapacity = &ResourcePoolServerCapacity{
				ResourcePoolId: serverCapPlanning.ResourcePoolId,
			}
		}
		// 判断指标是否有ecs按规则计算，单独处理ecs容量规划-按规格数量计算
		if serverCapPlanning.Type == 2 && util.IsNotBlank(serverCapPlanning.Special) {
			util.ToObject(serverCapPlanning.Special, &resourcePoolServerCapacity.CommonServerCapacityList)
			continue
		}
		resourcePoolServerCapacity.CommonServerCapacityList = append(resourcePoolServerCapacity.CommonServerCapacityList, &RequestServerCapacity{
			Id:            serverCapPlanning.CapacityBaselineId,
			Number:        serverCapPlanning.Number,
			FeatureNumber: serverCapPlanning.FeatureNumber})
		resourcePoolServerCapacityMap[serverCapPlanning.ResourcePoolId] = resourcePoolServerCapacity
	}
	// 查询容量指标基线数据map
	capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, err := getCapBaseline(db, capacityBaselineIdList)
	if err != nil {
		return nil, err
	}
	for resourcePoolId, resourcePoolServerCapacity := range resourcePoolServerCapacityMap {
		// 计算各资源池服务器数量
		capNumber, ecsServerPlanning, _, err := computing(db, resourcePoolServerCapacity, capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, resourcePoolServerPlanningMap[resourcePoolId], nodeRoleBaselineMap, serverBaselineMap)
		if err != nil {
			return nil, err
		}
		resourcePoolCapNumberMap[resourcePoolId] += capNumber
		if ecsServerPlanning != nil {
			resourcePoolCapNumberMap[resourcePoolId] += ecsServerPlanning.Number
		}
	}
	return resourcePoolCapNumberMap, nil
}

func computing(db *gorm.DB, resourcePoolServerCapacity *ResourcePoolServerCapacity, capConvertBaselineMap map[int64]*entity.CapConvertBaseline, capActualResBaselineMap map[string][]*entity.CapActualResBaseline, capServerCalcBaselineMap map[string]*entity.CapServerCalcBaseline,
	serverPlanning *entity.ServerPlanning, nodeRoleBaselineMap map[string]*entity.NodeRoleBaseline, serverBaselineMap map[int64]*entity.ServerBaseline) (int, *entity.ServerPlanning, *entity.ServerCapPlanning, error) {
	var resourcePoolCapNumber int
	if serverPlanning == nil {
		return resourcePoolCapNumber, nil, nil, nil
	}
	var expendResCodeFeatureMap = make(map[string]*ExpendResFeature)
	// 部分产品特殊处理
	var serverCapacityMap = make(map[int64]float64)
	for _, requestServerCapacity := range resourcePoolServerCapacity.CommonServerCapacityList {
		serverCapacityMap[requestServerCapacity.Id] = float64(requestServerCapacity.Number)
	}
	// 计算各资源池的总消耗map，消耗资源编码为key，总消耗为value
	var resourcePoolExpendResCodeMap = make(map[string]float64)
	// 按产品将容量输入参数分类
	var productCapMap = make(map[string][]*entity.CapConvertBaseline)
	for _, v := range capConvertBaselineMap {
		if _, ok := serverCapacityMap[v.Id]; ok {
			productCapMap[v.ProductCode] = append(productCapMap[v.ProductCode], v)
		}
	}
	specialCapActualResMap := SpecialCapacityComputing(serverCapacityMap, productCapMap, resourcePoolExpendResCodeMap)

	var versionId int64
	for _, capConvertBaseline := range capConvertBaselineMap {
		versionId = capConvertBaseline.VersionId
		break
	}
	if err := calcFixedNumber(db, resourcePoolServerCapacity, versionId, serverPlanning, expendResCodeFeatureMap, resourcePoolExpendResCodeMap); err != nil {
		return resourcePoolCapNumber, nil, nil, err
	}

	// 产品编码为key，容量输入列表为value
	var resourcePoolEcsResourceProductMap = make(map[string][]*RequestServerCapacity)
	var bgwCapPlanningInputNumber int
	var bmsCapPlanningInputNumber int
	var nodeRoleIdMap = make(map[int64]*entity.NodeRoleBaseline)
	// 计算所有消耗资源的总量
	for _, commonServerCapacity := range resourcePoolServerCapacity.CommonServerCapacityList {
		// 查询容量换算表
		capConvertBaseline := capConvertBaselineMap[commonServerCapacity.Id]
		if capConvertBaseline == nil {
			continue
		}
		if capConvertBaseline.ProductCode == constant.ProductCodeBGW {
			bgwCapPlanningInputNumber = commonServerCapacity.Number
			continue
		}
		if capConvertBaseline.ProductCode == constant.ProductCodeBMS {
			bmsCapPlanningInputNumber = commonServerCapacity.Number
			continue
		}
		// 查询容量实际资源消耗表
		capConvertBaselineKey := fmt.Sprintf("%v-%v-%v-%v-%v", capConvertBaseline.ProductCode, capConvertBaseline.SellSpecs, capConvertBaseline.ValueAddedService, capConvertBaseline.CapPlanningInput, capConvertBaseline.Features)
		capActualResBaselineList := capActualResBaselineMap[capConvertBaselineKey]
		// 处理CCR数据，单实例存储容量计算需要乘以实例数量得到表2的实际消耗
		if capConvertBaseline.ProductCode == constant.ProductCodeCCR && capConvertBaseline.CapPlanningInput == constant.CapPlanningInputSingleInstanceCapacity {
			capConvertBaselines := productCapMap[capConvertBaseline.ProductCode]
			if len(capConvertBaselines) != 0 {
				var instanceNumber, singleInstanceCapacity float64
				for _, productCapConvertBaseline := range capConvertBaselines {
					if productCapConvertBaseline.CapPlanningInput == constant.CapPlanningInputInstances {
						instanceNumber = serverCapacityMap[productCapConvertBaseline.Id]
					}
					if productCapConvertBaseline.CapPlanningInput == constant.CapPlanningInputSingleInstanceCapacity {
						singleInstanceCapacity = serverCapacityMap[productCapConvertBaseline.Id]
					}
				}
				commonServerCapacity.Number = int(instanceNumber * singleInstanceCapacity)
			}
		}
		if len(capActualResBaselineList) == 0 {
			if capConvertBaseline.ProductCode == constant.ProductCodeCKE {
				// 放入CKE容量规划表1有的但是表2没有的数据
				resourcePoolEcsResourceProductMap[capConvertBaseline.ProductCode] = append(resourcePoolEcsResourceProductMap[capConvertBaseline.ProductCode], commonServerCapacity)
			}
			// 判断是否是CNBH产品
			if capConvertBaseline.ProductCode == constant.ProductCodeCNBH {
				switch capConvertBaseline.CapPlanningInput {
				case constant.CapPlanningInputOneHundred:
					commonServerCapacity.Number = int(serverCapacityMap[capConvertBaseline.Id] * constant.CapPlanningInputOneHundredInt)
				case constant.CapPlanningInputFiveHundred:
					commonServerCapacity.Number = int(serverCapacityMap[capConvertBaseline.Id] * constant.CapPlanningInputFiveHundredInt)
				case constant.CapPlanningInputOneThousand:
					commonServerCapacity.Number = int(serverCapacityMap[capConvertBaseline.Id] * constant.CapPlanningInputOneThousandInt)
				}
				for key, capActualResBaselines := range capActualResBaselineMap {
					if strings.Contains(key, constant.ProductCodeCNBH) && strings.Contains(key, constant.CapPlanningInputOpsAssets) {
						capActualResBaselineList = capActualResBaselines
					}
				}
			}
			if capConvertBaseline.ProductCode == constant.ProductCodeNLB && capConvertBaseline.CapPlanningInput == constant.CapPlanningInputNetworkNLB {
				serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
				var copyNumber float64
				if serverBaseline.Arch == constant.CpuArchX86 {
					copyNumber = 2
				} else {
					copyNumber = 4
				}
				resourcePoolExpendResCodeMap[constant.ExpendResCodeNFVVCpu] += 2 * copyNumber * serverCapacityMap[capConvertBaseline.Id]
				resourcePoolExpendResCodeMap[constant.ExpendResCodeNFVMemory] += 2.5 * copyNumber * serverCapacityMap[capConvertBaseline.Id]
			}
		}
		for _, capActualResBaseline := range capActualResBaselineList {
			if resourcePoolServerCapacity.EcsCapacity != nil {
				// 如果ecs容量规划-按规格数量计算，则将CKE、ECS_VCPU的容量输入信息放入，不判断ECS_MEM的目的是为了不造成数据重复
				if capConvertBaseline.ProductCode == constant.ProductCodeCKE || capActualResBaseline.ExpendResCode == constant.ExpendResCodeECSVCpu {
					resourcePoolEcsResourceProductMap[capConvertBaseline.ProductCode] = append(resourcePoolEcsResourceProductMap[capConvertBaseline.ProductCode], commonServerCapacity)
					continue
				}
				// 过滤ecs内存，以免下面会计算节点消耗
				if capActualResBaseline.ExpendResCode == constant.ExpendResCodeECSMemory {
					continue
				}
			}
			// 过滤特殊产品特殊计算
			if _, ok := SpecialProduct[capConvertBaseline.ProductCode]; ok {
				continue
			}
			var capActualResNumber float64
			expendResCodeSplits := strings.Split(capActualResBaseline.ExpendResCode, constant.Underline)
			if capConvertBaseline.ProductCode == expendResCodeSplits[0] && capActualResBaseline.Features != "" {
				// 消耗资源编码的下划线前半部分和自身的产品编码相同，且特性不为空，先不按照特性计算，等累加后再按照这个特性计算，这种做法是因为有其他云产品可能会用到这个云产品的资源，需要先把原始的进行累加计算，最后再来按照该特性计算
				expendResCodeFeatureMap[capActualResBaseline.ExpendResCode] = &ExpendResFeature{
					CapActualResBaseline: *capActualResBaseline,
					FeatureNumber:        commonServerCapacity.FeatureNumber,
				}
				capActualResNumber = float64(commonServerCapacity.Number)
			} else {
				capActualResNumber = capActualRes(float64(commonServerCapacity.Number), float64(commonServerCapacity.FeatureNumber), capActualResBaseline)
			}
			resourcePoolExpendResCodeMap[capActualResBaseline.ExpendResCode] += capActualResNumber
		}
	}
	// 计算每个角色节点的服务器数量
	for k, capActualResNumber := range resourcePoolExpendResCodeMap {
		specialCapActualResNum := specialCapActualResMap[k]
		if specialCapActualResNum != 0 {
			// 加上特殊产品计算的总消耗
			capActualResNumber += specialCapActualResMap[k]
		}
		expendResCodeFeature, ok := expendResCodeFeatureMap[k]
		if ok {
			capActualResNumber = capActualRes(capActualResNumber, float64(expendResCodeFeature.FeatureNumber), &expendResCodeFeature.CapActualResBaseline)
		}
		capServerCalcBaseline := capServerCalcBaselineMap[k]
		if capServerCalcBaseline == nil {
			continue
		}
		nodeRoleBaseline := nodeRoleBaselineMap[capServerCalcBaseline.ExpendNodeRoleCode]
		if nodeRoleBaseline == nil {
			continue
		}
		nodeRoleIdMap[nodeRoleBaseline.Id] = nodeRoleBaseline
		// 计算各角色节点单个服务器消耗
		capServerCalcNumber := capServerCalc(k, capServerCalcBaseline, serverBaselineMap[serverPlanning.ServerBaselineId])
		// 总消耗除以单个服务器消耗，等于服务器数量
		serverNumber := math.Ceil(capActualResNumber / capServerCalcNumber)
		if resourcePoolCapNumber < int(serverNumber) {
			resourcePoolCapNumber = int(serverNumber)
		}
	}
	var ecsServerPlanning *entity.ServerPlanning
	var ecsServerCapPlanning *entity.ServerCapPlanning
	// 单独处理ecs容量规划-按规格数量计算
	if resourcePoolServerCapacity.EcsCapacity != nil {
		var minimumNum int
		// 根据计算节点角色id查询服务器规划数据
		computeNodeRoleBaseline := nodeRoleBaselineMap[constant.NodeRoleCodeCompute]
		if computeNodeRoleBaseline != nil {
			// TODO 这期先按照ECS只有一个资源池做计算，后续做多资源池的时候需要修改
			ecsServerPlanning = serverPlanning
			minimumNum = computeNodeRoleBaseline.MinimumNum
		}
		// 查询计算节点的服务器基线配置
		if ecsServerPlanning != nil {
			serverBaseline := serverBaselineMap[ecsServerPlanning.ServerBaselineId]
			number := handleEcsData(resourcePoolServerCapacity.EcsCapacity, serverBaseline, resourcePoolEcsResourceProductMap, capConvertBaselineMap, capServerCalcBaselineMap, minimumNum)
			ecsServerPlanning.Number = number
			ecsServerCapPlanning = &entity.ServerCapPlanning{
				PlanId:         serverPlanning.PlanId,
				ProductCode:    constant.ProductCodeECS,
				Type:           2,
				FeatureNumber:  resourcePoolServerCapacity.EcsCapacity.FeatureNumber,
				VersionId:      serverBaseline.VersionId,
				Special:        util.ToString(resourcePoolServerCapacity.EcsCapacity),
				ResourcePoolId: serverPlanning.ResourcePoolId,
			}
		}
	}
	// 处理BGW云产品
	if bgwCapPlanningInputNumber != 0 {
		nodeRoleBaseline := nodeRoleBaselineMap[constant.NodeRoleCodeNETWORK]
		if nodeRoleBaseline != nil {
			serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
			if serverBaseline.NetworkInterface == constant.NetworkInterface10GE {
				resourcePoolCapNumber += int(math.Ceil(float64(bgwCapPlanningInputNumber) / 0.8 / 20))
			} else if serverBaseline.NetworkInterface == constant.NetworkInterface25GE {
				resourcePoolCapNumber += int(math.Ceil(float64(bgwCapPlanningInputNumber) / 0.8 / 50))
			} else {
				log.Errorf("can not find network interface enum: %s", serverBaseline.NetworkInterface)
			}
		}
	}
	// 处理BMS云产品
	if bmsCapPlanningInputNumber != 0 {
		bmsNodeRoleBaseline := nodeRoleBaselineMap[constant.NodeRoleCodeBMS]
		// TODO 由于资源池改动，导致BMS的资源池和BMSGW资源池不一致，后续再讨论怎么处理
		// bmsGwNodeRoleBaseline := nodeRoleBaselineMap[constant.NodeRoleCodeBMSGW]
		if bmsNodeRoleBaseline != nil {
			resourcePoolCapNumber += bmsCapPlanningInputNumber
		}
		// if bmsGwNodeRoleBaseline != nil {
		// 	resourcePoolCapNumber += int(math.Ceil(float64(bmsCapPlanningInputNumber)/30)) * 2
		// }
	}
	return resourcePoolCapNumber, ecsServerPlanning, ecsServerCapPlanning, nil
}

func calcFixedNumber(db *gorm.DB, resourcePoolServerCapacity *ResourcePoolServerCapacity, versionId int64, serverPlanning *entity.ServerPlanning, expendResCodeFeatureMap map[string]*ExpendResFeature, resourcePoolExpendResCodeMap map[string]float64) error {
	// 处理容量规划没有输入，按照固定数量计算的数据，DSP、CWP、DES，查询云产品规划表，看看是否包含了这3个云产品
	extraCalcProductCodes := []string{constant.ProductCodeDSP, constant.ProductCodeCWP, constant.ProductCodeDES}
	var extraCloudProductBaselineList []*entity.CloudProductBaseline
	if err := db.Table(entity.CloudProductBaselineTable).Where("version_id = ? and product_code in (?) and id in (select product_id from cloud_product_planning where plan_id = ?)", versionId, extraCalcProductCodes, serverPlanning.PlanId).Find(&extraCloudProductBaselineList).Error; err != nil {
		return err
	}
	extraCloudProductCodeResourcePoolIdMap := make(map[string]int64)
	var extraProductCodes []string
	for _, cloudProductBaseline := range extraCloudProductBaselineList {
		var cloudProductNodeRoleIds []int64
		extraProductCodes = append(extraProductCodes, cloudProductBaseline.ProductCode)
		if err := db.Table(entity.CloudProductNodeRoleRelTable).Select("node_role_id").Where("product_id = ?", cloudProductBaseline.Id).Find(&cloudProductNodeRoleIds).Error; err != nil {
			return err
		}
		for _, cloudProductNodeRoleId := range cloudProductNodeRoleIds {
			// TODO 这里因为固定的几个产品都没用到DPDK，所以在这里先写死
			if cloudProductNodeRoleId == serverPlanning.ServerBaselineId && serverPlanning.OpenDpdk == 0 {
				extraCloudProductCodeResourcePoolIdMap[cloudProductBaseline.ProductCode] = resourcePoolServerCapacity.ResourcePoolId
				// 加break是因为对于同一个节点角色，默认只算给第一个资源池
				break
			}
		}
	}
	if len(extraProductCodes) > 0 {
		// 查询容量实际资源消耗表
		var capActualResBaselineList []*entity.CapActualResBaseline
		if err := db.Where("version_id = ? AND product_code IN (?)", versionId, extraProductCodes).Find(&capActualResBaselineList).Error; err != nil {
			return err
		}
		extraCapActualResBaselineMap := make(map[string][]*entity.CapActualResBaseline)
		for _, capActualResBaseline := range capActualResBaselineList {
			extraCapActualResBaselineMap[capActualResBaseline.ProductCode] = append(extraCapActualResBaselineMap[capActualResBaseline.ProductCode], capActualResBaseline)
		}
		for _, extraProductCode := range extraProductCodes {
			capActualResBaselines := extraCapActualResBaselineMap[extraProductCode]
			if _, ok := extraCloudProductCodeResourcePoolIdMap[extraProductCode]; ok {
				for _, capActualResBaseline := range capActualResBaselines {
					var capActualResNumber float64
					expendResCodeSplits := strings.Split(capActualResBaseline.ExpendResCode, constant.Underline)
					if extraProductCode == expendResCodeSplits[0] && capActualResBaseline.Features != "" {
						// 消耗资源编码的下划线前半部分和自身的产品编码相同，且特性不为空，先不按照特性计算，等累加后再按照这个特性计算，这种做法是因为有其他云产品可能会用到这个云产品的资源，需要先把原始的进行累加计算，最后再来按照该特性计算
						expendResCodeFeatureMap[capActualResBaseline.ExpendResCode] = &ExpendResFeature{
							CapActualResBaseline: *capActualResBaseline,
							FeatureNumber:        1,
						}
						capActualResNumber = 1
					} else {
						capActualResNumber = capActualRes(1, 0, capActualResBaseline)
					}
					resourcePoolExpendResCodeMap[capActualResBaseline.ExpendResCode] += capActualResNumber
				}
			}
		}
	}
	return nil
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
	for _, v := range capActualResBaselineList {
		key := fmt.Sprintf("%v-%v-%v-%v-%v", v.ProductCode, v.SellSpecs, v.ValueAddedService, v.SellUnit, v.Features)
		capActualResBaselineListMap[key] = append(capActualResBaselineListMap[key], v)
	}
	// 查询容量服务器数量计算
	var capServerCalcBaselineList []*entity.CapServerCalcBaseline
	if err := db.Where("version_id = ?", capConvertBaselineList[0].VersionId).Find(&capServerCalcBaselineList).Error; err != nil {
		return nil, nil, nil, err
	}
	// 消耗资源编码为key
	var capServerCalcBaselineMap = make(map[string]*entity.CapServerCalcBaseline)
	for _, v := range capServerCalcBaselineList {
		capServerCalcBaselineMap[v.ExpendResCode] = v
	}
	return capConvertBaselineMap, capActualResBaselineListMap, capServerCalcBaselineMap, nil
}

func handleEcsData(ecsCapacity *EcsCapacity, serverBaseline *entity.ServerBaseline, ecsResourceProductMap map[string][]*RequestServerCapacity, capConvertBaselineMap map[int64]*entity.CapConvertBaseline, capServerCalcBaselineMap map[string]*entity.CapServerCalcBaseline, minimumNumber int) int {
	// 计算其它计算节点相关的产品消耗的ECS实例数量
	var ecsCapacityList []*EcsSpecs
	ecsCapacityList = append(ecsCapacityList, ecsCapacity.List...)
	for k, v := range ecsResourceProductMap {
		switch k {
		case constant.ProductCodeCKE:
			var vCpu, cluster float64
			for _, requestCapacity := range v {
				if capConvertBaselineMap[requestCapacity.Id].CapPlanningInput == constant.CapPlanningInputVCpu {
					vCpu = float64(requestCapacity.Number)
				}
				if capConvertBaselineMap[requestCapacity.Id].CapPlanningInput == constant.CapPlanningInputContainerCluster {
					cluster = float64(requestCapacity.Number)
				}
			}
			waterLevel, _ := strconv.ParseFloat(capServerCalcBaselineMap[constant.ExpendResCodeECSVCpu].WaterLevel, 64)
			ecsCapacityList = append(ecsCapacityList, &EcsSpecs{
				CpuNumber:    16,
				MemoryNumber: 32,
				Count:        int(math.Ceil(vCpu/waterLevel/14.6 + cluster*3)),
			})
		case constant.ProductCodeCNBH:
			// 前面已经算好了CNBH每个资产类型的数量，这里不需要计算了
			var assetsNumber float64
			for _, requestCapacity := range v {
				assetsNumber += float64(requestCapacity.Number)
			}
			ecsCapacityList = append(ecsCapacityList, &EcsSpecs{
				CpuNumber:    4,
				MemoryNumber: 8,
				Count:        int(math.Ceil(assetsNumber / 50)),
			})
		case constant.ProductCodeCNFW:
			for _, requestCapacity := range v {
				if capConvertBaselineMap[requestCapacity.Id].SellSpecs == constant.SellSpecsStandardEdition {
					ecsCapacityList = append(ecsCapacityList, &EcsSpecs{
						CpuNumber:    4,
						MemoryNumber: 8,
						Count:        requestCapacity.Number,
					})
				}
				if capConvertBaselineMap[requestCapacity.Id].SellSpecs == constant.SellSpecsUltimateEdition {
					ecsCapacityList = append(ecsCapacityList, &EcsSpecs{
						CpuNumber:    8,
						MemoryNumber: 16,
						Count:        requestCapacity.Number,
					})
				}
			}
		case constant.ProductCodeCCR:
			var instanceNumber int
			for _, requestCapacity := range v {
				instanceNumber += requestCapacity.Number
			}
			ecsCapacityList = append(ecsCapacityList, &EcsSpecs{
				CpuNumber:    4,
				MemoryNumber: 4,
				Count:        instanceNumber,
			})
		}
	}
	// 计算ecs小箱子
	var items []util.Item
	for _, v := range ecsCapacityList {
		width := float64(v.CpuNumber)
		// 每个实例额外消耗内存（单位：M）=138M(libvirt)+8M(IO)+16M(GPU)+8M*vCPU+内存(M)/512，ARM额外128M
		extraMemorySpent := float64(138 + 8 + 16 + 8*v.CpuNumber + v.MemoryNumber*1024/512)
		if serverBaseline.Arch == constant.CpuArchARM {
			extraMemorySpent += 128
		}
		height := float64(v.MemoryNumber) + extraMemorySpent/1024
		items = append(items, util.Item{Size: util.Rectangle{Width: width, Height: height}, Number: v.Count})
	}
	// 节点固定开销5C8G，则单节点可用vCPU=(节点总vCPU*-5）*70%*超分系数N；单节点可用内存=(节点总内存*-8）*70%，为大箱子的长宽
	var boxSize = util.Rectangle{Width: float64((serverBaseline.Cpu-5)*ecsCapacity.FeatureNumber) * 0.7, Height: float64(serverBaseline.Memory-8) * 0.7}
	boxes := util.Pack(items, boxSize)
	// TODO 如果不是单独部署，则不能使用最小数量
	if minimumNumber > len(boxes) {
		return minimumNumber
	}
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
			ResourcePoolId:   v.ResourcePoolId,
		})
	}
	if err := db.Create(&serverPlanningEntityList).Error; err != nil {
		return nil, err
	}
	return serverPlanningEntityList, nil
}

func handleServerPlanning(db *gorm.DB, serverPlanningList []*entity.ServerPlanning) (map[int64]*entity.ServerPlanning, map[string]*entity.NodeRoleBaseline, map[int64]*entity.NodeRoleBaseline, map[int64]*entity.ServerBaseline, error) {
	// 将服务器规划数据封装成map，角色节点id为key
	var resourcePoolServerPlanningMap = make(map[int64]*entity.ServerPlanning)
	var nodeRoleIdIdList []int64
	var serverBaselineIdList []int64
	for _, serverPlanning := range serverPlanningList {
		resourcePoolServerPlanningMap[serverPlanning.ResourcePoolId] = serverPlanning
		nodeRoleIdIdList = append(nodeRoleIdIdList, serverPlanning.NodeRoleId)
		serverBaselineIdList = append(serverBaselineIdList, serverPlanning.ServerBaselineId)
	}
	// 查询节点角色，为了计算容量计算表3的数据
	var nodeRoleBaselineList []*entity.NodeRoleBaseline
	if err := db.Where("id IN (?)", nodeRoleIdIdList).Find(&nodeRoleBaselineList).Error; err != nil {
		return nil, nil, nil, nil, err
	}
	var nodeRoleCodeBaselineMap = make(map[string]*entity.NodeRoleBaseline)
	var nodeRoleIdBaselineMap = make(map[int64]*entity.NodeRoleBaseline)
	for _, nodeRoleBaseline := range nodeRoleBaselineList {
		nodeRoleCodeBaselineMap[nodeRoleBaseline.NodeRoleCode] = nodeRoleBaseline
		nodeRoleIdBaselineMap[nodeRoleBaseline.Id] = nodeRoleBaseline
	}
	// 查询服务器基线数据，为了计算容量计算表3的数据
	var serverBaselineList []*entity.ServerBaseline
	if err := db.Where("id IN (?)", serverBaselineIdList).Find(&serverBaselineList).Error; err != nil {
		return nil, nil, nil, nil, err
	}
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	for _, v := range serverBaselineList {
		serverBaselineMap[v.Id] = v
	}
	return resourcePoolServerPlanningMap, nodeRoleCodeBaselineMap, nodeRoleIdBaselineMap, serverBaselineMap, nil
}

var SpecialProduct = map[string]interface{}{constant.ProductCodeCKE: nil}

func SpecialCapacityComputing(serverCapacityMap map[int64]float64, productCapMap map[string][]*entity.CapConvertBaseline, expendResCodeMap map[string]float64) map[string]float64 {
	var capActualResMap = make(map[string]float64)
	for productCode, capConvertBaselineList := range productCapMap {
		switch productCode {
		case constant.ProductCodeCKE:
			var vCpu, memory, cluster float64
			for _, capConvertBaseline := range capConvertBaselineList {
				switch capConvertBaseline.CapPlanningInput {
				case constant.CapPlanningInputVCpu:
					vCpu = serverCapacityMap[capConvertBaseline.Id]
				case constant.CapPlanningInputMemory:
					memory = serverCapacityMap[capConvertBaseline.Id]
				case constant.CapPlanningInputContainerCluster:
					cluster = serverCapacityMap[capConvertBaseline.Id]
				}
			}
			capActualResMap[constant.ExpendResCodeECSVCpu] = 48*cluster + 16*vCpu/0.7/14.6
			capActualResMap[constant.ExpendResCodeECSMemory] = 96*cluster + 32*memory/0.7/29.4
			break
		case constant.ProductCodeKAFKA:
			broker, standardEdition, professionalEdition, enterpriseEdition, platinumEdition, diskCapacity := handlePAASCapPlanningInput(serverCapacityMap, capConvertBaselineList)
			expendResCodeMap[constant.ExpendResCodePAASDataVCpu] += (2*broker+2.3)*standardEdition + (4*broker+2.3)*professionalEdition + (8*broker+2.3)*enterpriseEdition + (16*broker+2.3)*platinumEdition
			expendResCodeMap[constant.ExpendResCodePAASDataMemory] += (4*broker+4.5)*standardEdition + (8*broker+4.5)*professionalEdition + (16*broker+4.5)*enterpriseEdition + (32*broker+4.5)*platinumEdition
			expendResCodeMap[constant.ExpendResCodePAASDataDisk] += diskCapacity
			break
		case constant.ProductCodeROCKETMQ:
			_, standardEdition, professionalEdition, enterpriseEdition, platinumEdition, diskCapacity := handlePAASCapPlanningInput(serverCapacityMap, capConvertBaselineList)
			expendResCodeMap[constant.ExpendResCodePAASDataVCpu] += 14.8*standardEdition + 26.8*professionalEdition + 38.8*enterpriseEdition + 50.8*platinumEdition
			expendResCodeMap[constant.ExpendResCodePAASDataMemory] += 29.8*standardEdition + 53.5*professionalEdition + 77.5*enterpriseEdition + 101.5*platinumEdition
			expendResCodeMap[constant.ExpendResCodePAASDataDisk] += diskCapacity
			break
		case constant.ProductCodeRABBITMQ:
			broker, standardEdition, professionalEdition, enterpriseEdition, platinumEdition, diskCapacity := handlePAASCapPlanningInput(serverCapacityMap, capConvertBaselineList)
			expendResCodeMap[constant.ExpendResCodePAASDataVCpu] += (2*broker+0.5)*standardEdition + (4*broker+0.5)*professionalEdition + (8*broker+0.5)*enterpriseEdition + (16*broker+0.5)*platinumEdition
			expendResCodeMap[constant.ExpendResCodePAASDataMemory] += (4*broker+1)*standardEdition + (8*broker+1)*professionalEdition + (16*broker+1)*enterpriseEdition + (32*broker+1)*platinumEdition
			expendResCodeMap[constant.ExpendResCodePAASDataDisk] += diskCapacity
			break
		case constant.ProductCodeAPIM:
			_, standardEdition, professionalEdition, enterpriseEdition, _, _ := handlePAASCapPlanningInput(serverCapacityMap, capConvertBaselineList)
			expendResCodeMap[constant.ExpendResCodePAASComputeVCpu] += 3*standardEdition + 6*professionalEdition + 12*enterpriseEdition
			expendResCodeMap[constant.ExpendResCodePAASComputeMemory] += 6*standardEdition + 12*professionalEdition + 24*enterpriseEdition
			break
		case constant.ProductCodeCONNECT:
			_, standardEdition, professionalEdition, enterpriseEdition, _, _ := handlePAASCapPlanningInput(serverCapacityMap, capConvertBaselineList)
			expendResCodeMap[constant.ExpendResCodePAASComputeVCpu] += 4*standardEdition + 8*professionalEdition + 24*enterpriseEdition
			expendResCodeMap[constant.ExpendResCodePAASComputeMemory] += 16*standardEdition + 32*professionalEdition + 96*enterpriseEdition
			break
		case constant.ProductCodeCLCP:
			_, standardEdition, professionalEdition, enterpriseEdition, _, _ := handlePAASCapPlanningInput(serverCapacityMap, capConvertBaselineList)
			expendResCodeMap[constant.ExpendResCodePAASComputeVCpu] += 16*standardEdition + 64*professionalEdition + 96*enterpriseEdition
			expendResCodeMap[constant.ExpendResCodePAASComputeMemory] += 32*standardEdition + 128*professionalEdition + 196*enterpriseEdition
			break
		case constant.ProductCodeCLS:
			for _, capConvertBaseline := range capConvertBaselineList {
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputLogStorage {
					expendResCodeMap[constant.ExpendResCodePAASDataDisk] += serverCapacityMap[capConvertBaseline.Id] * 1024
					break
				}
			}
			break
		case constant.ProductCodeSLB:
			var overallocation, basicType, standardType, highOrderType float64
			for _, capConvertBaseline := range capConvertBaselineList {
				switch capConvertBaseline.CapPlanningInput {
				case constant.CapPlanningInputOverallocation:
					overallocation = serverCapacityMap[capConvertBaseline.Id]
				case constant.CapPlanningInputBasicType:
					basicType = serverCapacityMap[capConvertBaseline.Id]
				case constant.CapPlanningInputStandardType:
					standardType = serverCapacityMap[capConvertBaseline.Id]
				case constant.CapPlanningInputHighOrderType:
					highOrderType = serverCapacityMap[capConvertBaseline.Id]
				}
			}
			if overallocation == 0 {
				overallocation = 1
			}
			expendResCodeMap[constant.ExpendResCodeNFVVCpu] += (8*basicType + 16*standardType + 32*highOrderType) / overallocation
			expendResCodeMap[constant.ExpendResCodeNFVMemory] += (8*basicType + 16*standardType + 32*highOrderType) / overallocation
			break
		case constant.ProductCodeCEASQLDW, constant.ProductCodeCEASQLCK, constant.ProductCodeMYSQL, constant.ProductCodeCEASQLTX, constant.ProductCodePOSTGRESQL, constant.ProductCodeMONGODB, constant.ProductCodeINFLUXDB, constant.ProductCodeES:
			var vCpu, memory, disk, copyNumber float64
			for _, capConvertBaseline := range capConvertBaselineList {
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputVCpuTotal {
					vCpu = serverCapacityMap[capConvertBaseline.Id]
				}
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputMemTotal {
					memory = serverCapacityMap[capConvertBaseline.Id]
				}
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputBusinessDataVolume {
					disk = serverCapacityMap[capConvertBaseline.Id]
				}
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputCopy {
					copyNumber = serverCapacityMap[capConvertBaseline.Id]
				}
			}
			expendResCodeMap[constant.ExpendResCodeDBVCpu] += vCpu * copyNumber
			expendResCodeMap[constant.ExpendResCodeDBMemory] += memory * copyNumber
			expendResCodeMap[constant.ExpendResCodeDBDisk] += disk * copyNumber
			break
		case constant.ProductCodeREDIS:
			var vCpu, memory, copyNumber float64
			for _, capConvertBaseline := range capConvertBaselineList {
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputVCpuTotal {
					vCpu = serverCapacityMap[capConvertBaseline.Id]
				}
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputMemTotal {
					memory = serverCapacityMap[capConvertBaseline.Id]
				}
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputCopy {
					copyNumber = serverCapacityMap[capConvertBaseline.Id]
				}
			}
			expendResCodeMap[constant.ExpendResCodeDBVCpu] += vCpu * copyNumber
			expendResCodeMap[constant.ExpendResCodeDBMemory] += memory * copyNumber
			// REDIS的磁盘量等于内存量
			expendResCodeMap[constant.ExpendResCodeDBDisk] += memory * copyNumber
			break
		case constant.ProductCodeCIK:
			var vCpu, memory, disk, copyNumber float64
			for _, capConvertBaseline := range capConvertBaselineList {
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputVCpuTotal {
					vCpu = serverCapacityMap[capConvertBaseline.Id]
				}
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputMemTotal {
					memory = serverCapacityMap[capConvertBaseline.Id]
				}
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputBusinessDataVolume {
					disk = serverCapacityMap[capConvertBaseline.Id]
				}
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputCopy {
					copyNumber = serverCapacityMap[capConvertBaseline.Id]
				}
			}
			expendResCodeMap[constant.ExpendResCodeBDVCpu] += vCpu * copyNumber
			expendResCodeMap[constant.ExpendResCodeBDMemory] += memory * copyNumber
			expendResCodeMap[constant.ExpendResCodeBDDisk] += disk * copyNumber
			break
		case constant.ProductCodeDTS:
			var small, middle, large float64
			for _, capConvertBaseline := range capConvertBaselineList {
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputSmall {
					small = serverCapacityMap[capConvertBaseline.Id]
				}
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputMiddle {
					middle = serverCapacityMap[capConvertBaseline.Id]
				}
				if capConvertBaseline.CapPlanningInput == constant.CapPlanningInputLarge {
					large = serverCapacityMap[capConvertBaseline.Id]
				}
			}
			expendResCodeMap[constant.ExpendResCodeDBVCpu] += small*5 + middle*8 + large*14
			expendResCodeMap[constant.ExpendResCodeDBMemory] += small*16 + middle*30 + large*40
			expendResCodeMap[constant.ExpendResCodeDBDisk] += (small*3 + middle*3 + large*3) * 1024
			break
		default:
			break
		}
	}
	return capActualResMap
}

// 处理PAAS云产品的容量规划表1的输入参数
func handlePAASCapPlanningInput(serverCapacityMap map[int64]float64, capConvertBaselines []*entity.CapConvertBaseline) (float64, float64, float64, float64, float64, float64) {
	var broker, standardEdition, professionalEdition, enterpriseEdition, platinumEdition, diskCapacity float64
	for _, capConvertBaseline := range capConvertBaselines {
		switch capConvertBaseline.CapPlanningInput {
		case constant.CapPlanningInputBroker:
			broker = serverCapacityMap[capConvertBaseline.Id]
		case constant.CapPlanningInputStandardEdition:
			standardEdition = serverCapacityMap[capConvertBaseline.Id]
		case constant.CapPlanningInputProfessionalEdition:
			professionalEdition = serverCapacityMap[capConvertBaseline.Id]
		case constant.CapPlanningInputEnterpriseEdition:
			enterpriseEdition = serverCapacityMap[capConvertBaseline.Id]
		case constant.CapPlanningInputPlatinumEdition:
			platinumEdition = serverCapacityMap[capConvertBaseline.Id]
		case constant.CapPlanningInputDiskCapacity:
			diskCapacity = serverCapacityMap[capConvertBaseline.Id]
		}
	}
	return broker, standardEdition, professionalEdition, enterpriseEdition, platinumEdition, diskCapacity
}

func capActualRes(number, featureNumber float64, capActualResBaseline *entity.CapActualResBaseline) float64 {
	if featureNumber <= 0 {
		featureNumber = 1
	}
	numerator, _ := strconv.ParseFloat(capActualResBaseline.OccRatioNumerator, 64)
	if numerator == 0 {
		numerator = featureNumber
	}
	denominator, _ := strconv.ParseFloat(capActualResBaseline.OccRatioDenominator, 64)
	if denominator == 0 {
		denominator = featureNumber
	}
	// 总消耗
	return number / numerator * denominator
}

func capServerCalc(expendResCode string, capServerCalcBaseline *entity.CapServerCalcBaseline, serverBaseline *entity.ServerBaseline) float64 {
	// 判断用哪个容量参数
	var singleCapacity int
	if strings.Contains(expendResCode, constant.ExpendResCodeEndOfVCpu) {
		singleCapacity = serverBaseline.Cpu
	}
	if strings.Contains(expendResCode, constant.ExpendResCodeEndOfMem) {
		singleCapacity = serverBaseline.Memory
	}
	if strings.Contains(expendResCode, constant.ExpendResCodeEndOfDisk) && capServerCalcBaseline.NodeWastageCalcType != 3 {
		singleCapacity = serverBaseline.StorageDiskNum * serverBaseline.StorageDiskCapacity
	}
	nodeWastage, _ := strconv.ParseFloat(capServerCalcBaseline.NodeWastage, 64)
	waterLevel, _ := strconv.ParseFloat(capServerCalcBaseline.WaterLevel, 64)
	// 单个服务器消耗
	if capServerCalcBaseline.NodeWastageCalcType == 1 {
		return (float64(singleCapacity) - nodeWastage) * waterLevel
	}
	if capServerCalcBaseline.NodeWastageCalcType == 2 {
		return (float64(singleCapacity) * (1 - nodeWastage)) * waterLevel
	}
	return ((float64(serverBaseline.StorageDiskNum) - nodeWastage) * float64(serverBaseline.StorageDiskCapacity)) * waterLevel
}
