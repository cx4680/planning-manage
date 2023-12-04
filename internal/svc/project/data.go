package project

import (
	"code.cestc.cn/zhangzhi/planning-manage/internal/data"
	"code.cestc.cn/zhangzhi/planning-manage/internal/entity"
	"code.cestc.cn/zhangzhi/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/zhangzhi/planning-manage/internal/pkg/util"
	"errors"
)

func PageProject(request *Request) ([]*entity.ProjectManage, int64, error) {
	screenSql, screenParams, orderSql := " delete_state = ? AND id customer_id = ? ", []interface{}{0, request.CustomerId}, " update_time "
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
	if err := data.DB.Model(&entity.ProjectManage{}).Where(screenSql, screenParams...).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	var list []*entity.ProjectManage
	if err := data.DB.Where(screenSql, screenParams...).Order(orderSql).Offset((request.Current - 1) * request.PageSize).Limit(request.PageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, count, nil
}

func CreateProject(request *Request) error {
	if err := checkBusiness(request, true); err != nil {
		return err
	}
	now := datetime.GetNow()
	projectEntity := &entity.ProjectManage{
		Name:         request.Name,
		RegionId:     request.RegionId,
		AzId:         request.AzId,
		CellId:       request.CellId,
		CustomerId:   request.CustomerId,
		Type:         request.Type,
		Stage:        "planning",
		DeleteState:  0,
		CreateUserId: request.UserId,
		CreateTime:   now,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
	}
	if err := data.DB.Create(&projectEntity).Error; err != nil {
		return err
	}
	return nil
}

func UpdateProject(request *Request) error {
	if err := checkBusiness(request, false); err != nil {
		return err
	}
	now := datetime.GetNow()
	projectEntity := &entity.ProjectManage{
		Id:           request.Id,
		Name:         request.Name,
		Stage:        request.Stage,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
	}
	if err := data.DB.Updates(&projectEntity).Error; err != nil {
		return err
	}
	return nil
}

func DeleteProject(request *Request) error {
	//校验该项目下是否有方案
	var planCount int64
	if err := data.DB.Model(&entity.PlanManage{}).Where("project_id = ?", request.Id).Count(&planCount).Error; err != nil {
		return err
	}
	if planCount != 0 {
		return errors.New("无法删除，该项目有方案未删除")
	}
	now := datetime.GetNow()
	projectEntity := &entity.ProjectManage{
		Id:           request.Id,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  1,
	}
	if err := data.DB.Updates(&projectEntity).Error; err != nil {
		return err
	}
	return nil
}

func checkBusiness(request *Request, isCreate bool) error {
	if isCreate {
		//校验projectType
		var projectTypeCount int64
		if err := data.DB.Model(&entity.ConfigItem{}).Where("p_id = ? AND code = ?", "3", request.Type).Count(&projectTypeCount).Error; err != nil {
			return err
		}
		if projectTypeCount == 0 {
			return errors.New("type参数错误")
		}
		//校验cloudPlatform
		var cloudPlatform = &entity.CloudPlatformManage{}
		if err := data.DB.Where("id = ? AND delete_state = ?", request.CloudPlatformId, 0).Find(&cloudPlatform).Error; err != nil {
			return err
		}
		if cloudPlatform.Id == 0 {
			return errors.New("cloudPlatformId参数错误")
		}
		if util.IsBlank(cloudPlatform.Type) && util.IsBlank(request.CloudPlatformType) {
			return errors.New("云平台类型未设置")
		}
		//校验regionId
		var regionCount int64
		if err := data.DB.Model(&entity.RegionManage{}).Where("id = ? AND delete_state = ?", request.RegionId, 0).Count(&regionCount).Error; err != nil {
			return err
		}
		if regionCount == 0 {
			return errors.New("region不存在")
		}
		//校验azId
		var azCount int64
		if err := data.DB.Model(&entity.AzManage{}).Where("id = ? AND delete_state = ?", request.AzId, 0).Count(&azCount).Error; err != nil {
			return err
		}
		if azCount == 0 {
			return errors.New("az不存在")
		}
	} else {
		//校验projectId
		var projectCount int64
		if err := data.DB.Model(&entity.ProjectManage{}).Where("id = ? AND delete_state = ?", request.Id, 0).Count(&projectCount).Error; err != nil {
			return err
		}
		if projectCount == 0 {
			return errors.New("project不存在")
		}
		//校验projectStage
		if util.IsNotBlank(request.Stage) {
			var projectStageCount int64
			if err := data.DB.Model(&entity.ConfigItem{}).Where("p_id = ? AND code = ?", "4", request.Stage).Count(&projectStageCount).Error; err != nil {
				return err
			}
			if projectStageCount == 0 {
				return errors.New("stage参数错误")
			}
		}
	}
	return nil
}
