package plan

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"errors"
)

func PagePlan(request *Request) ([]*entity.PlanManage, int64, error) {
	screenSql, screenParams, orderSql := " delete_state = ? AND project_id = ? ", []interface{}{0, request.ProjectId}, " update_time "
	switch request.SortField {
	case "createTime":
		orderSql = " create_time "
	case "updateTime":
		orderSql = " update_time "
	}
	switch request.Sort {
	case "asc":
		orderSql += " asc "
	case "desc":
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
