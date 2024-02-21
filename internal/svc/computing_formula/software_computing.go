package computing_formula

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"fmt"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"
	"math"
)

func SaveSoftwareBomPlanning(db *gorm.DB, planId int64) error {
	var softwareBomPlanningList []*entity.SoftwareBomPlanning
	softwareData, err := getSoftwareBomPlanningData(db, planId)
	if err != nil {
		return err
	}
	var cpuNumber int
	var softwareBomMap = make(map[string]*entity.SoftwareBomPlanning)
	for _, v := range softwareData.CloudProductNodeRoleRelList {
		cloudProductBaseline := softwareData.CloudProductBaselineMap[v.ProductId]
		serverPlanning := softwareData.ServerPlanningMap[v.NodeRoleId]
		serverBaseline := softwareData.ServerBaselineMap[serverPlanning.ServerBaselineId]
		cpuNumber += serverPlanning.Number * serverBaseline.CpuNum
		list := softwareData.SoftwareBomLicenseBaselineListMap[fmt.Sprintf("%v-%v-%v", cloudProductBaseline.ProductCode, cloudProductBaseline.SellSpecs, serverBaseline.Arch)]
		if len(list) == 0 {
			//部分软件bom只有一个，不区分硬件架构
			list = softwareData.SoftwareBomLicenseBaselineListMap[fmt.Sprintf("%v-%v-", cloudProductBaseline.ProductCode, cloudProductBaseline.SellSpecs)]
			if len(list) == 0 {
				continue
			}
		}
		for i, softwareBomLicenseBaseline := range list {
			log.Info("softwareBomLicenseBaseline:", softwareBomLicenseBaseline)
			if _, ok := softwareBomMap[softwareBomLicenseBaseline.BomId]; ok {
				softwareBomMap[softwareBomLicenseBaseline.BomId].Number += SoftwareComputing(i, softwareBomLicenseBaseline, softwareData.CloudProductPlanningList[0], serverPlanning, serverBaseline, softwareData.ServerCapPlanningMap)
			} else {
				softwareBomMap[softwareBomLicenseBaseline.BomId] = &entity.SoftwareBomPlanning{
					PlanId:             planId,
					SoftwareBaselineId: softwareBomLicenseBaseline.Id,
					BomId:              softwareBomLicenseBaseline.BomId,
					Number:             SoftwareComputing(i, softwareBomLicenseBaseline, softwareData.CloudProductPlanningList[0], serverPlanning, serverBaseline, softwareData.ServerCapPlanningMap),
					CloudService:       softwareBomLicenseBaseline.CloudService,
					ServiceCode:        softwareBomLicenseBaseline.ServiceCode,
					SellSpecs:          softwareBomLicenseBaseline.SellSpecs,
					AuthorizedUnit:     softwareBomLicenseBaseline.AuthorizedUnit,
					SellType:           softwareBomLicenseBaseline.SellType,
					HardwareArch:       softwareBomLicenseBaseline.HardwareArch,
				}
			}
		}
	}
	//平台规模授权：0100115148387809，按云平台下服务器数量计算，N=整网所有服务器的物理CPU数量之和-管理减免（10）；N大于等于0
	cpuNumber = cpuNumber - 10
	if cpuNumber < 0 {
		cpuNumber = 0
	}
	softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: PlatformBom, CloudService: PlatformName, ServiceCode: PlatformCode, Number: cpuNumber - 10})
	//软件base：0100115150861886，默认1套
	softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: SoftwareBaseBom, CloudService: SoftwareName, ServiceCode: SoftwareCode, Number: 1})
	//平台升级维保：根据选择年限对应不同BOM
	softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: ServiceYearBom[softwareData.CloudProductPlanningList[0].ServiceYear], CloudService: ServiceYearName, ServiceCode: ServiceYearCode, Number: 1})

	for _, v := range softwareBomMap {
		softwareBomPlanningList = append(softwareBomPlanningList, v)
	}
	// 保存云产品规划bom表
	if err = db.Delete(&entity.SoftwareBomPlanning{}, "plan_id = ?", planId).Error; err != nil {
		return err
	}
	if err = db.Create(softwareBomPlanningList).Error; err != nil {
		return err
	}
	return nil
}

func getSoftwareBomPlanningData(db *gorm.DB, planId int64) (*SoftwareData, error) {
	//查询云产品规划表
	var cloudProductPlanningList []*entity.CloudProductPlanning
	if err := db.Where("plan_id = ?", planId).Find(&cloudProductPlanningList).Error; err != nil {
		return nil, err
	}
	var productIdList []int64
	for _, v := range cloudProductPlanningList {
		productIdList = append(productIdList, v.ProductId)
	}
	//查询云产品和角色关联表
	var cloudProductNodeRoleRelList []*entity.CloudProductNodeRoleRel
	if err := db.Where("product_id IN (?)", productIdList).Find(&cloudProductNodeRoleRelList).Error; err != nil {
		return nil, err
	}
	//查询云产品基线
	var cloudProductBaselineList []*entity.CloudProductBaseline
	if err := db.Where("id IN (?)", productIdList).Find(&cloudProductBaselineList).Error; err != nil {
		return nil, err
	}
	var cloudProductBaselineMap = make(map[int64]*entity.CloudProductBaseline)
	for _, v := range cloudProductBaselineList {
		cloudProductBaselineMap[v.Id] = v
	}
	var productCodeList []string
	for _, v := range cloudProductBaselineList {
		productCodeList = append(productCodeList, v.ProductCode)
	}
	//查询服务器规划
	var serverPlanningList []*entity.ServerPlanning
	if err := db.Where("plan_id = ?", planId).Find(&serverPlanningList).Error; err != nil {
		return nil, err
	}
	var serverPlanningMap = make(map[int64]*entity.ServerPlanning)
	var serverBaselineIdList []int64
	for _, v := range serverPlanningList {
		serverPlanningMap[v.NodeRoleId] = v
		serverBaselineIdList = append(serverBaselineIdList, v.ServerBaselineId)
	}
	//查询服务器基线表
	var serverBaselineList []*entity.ServerBaseline
	if err := db.Where("id IN (?)", serverBaselineIdList).Find(&serverBaselineList).Error; err != nil {
		return nil, err
	}
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	for _, v := range serverBaselineList {
		serverBaselineMap[v.Id] = v
	}
	//查询容量规划
	var serverCapPlanningList []*entity.ServerCapPlanning
	if err := db.Where("plan_id = ?", planId).Find(&serverCapPlanningList).Error; err != nil {
		return nil, err
	}
	var serverCapPlanningMap = make(map[string]*entity.ServerCapPlanning)
	for _, v := range serverCapPlanningList {
		serverCapPlanningMap[fmt.Sprintf("%v-%v", v.ProductCode, v.CapPlanningInput)] = v
	}
	//查询软件bom表
	var softwareBomLicenseBaselineList []*entity.SoftwareBomLicenseBaseline
	if err := db.Where("service_code IN (?) AND version_id = ?", productCodeList, cloudProductPlanningList[0].VersionId).Find(&softwareBomLicenseBaselineList).Error; err != nil {
		return nil, err
	}
	var softwareBomLicenseBaselineListMap = make(map[string][]*entity.SoftwareBomLicenseBaseline)
	for _, v := range softwareBomLicenseBaselineList {
		if v.HardwareArch == "xc" {
			v.HardwareArch = "ARM"
		}
		softwareBomLicenseBaselineListMap[fmt.Sprintf("%v-%v-%v", v.ServiceCode, v.SellSpecs, v.HardwareArch)] = append(softwareBomLicenseBaselineListMap[fmt.Sprintf("%v-%v-%v", v.ServiceCode, v.SellSpecs, v.HardwareArch)], v)
	}
	return &SoftwareData{
		CloudProductPlanningList:          cloudProductPlanningList,
		CloudProductNodeRoleRelList:       cloudProductNodeRoleRelList,
		CloudProductBaselineMap:           cloudProductBaselineMap,
		ServerPlanningMap:                 serverPlanningMap,
		ServerBaselineMap:                 serverBaselineMap,
		ServerCapPlanningMap:              serverCapPlanningMap,
		SoftwareBomLicenseBaselineListMap: softwareBomLicenseBaselineListMap,
	}, nil
}

func SoftwareComputing(i int, softwareBomLicenseBaseline *entity.SoftwareBomLicenseBaseline, cloudProductPlanning *entity.CloudProductPlanning, serverPlanning *entity.ServerPlanning, serverBaseline *entity.ServerBaseline, serverCapPlanningMap map[string]*entity.ServerCapPlanning) int {
	switch softwareBomLicenseBaseline.ServiceCode {
	case "ECS", "BMS", "VPC":
		return serverPlanning.Number * serverBaseline.CpuNum
	case "CKE":
		serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", softwareBomLicenseBaseline.ServiceCode, "vCPU")]
		if serverCapPlanning != nil {
			return serverCapPlanning.Number / 100
		}
		return 1
	case "EBS", "EFS", "OSS", "CBR":
		//todo 暂时写死
		serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", softwareBomLicenseBaseline.ServiceCode, "容量")]
		if serverCapPlanning != nil && serverCapPlanning.Features == "三副本" {
			return int(math.Ceil(float64(serverPlanning.Number*serverBaseline.StorageDiskNum*serverBaseline.StorageDiskCapacity) / 1024 * 0.9 * 0.91 * 1 / 3))
		}
		return int(math.Ceil(float64(serverPlanning.Number*serverBaseline.StorageDiskNum*serverBaseline.StorageDiskCapacity) / 1024 * 0.9 * 0.91 * 2 / 3))
	case "CNFW", "CWAF", "CNBH", "CWP", "DES":
		if softwareBomLicenseBaseline.SellType == "升级维保" {
			return cloudProductPlanning.ServiceYear
		}
		return 1
	case "KAFKA", "ROCKETMQ", "APIM", "CONNECT":
		if i > 1 {
			serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", softwareBomLicenseBaseline.ServiceCode, "vCPU")]
			if serverCapPlanning != nil && serverCapPlanning.Number > 200 {
				return int(math.Ceil(float64(serverCapPlanning.Number-200) / 100))
			}
			return 0
		}
		return 1
	case "CSP":
		if i > 1 {
			serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", softwareBomLicenseBaseline.ServiceCode, "vCPU")]
			if serverCapPlanning != nil && serverCapPlanning.Number > 500 {
				return int(math.Ceil(float64(serverCapPlanning.Number-500) / 100))
			}
			return 0
		}
		return 1
	default:
		return 1
	}
}

const (
	PlatformName    = "平台规模授权"
	PlatformCode    = "Platform"
	PlatformBom     = "0100115148387809"
	SoftwareName    = "软件base"
	SoftwareCode    = "SoftwareBase"
	SoftwareBaseBom = "0100115150861886"
	ServiceYearName = "平台升级维保"
	ServiceYearCode = "ServiceYear"
)

var ServiceYearBom = map[int]string{1: "0100115152958526", 2: "0100115153975617", 3: "0100115154780568", 4: "0100115155303482", 5: "0100115156784743"}
