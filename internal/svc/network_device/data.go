package network_device

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"

	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
)

func SearchDevicePlanByPlanId(planId int64) (*entity.NetworkDevicePlanning, error) {
	var devicePlan entity.NetworkDevicePlanning
	if err := data.DB.Where("plan_id = ?", planId).Find(&devicePlan).Error; err != nil {
		log.Errorf("[searchDevicePlanByPlanId] query device plan error, %v", err)
		return nil, err
	}
	return &devicePlan, nil
}

func SearchDeviceListByPlanId(planId int64) ([]entity.NetworkDeviceList, error) {
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

func SaveSoftwareBomPlanning(db *gorm.DB, planId int64) error {
	cloudProductPlanningList, cloudProductNodeRoleRelList, cloudProductBaselineMap, serverPlanningMap, serverBaselineMap, softwareBomLicenseBaselineListMap, err := getSoftwareBomPlanningData(db, planId)
	if err != nil {
		return err
	}
	var cpuNumber, serviceYearBomNumber int
	for _, v := range serverPlanningMap {
		switch v.CpuType {
		case "Intel", "Hygon", "Kunpeng":
			cpuNumber += v.Number * 2
		case "FT2000+":
			cpuNumber += v.Number
		}
	}
	var softwareBomPlanningList []*entity.SoftwareBomPlanning
	//平台规模授权：0100115148387809，按云平台下服务器数量计算，N=整网所有服务器的物理CPU数量之和-管理减免（10）；N大于等于0
	softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: constant.PlatformBom, Number: cpuNumber})
	//软件base：0100115150861886，默认1套
	softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: constant.SoftwareBaseBom, Number: 1})
	for _, v := range cloudProductNodeRoleRelList {
		cloudProductBaseline := cloudProductBaselineMap[v.ProductId]
		serverPlanning := serverPlanningMap[v.NodeRoleId]
		serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
		list := softwareBomLicenseBaselineListMap[fmt.Sprintf("%v-%v-%v", cloudProductBaseline.ProductCode, cloudProductBaseline.SellSpecs, serverBaseline.Arch)]
		if len(list) == 0 {
			//部分软件bom只有一个，不区分硬件架构
			list = softwareBomLicenseBaselineListMap[fmt.Sprintf("%v-%v-", cloudProductBaseline.ProductCode, cloudProductBaseline.SellSpecs)]
			if len(list) == 0 {
				continue
			}
		}
		for _, softwareBomLicenseBaseline := range list {
			cloudProductPlanningBom := &entity.SoftwareBomPlanning{
				PlanId:             planId,
				SoftwareBaselineId: softwareBomLicenseBaseline.Id,
				BomId:              softwareBomLicenseBaseline.BomId,
				Number:             1,
				CloudService:       softwareBomLicenseBaseline.CloudService,
				ServiceCode:        softwareBomLicenseBaseline.ServiceCode,
				SellSpecs:          softwareBomLicenseBaseline.SellSpecs,
				AuthorizedUnit:     softwareBomLicenseBaseline.AuthorizedUnit,
				SellType:           softwareBomLicenseBaseline.SellType,
				HardwareArch:       softwareBomLicenseBaseline.HardwareArch,
			}
			softwareBomPlanningList = append(softwareBomPlanningList, cloudProductPlanningBom)
			if softwareBomLicenseBaseline.SellType == "升级维保" {
				serviceYearBomNumber++
			}
		}
	}
	//平台升级维保：根据选择年限对应不同BOM
	softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: constant.ServiceYearBom[cloudProductPlanningList[0].ServiceYear], Number: serviceYearBomNumber})
	// 保存云产品规划bom表
	if err := db.Delete(&entity.SoftwareBomPlanning{}, "plan_id = ?", planId).Error; err != nil {
		return err
	}
	if err := db.Create(softwareBomPlanningList).Error; err != nil {
		return err
	}
	return nil
}

func getSoftwareBomPlanningData(db *gorm.DB, planId int64) ([]*entity.CloudProductPlanning, []*entity.CloudProductNodeRoleRel, map[int64]*entity.CloudProductBaseline, map[int64]*entity.ServerPlanning, map[int64]*entity.ServerBaseline, map[string][]*entity.SoftwareBomLicenseBaseline, error) {
	//查询云产品规划表
	var cloudProductPlanningList []*entity.CloudProductPlanning
	if err := db.Where("plan_id = ?", planId).Find(&cloudProductPlanningList).Error; err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	var productIdList []int64
	for _, v := range cloudProductPlanningList {
		productIdList = append(productIdList, v.ProductId)
	}
	//查询云产品和角色关联表
	var cloudProductNodeRoleRelList []*entity.CloudProductNodeRoleRel
	if err := db.Where("product_id IN (?)", productIdList).Find(&cloudProductNodeRoleRelList).Error; err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	//查询服务器规划
	var serverPlanningList []*entity.ServerPlanning
	if err := db.Where("plan_id = ?", planId).Find(&serverPlanningList).Error; err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	var serverPlanningMap = make(map[int64]*entity.ServerPlanning)
	var serverBaselineIdList []int64
	for _, v := range serverPlanningList {
		serverPlanningMap[v.NodeRoleId] = v
		serverBaselineIdList = append(serverBaselineIdList, v.ServerBaselineId)
	}
	//查询容量规划
	var serverCapPlanningList []*entity.ServerCapPlanning
	if err := db.Where("plan_id = ?", planId).Find(&serverCapPlanningList).Error; err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	var serverCapPlanningMap = make(map[int64]*entity.ServerCapPlanning)
	for _, v := range serverCapPlanningList {
		serverCapPlanningMap[v.NodeRoleId] = v
	}
	//查询服务器基线表
	var serverBaselineList []*entity.ServerBaseline
	if err := db.Where("id IN (?)", serverBaselineIdList).Find(&serverBaselineList).Error; err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	for _, v := range serverBaselineList {
		serverBaselineMap[v.Id] = v
	}
	//查询云产品基线
	var cloudProductBaselineList []*entity.CloudProductBaseline
	if err := db.Where("id IN (?)", productIdList).Find(&cloudProductBaselineList).Error; err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	var cloudProductBaselineMap = make(map[int64]*entity.CloudProductBaseline)
	for _, v := range cloudProductBaselineList {
		cloudProductBaselineMap[v.Id] = v
	}
	var productCodeList []string
	for _, v := range cloudProductBaselineList {
		productCodeList = append(productCodeList, v.ProductCode)
	}
	//查询软件bom表
	var softwareBomLicenseBaselineList []*entity.SoftwareBomLicenseBaseline
	if err := db.Where("service_code IN (?) AND version_id = ?", productCodeList, cloudProductPlanningList[0].VersionId).Find(&softwareBomLicenseBaselineList).Error; err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	var softwareBomLicenseBaselineListMap = make(map[string][]*entity.SoftwareBomLicenseBaseline)
	for _, v := range softwareBomLicenseBaselineList {
		if v.HardwareArch == "xc" {
			v.HardwareArch = "ARM"
		}
		softwareBomLicenseBaselineListMap[fmt.Sprintf("%v-%v-%v", v.ServiceCode, v.SellSpecs, v.HardwareArch)] = append(softwareBomLicenseBaselineListMap[fmt.Sprintf("%v-%v-%v", v.ServiceCode, v.SellSpecs, v.HardwareArch)], v)
	}
	return cloudProductPlanningList, cloudProductNodeRoleRelList, cloudProductBaselineMap, serverPlanningMap, serverBaselineMap, softwareBomLicenseBaselineListMap, nil
}

func ExpireDeviceListByPlanId(tx *gorm.DB, planId int64) error {
	if err := tx.Model(&entity.NetworkDeviceList{}).Where("plan_id = ?", planId).Updates(map[string]interface{}{"delete_state": 1, "update_time": time.Now()}).Error; err != nil {
		log.Errorf("[expireDeviceListByPlanId] expire device list error, %v", err)
		return err
	}
	return nil
}

func SearchDeviceRoleBaselineByVersionId(versionId int64) ([]entity.NetworkDeviceRoleBaseline, error) {
	var deviceRoleBaselineList []entity.NetworkDeviceRoleBaseline
	if err := data.DB.Where("version_id = ?", versionId).Find(&deviceRoleBaselineList).Error; err != nil {
		log.Errorf("[searchDeviceRoleBaselineByVersionId] query device role baseline list error, %v", err)
		return nil, err
	}
	return deviceRoleBaselineList, nil
}

func SearchModelRoleRelByRoleIdAndNetworkModel(roleId int64, networkModel int) ([]entity.NetworkModelRoleRel, error) {
	var modelRoleRel []entity.NetworkModelRoleRel
	if err := data.DB.Where("network_device_role_id = ? AND network_model = ?", roleId, networkModel).Find(&modelRoleRel).Error; err != nil {
		log.Errorf("[searchModelRoleRelByRoleIdAndNetworkModel] error, %v", err)
		return nil, err
	}
	return modelRoleRel, nil
}

func CreateDevicePlan(request *Request) error {
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

func GetBrandsByVersionId(versionId int64) ([]string, error) {
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
		Where("plan_id = ? and delete_state = 0", planId).Group("network_device_role_id").Find(&roleNum).Error; err != nil {
		log.Errorf("[getDeviceRoleGroupNumByPlanId] error, %v", err)
		return nil, err
	}
	return roleNum, nil
}

func GetDeviceRoleLogicGroupByPlanId(tx *gorm.DB, planId int64) ([]*DeviceRoleLogicGroup, error) {
	var logicGroups []*DeviceRoleLogicGroup
	if err := tx.Table(entity.NetworkDeviceListTable).Select("DISTINCT logical_grouping", "network_device_role_id").
		Where("plan_id = ? and delete_state = 0", planId).Find(&logicGroups).Error; err != nil {
		log.Errorf("[getDeviceRoleLogicGroupByPlanId] error, %v", err)
		return nil, err
	}
	return logicGroups, nil
}

func GetModelsByVersionIdAndRoleAndBrand(versionId int64, id int64, brand string, deviceType int) ([]NetworkDeviceModel, error) {
	var deviceModel []NetworkDeviceModel
	if err := data.DB.Table(entity.NetworkDeviceBaselineTable+" a").Select("a.device_model", "a.conf_overview", "a.bom_id").
		Joins("left join network_device_role_rel b on a.id = b.device_id").
		Where("a.version_id = ? and b.device_role_id = ? and a.manufacturer = ? and a.device_type = ?", versionId, id, brand, deviceType).
		Find(&deviceModel).Error; err != nil {
		log.Errorf("[getModelsByVersionIdAndRoleAndBrandAndNetworkConfig] query device model error, %v", err)
		return nil, err
	}
	return deviceModel, nil
}

func UpdateDevicePlan(request *Request, devicePlanning entity.NetworkDevicePlanning) error {
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

func ExportNetworkDeviceListByPlanId(planId int64) (string, []NetworkDeviceListExportResponse, error) {
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
		networkDevice, _ := GetNetworkDeviceListByPlanIdAndRoleId(planId, roleId)
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

func GetNetworkDeviceListByPlanIdAndRoleId(planId int64, roleId int64) (entity.NetworkDeviceList, error) {
	var networkDevice entity.NetworkDeviceList
	if err := data.DB.Where("plan_id = ? AND delete_state = 0 AND network_device_role_id = ?", planId, roleId).Find(&networkDevice).Error; err != nil {
		log.Errorf("[getNetworkDeviceListByPlanIdAndRoleId] query db error, %v", err)
		return networkDevice, err
	}
	return networkDevice, nil
}

func GetNetworkShelveList(planId int64) ([]*entity.NetworkDeviceShelve, error) {
	var networkDeviceShelve []*entity.NetworkDeviceShelve
	if err := data.DB.Where("plan_id = ?", planId).Find(&networkDeviceShelve).Error; err != nil {
		return nil, err
	}
	return networkDeviceShelve, nil
}

func GetDownloadNetworkShelveTemplate(planId int64) ([]NetworkDeviceShelveDownload, string, error) {
	var networkDeviceList []*entity.NetworkDeviceList
	if err := data.DB.Where("plan_id = ?", planId).Find(&networkDeviceList).Error; err != nil {
		return nil, "", err
	}
	if len(networkDeviceList) == 0 {
		return nil, "", errors.New("网络设备未规划")
	}
	// 构建返回体
	var response []NetworkDeviceShelveDownload
	for _, v := range networkDeviceList {
		response = append(response, NetworkDeviceShelveDownload{
			DeviceLogicalId: v.LogicalGrouping,
			DeviceId:        v.DeviceId,
		})
	}
	// 构建文件名称
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
	fileName := projectManage.Name + "-" + planManage.Name + "-" + "网络设备上架模板"
	return response, fileName, nil
}

func UploadNetworkShelve(planId int64, networkDeviceShelveDownload []NetworkDeviceShelveDownload, userId string) error {
	if len(networkDeviceShelveDownload) == 0 {
		return errors.New("数据为空")
	}
	// 网络设备列表
	var networkDeviceList []*entity.NetworkDeviceList
	if err := data.DB.Where("plan_id = ?", planId).Find(&networkDeviceList).Error; err != nil {
		return err
	}
	if len(networkDeviceList) == 0 {
		return errors.New("网络设备未规划")
	}
	var networkDeviceMap = make(map[string]*entity.NetworkDeviceList)
	for _, v := range networkDeviceList {
		networkDeviceMap[fmt.Sprintf("%v-%v", v.LogicalGrouping, v.DeviceId)] = v
	}
	// 查询机柜信息
	var cabinetInfoList []*entity.CabinetInfo
	if err := data.DB.Where("plan_id = ?", planId).Find(&cabinetInfoList).Error; err != nil {
		return err
	}
	var cabinetInfoMap = make(map[string]*entity.CabinetInfo)
	for _, v := range cabinetInfoList {
		cabinetInfoMap[fmt.Sprintf("%v-%v-%v-%v", v.CabinetAsw, v.MachineRoomAbbr, v.MachineRoomNum, v.CabinetNum)] = v
	}
	var networkDeviceShelveList []*entity.NetworkDeviceShelve
	for _, v := range networkDeviceShelveDownload {
		if util.IsBlank(v.DeviceLogicalId) || util.IsBlank(v.DeviceId) || util.IsBlank(v.Sn) || v.UNumber == 0 {
			return errors.New("表单所有参数不能为空")
		}
		key := fmt.Sprintf("%v-%v-%v-%v", v.DeviceLogicalId, v.MachineRoomAbbr, v.MachineRoomNumber, v.CabinetNumber)
		cabinetInfo := cabinetInfoMap[key]
		if cabinetInfo == nil {
			return errors.New("机柜信息错误：" + key)
		}
		networkDeviceShelveList = append(networkDeviceShelveList, &entity.NetworkDeviceShelve{
			PlanId:              planId,
			DeviceLogicalId:     v.DeviceLogicalId,
			DeviceId:            v.DeviceId,
			Sn:                  v.Sn,
			NetworkDeviceRoleId: networkDeviceMap[fmt.Sprintf("%v-%v", v.DeviceLogicalId, v.DeviceId)].NetworkDeviceRoleId,
			CabinetId:           cabinetInfo.Id,
			MachineRoomAbbr:     v.MachineRoomAbbr,
			MachineRoomNumber:   v.MachineRoomNumber,
			CabinetNumber:       v.CabinetNumber,
			SlotPosition:        v.SlotPosition,
			UNumber:             v.UNumber,
			CreateUserId:        userId,
			CreateTime:          datetime.GetNow(),
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

func SaveNetworkShelve(request *Request) error {
	// 查询网络设备上架表
	var cabinetIdList []string
	if err := data.DB.Model(&entity.NetworkDeviceShelve{}).Select("cabinet_id").Where("plan_id = ?", request.PlanId).Group("cabinet_id").Find(&cabinetIdList).Error; err != nil {
		return err
	}
	if len(cabinetIdList) == 0 {
		return errors.New("网络设备未上架")
	}
	// 查询机柜信息
	var cabinetCount int64
	if err := data.DB.Model(&entity.CabinetInfo{}).Where("id IN (?)", cabinetIdList).Count(&cabinetCount).Error; err != nil {
		return err
	}
	if int64(len(cabinetIdList)) != cabinetCount {
		return errors.New("机房信息已修改，请重新下载网络设备上架模板并上传")
	}
	if err := data.DB.Updates(&entity.PlanManage{Id: request.PlanId, DeliverPlanStage: constant.DeliverPlanningServer}).Error; err != nil {
		return err
	}
	return nil
}

func GetDownloadNetworkShelve(planId int64) ([]NetworkDeviceShelveDownload, string, error) {
	var networkDeviceShelve []*entity.NetworkDeviceShelve
	if err := data.DB.Where("plan_id = ?", planId).Find(&networkDeviceShelve).Error; err != nil {
		return nil, "", err
	}
	if len(networkDeviceShelve) == 0 {
		return nil, "", errors.New("网络设备未上架")
	}
	// 构建返回体
	var response []NetworkDeviceShelveDownload
	for _, v := range networkDeviceShelve {
		response = append(response, NetworkDeviceShelveDownload{
			DeviceLogicalId:   v.DeviceLogicalId,
			DeviceId:          v.DeviceId,
			Sn:                v.Sn,
			MachineRoomAbbr:   v.MachineRoomAbbr,
			MachineRoomNumber: v.MachineRoomNumber,
			CabinetNumber:     v.CabinetNumber,
			SlotPosition:      v.SlotPosition,
			UNumber:           v.UNumber,
		})
	}
	// 构建文件名称
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
	fileName := projectManage.Name + "-" + planManage.Name + "-" + "网络设备上架清单"
	return response, fileName, nil
}
