package machine_room

import (
	"github.com/opentrx/seata-golang/v2/pkg/util/log"

	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
)

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
