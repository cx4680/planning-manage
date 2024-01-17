package global_config

import (
	"time"

	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
)

func QueryVlanIdConfigByPlanId(planId int64) (entity.VlanIdConfig, error) {
	var vlanIdConfig entity.VlanIdConfig
	if err := data.DB.Table(entity.VlanIdConfigTable).Where("plan_id = ?", planId).Find(&vlanIdConfig).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("[queryVlanIdConfigByPlanId] query vlan id config error, %v", err)
		return vlanIdConfig, err
	}
	return vlanIdConfig, nil
}

func InsertVlanIdConfig(userId string, request VlanIdConfigRequest) error {
	now := time.Now()
	vlanIdConfig := entity.VlanIdConfig{
		PlanId:             request.PlanId,
		InBandMgtVlanId:    request.InBandMgtVlanId,
		LocalStorageVlanId: request.LocalStorageVlanId,
		BizIntranetVlanId:  request.BizIntranetVlanId,
		CreateUserId:       userId,
		CreateTime:         now,
		UpdateUserId:       userId,
		UpdateTime:         now,
	}
	if err := data.DB.Table(entity.VlanIdConfigTable).Create(&vlanIdConfig).Error; err != nil {
		log.Errorf("[insertVlanIdConfigByPlanId] insert vlan id config error, %v", err)
		return err
	}
	return nil
}

func QueryVlanIdConfigById(id int64) (entity.VlanIdConfig, error) {
	var vlanIdConfig entity.VlanIdConfig
	if err := data.DB.Table(entity.VlanIdConfigTable).Where("id = ?", id).Find(&vlanIdConfig).Error; err != nil {
		log.Errorf("[queryVlanIdConfigById] query vlan id config error, %v", err)
		return vlanIdConfig, err
	}
	return vlanIdConfig, nil
}

func UpdateVlanIdConfigById(userId string, id int64, request VlanIdConfigRequest, originVlanIdConfig entity.VlanIdConfig) error {
	vlanIdConfig := entity.VlanIdConfig{
		Id:                 id,
		PlanId:             request.PlanId,
		InBandMgtVlanId:    request.InBandMgtVlanId,
		LocalStorageVlanId: request.LocalStorageVlanId,
		BizIntranetVlanId:  request.BizIntranetVlanId,
		CreateUserId:       originVlanIdConfig.CreateUserId,
		CreateTime:         originVlanIdConfig.CreateTime,
		UpdateUserId:       userId,
		UpdateTime:         time.Now(),
	}
	if err := data.DB.Table(entity.VlanIdConfigTable).Save(&vlanIdConfig).Error; err != nil {
		log.Errorf("[updateVlanIdConfigByPlanId] update vlan id config error, %v", err)
		return err
	}
	return nil
}

func QueryRegionAzCellByPlanId(planId int64) (RegionAzCell, error) {
	var regionAzCell RegionAzCell
	if err := data.DB.Table(entity.PlanManageTable+" plan").Select("region.id as region_id, region.code as region_code, region.type as region_type, az.id as az_id, az.code as az_code, cell.id as cell_id, cell.type as cell_type, cell.name as cell_name").
		Joins("left join project_manage project on plan.project_id = project.id").
		Joins("left join region_manage region on project.region_id = region.id").
		Joins("left join az_manage az on project.az_id = az.id").
		Joins("left join cell_manage cell on project.cell_id = cell.id").
		Where("plan.id = ? and plan.delete_state = 0", planId).
		Find(&regionAzCell).Error; err != nil {
		log.Errorf("[queryRegionAzCellByPlanId] query region az cell error, %v", err)
		return regionAzCell, err
	}
	return regionAzCell, nil
}

func QueryCellConfigByPlanId(planId int64) (entity.CellConfig, error) {
	var cellConfig entity.CellConfig
	if err := data.DB.Table(entity.CellConfigTable).Where("plan_id = ?", planId).Find(&cellConfig).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("[queryCellConfigByPlanId] query vlan id config error, %v", err)
		return cellConfig, err
	}
	return cellConfig, nil
}

func InsertCellConfig(userId string, request CellConfigReq) error {
	now := time.Now()
	cellConfig := entity.CellConfig{
		PlanId:                   request.PlanId,
		BizRegionAbbr:            request.BizRegionAbbr,
		CellSelfMgt:              request.CellSelfMgt,
		MgtGlobalDnsRootDomain:   request.MgtGlobalDnsRootDomain,
		GlobalDnsSvcAddress:      request.GlobalDnsSvcAddress,
		CellVip:                  request.CellVip,
		CellVipIpv6:              request.CellVipIpv6,
		ExternalNtpIp:            request.ExternalNtpIp,
		NetworkMode:              request.NetworkMode,
		CellContainerNetwork:     request.CellContainerNetwork,
		CellContainerNetworkIpv6: request.CellContainerNetworkIpv6,
		CellSvcNetwork:           request.CellSvcNetwork,
		CellSvcNetworkIpv6:       request.CellSvcNetworkIpv6,
		AddCellNodeSshPublicKey:  request.AddCellNodeSshPublicKey,
		CreateUserId:             userId,
		CreateTime:               now,
		UpdateUserId:             userId,
		UpdateTime:               now,
	}
	if err := data.DB.Table(entity.CellConfigTable).Create(&cellConfig).Error; err != nil {
		log.Errorf("[insertCellConfigByPlanId] insert cell config error, %v", err)
		return err
	}
	return nil
}

func QueryCellConfigById(id int64) (entity.CellConfig, error) {
	var cellConfig entity.CellConfig
	if err := data.DB.Table(entity.CellConfigTable).Where("id = ?", id).Find(&cellConfig).Error; err != nil {
		log.Errorf("[queryCellConfigById] query cell config error, %v", err)
		return cellConfig, err
	}
	return cellConfig, nil
}

func UpdateCellConfigById(userId string, id int64, request CellConfigReq, originCellConfig entity.CellConfig) error {
	cellConfig := entity.CellConfig{
		Id:                       id,
		PlanId:                   request.PlanId,
		BizRegionAbbr:            request.BizRegionAbbr,
		CellSelfMgt:              request.CellSelfMgt,
		MgtGlobalDnsRootDomain:   request.MgtGlobalDnsRootDomain,
		GlobalDnsSvcAddress:      request.GlobalDnsSvcAddress,
		CellVip:                  request.CellVip,
		CellVipIpv6:              request.CellVipIpv6,
		ExternalNtpIp:            request.ExternalNtpIp,
		NetworkMode:              request.NetworkMode,
		CellContainerNetwork:     request.CellContainerNetwork,
		CellContainerNetworkIpv6: request.CellContainerNetworkIpv6,
		CellSvcNetwork:           request.CellSvcNetwork,
		CellSvcNetworkIpv6:       request.CellSvcNetworkIpv6,
		AddCellNodeSshPublicKey:  request.AddCellNodeSshPublicKey,
		CreateUserId:             originCellConfig.CreateUserId,
		CreateTime:               originCellConfig.CreateTime,
		UpdateUserId:             userId,
		UpdateTime:               time.Now(),
	}
	if err := data.DB.Table(entity.CellConfigTable).Save(&cellConfig).Error; err != nil {
		log.Errorf("[updateCellConfigByPlanId] update cell config error, %v", err)
		return err
	}
	return nil
}

func UpdateRegionAzCellByPlanId(planId int64, userId string, regionAzCell RegionAzCell) error {
	originRegionAzCell, err := QueryRegionAzCellByPlanId(planId)
	if err != nil {
		return err
	}
	now := datetime.GetNow()
	regionManage := entity.RegionManage{
		Id:           originRegionAzCell.RegionId,
		Code:         regionAzCell.RegionCode,
		Type:         regionAzCell.RegionType,
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
		Type:         regionAzCell.CellType,
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

func QueryRoutePlanningConfigByPlanId(planId int64) (entity.RoutePlanningConfig, error) {
	var routePlanningConfig entity.RoutePlanningConfig
	if err := data.DB.Table(entity.RoutePlanningConfigTable).Where("plan_id = ?", planId).Find(&routePlanningConfig).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("[queryRoutePlanningConfigByPlanId] query route planning config error, %v", err)
		return routePlanningConfig, err
	}
	return routePlanningConfig, nil
}

func InsertRoutePlanningConfig(userId string, request RoutePlanningConfigReq) error {
	now := time.Now()
	routePlanningConfig := ConvertRoutePlanningReq2Entity(userId, now, request)
	routePlanningConfig.CreateUserId = userId
	routePlanningConfig.CreateTime = now
	if err := data.DB.Table(entity.RoutePlanningConfigTable).Create(&routePlanningConfig).Error; err != nil {
		log.Errorf("[insertRoutePlanningConfigByPlanId] insert route planning config error, %v", err)
		return err
	}
	return nil
}

func QueryRoutePlanningConfigById(id int64) (entity.RoutePlanningConfig, error) {
	var routePlanningConfig entity.RoutePlanningConfig
	if err := data.DB.Table(entity.RoutePlanningConfigTable).Where("id = ?", id).Find(&routePlanningConfig).Error; err != nil {
		log.Errorf("[queryRoutePlanningConfigById] query route planning config error, %v", err)
		return routePlanningConfig, err
	}
	return routePlanningConfig, nil
}

func UpdateRoutePlanningConfigById(userId string, id int64, request RoutePlanningConfigReq, originRoutePlanningConfig entity.RoutePlanningConfig) error {
	routePlanningConfig := ConvertRoutePlanningReq2Entity(userId, time.Now(), request)
	routePlanningConfig.Id = id
	routePlanningConfig.CreateUserId = originRoutePlanningConfig.CreateUserId
	routePlanningConfig.CreateTime = originRoutePlanningConfig.CreateTime
	if err := data.DB.Table(entity.RoutePlanningConfigTable).Save(&routePlanningConfig).Error; err != nil {
		log.Errorf("[updateRoutePlanningConfigByPlanId] update route planning config error, %v", err)
		return err
	}
	return nil
}

func QueryLargeNetworkSegmentConfigByPlanId(planId int64) (entity.LargeNetworkSegmentConfig, error) {
	var largeNetworkSegmentConfig entity.LargeNetworkSegmentConfig
	if err := data.DB.Table(entity.LargeNetworkSegmentConfigTable).Where("plan_id = ?", planId).Find(&largeNetworkSegmentConfig).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("[queryLargeNetworkSegmentConfigByPlanId] query large network segment config error, %v", err)
		return largeNetworkSegmentConfig, err
	}
	return largeNetworkSegmentConfig, nil
}

func InsertLargeNetworkSegmentConfig(userId string, request LargeNetworkSegmentConfigReq) error {
	now := time.Now()
	largeNetworkSegmentConfig := entity.LargeNetworkSegmentConfig{
		PlanId:                         request.PlanId,
		StorageNetworkSegmentRoute:     request.StorageNetworkSegmentRoute,
		BizIntranetNetworkSegmentRoute: request.BizIntranetNetworkSegmentRoute,
		BizExternalLargeNetworkSegment: request.BizExternalLargeNetworkSegment,
		BmcNetworkSegmentRoute:         request.BmcNetworkSegmentRoute,
		CreateUserId:                   userId,
		CreateTime:                     now,
		UpdateUserId:                   userId,
		UpdateTime:                     now,
	}
	if err := data.DB.Table(entity.LargeNetworkSegmentConfigTable).Create(&largeNetworkSegmentConfig).Error; err != nil {
		log.Errorf("[insertLargeNetworkSegmentConfig] insert large network segment config error, %v", err)
		return err
	}
	return nil
}

func QueryLargeNetworkSegmentConfigById(id int64) (entity.LargeNetworkSegmentConfig, error) {
	var largeNetworkSegmentConfig entity.LargeNetworkSegmentConfig
	if err := data.DB.Table(entity.LargeNetworkSegmentConfigTable).Where("id = ?", id).Find(&largeNetworkSegmentConfig).Error; err != nil {
		log.Errorf("[queryLargeNetworkSegmentConfigById] query large network segment config error, %v", err)
		return largeNetworkSegmentConfig, err
	}
	return largeNetworkSegmentConfig, nil
}

func UpdateLargeNetworkSegmentConfigById(userId string, id int64, request LargeNetworkSegmentConfigReq, originLargeNetworkSegmentConfig entity.LargeNetworkSegmentConfig) error {
	largeNetworkSegmentConfig := entity.LargeNetworkSegmentConfig{
		Id:                             id,
		PlanId:                         request.PlanId,
		StorageNetworkSegmentRoute:     request.StorageNetworkSegmentRoute,
		BizIntranetNetworkSegmentRoute: request.BizIntranetNetworkSegmentRoute,
		BizExternalLargeNetworkSegment: request.BizExternalLargeNetworkSegment,
		BmcNetworkSegmentRoute:         request.BmcNetworkSegmentRoute,
		CreateUserId:                   originLargeNetworkSegmentConfig.CreateUserId,
		CreateTime:                     originLargeNetworkSegmentConfig.CreateTime,
		UpdateUserId:                   userId,
		UpdateTime:                     time.Now(),
	}
	if err := data.DB.Table(entity.LargeNetworkSegmentConfigTable).Save(&largeNetworkSegmentConfig).Error; err != nil {
		log.Errorf("[updateLargeNetworkSegmentConfigById] update large network segment config error, %v", err)
		return err
	}
	return nil
}
