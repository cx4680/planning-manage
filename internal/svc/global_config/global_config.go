package global_config

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/excel"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/user"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/config_item"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/network_device"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/plan"
)

func GetVlanIdConfigByPlanId(context *gin.Context) {
	planId, err := strconv.ParseInt(context.Param("planId"), 10, 64)
	if err != nil {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	vlanIdConfig, err := QueryVlanIdConfigByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, vlanIdConfig)
	return
}

func CreateVlanIdConfig(context *gin.Context) {
	var request VlanIdConfigRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		log.Errorf("create vlan id config bind param error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	originVlanIdConfig, err := QueryVlanIdConfigByPlanId(request.PlanId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if originVlanIdConfig.Id != 0 {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	userId := user.GetUserId(context)
	if err = InsertVlanIdConfig(userId, request); err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, nil)
	return
}

func UpdateVlanIdConfig(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	var request VlanIdConfigRequest
	if err = context.ShouldBindJSON(&request); err != nil {
		log.Errorf("update vlan id config bind param error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	originVlanIdConfig, err := QueryVlanIdConfigById(id)
	if err == gorm.ErrRecordNotFound || request.PlanId != originVlanIdConfig.PlanId {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	userId := user.GetUserId(context)
	if err = UpdateVlanIdConfigById(userId, id, request, originVlanIdConfig); err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, nil)
	return
}

func GetCellConfigByPlanId(context *gin.Context) {
	planId, err := strconv.ParseInt(context.Param("planId"), 10, 64)
	if err != nil {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	cellConfig, err := QueryCellConfigByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	regionAzCell, err := QueryRegionAzCellByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, CellConfigResp{
		Id:                       cellConfig.Id,
		PlanId:                   cellConfig.PlanId,
		RegionCode:               regionAzCell.RegionCode,
		RegionType:               regionAzCell.RegionType,
		BizRegionAbbr:            cellConfig.BizRegionAbbr,
		AzCode:                   regionAzCell.AzCode,
		CellType:                 regionAzCell.CellType,
		CellName:                 regionAzCell.CellName,
		CellSelfMgt:              cellConfig.CellSelfMgt,
		MgtGlobalDnsRootDomain:   cellConfig.MgtGlobalDnsRootDomain,
		GlobalDnsSvcAddress:      cellConfig.GlobalDnsSvcAddress,
		CellVip:                  cellConfig.CellVip,
		CellVipIpv6:              cellConfig.CellVipIpv6,
		ExternalNtpIp:            cellConfig.ExternalNtpIp,
		NetworkMode:              cellConfig.NetworkMode,
		CellContainerNetwork:     cellConfig.CellContainerNetwork,
		CellContainerNetworkIpv6: cellConfig.CellContainerNetworkIpv6,
		CellSvcNetwork:           cellConfig.CellSvcNetwork,
		CellSvcNetworkIpv6:       cellConfig.CellSvcNetworkIpv6,
		AddCellNodeSshPublicKey:  cellConfig.AddCellNodeSshPublicKey,
	})
	return
}

func CreateCellConfig(context *gin.Context) {
	var request CellConfigReq
	if err := context.ShouldBindJSON(&request); err != nil {
		log.Errorf("create cell config bind param error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	originCellConfig, err := QueryCellConfigByPlanId(request.PlanId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if originCellConfig.Id != 0 {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	userId := user.GetUserId(context)
	regionAzCellReq := RegionAzCell{
		RegionCode: request.RegionCode,
		RegionType: request.RegionType,
		AzCode:     request.AzCode,
		CellType:   request.CellType,
		CellName:   request.CellName,
	}
	if err = UpdateRegionAzCellByPlanId(request.PlanId, userId, regionAzCellReq); err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if err = InsertCellConfig(userId, request); err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, nil)
	return
}

func UpdateCellConfig(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	var request CellConfigReq
	if err = context.ShouldBindJSON(&request); err != nil {
		log.Errorf("update cell config bind param error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	originCellConfig, err := QueryCellConfigById(id)
	if err == gorm.ErrRecordNotFound || request.PlanId != originCellConfig.PlanId {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	userId := user.GetUserId(context)
	regionAzCellReq := RegionAzCell{
		RegionCode: request.RegionCode,
		RegionType: request.RegionType,
		AzCode:     request.AzCode,
		CellType:   request.CellType,
		CellName:   request.CellName,
	}
	if err = UpdateRegionAzCellByPlanId(request.PlanId, userId, regionAzCellReq); err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if err = UpdateCellConfigById(userId, id, request, originCellConfig); err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, nil)
	return
}

func GetRoutePlanningConfigByPlanId(context *gin.Context) {
	planId, err := strconv.ParseInt(context.Param("planId"), 10, 64)
	if err != nil {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	routePlanningConfig, err := QueryRoutePlanningConfigByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, routePlanningConfig)
	return
}

func CreateRoutePlanningConfig(context *gin.Context) {
	var request RoutePlanningConfigReq
	if err := context.ShouldBindJSON(&request); err != nil {
		log.Errorf("create route planning config bind param error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	originRoutePlanningConfig, err := QueryRoutePlanningConfigByPlanId(request.PlanId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if originRoutePlanningConfig.Id != 0 {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	userId := user.GetUserId(context)
	if err = InsertRoutePlanningConfig(userId, request); err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, nil)
	return
}

func UpdateRoutePlanningConfig(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	var request RoutePlanningConfigReq
	if err = context.ShouldBindJSON(&request); err != nil {
		log.Errorf("update route planning config bind param error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	originRoutePlanningConfig, err := QueryRoutePlanningConfigById(id)
	if err == gorm.ErrRecordNotFound || request.PlanId != originRoutePlanningConfig.PlanId {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	userId := user.GetUserId(context)
	if err = UpdateRoutePlanningConfigById(userId, id, request, originRoutePlanningConfig); err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, nil)
	return
}

func GetLargeNetworkConfigByPlanId(context *gin.Context) {
	planId, err := strconv.ParseInt(context.Param("planId"), 10, 64)
	if err != nil {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	largeNetworkSegmentConfig, err := QueryLargeNetworkSegmentConfigByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, largeNetworkSegmentConfig)
	return
}

func CreateLargeNetworkConfig(context *gin.Context) {
	var request LargeNetworkSegmentConfigReq
	if err := context.ShouldBindJSON(&request); err != nil {
		log.Errorf("create large network segment config bind param error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	originLargeNetworkSegmentConfig, err := QueryLargeNetworkSegmentConfigByPlanId(request.PlanId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if originLargeNetworkSegmentConfig.Id != 0 {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	userId := user.GetUserId(context)
	if err = InsertLargeNetworkSegmentConfig(userId, request); err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, nil)
	return
}

func UpdateLargeNetworkConfig(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	var request LargeNetworkSegmentConfigReq
	if err = context.ShouldBindJSON(&request); err != nil {
		log.Errorf("update large network segment config bind param error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	originLargeNetworkSegmentConfig, err := QueryLargeNetworkSegmentConfigById(id)
	if err == gorm.ErrRecordNotFound || request.PlanId != originLargeNetworkSegmentConfig.PlanId {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	userId := user.GetUserId(context)
	if err = UpdateLargeNetworkSegmentConfigById(userId, id, request, originLargeNetworkSegmentConfig); err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, nil)
	return
}

func CompleteGlobalConfig(context *gin.Context) {
	planId, err := strconv.ParseInt(context.Param("planId"), 10, 64)
	if err != nil {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if CheckParamCompleteGlobalConfig(context, planId) {
		return
	}
	userId := user.GetUserId(context)
	if err = plan.UpdatePlanStage(data.DB, planId, constant.PlanStageDelivered, userId, constant.DeliverPlanningEnd); err != nil {
		log.Errorf("[CompleteGlobalConfig] complete global config error, %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, nil)
	return
}

func CheckParamCompleteGlobalConfig(context *gin.Context, planId int64) bool {
	vlanIdConfig, err := QueryVlanIdConfigByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return true
	}
	if vlanIdConfig.Id == 0 {
		result.FailureWithMsg(context, errorcodes.InvalidParam, http.StatusBadRequest, errorcodes.VlanIdConfigEmpty)
		return true
	}
	cellConfig, err := QueryCellConfigByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return true
	}
	if cellConfig.Id == 0 {
		result.FailureWithMsg(context, errorcodes.InvalidParam, http.StatusBadRequest, errorcodes.CellConfigEmpty)
		return true
	}
	routePlanningConfig, err := QueryRoutePlanningConfigByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return true
	}
	if routePlanningConfig.Id == 0 {
		result.FailureWithMsg(context, errorcodes.InvalidParam, http.StatusBadRequest, errorcodes.RoutePlanningConfigEmpty)
		return true
	}
	largeNetworkSegmentConfig, err := QueryLargeNetworkSegmentConfigByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return true
	}
	if largeNetworkSegmentConfig.Id == 0 {
		result.FailureWithMsg(context, errorcodes.InvalidParam, http.StatusBadRequest, errorcodes.LargeNetworkSegmentConfigEmpty)
		return true
	}
	return false
}

func DownloadPlanningFile(context *gin.Context) {
	planId, err := strconv.ParseInt(context.Param("planId"), 10, 64)
	if err != nil {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	file, err := excelize.OpenFile("template/规划文件模版.xlsx")
	if err != nil {
		log.Errorf("open planning file template error: %v", err)
		if err = file.Close(); err != nil {
			log.Errorf("excelize close error: %v", err)
		}
		return
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Errorf("excelize close error: %v", err)
		}
	}()
	globalConfigExcel, err := ConvertGlobalConfigExcel(context, planId)
	if err != nil {
		return
	}
	excelFile := excel.Excel{F: file}
	if err = excel.ExportExcelByAssignCell("Sheet3", "", false, *globalConfigExcel, &excelFile); err != nil {
		log.Errorf("export excel error: %v", err)
		return
	}
	networkDeviceIps, err := QueryNetworkDeviceIpByPlanId(planId)
	if err != nil {
		log.Errorf("query network device ip by plan id error: %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	netWorkShelves, err := network_device.GetNetworkShelveList(planId)
	if err != nil {
		log.Errorf("query network device shelve list error: %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	networkShelveMap := make(map[string]entity.NetworkDeviceShelve)
	for _, netWorkShelf := range netWorkShelves {
		_, ok := networkShelveMap[netWorkShelf.DeviceLogicalId]
		if !ok {
			networkShelveMap[netWorkShelf.DeviceLogicalId] = *netWorkShelf
		}
	}
	var globalConfigNetworkDeviceExcels []GlobalConfigNetworkDeviceExcel
	for _, networkDeviceIp := range networkDeviceIps {
		logicalGrouping := networkDeviceIp.LogicalGrouping
		networkDeviceShelve := networkShelveMap[logicalGrouping]
		globalConfigNetworkDeviceExcels = append(globalConfigNetworkDeviceExcels, GlobalConfigNetworkDeviceExcel{
			LogicalGrouping:            logicalGrouping,
			MachineRoomAbbr:            networkDeviceShelve.MachineRoomAbbr,
			CabinetNum:                 networkDeviceShelve.CabinetNumber,
			SlotNum:                    networkDeviceShelve.SlotPosition,
			CabinetAsw:                 "",
			PxeSubnet:                  networkDeviceIp.PxeSubnet,
			PxeSubnetRange:             networkDeviceIp.PxeSubnetRange,
			PxeNetworkGateway:          networkDeviceIp.PxeNetworkGateway,
			ManageSubnet:               networkDeviceIp.ManageSubnet,
			ManageNetworkGateway:       networkDeviceIp.ManageNetworkGateway,
			ManageIpv6Subnet:           networkDeviceIp.ManageIpv6Subnet,
			ManageIpv6NetworkGateway:   networkDeviceIp.ManageIpv6NetworkGateway,
			BizSubnet:                  networkDeviceIp.BizSubnet,
			BizNetworkGateway:          networkDeviceIp.BizNetworkGateway,
			StorageFrontNetwork:        networkDeviceIp.StorageFrontNetwork,
			StorageFrontNetworkGateway: networkDeviceIp.StorageFrontNetworkGateway,
		})
	}
	if err = excel.ExportExcelByExistHeader("Sheet1", "", 3, false, globalConfigNetworkDeviceExcels, &excelFile); err != nil {
		log.Errorf("export excel error: %v", err)
		return
	}
	// if err = excelFile.F.SaveAs("/Users/blue/Desktop/规划文件模板.xlsx"); err != nil {
	// 	log.Errorf("excelize save error: %v", err)
	// 	return
	// }
	excel.DownLoadExcel("规划文件", context.Writer, file)
	return
}

func ConvertGlobalConfigExcel(context *gin.Context, planId int64) (*GlobalConfigExcel, error) {
	vlanIdConfig, err := QueryVlanIdConfigByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return nil, err
	}
	if vlanIdConfig.Id == 0 {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return nil, err
	}
	cellConfig, err := QueryCellConfigByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return nil, err
	}
	if cellConfig.Id == 0 {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return nil, err
	}
	regionAzCell, err := QueryRegionAzCellByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return nil, err
	}
	regionType := regionAzCell.RegionType
	cellType := regionAzCell.CellType
	regionTypeList, err := config_item.ListConfigItem(constant.RegionTypeCode)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return nil, err
	}
	for _, configItem := range regionTypeList {
		if configItem.Code == regionType {
			regionType = configItem.Name
			break
		}
	}
	cellTypeList, err := config_item.ListConfigItem(constant.CellTypeCode)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return nil, err
	}
	for _, configItem := range cellTypeList {
		if configItem.Code == cellType {
			cellType = configItem.Name
			break
		}
	}
	cellSelfMgt := constant.NoCn
	if cellConfig.CellSelfMgt == constant.Yes {
		cellSelfMgt = constant.YesCn
	}
	networkDevicePlanning, err := network_device.SearchDevicePlanByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return nil, err
	}
	dualStackDeploy := constant.Ipv6NoCn
	if networkDevicePlanning.Ipv6 == constant.Ipv6Yes {
		dualStackDeploy = constant.Ipv6YesCn
	}
	networkMode := constant.NetworkModeStandardCn
	if cellConfig.NetworkMode == constant.NetworkMode2Network {
		networkMode = constant.NetworkMode2NetworkCn
	}
	routePlanningConfig, err := QueryRoutePlanningConfigByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return nil, err
	}
	if routePlanningConfig.Id == 0 {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return nil, err
	}
	deployUseBgp := constant.NoCn
	if routePlanningConfig.DeployUseBgp == constant.Yes {
		deployUseBgp = constant.YesCn
	}
	largeNetworkSegmentConfig, err := QueryLargeNetworkSegmentConfigByPlanId(planId)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return nil, err
	}
	if largeNetworkSegmentConfig.Id == 0 {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return nil, err
	}
	globalConfigExcel := GlobalConfigExcel{
		InBandMgtVlanId:                vlanIdConfig.InBandMgtVlanId,
		LocalStorageVlanId:             vlanIdConfig.LocalStorageVlanId,
		BizIntranetVlanId:              vlanIdConfig.BizIntranetVlanId,
		RegionCode:                     regionAzCell.RegionCode,
		AzCode:                         regionAzCell.AzCode,
		RegionType:                     regionType,
		CellType:                       cellType,
		CellSelfMgt:                    cellSelfMgt,
		CellName:                       regionAzCell.CellName,
		MgtGlobalDnsRootDomain:         cellConfig.MgtGlobalDnsRootDomain,
		GlobalDnsSvcAddress:            cellConfig.GlobalDnsSvcAddress,
		DualStackDeploy:                dualStackDeploy,
		CellVip:                        cellConfig.CellVip,
		CellVipIpv6:                    cellConfig.CellVipIpv6,
		ExternalNtpIp:                  cellConfig.ExternalNtpIp,
		NetworkMode:                    networkMode,
		CellContainerNetwork:           cellConfig.CellContainerNetwork,
		CellContainerNetworkIpv6:       cellConfig.CellContainerNetworkIpv6,
		CellSvcNetwork:                 cellConfig.CellSvcNetwork,
		CellSvcNetworkIpv6:             cellConfig.CellSvcNetworkIpv6,
		AddCellNodeSshPublicKey:        cellConfig.AddCellNodeSshPublicKey,
		DeployUseBgp:                   deployUseBgp,
		DeployMachSwitchSelfNum:        routePlanningConfig.DeployMachSwitchSelfNum,
		DeployMachSwitchIp:             routePlanningConfig.DeployMachSwitchIp,
		SvcExternalAccessAddress:       routePlanningConfig.SvcExternalAccessAddress,
		BgpNeighbor:                    routePlanningConfig.BgpNeighbor,
		CellDnsSvcAddress:              routePlanningConfig.CellDnsSvcAddress,
		RegionDnsSvcAddress:            routePlanningConfig.RegionDnsSvcAddress,
		OpsCenterIp:                    routePlanningConfig.OpsCenterIp,
		OpsCenterIpv6:                  routePlanningConfig.OpsCenterIpv6,
		OpsCenterPort:                  routePlanningConfig.OpsCenterPort,
		OpsCenterDomain:                routePlanningConfig.OpsCenterDomain,
		OperationCenterIp:              routePlanningConfig.OperationCenterIp,
		OperationCenterIpv6:            routePlanningConfig.OperationCenterIpv6,
		OperationCenterPort:            routePlanningConfig.OperationCenterPort,
		OperationCenterDomain:          routePlanningConfig.OperationCenterDomain,
		OpsCenterInitUserName:          routePlanningConfig.OpsCenterInitUserName,
		OpsCenterInitUserPwd:           routePlanningConfig.OpsCenterInitUserPwd,
		OperationCenterInitUserName:    routePlanningConfig.OperationCenterInitUserName,
		OperationCenterInitUserPwd:     routePlanningConfig.OperationCenterInitUserPwd,
		StorageNetworkSegmentRoute:     largeNetworkSegmentConfig.StorageNetworkSegmentRoute,
		BizIntranetNetworkSegmentRoute: largeNetworkSegmentConfig.BizIntranetNetworkSegmentRoute,
		BizExternalLargeNetworkSegment: largeNetworkSegmentConfig.BizExternalLargeNetworkSegment,
		BmcNetworkSegmentRoute:         largeNetworkSegmentConfig.BmcNetworkSegmentRoute,
	}
	return &globalConfigExcel, nil
}
