package az

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"errors"
	"gorm.io/gorm"
)

func ListAz(request *Request) ([]*Response, error) {
	var response []*Response
	screenSql, screenParams, orderSql := " delete_state = ? AND region_id = ? ", []interface{}{0, request.RegionId}, " create_time "
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
	var azList []*entity.AzManage
	if err := data.DB.Where(screenSql, screenParams...).Order(orderSql).Find(&azList).Error; err != nil {
		return nil, err
	}
	var azIdList []int64
	for _, v := range azList {
		response = append(response, &Response{Az: v})
		azIdList = append(azIdList, v.Id)
	}
	//查询机房表
	var machineRoomList []*entity.MachineRoom
	if err := data.DB.Model(&entity.MachineRoom{}).Where("az_id IN (?)", azIdList).Order("sort asc").Find(&machineRoomList).Error; err != nil {
		return nil, err
	}
	var azIdMachineRoomMap = make(map[int64][]*entity.MachineRoom)
	for _, v := range machineRoomList {
		azIdMachineRoomMap[v.AzId] = append(azIdMachineRoomMap[v.AzId], v)
	}
	for i, v := range response {
		response[i].MachineRoomList = azIdMachineRoomMap[v.Az.Id]
	}
	return response, nil
}

func CreateAz(request *Request) error {
	if err := checkBusiness(request, true); err != nil {
		return err
	}
	now := datetime.GetNow()
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		azEntity := &entity.AzManage{
			Code:         request.Code,
			RegionId:     request.RegionId,
			CreateUserId: request.UserId,
			CreateTime:   now,
			UpdateUserId: request.UserId,
			UpdateTime:   now,
			DeleteState:  0,
		}
		if err := tx.Create(&azEntity).Error; err != nil {
			return err
		}
		var machineRoomList []*entity.MachineRoom
		for i, v := range request.MachineRoomList {
			machineRoomList = append(machineRoomList, &entity.MachineRoom{
				AzId:     azEntity.Id,
				Name:     v.Name,
				Abbr:     v.Abbr,
				Province: v.Province,
				City:     v.City,
				Address:  v.Address,
				Sort:     i + 1,
			})
		}
		if err := tx.Create(&machineRoomList).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func UpdateAz(request *Request) error {
	if err := checkBusiness(request, false); err != nil {
		return err
	}
	now := datetime.GetNow()
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		azEntity := &entity.AzManage{
			Id:           request.Id,
			Code:         request.Code,
			UpdateUserId: request.UserId,
			UpdateTime:   now,
		}
		if err := tx.Updates(&azEntity).Error; err != nil {
			return err
		}
		if err := tx.Delete(&entity.MachineRoom{}, "az_id = ?", request.Id).Error; err != nil {
			return err
		}
		var machineRoomList []*entity.MachineRoom
		for i, v := range request.MachineRoomList {
			machineRoomList = append(machineRoomList, &entity.MachineRoom{
				AzId:     request.Id,
				Name:     v.Name,
				Abbr:     v.Abbr,
				Province: v.Province,
				City:     v.City,
				Address:  v.Address,
				Sort:     i + 1,
			})
		}
		if err := tx.Create(&machineRoomList).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func DeleteAz(request *Request) error {
	now := datetime.GetNow()
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		azEntity := &entity.AzManage{
			Id:           request.Id,
			UpdateUserId: request.UserId,
			UpdateTime:   now,
			DeleteState:  1,
		}
		if err := tx.Updates(&azEntity).Error; err != nil {
			return err
		}
		if err := tx.Delete(&entity.MachineRoom{}, "az_id = ?", request.Id).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func checkBusiness(request *Request, isCreate bool) error {
	if isCreate {
		//校验regionId
		var regionCount int64
		if err := data.DB.Model(&entity.RegionManage{}).Where("id = ? AND delete_state = ?", request.RegionId, 0).Count(&regionCount).Error; err != nil {
			return err
		}
		if regionCount == 0 {
			return errors.New("region不存在")
		}
	} else {
		//校验azId
		var azCount int64
		if err := data.DB.Model(&entity.AzManage{}).Where("id = ? AND delete_state = ?", request.Id, 0).Count(&azCount).Error; err != nil {
			return err
		}
		if azCount == 0 {
			return errors.New("az不存在")
		}
	}
	return nil
}
