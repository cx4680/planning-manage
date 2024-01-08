package machine_room

import (
	"github.com/opentrx/seata-golang/v2/pkg/util/log"

	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
)

func QueryCabinetsPage(request *PageRequest) ([]entity.CabinetInfo, int64, error) {
	var cabinets []entity.CabinetInfo
	var count int64
	if err := data.DB.Where("plan_id = ?", request.PlanId).Count(&count).Error; err != nil {
		log.Errorf("[pageCabinets] error, %v", err)
		return nil, 0, err
	}
	if err := data.DB.Where("plan_id =?", request.PlanId).Order("create_time desc, update_time desc").Offset((request.Current - 1) * request.PageSize).Limit(request.PageSize).Find(&cabinets).Error; err != nil {
		log.Errorf("[pageCabinets] error, %v", err)
		return nil, 0, err
	}
	return cabinets, count, nil
}
