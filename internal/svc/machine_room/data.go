package machine_room

import (
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
)

func QueryRegionAzCellByPlanId(planId int64) (RegionAzCell, error) {
	var regionAzCell RegionAzCell
	if err := data.DB.Table(entity.PlanManageTable+" plan").Select("plan.id as planId, region.id as regionId, region.code as regionCode, az.id as azId, az.code as azCode, cell.id as cellId, cell.name as cellName").
		Joins("project_manage project on plan.project_id = project.id").
		Joins("region_manage region on project.region_id = region.id").
		Joins("az_manage az on project.az_id = az.id").
		Joins("cell_manage cell on project.cell_id = cell.id").
		Where("plan.id = ?", planId).
		Find(&regionAzCell).Error; err != nil {
		log.Errorf("[queryRegionAzCellByPlanId] query region az cell error, %v", err)
		return regionAzCell, err
	}
	return regionAzCell, nil
}

func UpdateRegionAzCellByPlanId(regionAzCell RegionAzCell, userId string) error {
	originRegionAzCell, err := QueryRegionAzCellByPlanId(regionAzCell.PlanId)
	if err != nil {
		return err
	}
	now := datetime.GetNow()
	regionManage := entity.RegionManage{
		Id:           originRegionAzCell.RegionId,
		Code:         regionAzCell.RegionCode,
		UpdateTime:   now,
		UpdateUserId: userId,
	}
	azManage := entity.AzManage{
		Id:           originRegionAzCell.AzId,
		Code:         regionAzCell.AzCode,
		UpdateTime:   now,
		UpdateUserId: userId,
	}
	cellManage := entity.CellManage{
		Id:           originRegionAzCell.CellId,
		Name:         regionAzCell.CellName,
		UpdateTime:   now,
		UpdateUserId: userId,
	}
	if err = data.DB.Transaction(func(tx *gorm.DB) error {
		if err = tx.Updates(&regionManage).Error; err != nil {
			return err
		}
		if err = tx.Updates(&azManage).Error; err != nil {
			return err
		}
		if err = tx.Updates(&cellManage).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func QueryMachineRoomByPlanId(planId int64) ([]entity.MachineRoom, error) {
	var machineRooms []entity.MachineRoom
	var azId int64
	if err := data.DB.Table(entity.PlanManageTable+" plan").Select("project.az_id").
		Joins("left join project_manage project on plan.project_id = project.id").
		Where("plan.id = ?", planId).Find(&azId).Error; err != nil {
		log.Errorf("[queryAzIdByPlanId] query azId error, %v", err)
	}
	if err := data.DB.Table(entity.MachineRoomTable).
		Where("az_id = ?", azId).
		Find(&machineRooms).Error; err != nil {
		log.Errorf("[queryMachineRoomByAzId] query machine room error, %v", err)
	}
	return machineRooms, nil
}

func UpdateMachineRoomByPlanId(planId int64, machineRooms []entity.MachineRoom) error {
	var azId int64
	if err := data.DB.Table(entity.PlanManageTable+" plan").Select("project.az_id").
		Joins("left join project_manage project on plan.project_id = project.id").
		Where("plan.id = ?", planId).Find(&azId).Error; err != nil {
		log.Errorf("[queryAzIdByPlanId] query azId error, %v", err)
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&entity.MachineRoom{}, "az_id = ?", azId).Error; err != nil {
			return err
		}
		for i := range machineRooms {
			machineRooms[i].AzId = azId
			machineRooms[i].Sort = i + 1
		}
		if err := tx.Create(&machineRooms).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func QueryCabinetsPage(request *PageRequest) ([]entity.CabinetInfo, int64, error) {
	var cabinets []entity.CabinetInfo
	var count int64
	if err := data.DB.Table(entity.CabinetInfoTable).Where("plan_id = ?", request.PlanId).Count(&count).Error; err != nil {
		log.Errorf("[pageCabinets] error, %v", err)
		return nil, 0, err
	}
	if err := data.DB.Table(entity.CabinetInfoTable).Where("plan_id =?", request.PlanId).Order("id asc").Offset((request.Current - 1) * request.PageSize).Limit(request.PageSize).Find(&cabinets).Error; err != nil {
		log.Errorf("[pageCabinets] error, %v", err)
		return nil, 0, err
	}
	return cabinets, count, nil
}

func QueryCabinetsByPlanId(planId int64) ([]entity.CabinetInfo, error) {
	var cabinets []entity.CabinetInfo
	if err := data.DB.Table(entity.CabinetInfoTable).Where("plan_id =?", planId).Order("id asc").Find(&cabinets).Error; err != nil {
		log.Errorf("[queryCabinetsByPlanId] error, %v", err)
		return nil, err
	}
	return cabinets, nil
}

func DeleteCabinets(cabinets []entity.CabinetInfo) error {
	if len(cabinets) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.CabinetInfoTable).Delete(&cabinets).Error; err != nil {
		log.Errorf("[deleteCabinets] error, %v", err)
		return err
	}
	return nil
}

func BatchCreateCabinets(cabinets []entity.CabinetInfo) error {
	if len(cabinets) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.CabinetInfoTable).Create(&cabinets).Error; err != nil {
		log.Errorf("batch insert cabinetInfo error: %v", err)
		return err
	}
	return nil
}

func CreateCabinet(cabinet *entity.CabinetInfo) error {
	if err := data.DB.Table(entity.CabinetInfoTable).Create(&cabinet).Error; err != nil {
		log.Errorf("insert cabinetInfo error: %v", err)
		return err
	}
	return nil
}

func DeleteCabinetIdleSlotRel(cabinetIds []int64) error {
	if err := data.DB.Table(entity.CabinetIdleSlotRelTable).Where("cabinet_id in (?)", cabinetIds).Delete(&entity.CabinetIdleSlotRel{}).Error; err != nil {
		log.Errorf("[deleteCabinetIdleSlotRel] error, %v", err)
		return err
	}
	return nil
}

func BatchCreateCabinetIdleSlotRel(cabinetIdleSlots []entity.CabinetIdleSlotRel) error {
	if len(cabinetIdleSlots) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.CabinetIdleSlotRelTable).Create(&cabinetIdleSlots).Error; err != nil {
		log.Errorf("batch insert cabinetIdleSlots error: %v", err)
		return err
	}
	return nil
}

func DeleteCabinetRackServerRel(cabinetIds []int64) error {
	if err := data.DB.Table(entity.CabinetRackServerSlotRelTable).Where("cabinet_id in (?)", cabinetIds).Delete(&entity.CabinetRackServerSlotRel{}).Error; err != nil {
		log.Errorf("[deleteCabinetRackServerRel] error, %v", err)
		return err
	}
	return nil
}

func BatchCreateCabinetRackServerRel(cabinetRackServerSlots []entity.CabinetRackServerSlotRel) error {
	if len(cabinetRackServerSlots) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.CabinetRackServerSlotRelTable).Create(&cabinetRackServerSlots).Error; err != nil {
		log.Errorf("batch insert cabinetRackServerSlots error: %v", err)
		return err
	}
	return nil
}

func DeleteCabinetRackAswPortRel(cabinetIds []int64) error {
	if err := data.DB.Table(entity.CabinetRackAswPortRelTable).Where("cabinet_id in (?)", cabinetIds).Delete(&entity.CabinetRackAswPortRel{}).Error; err != nil {
		log.Errorf("[deleteCabinetRackAswPortRel] error, %v", err)
		return err
	}
	return nil
}

func BatchCreateCabinetRackAswPortRel(cabinetRackAswPorts []entity.CabinetRackAswPortRel) error {
	if len(cabinetRackAswPorts) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.CabinetRackAswPortRelTable).Create(&cabinetRackAswPorts).Error; err != nil {
		log.Errorf("batch insert cabinetRackAswPorts error: %v", err)
		return err
	}
	return nil
}
