package region

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"errors"
)

func ListRegion(request *Request) ([]*entity.RegionManage, error) {
	screenSql, screenParams, orderSql := " delete_state = ? AND cloud_platform_id = ? ", []interface{}{0, request.CloudPlatformId}, " create_time "
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
		orderSql += " asc "
	}
	var list []*entity.RegionManage
	if err := data.DB.Model(&entity.RegionManage{}).Where(screenSql, screenParams...).Order(orderSql).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func CreateRegion(request *Request) error {
	if err := checkBusiness(request, true); err != nil {
		return err
	}
	now := datetime.GetNow()
	regionEntity := &entity.RegionManage{
		Name:         request.Name,
		Code:         request.Code,
		Type:         request.Type,
		CreateUserId: request.UserId,
		CreateTime:   now,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  0,
	}
	if err := data.DB.Create(regionEntity).Error; err != nil {
		return err
	}
	return nil
}

func UpdateRegion(request *Request) error {
	if err := checkBusiness(request, false); err != nil {
		return err
	}
	now := datetime.GetNow()
	regionEntity := &entity.RegionManage{
		Id:              request.Id,
		Name:            request.Name,
		Code:            request.Code,
		Type:            request.Type,
		CloudPlatformId: request.CloudPlatformId,
		UpdateUserId:    request.UserId,
		UpdateTime:      now,
	}
	if err := data.DB.Updates(regionEntity).Error; err != nil {
		return err
	}
	return nil
}

func DeleteRegion(request *Request) error {
	//校验该region下是否有az
	var azCount int64
	if err := data.DB.Model(&entity.AzManage{}).Where("region_id = ?", request.Id).Count(&azCount).Error; err != nil {
		return err
	}
	if azCount != 0 {
		return errors.New("无法删除，该region有az未删除")
	}
	now := datetime.GetNow()
	regionEntity := &entity.RegionManage{
		Id:           request.Id,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  1,
	}
	if err := data.DB.Updates(regionEntity).Error; err != nil {
		return err
	}
	return nil
}

func checkBusiness(request *Request, isCreate bool) error {
	if !isCreate {
		//校验regionId
		var regionCount int64
		if err := data.DB.Model(&entity.RegionManage{}).Where("id = ? AND delete_state = ?", request.Id, 0).Count(&regionCount).Error; err != nil {
			return err
		}
		if regionCount == 0 {
			return errors.New("az不存在")
		}
	}
	//校验regionCode
	var regionCodeCount int64
	if err := data.DB.Model(&entity.RegionManage{}).Where("code = ? AND cloud_platform_id = ? AND delete_state = ?", request.Code, request.CloudPlatformId, 0).Count(&regionCodeCount).Error; err != nil {
		return err
	}
	if regionCodeCount != 0 {
		return errors.New("RegionID重复")
	}
	//校验cloudPlatformId
	var cloudPlatformCount int64
	if err := data.DB.Model(&entity.CloudPlatformManage{}).Where("id = ? AND delete_state = ?", request.CloudPlatformId, 0).Count(&cloudPlatformCount).Error; err != nil {
		return err
	}
	if cloudPlatformCount == 0 {
		return errors.New("cloudPlatformId参数错误")
	}
	//regionType
	var regionTypeCount int64
	if err := data.DB.Model(&entity.ConfigItem{}).Where("p_id = ? AND code = ?", "2", request.Type).Count(&regionTypeCount).Error; err != nil {
		return err
	}
	if regionTypeCount == 0 {
		return errors.New("type参数错误")
	}
	return nil
}
