package cell

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
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
	if util.IsBlank(request.Type) {
		request.Type = constant.CellTypeControl
	}
	now := datetime.GetNow()
	cellEntity := &entity.CellManage{
		Name:         request.Name,
		AzId:         request.AzId,
		Type:         request.Type,
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
		Type:         request.Type,
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
	var azManage = &entity.AzManage{}
	if isCreate {
		//校验azId
		if err := data.DB.Where("id = ? AND delete_state = ?", request.AzId, 0).Find(&azManage).Error; err != nil {
			return err
		}
		if azManage.Id == 0 {
			return errors.New("azId错误")
		}
	} else {
		//校验cellId
		var cellManage = &entity.CellManage{}
		if err := data.DB.Where("id = ? AND delete_state = ?", request.Id, 0).Find(&cellManage).Error; err != nil {
			return err
		}
		if cellManage.Id == 0 {
			return errors.New("cell不存在")
		}
		//查询az
		if err := data.DB.Where("id = ? AND delete_state = ?", cellManage.AzId, 0).Find(&azManage).Error; err != nil {
			return err
		}
	}
	//校验控制集群为region下唯一
	if request.Type == constant.CellTypeControl {
		var azIdList []string
		if err := data.DB.Model(&entity.AzManage{}).Select("id").Where("region_id = ? AND delete_state = ?", azManage.RegionId, 0).Find(&azIdList).Error; err != nil {
			return err
		}
		var cellTypeCount int64
		if err := data.DB.Model(&entity.CellManage{}).Where("az_id IN (?) AND type = ? AND delete_state = ?", azIdList, constant.CellTypeControl, 0).Count(&cellTypeCount).Error; err != nil {
			return err
		}
		if cellTypeCount != 0 {
			return errors.New("已存在控制集群")
		}
	}
	return nil
}
