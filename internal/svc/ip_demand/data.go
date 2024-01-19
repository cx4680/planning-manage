package ip_demand

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
)

func SearchIpDemandPlanningByPlanId(planId int64) ([]entity.IpDemandPlanning, error) {
	var ipDemands []entity.IpDemandPlanning
	if err := data.DB.Where("plan_id = ?", planId).Find(&ipDemands).Error; err != nil {
		log.Errorf("[searchIpDemandPlanningByPlanId] error, %v", err)
		return nil, err
	}
	return ipDemands, nil
}

func DeleteIpDemandPlanningByPlanId(tx *gorm.DB, planId int64) error {
	if err := tx.Delete(&entity.IpDemandPlanning{}, "plan_id = ?", planId).Error; err != nil {
		log.Errorf("[DeleteIpDemandPlanningByPlanId] error, %v", err)
		return err
	}
	return nil
}

func GetIpDemandBaselineByVersionId(versionId int64) ([]*IpDemandBaselineDto, error) {
	var demandBaseline []*IpDemandBaselineDto
	if err := data.DB.Table(entity.IPDemandBaselineTable+" a").Select("a.id", "a.version_id", "a.vlan", "a.explain", "a.network_type", "a.description", "a.ip_suggestion", "a.assign_num", "a.remark", "b.device_role_id").
		Joins("left join ip_demand_device_role_rel b on a.id = b.ip_demand_id").
		Where("a.version_id = ?", versionId).Find(&demandBaseline).Error; err != nil {
		log.Errorf("[GetIpDemandBaselineByVersionId] error, %v", err)
		return nil, err
	}
	return demandBaseline, nil
}

func SaveBatch(tx *gorm.DB, demandPlannings []*entity.IpDemandPlanning) error {
	if err := tx.Create(&demandPlannings).Error; err != nil {
		log.Errorf("batch insert demandPlannings error: ", err)
		return err
	}
	return nil
}

func exportIpDemandPlanningByPlanId(planId int64) (string, []IpDemandPlanningExportResponse, error) {
	var planManage entity.PlanManage
	if err := data.DB.Where("id = ? AND delete_state = 0", planId).Find(&planManage).Error; err != nil {
		log.Errorf("[exportIpDemandPlanningByPlanId] get planManage by id err, %v", err)
		return "", nil, err
	}
	var projectManage entity.ProjectManage
	if err := data.DB.Where("id = ? AND delete_state = 0", planManage.ProjectId).Find(&projectManage).Error; err != nil {
		log.Errorf("[exportIpDemandPlanningByPlanId] get projectManage by id err, %v", err)
		return "", nil, err
	}
	var ipDemandPlanningList []*entity.IpDemandPlanning
	if err := data.DB.Where("plan_id = ?", planId).Find(&ipDemandPlanningList).Error; err != nil {
		log.Errorf("[exportIpDemandPlanningByPlanId] query db error, %v", err)
		return "", nil, err
	}
	var response []IpDemandPlanningExportResponse
	for _, v := range ipDemandPlanningList {
		networkType := constant.IpDemandNetworkTypeIpv4Cn
		if v.NetworkType == constant.IpDemandNetworkTypeIpv6 {
			networkType = constant.IpDemandNetworkTypeIpv6Cn
		}
		response = append(response, IpDemandPlanningExportResponse{
			LogicalGrouping: v.LogicalGrouping,
			SegmentType:     v.SegmentType,
			NetworkType:     networkType,
			Describe:        v.Describe,
			Vlan:            v.Vlan,
			CNum:            v.CNum,
			Address:         v.Address,
			AddressPlanning: v.AddressPlanning,
		})
	}
	return projectManage.Name + "-" + planManage.Name + "-" + "IP需求清单", response, nil
}

func getIpDemandPlanningList(planId int64) ([]*entity.IpDemandPlanning, error) {
	var list []*entity.IpDemandPlanning
	if err := data.DB.Where("plan_id = ?", planId).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func uploadIpDemand(planId int64, ipDemandPlanningExportResponse []IpDemandPlanningExportResponse) error {
	var ipDemandPlanningList []*entity.IpDemandPlanning
	now := datetime.GetNow()
	for _, v := range ipDemandPlanningExportResponse {
		networkType := constant.IpDemandNetworkTypeIpv4
		if v.NetworkType == constant.IpDemandNetworkTypeIpv6Cn {
			networkType = constant.IpDemandNetworkTypeIpv6
		}
		ipDemandPlanningList = append(ipDemandPlanningList, &entity.IpDemandPlanning{
			PlanId:          planId,
			LogicalGrouping: v.LogicalGrouping,
			SegmentType:     v.SegmentType,
			NetworkType:     networkType,
			Vlan:            v.Vlan,
			CNum:            v.CNum,
			Address:         v.Address,
			Describe:        v.Describe,
			AddressPlanning: v.AddressPlanning,
			CreateTime:      now,
			UpdateTime:      now,
		})
	}
	if err := data.DB.Create(ipDemandPlanningList).Error; err != nil {
		return err
	}
	return nil
}
