package plan

import (
	"bytes"
	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/httpcall"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"
	"strings"
	"time"
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
	//校验该项目下是否有方案
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
		//校验projectId
		var projectCount int64
		if err := data.DB.Model(&entity.ProjectManage{}).Where("id = ? AND delete_state = ?", request.ProjectId, 0).Count(&projectCount).Error; err != nil {
			return err
		}
		if projectCount == 0 {
			return errors.New("项目不存在")
		}
	} else {
		//校验planId
		var planCount int64
		if err := data.DB.Model(&entity.PlanManage{}).Where("id = ? AND delete_state = ?", request.Id, 0).Count(&planCount).Error; err != nil {
			return err
		}
		if planCount == 0 {
			return errors.New("方案不存在")
		}
		//校验planType
		if util.IsNotBlank(request.Type) {
			var planTypeCount int64
			if err := data.DB.Model(&entity.ConfigItem{}).Where("p_id = ? AND code = ?", "5", request.Type).Count(&planTypeCount).Error; err != nil {
				return err
			}
			if planTypeCount == 0 {
				return errors.New("type参数错误")
			}
		}
		//校验planStage
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

func SendPlan(planId int64) (*SendBomsRequest, error) {
	var result []*SendBomsRequestStep

	// 查ProductConfigLibId
	var quotationId string
	if err := data.DB.Table("plan_manage").Select("quotation_no").
		Joins("LEFT JOIN project_manage on plan_manage.project_id = project_manage.id").
		Where("plan_manage.id = ?", planId).
		Find(&quotationId).Error; err != nil {
		return nil, err
	}

	stepProductRequest := SendBomsRequestStep{
		StepName: "Step1 - Cloud Product",
	}
	productFeatures, err := buildCloudProductFeatures(planId)
	if err != nil {
		return nil, err
	}
	stepProductRequest.Features = productFeatures
	result = append(result, &stepProductRequest)

	stepServerRequest := SendBomsRequestStep{
		StepName: "Step2 - Server",
	}
	serverFeatures, err := buildServerFeatures(planId)
	if err != nil {
		return nil, err
	}
	stepServerRequest.Features = serverFeatures
	result = append(result, &stepServerRequest)

	stepNetworkRequest := SendBomsRequestStep{
		StepName: "Step3 - Network Device",
	}
	netDeviceFeatures, err := buildNetDeviceFeatures(planId)
	if err != nil {
		return nil, err
	}
	stepNetworkRequest.Features = netDeviceFeatures
	result = append(result, &stepNetworkRequest)

	// POST
	request := SendBomsRequest{
		ProductConfigLibId: quotationId,
		Steps:              result,
	}
	reqJson, err := json.Marshal(request)
	// 正式环境: https://cbs.cestc.cn
	// 测试环境: http://bom.cestcdev.cn
	// 预演环境: http://cbs.cestcpre.cn
	url := "http://bom.cestcdev.cn" + "/api/v1/product/receiveConfig"
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

	return &request, nil
}

func buildCloudProductFeatures(planId int64) ([]*SendBomsRequestFeature, error) {
	featureMap := make(map[string]*SendBomsRequestFeature)
	var planList []entity.SoftwareBomPlanning
	if err := data.DB.Model(&entity.SoftwareBomPlanning{}).Where("plan_id = ?", planId).Find(&planList).Error; err != nil {
		return nil, err
	}
	for _, plan := range planList {
		// bomId
		var baseLine entity.SoftwareBomLicenseBaseline
		if err := data.DB.Model(&entity.SoftwareBomLicenseBaseline{}).Where("id = ?", plan.SoftwareBaselineId).Find(&baseLine).Error; err != nil {
			return nil, err
		}

		if plan.BomId == "" {
			continue
		}

		// feature code/name
		var featureInfo entity.FeatureNameCodeRel
		if err := data.DB.Table("cloud_product_baseline cpb").Select("f.feature_name, f.feature_code").
			Joins("left join feature_name_code_rel f on cpb.product_type = f.feature_name").
			Where("cpb.product_code = ?", plan.ServiceCode).
			Find(&featureInfo).Error; err != nil {
			return nil, err
		}

		// 拼request先用字典
		if _, flag := featureMap[featureInfo.FeatureCode]; flag {
			featureMap[featureInfo.FeatureCode].Boms = append(featureMap[featureInfo.FeatureCode].Boms, &SendBomsRequestBom{
				Code:  plan.BomId,
				Count: plan.Number,
			})
		} else {
			featureMap[featureInfo.FeatureCode] = &SendBomsRequestFeature{
				FeatureCode: featureInfo.FeatureCode,
				FeatureName: featureInfo.FeatureName,
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

func buildServerFeatures(planId int64) ([]*SendBomsRequestFeature, error) {
	featureMap := make(map[string]*SendBomsRequestFeature)
	var planList []entity.ServerPlanning
	if err := data.DB.Model(&entity.ServerPlanning{}).Where("plan_id = ? and delete_state = 0", planId).Find(&planList).Error; err != nil {
		return nil, err
	}
	for _, plan := range planList {
		// bomId
		var baseLine entity.ServerBaseline
		if err := data.DB.Model(&entity.ServerBaseline{}).Where("id = ? ", plan.ServerBaselineId).Find(&baseLine).Error; err != nil {
			return nil, err
		}

		if baseLine.BomCode == "" {
			continue
		}

		// feature code/name
		var featureInfo entity.FeatureNameCodeRel
		if err := data.DB.Table("node_role_baseline").Select("f.feature_name, f.feature_code").
			Joins("LEFT JOIN feature_name_code_rel f on node_role_baseline.classify = f.feature_name").
			Where("node_role_baseline.id = ?", plan.NodeRoleId).
			Find(&featureInfo).Error; err != nil {
			return nil, err
		}

		if _, flag := featureMap[featureInfo.FeatureCode]; flag {
			featureMap[featureInfo.FeatureCode].Boms = append(featureMap[featureInfo.FeatureCode].Boms, &SendBomsRequestBom{
				Code:  baseLine.BomCode,
				Count: plan.Number,
			})
		} else {
			featureMap[featureInfo.FeatureCode] = &SendBomsRequestFeature{
				FeatureCode: featureInfo.FeatureCode,
				FeatureName: featureInfo.FeatureName,
				Boms: []*SendBomsRequestBom{
					{
						Code:  baseLine.BomCode,
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

func buildNetDeviceFeatures(planId int64) ([]*SendBomsRequestFeature, error) {
	featureMap := make(map[string]*SendBomsRequestFeature)
	var planList []entity.NetworkDeviceSelect
	if err := data.DB.Table("network_device_list").Select("plan_id, network_device_role, network_device_role_name, bom_id, count(*) as number").
		Where("delete_state = 0 and plan_id = ? ", planId).Group("plan_id, network_device_role, bom_id, network_device_role_name").
		Find(&planList).Error; err != nil {
		return nil, err
	}

	for _, plan := range planList {

		//if plan.BomId == ""{
		//	continue
		//}

		if _, flag := featureMap[plan.NetworkDeviceRole]; flag {
			featureMap[plan.NetworkDeviceRole].Boms = append(featureMap[plan.NetworkDeviceRole].Boms, &SendBomsRequestBom{
				Code:  plan.BomId,
				Count: plan.Number,
			})
		} else {
			featureMap[plan.NetworkDeviceRole] = &SendBomsRequestFeature{
				FeatureCode: plan.NetworkDeviceRole,
				FeatureName: plan.NetworkDeviceRoleName,
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
