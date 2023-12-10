package plan

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"errors"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"
	"time"
)

func PagePlan(request *Request) ([]*entity.PlanManage, int64, error) {
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
		screenSql += " AND type = ? "
		screenParams = append(screenParams, request.Type)
	}
	if util.IsNotBlank(request.Stage) {
		screenSql += " AND stage = ? "
		screenParams = append(screenParams, request.Stage)
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
	var list []*entity.PlanManage
	if err := data.DB.Where(screenSql, screenParams...).Order(orderSql).Offset((request.Current - 1) * request.PageSize).Limit(request.PageSize).Find(&list).Error; err != nil {
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
		Name:         request.Name,
		ProjectId:    request.ProjectId,
		Type:         "general",
		Stage:        "plan",
		DeleteState:  0,
		CreateUserId: request.UserId,
		CreateTime:   now,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
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
	planEntity := &entity.PlanManage{
		Id:           request.Id,
		Type:         request.Type,
		Stage:        request.Stage,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
	}
	if err := data.DB.Updates(&planEntity).Error; err != nil {
		return err
	}
	return nil
}

func UpdatePlanStage(tx *gorm.DB, planId int64, stage string, userId string, planState int) error {
	if err := tx.Model(entity.PlanManage{}).Where("id = ?", planId).Updates(entity.PlanManage{
		Stage:            stage,
		BusinessPanStage: planState,
		UpdateUserId:     userId,
		UpdateTime:       time.Now(),
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
