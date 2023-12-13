package network_device

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"errors"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"
	"strconv"
	"time"
)

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

func SaveBatch(tx *gorm.DB, networkDeviceList []*entity.NetworkDeviceList) error {
	if err := tx.Table(entity.NetworkDeviceListTable).Create(&networkDeviceList).Scan(&networkDeviceList).Error; err != nil {
		log.Errorf("batch insert networkDeviceList error: ", err)
		return err
	}
	return nil
}

func expireDeviceListByPlanId(tx *gorm.DB, planId int64) error {
	if err := tx.Model(entity.NetworkDeviceList{}).Where("plan_id = ?", planId).Update("delete_state", 1).Update("update_time", time.Now().Unix()).Error; err != nil {
		log.Errorf("[expireDeviceListByPlanId] expire device list error, %v", err)
		return err
	}
	return nil
}

func searchDeviceRoleBaselineByVersionId(versionId int64) ([]entity.NetworkDeviceRoleBaseline, error) {
	var deviceRoleBaselineList []entity.NetworkDeviceRoleBaseline
	if err := data.DB.Table(entity.NetworkDeviceRoleBaselineTable).Where("version_id=?", versionId).Scan(&deviceRoleBaselineList).Error; err != nil {
		log.Errorf("[searchDeviceRoleBaselineByVersionId] query device role baseline list error, %v", err)
		return nil, err
	}
	return deviceRoleBaselineList, nil
}

func searchModelRoleRelByRoleIdAndNetworkModel(roleId int64, networkModel int) ([]entity.NetworkModelRoleRel, error) {
	var modelRoleRel []entity.NetworkModelRoleRel
	if err := data.DB.Table(entity.NetworkModelRoleRelTable).Where("network_device_role_id = ? and network_model=?", roleId, networkModel).Scan(&modelRoleRel).Error; err != nil {
		log.Errorf("[searchModelRoleRelByRoleIdAndNetworkModel] error, %v", err)
		return nil, err
	}
	return modelRoleRel, nil
}

func createDevicePlan(request *Request) error {
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
		DeviceType:            request.DeviceType,
	}
	if err := data.DB.Table(entity.NetworkDevicePlanningTable).Create(&networkPlan).Error; err != nil {
		log.Errorf("[createDevicePlan] insert db error", err)
		return err
	}
	return nil
}

func getBrandsByVersionIdAndNetworkVersion(versionId int64, networkVersion string) ([]string, error) {
	var brands []string
	if err := data.DB.Raw("select distinct manufacturer from network_device_baseline where version_id=? and network_model = ?", versionId, networkVersion).Scan(&brands).Error; err != nil {
		log.Errorf("[getBrandsByVersionIdAndNetworkVersion] query device brands error, %v", err)
		return nil, err
	}
	return brands, nil
}

func getDeviceRoleGroupNumByPlanId(tx *gorm.DB, planId int64) ([]*DeviceRoleGroupNum, error) {
	var roleNum []*DeviceRoleGroupNum
	if err := tx.Raw("SELECT count(DISTINCT logical_grouping) as groupNum,network_device_role_id FROM network_device_list where plan_id=? GROUP BY network_device_role_id", planId).Scan(&roleNum).Error; err != nil {
		log.Errorf("[getDeviceRoleGroupNumByPlanId] error, %v", err)
		return nil, err
	}
	return roleNum, nil
}

func getModelsByVersionIdAndRoleAndBrandAndNetworkConfig(versionId int64, networkInterface string, id int64, brand string, deviceType int) ([]NetworkDeviceModel, error) {
	var deviceModel []NetworkDeviceModel
	if err := data.DB.Raw("select a.device_model,a.conf_overview from network_device_baseline a left join network_device_role_rel b on a.id = b.device_id where a.version_id = ? and b.device_role_id = ? and a.network_model = ? and a.manufacturer = ? and a.device_type = ?", versionId, id, networkInterface, brand, deviceType).Scan(&deviceModel).Error; err != nil {
		log.Errorf("[getModelsByVersionIdAndRoleAndBrandAndNetworkConfig] query device model error, %v", err)
		return nil, err
	}
	return deviceModel, nil
}

func updateDevicePlan(request *Request, devicePlanning entity.NetworkDevicePlanning) error {
	devicePlanning.UpdateTime = time.Now()
	devicePlanning.Brand = request.Brand
	devicePlanning.AwsServerNum = request.AwsServerNum
	devicePlanning.AwsBoxNum = request.AwsBoxNum
	devicePlanning.TotalBoxNum = request.TotalBoxNum
	devicePlanning.Ipv6 = request.Ipv6
	devicePlanning.NetworkModel = request.NetworkModel
	devicePlanning.DeviceType = request.DeviceType
	devicePlanning.ApplicationDispersion = request.ApplicationDispersion
	if err := data.DB.Table(entity.NetworkDevicePlanningTable).Updates(&devicePlanning).Error; err != nil {
		log.Errorf("[updateDevicePlan] update device planning error, %v", err)
		return err
	}
	return nil
}

func exportNetworkDeviceListByPlanId(planId int64) (string, []NetworkDeviceListExportResponse, error) {
	var planManage entity.PlanManage
	if err := data.DB.Table(entity.PlanManageTable).Where("id=? and delete_state = 0", planId).Scan(&planManage).Error; err != nil {
		log.Errorf("[exportNetworkDeviceListByPlanId] get planManage by id err, %v", err)
		return "", nil, err
	}
	var projectManage entity.ProjectManage
	if err := data.DB.Table(entity.ProjectManageTable).Where("id=? and delete_state = 0", planManage.ProjectId).Scan(&projectManage).Error; err != nil {
		log.Errorf("[exportNetworkDeviceListByPlanId] get projectManage by id err, %v", err)
		return "", nil, err
	}
	var roleIdNum []NetworkDeviceRoleIdNum
	if err := data.DB.Raw("select network_device_role_id,count(*) as num from network_device_list where plan_id = ? and delete_state = 0 group by network_device_role_id", planId).Scan(&roleIdNum).Error; err != nil {
		log.Errorf("[exportNetworkDeviceListByPlanId] query db error, %v", err)
		return "", nil, err
	}
	if len(roleIdNum) == 0 {
		return "", nil, errors.New("获取网络设备清单为空")
	}
	var response []NetworkDeviceListExportResponse
	for _, roleNum := range roleIdNum {
		roleId := roleNum.NetworkDeviceRoleId
		networkDevice, _ := getNetworkDeviceListByPlanIdAndRoleId(planId, roleId)
		var exportDto = NetworkDeviceListExportResponse{
			networkDevice.NetworkDeviceRoleName,
			networkDevice.NetworkDeviceRole,
			networkDevice.Brand,
			networkDevice.DeviceModel,
			networkDevice.ConfOverview,
			strconv.Itoa(roleNum.Num),
		}
		response = append(response, exportDto)
	}
	return projectManage.Name + "-" + planManage.Name + "-" + "网络设备清单", response, nil
}

func getNetworkDeviceListByPlanIdAndRoleId(planId int64, roleId int64) (entity.NetworkDeviceList, error) {
	var networkDevice entity.NetworkDeviceList
	if err := data.DB.Table(entity.NetworkDeviceListTable).Where("plan_id=? and delete_state = 0 and network_device_role_id = ? limit 1", planId, roleId).Scan(&networkDevice).Error; err != nil {
		log.Errorf("[getNetworkDeviceListByPlanIdAndRoleId] query db error, %v", err)
		return networkDevice, err
	}
	return networkDevice, nil
}
