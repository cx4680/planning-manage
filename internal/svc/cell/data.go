package cell

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"errors"
	"gorm.io/gorm"
)

func ListCell(request *Request) ([]*entity.CellManage, error) {
	var azCellRelList []*entity.AzCellRel
	if err := data.DB.Model(&entity.AzCellRel{}).Where("az_id = ?", request.AzId).Find(&azCellRelList).Error; err != nil {
		return nil, err
	}
	var cellIdList []int64
	for _, v := range azCellRelList {
		cellIdList = append(cellIdList, v.CellId)
	}
	screenSql, screenParams, orderSql := " delete_state = ? AND id IN (?) ", []interface{}{0, cellIdList}, " create_time "
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
		Code:         request.Code,
		CreateUserId: request.UserId,
		CreateTime:   now,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  0,
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&cellEntity).Error; err != nil {
			return err
		}
		var azCellRelList []*entity.AzCellRel
		for _, v := range request.AzIdList {
			azCellRelList = append(azCellRelList, &entity.AzCellRel{AzId: v, CellId: cellEntity.Id})
		}
		if err := tx.Create(&azCellRelList).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
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
		Code:         request.Code,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Updates(&cellEntity).Error; err != nil {
			return err
		}
		if err := tx.Where("cell_id = ?", request.Id).Delete(&entity.AzCellRel{}).Error; err != nil {
			return err
		}
		var azCellRelList []*entity.AzCellRel
		for _, v := range request.AzIdList {
			azCellRelList = append(azCellRelList, &entity.AzCellRel{AzId: v, CellId: request.Id})
		}
		if err := tx.Create(&azCellRelList).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
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
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Updates(&cellEntity).Error; err != nil {
			return err
		}
		if err := tx.Where("cell_id = ?", request.Id).Delete(&entity.AzCellRel{}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func checkBusiness(request *Request, isCreate bool) error {
	if !isCreate {
		//校验azId
		var cellCount int64
		if err := data.DB.Model(&entity.CellManage{}).Where("id = ? AND delete_state = ?", request.Id, 0).Count(&cellCount).Error; err != nil {
			return err
		}
		if cellCount == 0 {
			return errors.New("az不存在")
		}
	}
	//校验azId
	var azCount int64
	if err := data.DB.Model(&entity.AzManage{}).Where("id IN (?) AND delete_state = ?", request.AzIdList, 0).Count(&azCount).Error; err != nil {
		return err
	}
	if int(azCount) != len(request.AzIdList) {
		return errors.New("azIdList中有不存在的az")
	}
	return nil
}
