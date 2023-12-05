package ip_demand

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"
)

func SearchIpDemandPlanningByPlanId(planId int64) ([]entity.IPDemandPlanning, error) {
	var ipDemands []entity.IPDemandPlanning
	if err := data.DB.Table(entity.IPDemandPlanningTable).Where("plan_id=?", planId).Scan(&ipDemands).Error; err != nil {
		log.Errorf("[searchIpDemandPlanningByPlanId] error, %v", err)
		return nil, err
	}
	return ipDemands, nil
}

func DeleteIpDemandPlanningByPlanId(tx *gorm.DB, planId int64) error {
	if err := tx.Delete(&entity.IPDemandPlanning{}, "plan_id=?", planId).Error; err != nil {
		log.Errorf("[DeleteIpDemandPlanningByPlanId] error, %v", err)
		return err
	}
	return nil
}
