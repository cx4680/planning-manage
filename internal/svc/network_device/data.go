package network_device

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"time"
)

type BoxTotalResponse struct {
	Count int64 `json:"count"`
}

func searchDevicePlanByPlanId(planId int64) (*entity.NetworkDevicePlanning, error) {
	var devicePlan entity.NetworkDevicePlanning
	if err := data.DB.Table(entity.NetworkDevicePlanningTable).Where("plan_id=?", planId).Scan(&devicePlan).Error; err != nil {
		log.Errorf("[searchDevicePlanByPlanId] query device plan error, %v", err)
		return nil, err
	}
	return &devicePlan, nil
}

func searchDeviceListByPlanId(planId int64) ([]entity.NetworkDevicePlanning, error) {
	var deviceList []entity.NetworkDevicePlanning
	if err := data.DB.Table(entity.NetworkDeviceListTable).Where("plan_id=? and delete_state = 0", planId).Scan(&deviceList).Error; err != nil {
		log.Errorf("[searchDeviceListByPlanId] query device list error, %v", err)
		return nil, err
	}
	return deviceList, nil
}

func createDevicePlan(request Request) error {
	networkPlan := entity.NetworkDevicePlanning{
		PlanId:                request.PlanId,
		Brand:                 request.Brand,
		ApplicationDispersion: request.ApplicationDispersion,
		AwsServerNum:          request.AwsServerNum,
		AwsBoxNum:             request.AwsBoxNum,
		TotalBoxNum:           request.TotalBoxNum,
		CreateTime:            time.Now(),
		UpdateTime:            time.Now(),
		Ipv6:                  request.Ipv6,
		NetworkModel:          request.NetworkModel,
		OpenDpdk:              request.OpenDpdk,
	}
	var devicePlan entity.NetworkDevicePlanning
	if err := data.DB.Table(entity.NetworkDevicePlanningTable).Create(&networkPlan).Scan(&devicePlan).Error; err != nil {
		log.Errorf("[createDevicePlan] insert db error", err)
		return err
	}
	return nil
}

func getBrandsByVersionIdAndNetworkVersion(versionId int64, networkVersion string) ([]string, error) {
	var deviceBaseline []entity.NetworkDeviceBaseline
	if err := data.DB.Table(entity.NetworkDeviceBaselineTable).Where("version_id=? and network_model = ?", versionId, networkVersion).Scan(&deviceBaseline).Error; err != nil {
		log.Errorf("[getBrandsByVersionIdAndNetworkVersion] query device brands error, %v", err)
		return nil, err
	}
	if len(deviceBaseline) == 0 {
		return nil, nil
	}
	var brands []string
	return brands, nil
}

func updateDevicePlan(request Request, devicePlanning entity.NetworkDevicePlanning) error {
	devicePlanning.UpdateTime = time.Now()
	devicePlanning.Brand = request.Brand
	devicePlanning.AwsServerNum = request.AwsServerNum
	devicePlanning.AwsBoxNum = request.AwsBoxNum
	devicePlanning.TotalBoxNum = request.TotalBoxNum
	devicePlanning.Ipv6 = request.Ipv6
	devicePlanning.NetworkModel = request.NetworkModel
	devicePlanning.OpenDpdk = request.OpenDpdk
	if err := data.DB.Table(entity.NetworkDevicePlanningTable).Updates(&devicePlanning).Error; err != nil {
		log.Errorf("[updateDevicePlan] update device planning error, %v", err)
		return err
	}
	return nil
}
