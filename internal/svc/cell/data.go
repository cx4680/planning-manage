package cell

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"errors"
)

func ListCell(request *Request) ([]*entity.CellManage, error) {
	screenSql, screenParams, orderSql := " delete_state = ? AND az_id = ? ", []interface{}{0, request.AzId}, " create_time "
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
	var list []*entity.CellManage
	if err := data.DB.Where(screenSql, screenParams...).Order(orderSql).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func CreateCell(request *Request) error {
	if err := checkBusiness(request, true); err != nil {
		return err
	}
	now := datetime.GetNow()
	cellEntity := &entity.CellManage{
		Name:         request.Name,
		AzId:         request.AzId,
		CreateUserId: request.UserId,
		CreateTime:   now,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  0,
	}
	if err := data.DB.Create(&cellEntity).Error; err != nil {
		return err
	}
	return nil
}

func UpdateCell(request *Request) error {
	if err := checkBusiness(request, false); err != nil {
		return err
	}
	now := datetime.GetNow()
	cellEntity := &entity.CellManage{
		Id:           request.Id,
		Name:         request.Name,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
	}
	if err := data.DB.Updates(&cellEntity).Error; err != nil {
		return err
	}
	return nil
}

func DeleteCell(request *Request) error {
	now := datetime.GetNow()
	cellEntity := &entity.CellManage{
		Id:           request.Id,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  1,
	}
	if err := data.DB.Updates(&cellEntity).Error; err != nil {
		return err
	}
	return nil
}

func checkBusiness(request *Request, isCreate bool) error {
	if !isCreate {
		//校验cellId
		var cellCount int64
		if err := data.DB.Model(&entity.CellManage{}).Where("id = ? AND delete_state = ?", request.Id, 0).Count(&cellCount).Error; err != nil {
			return err
		}
		if cellCount == 0 {
			return errors.New("cell不存在")
		}
	}
	//校验azId
	var azCount int64
	if err := data.DB.Model(&entity.AzManage{}).Where("id = ? AND delete_state = ?", request.AzId, 0).Count(&azCount).Error; err != nil {
		return err
	}
	if azCount == 0 {
		return errors.New("azId错误")
	}
	return nil
}
