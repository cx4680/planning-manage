package plan

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"code.cestc.cn/ccos/cnm/ops-base/logging"
	"github.com/acmestack/godkits/gox/stringsx"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/excel"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/httpcall"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
)

func PagePlan(request *Request) ([]*Plan, int64, error) {
	screenSql, screenParams, orderSql := " delete_state = ? ", []interface{}{0}, " CASE WHEN type = 'standby' OR type = 'delivery' THEN 0 ELSE 1 END ASC "
	if request.ProjectId != 0 {
		screenSql += " AND project_id = ? "
		screenParams = append(screenParams, request.ProjectId)
	}
	if request.Id != 0 {
		screenSql += " AND id = ? "
		screenParams = append(screenParams, request.Id)
	}
	if util.IsNotBlank(request.Name) {
		screenSql += " AND name LIKE CONCAT('%',?,'%') "
		screenParams = append(screenParams, request.Name)
	}
	if util.IsNotBlank(request.Type) {
		typeSplit := strings.Split(request.Type, ",")
		screenSql += " AND type IN (?) "
		screenParams = append(screenParams, typeSplit)
	}
	if util.IsNotBlank(request.Stage) {
		stageSplit := strings.Split(request.Stage, ",")
		screenSql += " AND stage IN (?) "
		screenParams = append(screenParams, stageSplit)
	}
	switch request.SortField {
	case "createTime":
		orderSql += ", create_time "
	case "updateTime":
		orderSql += ", update_time "
	default:
		orderSql += ", update_time "
	}
	switch request.Sort {
	case "ascend":
		orderSql += " asc "
	case "descend":
		orderSql += " desc "
	default:
		orderSql += " desc "
	}
	var count int64
	if err := data.DB.Model(&entity.PlanManage{}).Where(screenSql, screenParams...).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	var list []*Plan
	if err := data.DB.Model(&entity.PlanManage{}).Where(screenSql, screenParams...).Order(orderSql).Offset((request.Current - 1) * request.PageSize).Limit(request.PageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	var alternativeCount int64
	if err := data.DB.Model(&entity.PlanManage{}).Where(" delete_state = ? AND project_id = ? AND type = ?", 0, request.ProjectId, "alternate").Count(&alternativeCount).Error; err != nil {
		return nil, 0, err
	}
	if alternativeCount > 0 {
		for i := range list {
			list[i].Alternative = 1
		}
	}
	return list, count, nil
}

func CreatePlan(request *Request) error {
	if err := checkBusiness(request, true); err != nil {
		return err
	}
	now := datetime.GetNow()
	planEntity := &entity.PlanManage{
		Name:              request.Name,
		ProjectId:         request.ProjectId,
		Type:              constant.General,
		Stage:             constant.PlanStagePlan,
		BusinessPlanStage: constant.BusinessPlanningStart,
		DeliverPlanStage:  constant.DeliverPlanningStart,
		DeleteState:       0,
		CreateUserId:      request.UserId,
		CreateTime:        now,
		UpdateUserId:      request.UserId,
		UpdateTime:        now,
	}
	if err := data.DB.Create(&planEntity).Error; err != nil {
		return err
	}
	return nil
}

func UpdatePlan(request *Request) error {
	if err := checkBusiness(request, false); err != nil {
		return err
	}
	now := datetime.GetNow()
	var planManage = &entity.PlanManage{}
	if err := data.DB.Where("id = ?", request.Id).Find(planManage).Error; err != nil {
		return err
	}
	if planManage.Id == 0 {
		return errors.New("方案不存在")
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Updates(&entity.PlanManage{
			Id:           request.Id,
			Name:         request.Name,
			Type:         request.Type,
			Stage:        request.Stage,
			UpdateUserId: request.UserId,
			UpdateTime:   now,
		}).Error; err != nil {
			return err
		}
		if request.Stage == constant.PlanStageDelivering {
			if err := tx.Updates(&entity.ProjectManage{Id: planManage.ProjectId, Stage: constant.ProjectStageDelivery, UpdateUserId: request.UserId, UpdateTime: now}).Error; err != nil {
				return err
			}
		}
		if request.Stage == constant.PlanStageDelivered {
			if err := tx.Updates(&entity.ProjectManage{Id: planManage.ProjectId, Stage: constant.ProjectStageDelivered, UpdateUserId: request.UserId, UpdateTime: now}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func UpdatePlanStage(tx *gorm.DB, planId int64, stage string, userId string, businessPlanStage, deliverPlanStage int) error {
	if err := tx.Model(entity.PlanManage{}).Where("id = ?", planId).Updates(entity.PlanManage{
		Stage:             stage,
		BusinessPlanStage: businessPlanStage,
		DeliverPlanStage:  deliverPlanStage,
		UpdateUserId:      userId,
		UpdateTime:        time.Now(),
	}).Error; err != nil {
		log.Errorf("[UpdatePlanStage] update plan stage error, %v", err)
		return err
	}
	return nil
}

func DeletePlan(request *Request) error {
	// 校验该项目下是否有方案
	var plan = &entity.PlanManage{}
	if err := data.DB.Model(&entity.PlanManage{}).Where("id = ? AND delete_state = ?", request.Id, 0).First(&plan).Error; err != nil {
		return err
	}
	if plan.Type != "general" {
		return errors.New("无法删除，该项目为备选方案")
	}
	now := datetime.GetNow()
	planEntity := &entity.PlanManage{
		Id:           request.Id,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  1,
	}
	if err := data.DB.Updates(&planEntity).Error; err != nil {
		return err
	}
	return nil
}

func checkBusiness(request *Request, isCreate bool) error {
	if isCreate {
		// 校验projectId
		var projectCount int64
		if err := data.DB.Model(&entity.ProjectManage{}).Where("id = ? AND delete_state = ?", request.ProjectId, 0).Count(&projectCount).Error; err != nil {
			return err
		}
		if projectCount == 0 {
			return errors.New("项目不存在")
		}
		var planCount int64
		if err := data.DB.Model(&entity.PlanManage{}).Where("project_id = ? AND delete_state = ?", request.ProjectId, 0).Count(&planCount).Error; err != nil {
			return err
		}
		if planCount >= constant.PlanMaxCount {
			return errors.New("方案数量不能超过5个")
		}
	} else {
		// 校验planId
		var planCount int64
		if err := data.DB.Model(&entity.PlanManage{}).Where("id = ? AND delete_state = ?", request.Id, 0).Count(&planCount).Error; err != nil {
			return err
		}
		if planCount == 0 {
			return errors.New("方案不存在")
		}
		// 校验planType
		if util.IsNotBlank(request.Type) {
			var planTypeCount int64
			if err := data.DB.Model(&entity.ConfigItem{}).Where("p_id = ? AND code = ?", "5", request.Type).Count(&planTypeCount).Error; err != nil {
				return err
			}
			if planTypeCount == 0 {
				return errors.New("type参数错误")
			}
		}
		// 校验planStage
		if util.IsNotBlank(request.Stage) {
			var planStageCount int64
			if err := data.DB.Model(&entity.ConfigItem{}).Where("p_id = ? AND code = ?", "6", request.Stage).Count(&planStageCount).Error; err != nil {
				return err
			}
			if planStageCount == 0 {
				return errors.New("stage参数错误")
			}
		}
	}
	return nil
}

func SendPlan(planId int64) (*SendBomsResponse, error) {
	var result []*SendBomsRequestStep

	// 查ProductConfigLibId
	var quotationId entity.PlanQuotationId

	if err := data.DB.Table(entity.PlanManageTable).Select("quotation_no").
		Joins("LEFT JOIN project_manage on plan_manage.project_id = project_manage.id").
		Where("plan_manage.id = ?", planId).
		Find(&quotationId).Error; err != nil {
		return nil, err
	}

	var versionId int64
	if err := data.DB.Table(entity.CloudProductPlanningTable).Select("version_id").
		Where("plan_id = ?", planId).
		Limit(1).
		Find(&versionId).Error; err != nil {
		return nil, err
	}

	stepProductRequest := SendBomsRequestStep{
		StepName: "Step1 - 云产品配置",
	}
	productFeatures, err := buildCloudProductFeatures(planId, versionId)
	if err != nil {
		return nil, err
	}
	if len(productFeatures) > 0 {
		stepProductRequest.Features = productFeatures
		result = append(result, &stepProductRequest)
	}

	stepServerRequest := SendBomsRequestStep{
		StepName: "Step2 - 服务器规划",
	}
	serverFeatures, err := buildServerFeatures(planId, versionId)
	if err != nil {
		return nil, err
	}
	if len(serverFeatures) > 0 {
		stepServerRequest.Features = serverFeatures
		result = append(result, &stepServerRequest)
	}

	stepNetworkRequest := SendBomsRequestStep{
		StepName: "Step3 - 网络设备规划",
	}
	netDeviceFeatures, err := buildNetDeviceFeatures(planId)
	if err != nil {
		return nil, err
	}
	if len(netDeviceFeatures) > 0 {
		stepNetworkRequest.Features = netDeviceFeatures
		result = append(result, &stepNetworkRequest)
	}

	// POST
	request := SendBomsRequest{
		ProductConfigLibId: quotationId.QuotationNo,
		Steps:              result,
	}
	reqJson, err := json.Marshal(request)
	logging.Infof("Send Bom Req: %v", string(reqJson))
	// 正式环境: https://cbs.cestc.cn
	// 测试环境: http://bom.cestcdev.cn
	// 预演环境: http://cbs.cestcpre.cn
	bomUrl := stringsx.DefaultIfEmpty(os.Getenv("BOM_URL"), "http://bom.cestcdev.cn")
	url := bomUrl + "/api/v1/product/receiveConfig"
	response, err := httpcall.POSTResponse(httpcall.HttpRequest{
		Context: &gin.Context{},
		URI:     url,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: bytes.NewBuffer(reqJson),
	})
	if err != nil {
		log.Errorf("Do Post Err %v", err)
	}
	log.Infof("Send Plan Resp: %v", response)
	resByte, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Marshal json error: %v", err)
	}
	var responseData SendBomsResponse
	if err = json.Unmarshal(resByte, &responseData); err != nil {
		log.Errorf("Unmarshal resp json error: %v", err)
		return nil, err
	}

	return &responseData, nil
}

func buildCloudProductFeatures(planId int64, versionId int64) ([]*SendBomsRequestFeature, error) {
	featureMap := make(map[string]*SendBomsRequestFeature)
	var planList []entity.SoftwareBomPlanning
	if err := data.DB.Model(&entity.SoftwareBomPlanning{}).Where("plan_id = ?", planId).Find(&planList).Error; err != nil {
		return nil, err
	}

	// var softwareBomBaselineList []*entity.SoftwareBomLicenseBaseline
	// if err := data.DB.Where("version_id = ?", versionId).Find(&softwareBomBaselineList).Error; err != nil {
	// 	return nil, err
	// }
	// softwareBomBaselineMap := make(map[int64]*entity.SoftwareBomLicenseBaseline)
	// for _, softwareBomLicenseBaseline := range softwareBomBaselineList {
	// 	softwareBomBaselineMap[softwareBomLicenseBaseline.Id] = softwareBomLicenseBaseline
	// }

	var featureNameCodeRelList []*entity.FeatureNameCodeRel
	if err := data.DB.Where("feature_type = ?", constant.FeatureTypeCloudProduct).Find(&featureNameCodeRelList).Error; err != nil {
		return nil, err
	}
	featureNameCodeRelMap := make(map[string]*entity.FeatureNameCodeRel)
	for _, featureNameCodeRel := range featureNameCodeRelList {
		featureNameCodeRelMap[featureNameCodeRel.FeatureName] = featureNameCodeRel
	}

	var cloudProductBaselineList []*entity.CloudProductBaseline
	if err := data.DB.Where("version_id = ?", versionId).Find(&cloudProductBaselineList).Error; err != nil {
		return nil, err
	}
	productCodeFeatureNameCodeRelMap := make(map[string]*entity.FeatureNameCodeRel)
	for _, cloudProductBaseline := range cloudProductBaselineList {
		if featureNameCodeRel, ok := featureNameCodeRelMap[cloudProductBaseline.ProductType]; ok {
			productCodeFeatureNameCodeRelMap[cloudProductBaseline.ProductCode] = featureNameCodeRel
		}
	}

	for _, plan := range planList {
		if plan.BomId == "" {
			continue
		}
		// _, ok := softwareBomBaselineMap[plan.SoftwareBaselineId]
		// if !ok {
		// 	return nil, errors.New("软件BOM不存在")
		// }

		// 如果没查到所属类别，用自己的ServiceCode作为featureCode
		featureCode := plan.ServiceCode
		featureName := plan.CloudService
		if featureNameCodeRel, ok := productCodeFeatureNameCodeRelMap[plan.ServiceCode]; ok {
			featureCode = featureNameCodeRel.FeatureCode
			featureName = featureNameCodeRel.FeatureName
		}

		// 拼request先用字典
		if _, flag := featureMap[featureCode]; flag {
			featureMap[featureCode].Boms = append(featureMap[featureCode].Boms, &SendBomsRequestBom{
				Code:  plan.BomId,
				Count: plan.Number,
			})
		} else {
			featureMap[featureCode] = &SendBomsRequestFeature{
				FeatureCode: featureCode,
				FeatureName: featureName,
				Boms: []*SendBomsRequestBom{
					{
						Code:  plan.BomId,
						Count: plan.Number,
					},
				},
			}
		}
	}
	var result []*SendBomsRequestFeature
	for _, v := range featureMap {
		result = append(result, v)
	}
	return result, nil
}

func buildServerFeatures(planId int64, versionId int64) ([]*SendBomsRequestFeature, error) {
	featureMap := make(map[string]*SendBomsRequestFeature)

	// TODO 暂时不需要同步存储设备到BOM，后期要做的时候再放开
	var planList []entity.ServerPlanningSelect
	if err := data.DB.Table(entity.ServerPlanningTable+" sp").
		Select("sp.plan_id, sp.server_baseline_id, SUM(sp.number) as number, nrb.classify").
		Joins("LEFT JOIN node_role_baseline nrb on sp.node_role_id = nrb.id ").
		Where("sp.plan_id = ? and sp.delete_state = 0 and nrb.version_id = ? and nrb.node_role_code not in (?)", planId, versionId, []string{constant.NodeRoleCodeEBS, constant.NodeRoleCodeEFS, constant.NodeRoleCodeOSS}).
		Group("sp.plan_id, sp.server_baseline_id, nrb.classify").
		Find(&planList).Error; err != nil {
		return nil, err
	}

	var serverBaselineList []*entity.ServerBaseline
	if err := data.DB.Where("version_id = ?", versionId).Find(&serverBaselineList).Error; err != nil {
		return nil, err
	}
	serverBaselineMap := make(map[int64]*entity.ServerBaseline)
	for _, serverBaseline := range serverBaselineList {
		serverBaselineMap[serverBaseline.Id] = serverBaseline
	}

	var featureNameCodeRelList []*entity.FeatureNameCodeRel
	if err := data.DB.Where("feature_type = ?", constant.FeatureTypeServerNode).Find(&featureNameCodeRelList).Error; err != nil {
		return nil, err
	}
	featureNameCodeRelMap := make(map[string]*entity.FeatureNameCodeRel)
	for _, featureNameCodeRel := range featureNameCodeRelList {
		featureNameCodeRelMap[featureNameCodeRel.FeatureName] = featureNameCodeRel
	}

	for _, plan := range planList {
		serverBaseLine, ok := serverBaselineMap[plan.ServerBaselineId]
		if !ok {
			return nil, errors.New("server baseline not found")
		}

		if serverBaseLine.BomCode == "" {
			continue
		}

		featureNameCodeRel, ok := featureNameCodeRelMap[plan.Classify]
		if !ok {
			return nil, errors.New("feature name code rel not found")
		}

		if _, flag := featureMap[featureNameCodeRel.FeatureCode]; flag {
			featureMap[featureNameCodeRel.FeatureCode].Boms = append(featureMap[featureNameCodeRel.FeatureCode].Boms, &SendBomsRequestBom{
				Code:  serverBaseLine.BomCode,
				Count: float64(plan.Number),
			})
		} else {
			featureMap[featureNameCodeRel.FeatureCode] = &SendBomsRequestFeature{
				FeatureCode: featureNameCodeRel.FeatureCode,
				FeatureName: featureNameCodeRel.FeatureName,
				Boms: []*SendBomsRequestBom{
					{
						Code:  serverBaseLine.BomCode,
						Count: float64(plan.Number),
					},
				},
			}
		}
	}
	var result []*SendBomsRequestFeature
	for _, v := range featureMap {
		result = append(result, v)
	}
	return result, nil
}

func buildNetDeviceFeatures(planId int64) ([]*SendBomsRequestFeature, error) {
	featureMap := make(map[string]*SendBomsRequestFeature)
	var planList []entity.NetworkDeviceSelect
	if err := data.DB.Table(entity.NetworkDeviceListTable).Select("plan_id, network_device_role, network_device_role_name, bom_id, count(*) as number").
		Where("delete_state = 0 and plan_id = ? ", planId).Group("plan_id, network_device_role, bom_id, network_device_role_name").
		Find(&planList).Error; err != nil {
		return nil, err
	}

	for _, plan := range planList {

		if plan.BomId == "" {
			continue
		}

		if _, flag := featureMap[plan.NetworkDeviceRole]; flag {
			featureMap[plan.NetworkDeviceRole].Boms = append(featureMap[plan.NetworkDeviceRole].Boms, &SendBomsRequestBom{
				Code:  plan.BomId,
				Count: float64(plan.Number),
			})
		} else {
			featureMap[plan.NetworkDeviceRole] = &SendBomsRequestFeature{
				FeatureCode: plan.NetworkDeviceRole,
				FeatureName: plan.NetworkDeviceRoleName,
				Boms: []*SendBomsRequestBom{
					{
						Code:  plan.BomId,
						Count: float64(plan.Number),
					},
				},
			}
		}
	}
	var result []*SendBomsRequestFeature
	for _, v := range featureMap {
		result = append(result, v)
	}
	return result, nil
}

func CopyPlan(request *Request) error {
	now := datetime.GetNow()
	var planManage = &entity.PlanManage{}
	if err := data.DB.Where("id = ? AND delete_state = ?", request.Id, 0).Find(planManage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("方案不存在")
		}
		return err
	}
	if planManage.Type == constant.Delivery {
		return errors.New("只能复制售前方案")
	}
	if planManage.Type == constant.Alternate {
		planManage.Type = constant.General
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		planManage.Id = 0
		planManage.Name = planManage.Name + constant.CopyPlanEndOfName
		planManage.CreateUserId = request.UserId
		planManage.UpdateUserId = request.UserId
		planManage.CreateTime = now
		planManage.UpdateTime = now
		if err := tx.Create(&planManage).Error; err != nil {
			return err
		}
		// 复制云产品规划数据
		var cloudProductPlannings []*entity.CloudProductPlanning
		if err := tx.Where("plan_id = ?", request.Id).Find(&cloudProductPlannings).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil
			}
			return err
		}
		for i := range cloudProductPlannings {
			cloudProductPlannings[i].Id = 0
			cloudProductPlannings[i].PlanId = planManage.Id
			cloudProductPlannings[i].CreateTime = now
			cloudProductPlannings[i].UpdateTime = now
		}
		if err := tx.Create(&cloudProductPlannings).Error; err != nil {
			return err
		}
		// 复制资源池数据
		var resourcePoolList []*entity.ResourcePool
		if err := tx.Where("plan_id = ?", request.Id).Find(&resourcePoolList).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		originResourceIdNewMap := make(map[int64]int64)
		if len(resourcePoolList) > 0 {
			for _, resourcePool := range resourcePoolList {
				originResourcePoolId := resourcePool.Id
				resourcePool.Id = 0
				resourcePool.PlanId = planManage.Id
				if err := tx.Create(&resourcePool).Error; err != nil {
					return err
				}
				originResourceIdNewMap[originResourcePoolId] = resourcePool.Id
			}
		}
		// 复制服务器规划数据
		var serverPlannings []*entity.ServerPlanning
		// 这里没像云产品规划那样判断，如果数据为空就不走下面的逻辑的原因：在服务器规划页面只操作了容量规划，而没有点击下一步到网络设备规划，这样就不会将服务器规划数据保存到数据库了
		if err := tx.Where("plan_id = ? and delete_state = ?", request.Id, 0).Find(&serverPlannings).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if len(serverPlannings) > 0 {
			for i := range serverPlannings {
				serverPlannings[i].Id = 0
				serverPlannings[i].PlanId = planManage.Id
				serverPlannings[i].ResourcePoolId = originResourceIdNewMap[serverPlannings[i].ResourcePoolId]
				serverPlannings[i].CreateUserId = request.UserId
				serverPlannings[i].UpdateUserId = request.UserId
				serverPlannings[i].CreateTime = now
				serverPlannings[i].UpdateTime = now
			}
			if err := tx.Create(&serverPlannings).Error; err != nil {
				return err
			}
		}
		// 复制容量规划数据
		var serverCapPlannings []*entity.ServerCapPlanning
		if err := tx.Where("plan_id = ?", request.Id).Find(&serverCapPlannings).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if len(serverCapPlannings) > 0 {
			for i := range serverCapPlannings {
				serverCapPlannings[i].Id = 0
				serverCapPlannings[i].PlanId = planManage.Id
				serverCapPlannings[i].ResourcePoolId = originResourceIdNewMap[serverCapPlannings[i].ResourcePoolId]
			}
			if err := tx.Create(&serverCapPlannings).Error; err != nil {
				return err
			}
		}
		// 复制网络设备规划数据
		var networkDevicePlannings []*entity.NetworkDevicePlanning
		if err := tx.Where("plan_id = ?", request.Id).Find(&networkDevicePlannings).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if len(networkDevicePlannings) > 0 {
			for i := range networkDevicePlannings {
				networkDevicePlannings[i].Id = 0
				networkDevicePlannings[i].PlanId = planManage.Id
				networkDevicePlannings[i].CreateTime = now
				networkDevicePlannings[i].UpdateTime = now
			}
			if err := tx.Create(&networkDevicePlannings).Error; err != nil {
				return err
			}
		}
		// 复制网络设备清单数据
		var networkDeviceList []*entity.NetworkDeviceList
		if err := tx.Where("plan_id = ? and delete_state = ?", request.Id, 0).Find(&networkDeviceList).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if len(networkDeviceList) > 0 {
			for i := range networkDeviceList {
				networkDeviceList[i].Id = 0
				networkDeviceList[i].PlanId = planManage.Id
				networkDeviceList[i].CreateTime = now
				networkDeviceList[i].UpdateTime = now
			}
			if err := tx.Create(&networkDeviceList).Error; err != nil {
				return err
			}
		}
		// 复制IP需求规划数据
		var ipDemandPlannings []*entity.IpDemandPlanning
		if err := tx.Where("plan_id = ?", request.Id).Find(&ipDemandPlannings).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if len(ipDemandPlannings) > 0 {
			for i := range ipDemandPlannings {
				ipDemandPlannings[i].Id = 0
				ipDemandPlannings[i].PlanId = planManage.Id
				ipDemandPlannings[i].CreateTime = now
				ipDemandPlannings[i].UpdateTime = now
			}
			if err := tx.Create(&ipDemandPlannings).Error; err != nil {
				return err
			}
		}
		// 复制软件BOM规划数据
		var softwareBomPlannings []*entity.SoftwareBomPlanning
		if err := tx.Where("plan_id = ?", request.Id).Find(&softwareBomPlannings).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if len(softwareBomPlannings) > 0 {
			for i := range softwareBomPlannings {
				softwareBomPlannings[i].Id = 0
				softwareBomPlannings[i].PlanId = planManage.Id
			}
			if err := tx.Create(&softwareBomPlannings).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func DownloadPlanningConfigChecklist(planId int64) ([]excel.ExportSheet, string, error) {
	// 构建返回体
	var response []excel.ExportSheet
	var versionId int64
	// 查询云产品清单
	var cloudProductResponse []CloudProductPlanningExportResponse
	if err := data.DB.Table("cloud_product_planning cpp").Select("cpb.product_type,cpb.product_name, cpp.sell_spec, cpp.value_added_service, cpb.instructions").
		Joins("LEFT JOIN cloud_product_baseline cpb ON cpb.id = cpp.product_id").
		Where("cpp.plan_id=?", planId).
		Find(&cloudProductResponse).Error; err != nil {
		log.Errorf("[downloadPlanningConfigChecklist] query db error")
		return nil, "", err
	}
	response = append(response, excel.ExportSheet{SheetName: "云产品清单", Data: cloudProductResponse})

	// 查询服务器规划清单
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
	versionId = nodeRoleList[0].VersionId
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
	var serverResponse []ResponseDownloadServer
	var total int
	for _, v := range serverList {
		serverResponse = append(serverResponse, ResponseDownloadServer{
			NodeRole:   nodeRoleMap[v.NodeRoleId].NodeRoleName,
			ServerType: serverBaselineMap[v.ServerBaselineId].Arch,
			BomCode:    serverBaselineMap[v.ServerBaselineId].BomCode,
			Spec:       serverBaselineMap[v.ServerBaselineId].ConfigurationInfo,
			Number:     strconv.Itoa(v.Number),
		})
		total += v.Number
	}
	serverResponse = append(serverResponse, ResponseDownloadServer{
		Number: "总计：" + strconv.Itoa(total) + "台",
	})
	response = append(response, excel.ExportSheet{SheetName: "服务器规划清单", Data: serverResponse})

	// 查询网络设备规划清单
	var roleIdNum []NetworkDeviceRoleIdNum
	if err := data.DB.Table(entity.NetworkDeviceListTable).Select("network_device_role_id", "count(*) as num").
		Where("plan_id = ? AND delete_state = 0", planId).Group("network_device_role_id").Find(&roleIdNum).Error; err != nil {
		log.Errorf("[downloadPlanningConfigChecklist] query db error, %v", err)
		return nil, "", err
	}
	if len(roleIdNum) == 0 {
		return nil, "", errors.New("获取网络设备清单为空")
	}
	total = 0
	var networkDeviceResponse []NetworkDeviceListExportResponse
	for _, roleNum := range roleIdNum {
		roleId := roleNum.NetworkDeviceRoleId
		var networkDevice entity.NetworkDeviceList
		if err := data.DB.Where("plan_id = ? AND delete_state = 0 AND network_device_role_id = ?", planId, roleId).Find(&networkDevice).Error; err != nil {
			log.Errorf("[getNetworkDeviceListByPlanIdAndRoleId] query db error, %v", err)
			return nil, "", err
		}
		total += roleNum.Num
		networkDeviceResponse = append(networkDeviceResponse, NetworkDeviceListExportResponse{
			NetworkDeviceRoleName: networkDevice.NetworkDeviceRoleName,
			NetworkDeviceRole:     networkDevice.NetworkDeviceRole,
			Brand:                 networkDevice.Brand,
			DeviceModel:           networkDevice.DeviceModel,
			ConfOverview:          networkDevice.ConfOverview,
			Num:                   strconv.Itoa(roleNum.Num),
		})
	}
	// 手动添加合计行
	networkDeviceResponse = append(networkDeviceResponse, NetworkDeviceListExportResponse{
		Num: "总计:" + strconv.Itoa(total) + "台",
	})
	response = append(response, excel.ExportSheet{SheetName: "网络设备清单", Data: networkDeviceResponse})

	// 查询BOM清单
	var softwareBomPlannings []*entity.SoftwareBomPlanning
	if err := data.DB.Where("plan_id = ?", planId).Order("software_baseline_id, id").Find(&softwareBomPlannings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, "", errors.New("获取BOM清单为空")
		}
		return nil, "", err
	}
	var cloudProductBaselines []*entity.CloudProductBaseline
	if err := data.DB.Where("version_id = ?", versionId).Find(&cloudProductBaselines).Error; err != nil {
		return nil, "", err
	}
	cloudProductBaselineMap := make(map[string]*entity.CloudProductBaseline)
	for _, cloudProductBaseline := range cloudProductBaselines {
		cloudProductBaselineMap[cloudProductBaseline.ProductCode] = cloudProductBaseline
	}
	var bomListDownload []BomListDownload
	for _, softwareBomPlanning := range softwareBomPlannings {
		var category string
		cloudProductBaseline := cloudProductBaselineMap[softwareBomPlanning.ServiceCode]
		if cloudProductBaseline != nil {
			category = cloudProductBaseline.ProductType
		}
		bomListDownload = append(bomListDownload, BomListDownload{
			Category:       category,
			CloudProduct:   softwareBomPlanning.CloudService,
			SellType:       softwareBomPlanning.SellType,
			BomId:          softwareBomPlanning.BomId,
			AuthorizedUnit: softwareBomPlanning.AuthorizedUnit,
			Number:         strconv.FormatFloat(softwareBomPlanning.Number, 'f', -1, 64),
		})
	}
	response = append(response, excel.ExportSheet{SheetName: "BOM清单", Data: bomListDownload})

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
	fileName := projectManage.Name + "-" + planManage.Name + "-" + "规划配置清单"
	return response, fileName, nil
}
