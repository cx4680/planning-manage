package ip_demand

import (
	"errors"
	"time"

	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
)

func SearchIpDemandPlanningByPlanId(planId int64) ([]*entity.IpDemandPlanning, error) {
	var ipDemands []*entity.IpDemandPlanning
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

func ExportIpDemandPlanningByPlanId(planId int64) (string, []IpDemandPlanningExportResponse, error) {
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
			AddressPlanning: v.AddressPlanning,
		})
	}
	return projectManage.Name + "-" + planManage.Name + "-" + "IP需求清单", response, nil
}

func GetIpDemandPlanningList(planId int64) ([]*IpDemandPlanning, error) {
	ipDemandPlanningList, err := SearchIpDemandPlanningByPlanId(planId)
	if err != nil {
		return nil, err
	}
	var list []*IpDemandPlanning
	for _, v := range ipDemandPlanningList {
		networkType := constant.IpDemandNetworkTypeIpv4Cn
		if v.NetworkType == constant.IpDemandNetworkTypeIpv6 {
			networkType = constant.IpDemandNetworkTypeIpv6Cn
		}
		list = append(list, &IpDemandPlanning{
			IpDemandPlanning: v,
			NetworkTypeCn:    networkType,
		})
	}
	return list, nil
}

func UploadIpDemand(planId int64, ipDemandPlanningExportResponse []IpDemandPlanningExportResponse, userId string) error {
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		var ipDemandShelveList []*entity.IpDemandShelve
		for _, v := range ipDemandPlanningExportResponse {
			if util.IsBlank(v.Address) {
				return errors.New("地址段不能为空")
			}
			networkType := constant.IpDemandNetworkTypeIpv4
			if v.NetworkType == constant.IpDemandNetworkTypeIpv6Cn {
				networkType = constant.IpDemandNetworkTypeIpv6
			}
			ipDemandShelveList = append(ipDemandShelveList, &entity.IpDemandShelve{
				PlanId:          planId,
				LogicalGrouping: v.LogicalGrouping,
				SegmentType:     v.SegmentType,
				NetworkType:     networkType,
				Vlan:            v.Vlan,
				CNum:            v.CNum,
				Address:         v.Address,
				Describe:        v.Describe,
				AddressPlanning: v.AddressPlanning,
				CreateUserId:    userId,
				CreateTime:      time.Now(),
			})
		}
		if err := tx.Delete(&entity.IpDemandShelve{}, "plan_id = ?", planId).Error; err != nil {
			return err
		}
		if err := tx.Create(&ipDemandShelveList).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func SaveIpDemand(tx *gorm.DB, planId int64) error {
	if err := tx.Updates(&entity.PlanManage{Id: planId, DeliverPlanStage: constant.DeliverPlanningGlobalConfiguration}).Error; err != nil {
		return err
	}
	return nil
}

func GetNetworkShelveList(planId int64) ([]*entity.NetworkDeviceShelve, error) {
	var networkDeviceShelve []*entity.NetworkDeviceShelve
	if err := data.DB.Where("plan_id = ?", planId).Find(&networkDeviceShelve).Error; err != nil {
		return nil, err
	}
	return networkDeviceShelve, nil
}

func QueryNetworkDeviceIp(planId int64) ([]entity.NetworkDeviceIp, error) {
	var networkDeviceIps []entity.NetworkDeviceIp
	if err := data.DB.Table(entity.NetworkDeviceIpTable).Where("plan_id = ?", planId).Find(&networkDeviceIps).Error; err != nil && err != gorm.ErrRecordNotFound {
		return networkDeviceIps, err
	}
	return networkDeviceIps, nil
}

func CreateNetworkDeviceIp(tx *gorm.DB, networkDeviceIps []entity.NetworkDeviceIp) error {
	if len(networkDeviceIps) == 0 {
		return nil
	}
	if err := tx.Table(entity.NetworkDeviceIpTable).Create(&networkDeviceIps).Error; err != nil {
		return err
	}
	return nil
}

func UpdateNetworkDeviceIp(tx *gorm.DB, networkDeviceIps []entity.NetworkDeviceIp) error {
	for _, networkDeviceIp := range networkDeviceIps {
		if err := tx.Table(entity.NetworkDeviceIpTable).Save(&networkDeviceIp).Error; err != nil {
			return err
		}
	}
	return nil
}

func DeleteNetworkDeviceIp(tx *gorm.DB, networkDeviceIps []entity.NetworkDeviceIp) error {
	if len(networkDeviceIps) == 0 {
		return nil
	}
	if err := tx.Table(entity.NetworkDeviceIpTable).Delete(&networkDeviceIps).Error; err != nil {
		return err
	}
	return nil
}

func GetIpDemandShelve(planId int64) ([]*entity.IpDemandShelve, error) {
	var ipDemandShelves []*entity.IpDemandShelve
	if err := data.DB.Where("plan_id = ?", planId).Find(&ipDemandShelves).Error; err != nil {
		return nil, err
	}
	return ipDemandShelves, nil
}

func QueryInClusterNodeRoleBaselineByVersionId(versionId int64) ([]entity.NodeRoleBaseline, error) {
	var nodeRoleBaselines []entity.NodeRoleBaseline
	if err := data.DB.Table(entity.NodeRoleBaselineTable).Where("version_id = ? and deploy_method = ?", versionId, constant.NodeRoleDeployMethodInCluster).Find(&nodeRoleBaselines).Error; err != nil {
		return nodeRoleBaselines, err
	}
	return nodeRoleBaselines, nil
}

func QueryServerShelve(planId int64, nodeRoleIds []int64) ([]entity.ServerShelve, error) {
	var serverShelves []entity.ServerShelve
	if err := data.DB.Table(entity.ServerShelveTable).Where("plan_id = ? and node_role_id in (?)", planId, nodeRoleIds).Find(&serverShelves).Error; err != nil {
		return serverShelves, err
	}
	return serverShelves, nil
}

func QueryServerIp(planId int64) ([]entity.ServerIp, error) {
	var serverIps []entity.ServerIp
	if err := data.DB.Table(entity.ServerIpTable).Where("plan_id = ?", planId).Find(&serverIps).Error; err != nil && err != gorm.ErrRecordNotFound {
		return serverIps, err
	}
	return serverIps, nil
}

func CreateServerIp(tx *gorm.DB, serverIps []entity.ServerIp) error {
	if err := tx.Table(entity.ServerIpTable).Create(&serverIps).Error; err != nil {
		return err
	}
	return nil
}

func UpdateServerIp(tx *gorm.DB, serverIps []entity.ServerIp) error {
	for _, serverIp := range serverIps {
		if err := tx.Table(entity.ServerIpTable).Save(&serverIp).Error; err != nil {
			return err
		}
	}
	return nil
}

func DeleteServerIp(tx *gorm.DB, serverIps []entity.ServerIp) error {
	if len(serverIps) == 0 {
		return nil
	}
	if err := tx.Table(entity.ServerIpTable).Delete(&serverIps).Error; err != nil {
		return err
	}
	return nil
}
