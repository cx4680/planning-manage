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

func GetIpDemandBaselineByVersionId(versionId int64) ([]*IpDemandBaselineDto, error) {
	var demandBaseline []*IpDemandBaselineDto
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

func exportIpDemandPlanningByPlanId(planId int64) (string, []IpDemandPlanningExportResponse, error) {
	var planManage entity.PlanManage
	if err := data.DB.Table(entity.PlanManageTable).Where("id=? and delete_state = 0", planId).Scan(&planManage).Error; err != nil {
		log.Errorf("[exportIpDemandPlanningByPlanId] get planManage by id err, %v", err)
		return "", nil, err
	}
	var projectManage entity.ProjectManage
	if err := data.DB.Table(entity.ProjectManageTable).Where("id=? and delete_state = 0", planManage.ProjectId).Scan(&projectManage).Error; err != nil {
		log.Errorf("[exportIpDemandPlanningByPlanId] get projectManage by id err, %v", err)
		return "", nil, err
	}
	var response []IpDemandPlanningExportResponse
	if err := data.DB.Raw("select segment_type, `describe`, vlan, c_num, address, address_planning from ip_demand_planning where plan_id = ?", planId).Scan(&response).Error; err != nil {
		log.Errorf("[exportIpDemandPlanningByPlanId] query db error, %v", err)
		return "", nil, err
	}
	return projectManage.Name + "-" + planManage.Name + "-" + "IP需求清单", response, nil
}
