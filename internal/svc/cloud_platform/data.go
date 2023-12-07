package cloud_platform

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"errors"
	"gorm.io/gorm"
)

func ListCloudPlatform(request *Request) ([]*entity.CloudPlatformManage, error) {
	screenSql, screenParams, orderSql := " delete_state = ? ", []interface{}{0}, " create_time "
	if request.CustomerId != 0 {
		screenSql += " AND customer_id = ? "
		screenParams = append(screenParams, request.CustomerId)
	}
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
	var list []*entity.CloudPlatformManage
	if err := data.DB.Where(screenSql, screenParams...).Order(orderSql).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func CreateCloudPlatform(request *Request) error {
	if err := checkBusiness(request, true); err != nil {
		return err
	}
	now := datetime.GetNow()
	cloudPlatformEntity := &entity.CloudPlatformManage{
		Name:         request.Name,
		Type:         request.Type,
		CustomerId:   request.CustomerId,
		CreateUserId: request.UserId,
		CreateTime:   now,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  0,
	}
	if err := data.DB.Create(cloudPlatformEntity).Error; err != nil {
		return err
	}
	return nil
}

func UpdateCloudPlatform(request *Request) error {
	if err := checkBusiness(request, true); err != nil {
		return err
	}
	cloudPlatformEntity := &entity.CloudPlatformManage{
		Id:           request.Id,
		Name:         request.Name,
		Type:         request.Type,
		CustomerId:   request.CustomerId,
		UpdateUserId: request.UserId,
		UpdateTime:   datetime.GetNow(),
	}
	if err := data.DB.Updates(cloudPlatformEntity).Error; err != nil {
		return err
	}
	return nil
}

func CreateCloudPlatformByCustomerId(request *Request) error {
	now := datetime.GetNow()
	cloudPlatformEntity := &entity.CloudPlatformManage{
		Name:         "云平台1",
		Type:         "operational",
		CustomerId:   request.CustomerId,
		CreateUserId: request.UserId,
		CreateTime:   now,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  0,
	}
	regionEntity := &entity.RegionManage{
		Name:         "region1",
		Code:         "region1",
		Type:         "merge",
		CreateUserId: request.UserId,
		CreateTime:   now,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  0,
	}
	azEntity := &entity.AzManage{
		Name:         "zone1",
		Code:         "zone1",
		CreateUserId: request.UserId,
		CreateTime:   now,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  0,
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(cloudPlatformEntity).Error; err != nil {
			return err
		}
		regionEntity.CloudPlatformId = cloudPlatformEntity.Id
		if err := tx.Create(regionEntity).Error; err != nil {
			return err
		}
		azEntity.RegionId = regionEntity.Id
		if err := tx.Create(azEntity).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func checkBusiness(request *Request, isCreate bool) error {
	//校验cloudPlatformType
	if util.IsNotBlank(request.Type) {
		var cloudPlatformTypeCount int64
		if err := data.DB.Model(&entity.ConfigItem{}).Where("p_id = ? AND code = ?", "1", request.Type).Count(&cloudPlatformTypeCount).Error; err != nil {
			return err
		}
		if cloudPlatformTypeCount == 0 {
			return errors.New("type参数错误")
		}
	}
	if !isCreate {
		//校验cloudPlatformId
		var cloudPlatformCount int64
		if err := data.DB.Model(&entity.CloudPlatformManage{}).Where("id = ? AND delete_state = ?", request.Id, 0).Count(&cloudPlatformCount).Error; err != nil {
			return err
		}
		if cloudPlatformCount == 0 {
			return errors.New("云平台不存在")
		}
	}
	return nil
}
