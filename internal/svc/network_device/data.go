package network_device

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"errors"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func searchDevicePlanByPlanId(planId int64) (*entity.NetworkDevicePlanning, error) {
	var devicePlan entity.NetworkDevicePlanning
	if err := data.DB.Where("plan_id = ?", planId).Find(&devicePlan).Error; err != nil {
		log.Errorf("[searchDevicePlanByPlanId] query device plan error, %v", err)
		return nil, err
	}
	return &devicePlan, nil
}

func searchDeviceListByPlanId(planId int64) ([]entity.NetworkDeviceList, error) {
	var deviceList []entity.NetworkDeviceList
	if err := data.DB.Where("plan_id = ? AND delete_state = 0", planId).Find(&deviceList).Error; err != nil {
		log.Errorf("[searchDeviceListByPlanId] query device list error, %v", err)
		return nil, err
	}
	return deviceList, nil
}

func SaveBatch(tx *gorm.DB, networkDeviceList []*entity.NetworkDeviceList) error {
	if err := tx.Create(&networkDeviceList).Error; err != nil {
		log.Errorf("batch insert networkDeviceList error: ", err)
		return err
	}
	return nil
}

func expireDeviceListByPlanId(tx *gorm.DB, planId int64) error {
	if err := tx.Model(&entity.NetworkDeviceList{}).Where("plan_id = ?", planId).Updates(map[string]interface{}{"delete_state": 1, "update_time": time.Now()}).Error; err != nil {
		log.Errorf("[expireDeviceListByPlanId] expire device list error, %v", err)
		return err
	}
	return nil
}

func searchDeviceRoleBaselineByVersionId(versionId int64) ([]entity.NetworkDeviceRoleBaseline, error) {
	var deviceRoleBaselineList []entity.NetworkDeviceRoleBaseline
	if err := data.DB.Where("version_id = ?", versionId).Find(&deviceRoleBaselineList).Error; err != nil {
		log.Errorf("[searchDeviceRoleBaselineByVersionId] query device role baseline list error, %v", err)
		return nil, err
	}
	return deviceRoleBaselineList, nil
}

func searchModelRoleRelByRoleIdAndNetworkModel(roleId int64, networkModel int) ([]entity.NetworkModelRoleRel, error) {
	var modelRoleRel []entity.NetworkModelRoleRel
	if err := data.DB.Where("network_device_role_id = ? AND network_model = ?", roleId, networkModel).Find(&modelRoleRel).Error; err != nil {
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
	if err := data.DB.Create(&networkPlan).Error; err != nil {
		log.Errorf("[createDevicePlan] insert db error", err)
		return err
	}
	return nil
}

func getBrandsByVersionId(versionId int64) ([]string, error) {
	var brands []string
	if err := data.DB.Model(&entity.NetworkDeviceBaseline{}).Distinct("manufacturer").Where("version_id = ?", versionId).Find(&brands).Error; err != nil {
		log.Errorf("[getBrandsByVersionIdAndNetworkVersion] query device brands error, %v", err)
		return nil, err
	}
	return brands, nil
}

func getDeviceRoleGroupNumByPlanId(tx *gorm.DB, planId int64) ([]*DeviceRoleGroupNum, error) {
	var roleNum []*DeviceRoleGroupNum
	if err := tx.Table(entity.NetworkDeviceListTable).Select("count(DISTINCT logical_grouping) as groupNum", "network_device_role_id").
		Where("plan_id = ?", planId).Group("network_device_role_id").Find(&roleNum).Error; err != nil {
		log.Errorf("[getDeviceRoleGroupNumByPlanId] error, %v", err)
		return nil, err
	}
	return roleNum, nil
}

func getModelsByVersionIdAndRoleAndBrand(versionId int64, id int64, brand string, deviceType int) ([]NetworkDeviceModel, error) {
	var deviceModel []NetworkDeviceModel
	if err := data.DB.Table(entity.NetworkDeviceBaselineTable+" a").Select("a.device_model", "a.conf_overview").
		Joins("left join network_device_role_rel b on a.id = b.device_id").
		Where("a.version_id = ? and b.device_role_id = ? and a.manufacturer = ? and a.device_type = ?", versionId, id, brand, deviceType).
		Find(&deviceModel).Error; err != nil {
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
	if err := data.DB.Save(&devicePlanning).Error; err != nil {
		log.Errorf("[updateDevicePlan] update device planning error, %v", err)
		return err
	}
	return nil
}

func exportNetworkDeviceListByPlanId(planId int64) (string, []NetworkDeviceListExportResponse, error) {
	var planManage entity.PlanManage
	if err := data.DB.Where("id = ? and delete_state = 0", planId).Find(&planManage).Error; err != nil {
		log.Errorf("[exportNetworkDeviceListByPlanId] get planManage by id err, %v", err)
		return "", nil, err
	}
	var projectManage entity.ProjectManage
	if err := data.DB.Where("id = ? and delete_state = 0", planManage.ProjectId).Find(&projectManage).Error; err != nil {
		log.Errorf("[exportNetworkDeviceListByPlanId] get projectManage by id err, %v", err)
		return "", nil, err
	}
	var roleIdNum []NetworkDeviceRoleIdNum
	if err := data.DB.Table(entity.NetworkDeviceListTable).Select("network_device_role_id", "count(*) as num").
		Where("plan_id = ? AND delete_state = 0", planId).Group("network_device_role_id").Find(&roleIdNum).Error; err != nil {
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
	if err := data.DB.Where("plan_id = ? AND delete_state = 0 AND network_device_role_id = ?", planId, roleId).Find(&networkDevice).Error; err != nil {
		log.Errorf("[getNetworkDeviceListByPlanIdAndRoleId] query db error, %v", err)
		return networkDevice, err
	}
	return networkDevice, nil
}

func getNetworkShelveList(planId int64) ([]*entity.NetworkDeviceShelve, error) {
	var networkDeviceShelve []*entity.NetworkDeviceShelve
	if err := data.DB.Where("plan_id = ?", planId).Find(&networkDeviceShelve).Error; err != nil {
		return nil, err
	}
	return networkDeviceShelve, nil
}

func getNetworkShelveDownloadList(planId int64) ([]NetworkDeviceShelveDownload, string, error) {
	var networkDeviceList []*entity.NetworkDeviceList
	if err := data.DB.Where("plan_id = ?", planId).Find(&networkDeviceList).Error; err != nil {
		return nil, "", err
	}
	if len(networkDeviceList) == 0 {
		return nil, "", errors.New("网络设备未规划")
	}
	//构建返回体
	var response []NetworkDeviceShelveDownload
	for _, v := range networkDeviceList {
		response = append(response, NetworkDeviceShelveDownload{
			DeviceLogicalId: v.LogicalGrouping,
			DeviceId:        v.DeviceId,
		})
	}
	//构建文件名称
	var planManage = &entity.PlanManage{}
	if err := data.DB.Where("id = ? AND delete_state = ?", planId, 0).Find(&planManage).Error; err != nil {
		return nil, "", err
	}
	if planManage.Id == 0 {
		return nil, "", errors.New("方案不存在")
	}
	var projectManage = &entity.ProjectManage{}
	if err := data.DB.Where("id = ? AND delete_state = ?", planManage.ProjectId, 0).First(&projectManage).Error; err != nil {
		return nil, "", err
	}
	fileName := projectManage.Name + "-" + planManage.Name + "-" + "网络设备上架表"
	return response, fileName, nil
}

func uploadNetworkShelve(planId int64, networkDeviceShelveDownload []NetworkDeviceShelveDownload, userId string) error {
	if len(networkDeviceShelveDownload) == 0 {
		return errors.New("数据为空")
	}
	now := datetime.GetNow()
	var networkDeviceShelveList []*entity.NetworkDeviceShelve
	for _, v := range networkDeviceShelveDownload {
		if err := checkNetworkShelve(&v); err != nil {
			return err
		}
		networkDeviceShelveList = append(networkDeviceShelveList, &entity.NetworkDeviceShelve{
			PlanId:            planId,
			DeviceLogicalId:   v.DeviceLogicalId,
			DeviceId:          v.DeviceId,
			Sn:                v.Sn,
			MachineRoomAbbr:   v.MachineRoomAbbr,
			MachineRoomNumber: v.MachineRoomNumber,
			CabinetNumber:     v.CabinetNumber,
			SlotPosition:      v.SlotPosition,
			UNumber:           v.UNumber,
			CreateUserId:      userId,
			CreateTime:        now,
		})
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&entity.NetworkDeviceShelve{}, "plan_id = ?", planId).Error; err != nil {
			return err
		}
		if err := tx.CreateInBatches(&networkDeviceShelveList, 10).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func saveNetworkShelve(request *Request) error {
	if err := data.DB.Updates(&entity.PlanManage{Id: request.PlanId, DeliverPlanStage: constant.DeliverPlanningServer}).Error; err != nil {
		return err
	}
	return nil
}

func checkNetworkShelve(networkDeviceShelve *NetworkDeviceShelveDownload) error {
	if util.IsBlank(networkDeviceShelve.DeviceLogicalId) || util.IsBlank(networkDeviceShelve.DeviceId) || util.IsBlank(networkDeviceShelve.Sn) ||
		util.IsBlank(networkDeviceShelve.MachineRoomAbbr) || util.IsBlank(networkDeviceShelve.MachineRoomNumber) || util.IsBlank(networkDeviceShelve.CabinetNumber) ||
		util.IsBlank(networkDeviceShelve.SlotPosition) || networkDeviceShelve.UNumber == 0 {
		return errors.New("表单所有参数不能为空")
	}
	return nil
}
