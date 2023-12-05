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

func GetIpDemandBaselineByVersionId(versionId int64) ([]IpDemandBaselineDto, error) {
	var demandBaseline []IpDemandBaselineDto
	if err := data.DB.Raw("select a.id,a.version_id,a.vlan,a.explain,a.description,a.ip_suggestion,a.assign_num,a.remark,b.device_role_id from ip_demand_baseline a left join ip_demand_device_role_rel b on a.id = b.ip_demand_id where a.version_id = ?", versionId).Scan(&demandBaseline).Error; err != nil {
		log.Errorf("[GetIpDemandBaselineByVersionId] error, %v", err)
		return nil, err
	}
	return demandBaseline, nil
}

func SaveBatch(tx *gorm.DB, demandPlannings []*entity.IPDemandPlanning) error {
	if err := tx.Table(entity.IPDemandPlanningTable).Create(&demandPlannings).Scan(&demandPlannings).Error; err != nil {
		log.Errorf("batch insert demandPlannings error: ", err)
		return err
	}
	return nil
}
