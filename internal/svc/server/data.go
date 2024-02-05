package server

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"errors"
	"fmt"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"
	"math"
	"strconv"
	"strings"
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
	if err = db.Model(&entity.ServerPlanning{}).Where("plan_id = ?", request.PlanId).Find(&serverPlanningList).Error; err != nil {
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
		})
	}
	if err := db.Create(&serverPlanningEntityList).Error; err != nil {
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
	//单独处理ecs产品指标
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
		key := v.ProductCode + v.SellSpecs + v.CapPlanningInput
		if _, ok := capConvertBaselineMap[key]; !ok {
			responseCapConvert := &ResponseCapConvert{
				VersionId:        v.VersionId,
				ProductName:      v.ProductName,
				ProductCode:      v.ProductCode,
				SellSpecs:        v.SellSpecs,
				CapPlanningInput: v.CapPlanningInput,
				Unit:             v.Unit,
				FeatureMode:      v.FeaturesMode,
				Description:      v.Description,
			}
			responseCapConvert.FeatureId = v.Id
			responseCapConvertList = append(responseCapConvertList, responseCapConvert)
		}
		capConvertBaselineMap[key] = append(capConvertBaselineMap[key], &ResponseFeatures{Id: v.Id, Name: v.Features})
	}
	//整理容量指标的特性
	var classificationMap = make(map[string][]*ResponseCapConvert)
	var specialMap = make(map[string]*EcsCapacity)
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
		//单独处理ecs产品指标
		if v.ProductCode == "ECS" {
			specialMap[classification] = ecsCapacity
		}
	}
	//按产品分类
	var response []*ResponseCapClassification
	for k, v := range classificationMap {
		response = append(response, &ResponseCapClassification{
			Classification: k,
			ProductName:    v[0].ProductName,
			ProductCode:    v[0].ProductCode,
			CapConvert:     v,
			Special:        specialMap[k],
		})
	}
	return response, nil
}

func CountCapacity(request *RequestServerCapacityCount) (*ResponseCapCount, error) {
	//缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	//查询容量规划
	var serverCapPlanningList []*entity.ServerCapPlanning
	if err := db.Where("plan_id = ? AND node_role_id = ?", request.PlanId, request.NodeRoleId).Find(&serverCapPlanningList).Error; err != nil {
		return nil, err
	}
	var capacityBaselineIdList []int64
	for _, v := range serverCapPlanningList {
		capacityBaselineIdList = append(capacityBaselineIdList, v.CapacityBaselineId)
	}
	if len(capacityBaselineIdList) == 0 {
		return nil, nil
	}
	//查询容量指标基线
	capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, err := getCapBaseline(db, capacityBaselineIdList)
	if err != nil {
		return nil, err
	}
	var serverBaseline = &entity.ServerBaseline{}
	if err = db.Where("id = ?", request.ServerBaselineId).Find(&serverBaseline).Error; err != nil {
		return nil, err
	}
	if serverBaseline.Id == 0 {
		return nil, errors.New("服务器基线不存在")
	}
	//计算服务器数量
	var serverNumber int
	for _, v := range serverCapPlanningList {
		capConvertBaseline := capConvertBaselineMap[v.CapacityBaselineId]
		//查询容量实际资源消耗表
		capActualResBaseline := capActualResBaselineMap[fmt.Sprintf("%v-%v-%v-%v", capConvertBaseline.ProductCode, capConvertBaseline.SellSpecs, capConvertBaseline.CapPlanningInput, capConvertBaseline.Features)]
		if capActualResBaseline == nil {
			return nil, errors.New("服务器容量规划特性不存在")
		}
		//查询容量服务器数量计算
		capServerCalcBaseline := capServerCalcBaselineMap[capActualResBaseline.ExpendResCode]
		if capServerCalcBaseline == nil {
			return nil, errors.New("服务器数量计算数据不存在")
		}
		num := CountServerNumber(v.Number, v.FeatureNumber, capActualResBaseline, capServerCalcBaseline, serverBaseline)
		if serverNumber < num {
			serverNumber = num
		}
	}
	return &ResponseCapCount{Number: serverNumber}, nil
}

func SaveServerCapacity(request *Request) error {
	//缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	//保存服务器规划
	if err := CreateServerPlanning(db, request); err != nil {
		return err
	}
	//查询服务器容量数据map
	var serverCapacityIdList []int64
	for _, v := range request.ServerCapacityList {
		serverCapacityIdList = append(serverCapacityIdList, v.Id)
	}
	//单独处理ecs容量规划-按规格数量计算
	if request.EcsCapacity != nil {
		serverCapacityIdList = append(serverCapacityIdList, request.EcsCapacity.CapacityIdList...)
	}
	//查询容量指标基线
	capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, err := getCapBaseline(db, serverCapacityIdList)
	if err != nil {
		return err
	}
	//查询服务器规划是否有已保存数据
	var serverPlanningList []*entity.ServerPlanning
	if err = db.Where("plan_id = ?", request.PlanId).Find(&serverPlanningList).Error; err != nil {
		return err
	}
	var serverPlanningMap = make(map[int64]*entity.ServerPlanning)
	var nodeRoleIdIdList []int64
	var serverBaselineIdList []int64
	for _, v := range serverPlanningList {
		serverPlanningMap[v.NodeRoleId] = v
		nodeRoleIdIdList = append(nodeRoleIdIdList, v.NodeRoleId)
		serverBaselineIdList = append(serverBaselineIdList, v.ServerBaselineId)
	}
	//查询节点角色
	var nodeRoleBaselineList []*entity.NodeRoleBaseline
	if err = db.Where("id IN (?)", nodeRoleIdIdList).Find(&nodeRoleBaselineList).Error; err != nil {
		return err
	}
	var nodeRoleBaselineMap = make(map[string]*entity.NodeRoleBaseline)
	for _, v := range nodeRoleBaselineList {
		nodeRoleBaselineMap[v.NodeRoleCode] = v
	}
	//查询服务器基线数据
	var serverBaselineList []*entity.ServerBaseline
	if err = db.Where("id IN (?)", serverBaselineIdList).Find(&serverBaselineList).Error; err != nil {
		return err
	}
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	for _, v := range serverBaselineList {
		serverBaselineMap[v.Id] = v
	}

	var serverCapPlanningList []*entity.ServerCapPlanning
	var nodeRoleCapNumberMap = make(map[int64]int)
	if err = db.Transaction(func(tx *gorm.DB) error {
		for _, v := range request.ServerCapacityList {
			//处理参数
			capConvertBaseline, capActualResBaseline, capServerCalcBaseline, nodeRoleBaseline, serverPlanning, serverBaseline, err := handleCapCapacityParam(v.Id, capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, nodeRoleBaselineMap, serverPlanningMap, serverBaselineMap)
			if err != nil {
				log.Errorf("handleCapCapacityParam error: %v", err)
			}
			//计算服务器数量
			serverNumber := CountServerNumber(v.Number, v.FeatureNumber, capActualResBaseline, capServerCalcBaseline, serverBaseline)
			//保存服务器规划数据
			if nodeRoleCapNumberMap[nodeRoleBaseline.Id] < serverNumber {
				nodeRoleCapNumberMap[nodeRoleBaseline.Id] = serverNumber
			}
			serverPlanning.Number = nodeRoleCapNumberMap[nodeRoleBaseline.Id]
			if err = tx.Save(&serverPlanning).Error; err != nil {
				return err
			}
			//构建服务器容量规划
			serverCapPlanningList = append(serverCapPlanningList, &entity.ServerCapPlanning{
				PlanId:             request.PlanId,
				NodeRoleId:         nodeRoleBaseline.Id,
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
				ExpendResCode:      capActualResBaseline.ExpendResCode,
				Special:            "{}",
			})
		}
		//单独处理ecs容量规划-按规格数量计算
		if request.EcsCapacity != nil {
			serverCapPlanning, err := handleSpecialData(tx, request, capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, nodeRoleBaselineMap, serverPlanningMap, serverBaselineMap)
			if err != nil {
				return err
			}
			serverCapPlanningList = append(serverCapPlanningList, serverCapPlanning)
		}
		//保存服务器容量规划
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

func getCapBaseline(db *gorm.DB, serverCapacityIdList []int64) (map[int64]*entity.CapConvertBaseline, map[string]*entity.CapActualResBaseline, map[string]*entity.CapServerCalcBaseline, error) {
	var capConvertBaselineList []*entity.CapConvertBaseline
	if err := db.Where("id IN (?)", serverCapacityIdList).Find(&capConvertBaselineList).Error; err != nil {
		return nil, nil, nil, err
	}
	if len(capConvertBaselineList) == 0 {
		return nil, nil, nil, errors.New("服务器容量规划指标不存在")
	}
	var capConvertBaselineMap = make(map[int64]*entity.CapConvertBaseline)
	for _, v := range capConvertBaselineList {
		capConvertBaselineMap[v.Id] = v
	}
	//查询容量实际资源消耗表
	var capActualResBaselineList []*entity.CapActualResBaseline
	if err := db.Where("version_id = ?", capConvertBaselineList[0].VersionId).Find(&capActualResBaselineList).Error; err != nil {
		return nil, nil, nil, err
	}
	var capActualResBaselineMap = make(map[string]*entity.CapActualResBaseline)
	for _, v := range capActualResBaselineList {
		capActualResBaselineMap[fmt.Sprintf("%v-%v-%v-%v", v.ProductCode, v.SellSpecs, v.SellUnit, v.Features)] = v
	}
	//查询容量服务器数量计算
	var capServerCalcBaselineList []*entity.CapServerCalcBaseline
	if err := db.Where("version_id = ?", capConvertBaselineList[0].VersionId).Find(&capServerCalcBaselineList).Error; err != nil {
		return nil, nil, nil, err
	}
	var capServerCalcBaselineMap = make(map[string]*entity.CapServerCalcBaseline)
	for _, v := range capServerCalcBaselineList {
		capServerCalcBaselineMap[v.ExpendResCode] = v
	}
	return capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, nil
}

func CountServerNumber(number, featureNumber int, capActualResBaseline *entity.CapActualResBaseline, capServerCalcBaseline *entity.CapServerCalcBaseline, serverBaseline *entity.ServerBaseline) int {
	numerator, _ := strconv.ParseFloat(capActualResBaseline.OccRatioNumerator, 64)
	if numerator == 0 {
		numerator = float64(featureNumber)
	}
	denominator, _ := strconv.ParseFloat(capActualResBaseline.OccRatioDenominator, 64)
	if numerator == 0 || denominator == 0 {
		numerator = 1
		denominator = 1
	}
	//总消耗
	capacityNumber := float64(number) / numerator * denominator
	//判断用哪个容量参数
	var singleCapacity int
	switch capActualResBaseline.ExpendResCode {
	case "ECS_VCPU":
		singleCapacity = serverBaseline.Cpu
	case "ECS_MEM":
		singleCapacity = serverBaseline.Memory
	case "EBS_DISK", "EFS_DISK", "OSS_DISK":
		singleCapacity = serverBaseline.StorageDiskNum * serverBaseline.StorageDiskCapacity
	}
	nodeWastage, _ := strconv.ParseFloat(capServerCalcBaseline.NodeWastage, 64)
	waterLevel, _ := strconv.ParseFloat(capServerCalcBaseline.WaterLevel, 64)
	//单个服务器消耗
	var consumeNumber float64
	if capServerCalcBaseline.NodeWastageCalcType == 1 {
		consumeNumber = (float64(singleCapacity) - nodeWastage) * waterLevel
	} else {
		consumeNumber = (float64(singleCapacity) * (1 - nodeWastage)) * waterLevel
	}
	//总消耗除以单个服务器消耗，等于服务器数量
	serverNumber := math.Ceil(capacityNumber / consumeNumber)
	return int(serverNumber)
}

func handleCapCapacityParam(id int64, capConvertBaselineMap map[int64]*entity.CapConvertBaseline, capActualResBaselineMap map[string]*entity.CapActualResBaseline,
	capServerCalcBaselineMap map[string]*entity.CapServerCalcBaseline, nodeRoleBaselineMap map[string]*entity.NodeRoleBaseline,
	serverPlanningMap map[int64]*entity.ServerPlanning, serverBaselineMap map[int64]*entity.ServerBaseline) (*entity.CapConvertBaseline,
	*entity.CapActualResBaseline, *entity.CapServerCalcBaseline, *entity.NodeRoleBaseline, *entity.ServerPlanning, *entity.ServerBaseline, error) {
	capConvertBaseline := capConvertBaselineMap[id]
	//查询容量实际资源消耗表
	key := fmt.Sprintf("%v-%v-%v-%v", capConvertBaseline.ProductCode, capConvertBaseline.SellSpecs, capConvertBaseline.CapPlanningInput, capConvertBaseline.Features)
	capActualResBaseline := capActualResBaselineMap[key]
	if capActualResBaseline == nil {
		return nil, nil, nil, nil, nil, nil, errors.New("服务器容量规划特性不存在，key：" + key)
	}
	//查询容量服务器数量计算
	capServerCalcBaseline := capServerCalcBaselineMap[capActualResBaseline.ExpendResCode]
	if capServerCalcBaseline == nil {
		return nil, nil, nil, nil, nil, nil, errors.New("服务器数量计算数据不存在")
	}
	//查询服务器规划是否有已保存数据
	nodeRoleBaseline := nodeRoleBaselineMap[capServerCalcBaseline.ExpendNodeRoleCode]
	if nodeRoleBaseline == nil {
		return nil, nil, nil, nil, nil, nil, errors.New("节点角色数据不存在")
	}
	serverPlanning := serverPlanningMap[nodeRoleBaseline.Id]
	if serverPlanning == nil {
		return nil, nil, nil, nil, nil, nil, errors.New("方案不存在")
	}
	//查询服务器基线数据
	serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
	if serverBaseline == nil {
		return nil, nil, nil, nil, nil, nil, errors.New("服务器基线数据不存在")
	}
	return capConvertBaseline, capActualResBaseline, capServerCalcBaseline, nodeRoleBaseline, serverPlanning, serverBaseline, nil
}

func handleSpecialData(db *gorm.DB, request *Request, capConvertBaselineMap map[int64]*entity.CapConvertBaseline, capActualResBaselineMap map[string]*entity.CapActualResBaseline, capServerCalcBaselineMap map[string]*entity.CapServerCalcBaseline,
	nodeRoleBaselineMap map[string]*entity.NodeRoleBaseline, serverPlanningMap map[int64]*entity.ServerPlanning, serverBaselineMap map[int64]*entity.ServerBaseline) (*entity.ServerCapPlanning, error) {
	var number, cpuNumber, memoryNumber int
	for _, v := range request.EcsCapacity.CapacityIdList {
		capConvertBaseline, capActualResBaseline, capServerCalcBaseline, _, serverPlanning, serverBaseline, err := handleCapCapacityParam(v, capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, nodeRoleBaselineMap, serverPlanningMap, serverBaselineMap)
		if err != nil {
			return nil, err
		}
		for _, ecs := range request.EcsCapacity.List {
			if capConvertBaseline.CapPlanningInput == "vCPU" {
				cpuNumber += CountServerNumber(ecs.CpuNumber*ecs.Count, request.EcsCapacity.FeatureNumber, capActualResBaseline, capServerCalcBaseline, serverBaseline)
			}
			if capConvertBaseline.CapPlanningInput == "内存" {
				memoryTotal := 138 + 8 + 16 + 8*ecs.CpuNumber*ecs.Count + ecs.MemoryNumber*ecs.Count/512
				if serverBaseline.Arch == "ARM" {
					memoryTotal += 128
				}
				memoryNumber += CountServerNumber(memoryTotal/1024, 1, capActualResBaseline, capServerCalcBaseline, serverBaseline)
			}
		}
		if cpuNumber > memoryNumber {
			number = cpuNumber
		} else {
			number = memoryNumber
		}
		serverPlanning.Number = number
		if err := db.Save(&serverPlanning).Error; err != nil {
			return nil, err
		}
	}
	return &entity.ServerCapPlanning{PlanId: request.PlanId, NodeRoleId: nodeRoleBaselineMap["COMPUTE"].Id, ProductCode: "ECS", Type: 2, Special: util.ToString(request.EcsCapacity)}, nil
}

//func CountServerSpecialNumber(capacityId int64, capActualResBaselineMap map[string]*entity.CapActualResBaseline, capServerCalcBaselineMap map[string]*entity.CapServerCalcBaseline, nodeRoleBaselineMap map[string]*entity.NodeRoleBaseline,
//	serverPlanningMap map[int64]*entity.ServerPlanning, serverBaselineMap map[int64]*entity.ServerBaseline) (*entity.CapConvertBaseline,
//	*entity.CapActualResBaseline, *entity.CapServerCalcBaseline, *entity.NodeRoleBaseline, *entity.ServerPlanning, *entity.ServerBaseline) {
//
//	capConvertBaseline, capActualResBaseline, capServerCalcBaseline, _, _, serverBaseline, err := handleCapCapacityParam(capacityId, capConvertBaselineMap, capActualResBaselineMap, capServerCalcBaselineMap, nodeRoleBaselineMap, serverPlanningMap, serverBaselineMap)
//	if err != nil {
//		return nil, err
//	}
//	for _, ecs := range request.EcsCapacity.List {
//		if capConvertBaseline.CapPlanningInput == "vCPU" {
//			cpuNumber += CountServerNumber(ecs.CpuNumber*ecs.Count, request.EcsCapacity.FeatureNumber, capActualResBaseline, capServerCalcBaseline, serverBaseline)
//		}
//		if capConvertBaseline.CapPlanningInput == "内存" {
//			memoryTotal := 138 + 8 + 16 + 8*ecs.CpuNumber*ecs.Count + ecs.MemoryNumber*ecs.Count/512
//			if serverBaseline.Arch == "ARM" {
//				memoryTotal += 128
//			}
//			memoryNumber += CountServerNumber(memoryTotal/1024, 1, capActualResBaseline, capServerCalcBaseline, serverBaseline)
//		}
//	}
//	if cpuNumber > memoryNumber {
//		number = cpuNumber
//	} else {
//		number = memoryNumber
//	}
//}

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
			if (util.IsBlank(request.NetworkInterface) || serverBaseline.NetworkInterface == request.NetworkInterface) && (util.IsBlank(request.CpuType) || serverBaseline.CpuType == request.CpuType) {
				screenNodeRoleServerBaselineListMap[k] = append(screenNodeRoleServerBaselineListMap[k], serverBaseline)
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
	if serverCapPlanningList == nil || len(serverCapPlanningList) == 0 {
		return nil, nil
	}
	versionId := serverCapPlanningList[0].VersionId
	var nodeRoleCapMap = make(map[int64]int)
	if len(serverCapPlanningList) == 0 {
		return nodeRoleCapMap, nil
	}
	var capPlanningNodeRoleIdList []int64
	var expendResCodeList []string
	var productCodeList []string
	var ecsCapacity = &EcsCapacity{}
	for _, v := range serverCapPlanningList {
		capPlanningNodeRoleIdList = append(capPlanningNodeRoleIdList, v.NodeRoleId)
		expendResCodeList = append(expendResCodeList, v.ExpendResCode)
		productCodeList = append(productCodeList, v.ProductCode)
		//单独处理ecs容量规划-按规格数量计算
		if util.IsNotBlank(v.Special) {
			util.ToObject(v.Special, &ecsCapacity)
			expendResCodeList = append(expendResCodeList, "ECS_VCPU", "ECS_MEM")
		}
	}
	//查询服务器和角色关联表
	var capPlanningNodeRoleRelList []*entity.ServerNodeRoleRel
	if err := db.Where("node_role_id IN (?)", capPlanningNodeRoleIdList).Find(&capPlanningNodeRoleRelList).Error; err != nil {
		return nil, err
	}
	//查询容量特性表
	var capActualResBaselineList []*entity.CapActualResBaseline
	if err := db.Where("product_code IN (?) AND expend_res_code IN (?) AND version_id = ?", productCodeList, expendResCodeList, versionId).Find(&capActualResBaselineList).Error; err != nil {
		return nil, err
	}
	var capActualResBaselineMap = make(map[string]*entity.CapActualResBaseline)
	for _, v := range capActualResBaselineList {
		capActualResBaselineMap[fmt.Sprintf("%v-%v", v.ProductCode, v.ExpendResCode)] = v
	}
	//查询容量计算表
	var capServerCalcBaselineList []*entity.CapServerCalcBaseline
	if err := db.Where("expend_res_code IN (?) AND version_id = ?", expendResCodeList, versionId).Find(&capServerCalcBaselineList).Error; err != nil {
		return nil, err
	}
	var capServerCalcBaselineMap = make(map[string]*entity.CapServerCalcBaseline)
	for _, v := range capServerCalcBaselineList {
		capServerCalcBaselineMap[v.ExpendResCode] = v
	}
	for _, v := range serverCapPlanningList {
		for _, serverBaseline := range nodeRoleServerBaselineListMap[v.NodeRoleId] {
			if serverBaseline.CpuType != request.CpuType {
				continue
			}
			key := fmt.Sprintf("%v-%v", v.ProductCode, v.ExpendResCode)
			capActualResBaseline := capActualResBaselineMap[key]
			capServerCalcBaseline := capServerCalcBaselineMap[v.ExpendResCode]
			var num int
			if capActualResBaseline != nil || capServerCalcBaseline != nil {
				num = CountServerNumber(v.Number, v.FeatureNumber, capActualResBaseline, capServerCalcBaseline, &entity.ServerBaseline{Cpu: serverBaseline.Cpu, Memory: serverBaseline.Memory, StorageDiskNum: serverBaseline.StorageDiskNum, StorageDiskCapacity: serverBaseline.StorageDiskCapacity})
			} else {
				//单独处理ecs容量规划-按规格数量计算
				var cpuNumber, memoryNumber int
				for _, ecs := range ecsCapacity.List {
					cpuNumber += CountServerNumber(ecs.CpuNumber*ecs.Count, ecsCapacity.FeatureNumber, capActualResBaselineMap[fmt.Sprintf("%v-%v", "ECS", "ECS_VCPU")], capServerCalcBaselineMap["ECS_VCPU"], &entity.ServerBaseline{Cpu: serverBaseline.Cpu, Memory: serverBaseline.Memory, StorageDiskNum: serverBaseline.StorageDiskNum, StorageDiskCapacity: serverBaseline.StorageDiskCapacity})
					memoryTotal := 138 + 8 + 16 + 8*ecs.CpuNumber*ecs.Count + ecs.MemoryNumber*ecs.Count/512
					if serverBaseline.Arch == "ARM" {
						memoryTotal += 128
					}
					memoryNumber += CountServerNumber(memoryTotal/1024, 1, capActualResBaselineMap[fmt.Sprintf("%v-%v", "ECS", "ECS_VCPU")], capServerCalcBaselineMap["ECS_MEM"], &entity.ServerBaseline{Cpu: serverBaseline.Cpu, Memory: serverBaseline.Memory, StorageDiskNum: serverBaseline.StorageDiskNum, StorageDiskCapacity: serverBaseline.StorageDiskCapacity})
				}
				if cpuNumber > memoryNumber {
					num = cpuNumber
				} else {
					num = memoryNumber
				}
			}
			if nodeRoleCapMap[v.NodeRoleId] < num {
				nodeRoleCapMap[v.NodeRoleId] = num
			}
		}
	}
	return nodeRoleCapMap, nil
}

func getServerShelveDownloadTemplate(planId int64) ([]ShelveDownload, string, error) {
	//查询服务器规划表
	serverPlanningList, err := getServerShelvePlanningList(planId)
	if err != nil {
		return nil, "", err
	}
	if len(serverPlanningList) == 0 {
		return nil, "", errors.New("服务器未规划")
	}
	//查询机柜
	cabinetIdleSlotList, err := getCabinetInfo(planId)
	if err != nil {
		return nil, "", err
	}
	//构建返回体
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
	//构建文件名称
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
	//查询服务器规划表
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
	//查询节点角色表
	var nodeRoleList []*entity.NodeRoleBaseline
	if err := data.DB.Where("id IN (?)", nodeRoleIdList).Find(&nodeRoleList).Error; err != nil {
		return nil, err
	}
	var nodeRoleNameMap = make(map[int64]string)
	for _, v := range nodeRoleList {
		nodeRoleNameMap[v.Id] = v.NodeRoleName
	}
	//查询服务器基线表
	var serverBaseline []*entity.ServerBaseline
	if err := data.DB.Where("id IN (?)", serverBaselineIdList).Find(&serverBaseline).Error; err != nil {
		return nil, err
	}
	var serverBaselineMap = make(map[int64]string)
	for _, v := range serverBaseline {
		serverBaselineMap[v.Id] = v.BomCode
	}
	//查询服务器上架表
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
	var cabinetResidualRackServerNumMap = make(map[int64]int)    //剩余上架服务器数
	var networkDeviceShelveCabinetIdMap = make(map[string]int64) //网络设备与机柜的关联信息
	for _, v := range cabinetInfoList {
		cabinetMap[v.Id] = v
		cabinetIdList = append(cabinetIdList, v.Id)
		cabinetResidualRackServerNumMap[v.Id] = v.ResidualRackServerNum
		networkDeviceShelveCabinetIdMap[fmt.Sprintf("%v-%v-%v-%v", v.CabinetAsw, v.MachineRoomAbbr, v.MachineRoomNum, v.CabinetNum)] = v.Id
	}
	//查询机柜槽位
	var cabinetIdleSlotRelList []*entity.CabinetIdleSlotRel
	if err := data.DB.Where("cabinet_id IN (?)", cabinetIdList).Order("cabinet_id ASC, idle_slot_number ASC").Find(&cabinetIdleSlotRelList).Error; err != nil {
		return nil, err
	}
	if len(cabinetIdleSlotRelList) == 0 {
		return nil, errors.New("机柜槽位为空，请检查机房勘察表是否填写错误")
	}
	//查询网络设备占用槽位
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
		//1号槽位不上架
		if v.IdleSlotNumber == 1 {
			continue
		}
		//过滤网络设备占用的槽位
		if networkDeviceShelveSlotPositionMap[v.CabinetId] != nil {
			if _, ok := networkDeviceShelveSlotPositionMap[v.CabinetId][v.IdleSlotNumber]; ok {
				continue
			}
		}
		if cabinetIdleSlotNumberMap[v.CabinetId] == 0 {
			cabinetIdleSlotNumberMap[v.CabinetId] = v.IdleSlotNumber
		} else {
			//两个槽位相邻，且不超过机柜的剩余上架服务器数
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
	//跨机柜上架，将所有机柜槽位列表纵向排布，然后从低到高横向上架
	var cabinetIdleSlotList []*Cabinet
	var maxLength = 1 //槽位最多的机柜的槽位数量
	var index = 0     //下标
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
	//查询服务器规划表
	var serverPlanning []*entity.ServerPlanning
	if err := data.DB.Where("plan_id = ?", planId).Find(&serverPlanning).Error; err != nil {
		return err
	}
	var NodeRoleIdList []int64
	for _, v := range serverPlanning {
		NodeRoleIdList = append(NodeRoleIdList, v.NodeRoleId)
	}
	//查询节点角色表
	var nodeRoleList []*entity.NodeRoleBaseline
	if err := data.DB.Where("id IN (?)", NodeRoleIdList).Find(&nodeRoleList).Error; err != nil {
		return err
	}
	var nodeRoleNameMap = make(map[string]int64)
	for _, v := range nodeRoleList {
		nodeRoleNameMap[v.NodeRoleName] = v.Id
	}
	//查询机柜信息
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
	//查询服务器上架表
	var cabinetIdList []int64
	if err := data.DB.Model(&entity.ServerShelve{}).Select("cabinet_id").Where("plan_id = ?", request.PlanId).Group("cabinet_id").Find(&cabinetIdList).Error; err != nil {
		return err
	}
	if len(cabinetIdList) == 0 {
		return errors.New("服务器未上架")
	}
	//查询机柜信息
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
	//查询节点角色表
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
	fileName := projectManage.Name + "-" + planManage.Name + "-" + "服务器上架清单"
	return response, fileName, nil
}

func checkBusiness(db *gorm.DB, request *Request) error {
	//校验planId
	var planCount int64
	if err := db.Model(&entity.PlanManage{}).Where("id = ? AND delete_state = ?", request.PlanId, 0).Count(&planCount).Error; err != nil {
		return err
	}
	if planCount == 0 {
		return errors.New("方案不存在")
	}
	return nil
}
