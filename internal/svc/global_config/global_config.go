package global_config

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/user"
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
		AzCode:                   regionAzCell.AzCode,
		CellType:                 regionAzCell.CellType,
		CellName:                 regionAzCell.CellName,
		CellSelfMgt:              cellConfig.CellSelfMgt,
		MgtGlobalDnsRootDomain:   cellConfig.MgtGlobalDnsRootDomain,
		GlobalDnsSvcAddress:      cellConfig.GlobalDnsSvcAddress,
		CellVip:                  cellConfig.CellVip,
		CellVipIpv6:              cellConfig.CellVipIpv6,
		ExternalNtpIp:            cellConfig.ExternalNtpIp,
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
