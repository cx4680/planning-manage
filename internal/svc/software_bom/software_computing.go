package software_bom

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
)

func Save(c *gin.Context) {
	planIdString := c.Param("planId")
	planId, _ := strconv.ParseInt(planIdString, 10, 64)
	err := data.DB.Transaction(func(tx *gorm.DB) error {
		return SaveSoftwareBomPlanning(data.DB, planId)
	})
	if err != nil {
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func SaveSoftwareBomPlanning(db *gorm.DB, planId int64) error {
	softwareData, err := getSoftwareBomPlanningData(db, planId)
	if err != nil {
		return err
	}
	bomMap := ComputingSoftwareBom(softwareData)
	var softwareBomPlanningList []*entity.SoftwareBomPlanning
	var cpuNumber int
	for _, serverPlannings := range softwareData.ServerPlanningsMap {
		for _, serverPlanning := range serverPlannings {
			serverBaseline := softwareData.ServerBaselineMap[serverPlanning.ServerBaselineId]
			cpuNumber += serverPlanning.Number * serverBaseline.CpuNum
		}
	}
	for k, v := range bomMap {
		if k == DatabaseManagementBom {
			// 默认输出数据库管理平台授权，BOM iD：0100115140403032，单位：套
			softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: DatabaseManagementBom, CloudService: DatabaseManagementName, ServiceCode: DatabaseManagementCode, Number: v})
			continue
		}
		if k == SecurityMaintenanceBom {
			// 默认输出安全产品统一维保，BOM iD：0100115099084508，单位：年
			if softwareData.ServiceYear > 1 {
				softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: SecurityMaintenanceBom, CloudService: SecurityMaintenanceName, ServiceCode: SecurityMaintenanceCode, Number: softwareData.ServiceYear - 1})
			}
			continue
		}
		if k == CloudNativeSecBasicPkgBom {
			// 默认输出云原生安全-基础安全包，BOM iD：0100115808142197，单位：颗
			softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: CloudNativeSecBasicPkgBom, CloudService: CloudNativeSecBasicPkgName, ServiceCode: CloudNativeSecBasicPkgCode, Number: cpuNumber})
			continue
		}
		if k == BigDataManagementBom {
			// 默认输出中电云大数据开发管理平台，BOM iD：0100115228973107，单位：套
			softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: BigDataManagementBom, CloudService: BigDataManagementName, ServiceCode: BigDataManagementCode, Number: v})
			continue
		}
		if k == BigDataMaintenanceBom {
			// 默认输出中电云大数据平台(维保服务)，BOM iD：0100115230260255，单位：年
			if softwareData.ServiceYear > 1 {
				softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: BigDataMaintenanceBom, CloudService: BigDataMaintenanceName, ServiceCode: BigDataMaintenanceCode, Number: softwareData.ServiceYear - 1})
			}
			continue
		}
		if k == BigDataPlatformScaleBom {
			// 默认输出大数据CeaInsight-平台规模授权，BOM iD：0100115139411762，单位：vCPU
			softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: BigDataPlatformScaleBom, CloudService: BigDataPlatformScaleName, ServiceCode: BigDataPlatformScaleCode, Number: v})
			continue
		}
		softwareBomLicenseBaseline := softwareData.BomIdSoftwareBomLicenseBaselineMap[k]
		softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{
			PlanId:             planId,
			SoftwareBaselineId: softwareBomLicenseBaseline.Id,
			BomId:              softwareBomLicenseBaseline.BomId,
			Number:             v,
			CloudService:       softwareBomLicenseBaseline.CloudService,
			ServiceCode:        softwareBomLicenseBaseline.ServiceCode,
			SellSpecs:          softwareBomLicenseBaseline.SellSpecs,
			AuthorizedUnit:     softwareBomLicenseBaseline.AuthorizedUnit,
			SellType:           softwareBomLicenseBaseline.SellType,
			HardwareArch:       softwareBomLicenseBaseline.HardwareArch,
			ValueAddedService:  softwareBomLicenseBaseline.ValueAddedService,
		})
	}
	// 平台规模授权：0100115148387809，按云平台下服务器数量计算，N=整网所有服务器的物理CPU数量之和-管理减免（10）；N大于等于0
	cpuNumber = cpuNumber - 10
	if cpuNumber < 0 {
		cpuNumber = 0
	}
	softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: PlatformBom, CloudService: PlatformName, ServiceCode: PlatformCode, Number: cpuNumber})
	// 软件base：0100115150861886，默认1套
	softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: SoftwareBaseBom, CloudService: SoftwareName, ServiceCode: SoftwareCode, Number: 1})
	// 平台升级维保：根据选择年限对应不同BOM
	if softwareData.ServiceYear > 1 {
		softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: ServiceYearBom[softwareData.ServiceYear-1], CloudService: ServiceYearName, ServiceCode: ServiceYearCode, Number: 1})
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
	// 查询云产品规划表
	var cloudProductPlanningList []*entity.CloudProductPlanning
	if err := db.Where("plan_id = ?", planId).Find(&cloudProductPlanningList).Error; err != nil {
		return nil, err
	}
	var productIdList []int64
	// 产品id为key
	var cloudProductPlanningMap = make(map[int64]*entity.CloudProductPlanning)
	for _, cloudProductPlanning := range cloudProductPlanningList {
		productIdList = append(productIdList, cloudProductPlanning.ProductId)
		cloudProductPlanningMap[cloudProductPlanning.ProductId] = cloudProductPlanning
	}
	// 查询云产品和角色关联表
	var cloudProductNodeRoleRelList []*entity.CloudProductNodeRoleRel
	if err := db.Where("product_id IN (?)", productIdList).Find(&cloudProductNodeRoleRelList).Error; err != nil {
		return nil, err
	}
	var nodeRoleIdList []int64
	for _, cloudProductNodeRoleRel := range cloudProductNodeRoleRelList {
		nodeRoleIdList = append(nodeRoleIdList, cloudProductNodeRoleRel.NodeRoleId)
	}
	// 查询角色节点基线
	var nodeRoleBaselineList []*entity.NodeRoleBaseline
	if err := db.Where("id IN (?)", nodeRoleIdList).Find(&nodeRoleBaselineList).Error; err != nil {
		return nil, err
	}
	// 角色id为key
	var nodeRoleCodeMap = make(map[int64]string)
	for _, nodeRoleBaseline := range nodeRoleBaselineList {
		nodeRoleCodeMap[nodeRoleBaseline.Id] = nodeRoleBaseline.NodeRoleCode
	}
	// 查询服务器规划
	var serverPlanningList []*entity.ServerPlanning
	if err := db.Where("plan_id = ?", planId).Find(&serverPlanningList).Error; err != nil {
		return nil, err
	}
	// 角色节点code为key
	var roleCodeServerPlanningsMap = make(map[string][]*entity.ServerPlanning)
	var serverBaselineIdList []int64
	for _, serverPlanning := range serverPlanningList {
		nodeRoleCode := nodeRoleCodeMap[serverPlanning.NodeRoleId]
		roleCodeServerPlanningsMap[nodeRoleCode] = append(roleCodeServerPlanningsMap[nodeRoleCode], serverPlanning)
		serverBaselineIdList = append(serverBaselineIdList, serverPlanning.ServerBaselineId)
	}
	// 查询服务器基线表
	var serverBaselineList []*entity.ServerBaseline
	if err := db.Where("id IN (?)", serverBaselineIdList).Find(&serverBaselineList).Error; err != nil {
		return nil, err
	}
	// 服务器基线id为key
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	for _, serverBaseline := range serverBaselineList {
		if serverBaseline.Arch == constant.CpuArchARM {
			serverBaseline.Arch = constant.CpuArchXC
		}
		if strings.ToLower(serverBaseline.CpuType) == constant.CpuTypeHygon {
			serverBaseline.Arch = constant.CpuArchXC
		}
		serverBaselineMap[serverBaseline.Id] = serverBaseline
	}
	// 查询容量规划
	var serverCapPlanningList []*entity.ServerCapPlanning
	if err := db.Where("plan_id = ?", planId).Find(&serverCapPlanningList).Error; err != nil {
		return nil, err
	}
	var serverCapPlanningMap = make(map[string]*entity.ServerCapPlanning)
	for _, serverCapPlanning := range serverCapPlanningList {
		// 由于CSP和COS云产品没有关联节点角色，所以不加资源池id过滤
		if serverCapPlanning.ProductCode == constant.ProductCodeCSP || serverCapPlanning.ProductCode == constant.ProductCodeCOS {
			serverCapPlanningMap[fmt.Sprintf("%v-%v", serverCapPlanning.ProductCode, serverCapPlanning.CapPlanningInput)] = serverCapPlanning
			continue
		}
		serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", serverCapPlanning.ProductCode, serverCapPlanning.ResourcePoolId, serverCapPlanning.CapPlanningInput)] = serverCapPlanning
	}
	// 查询云产品基线
	var cloudProductBaselineList []*entity.CloudProductBaseline
	if err := db.Where("id IN (?)", productIdList).Find(&cloudProductBaselineList).Error; err != nil {
		return nil, err
	}
	// 产品id为key
	var cloudProductBaselineMap = make(map[int64]*entity.CloudProductBaseline)
	var productCodeList []string
	for _, cloudProductBaseline := range cloudProductBaselineList {
		productCodeList = append(productCodeList, cloudProductBaseline.ProductCode)
		cloudProductBaselineMap[cloudProductBaseline.Id] = cloudProductBaseline
	}
	// 查询软件bom表
	var softwareBomLicenseBaselineList []*entity.SoftwareBomLicenseBaseline
	if err := db.Where("service_code IN (?) AND version_id = ?", productCodeList, cloudProductPlanningList[0].VersionId).Find(&softwareBomLicenseBaselineList).Error; err != nil {
		return nil, err
	}
	// 根据产品编码-售卖规格、产品编码-增值服务、产品编码-硬件架构 筛选容量输入列表
	var screenCloudProductSellSpecMap = make(map[string]interface{})
	var screenCloudProductValueAddedServiceMap = make(map[string]interface{})
	for _, cloudProductPlanning := range cloudProductPlanningList {
		// 根据产品编码-售卖规格
		if util.IsNotBlank(cloudProductPlanning.SellSpec) {
			screenCloudProductSellSpecMap[fmt.Sprintf("%s-%s", cloudProductBaselineMap[cloudProductPlanning.ProductId].ProductCode, cloudProductPlanning.SellSpec)] = nil
		}
		// 产品编码-增值服务
		if util.IsNotBlank(cloudProductPlanning.ValueAddedService) {
			for _, valueAddedService := range strings.Split(cloudProductPlanning.ValueAddedService, ",") {
				screenCloudProductValueAddedServiceMap[fmt.Sprintf("%s-%s", cloudProductBaselineMap[cloudProductPlanning.ProductId].ProductCode, valueAddedService)] = nil
			}
		}
	}
	// 产品编码-售卖规格-增值服务-硬件架构为key
	var serviceCodeSoftwareBomLicenseBaselineMap = make(map[string][]*entity.SoftwareBomLicenseBaseline)
	// 软件bom为key
	var bomIdSoftwareBomLicenseBaselineMap = make(map[string]*entity.SoftwareBomLicenseBaseline)
	for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
		// 根据产品编码-售卖规格、产品编码-增值服务筛选容量输入列表
		if util.IsNotBlank(softwareBomLicenseBaseline.SellSpecs) {
			if _, ok := screenCloudProductSellSpecMap[fmt.Sprintf("%s-%s", softwareBomLicenseBaseline.ServiceCode, softwareBomLicenseBaseline.SellSpecs)]; !ok && softwareBomLicenseBaseline.ServiceCode != constant.ProductCodeCNBH {
				continue
			}
		}
		if util.IsNotBlank(softwareBomLicenseBaseline.ValueAddedService) {
			if _, ok := screenCloudProductValueAddedServiceMap[fmt.Sprintf("%s-%s", softwareBomLicenseBaseline.ServiceCode, softwareBomLicenseBaseline.ValueAddedService)]; !ok {
				continue
			}
		}
		if util.IsBlank(softwareBomLicenseBaseline.BomId) {
			continue
		}
		serviceCodeSoftwareBomLicenseBaselineMap[softwareBomLicenseBaseline.ServiceCode] = append(serviceCodeSoftwareBomLicenseBaselineMap[softwareBomLicenseBaseline.ServiceCode], softwareBomLicenseBaseline)
		bomIdSoftwareBomLicenseBaselineMap[softwareBomLicenseBaseline.BomId] = softwareBomLicenseBaseline
	}
	return &SoftwareData{
		ServiceYear:                              cloudProductPlanningList[0].ServiceYear,
		CloudProductBaselineList:                 cloudProductBaselineList,
		ServerPlanningsMap:                       roleCodeServerPlanningsMap,
		ServerBaselineMap:                        serverBaselineMap,
		ServerCapPlanningMap:                     serverCapPlanningMap,
		BomIdSoftwareBomLicenseBaselineMap:       bomIdSoftwareBomLicenseBaselineMap,
		ServiceCodeSoftwareBomLicenseBaselineMap: serviceCodeSoftwareBomLicenseBaselineMap,
	}, nil
}

func ComputingSoftwareBom(softwareData *SoftwareData) map[string]int {
	// bomId为key，数量为value
	var bomMap = make(map[string]int)
	serverPlanningsMap := softwareData.ServerPlanningsMap
	serverBaselineMap := softwareData.ServerBaselineMap
	serverCapPlanningMap := softwareData.ServerCapPlanningMap
	// 升级维保 = 实例数 * 年限
	for _, v := range softwareData.CloudProductBaselineList {
		productCode := v.ProductCode
		softwareBomLicenseBaselineList := softwareData.ServiceCodeSoftwareBomLicenseBaselineMap[productCode]
		if len(softwareBomLicenseBaselineList) == 0 && productCode != constant.ProductCodeCIK {
			continue
		}
		switch productCode {
		case constant.ProductCodeECS:
			// COMPUTE节点的CPU数量
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodeCompute]
			if len(serverPlannings) == 0 {
				continue
			}
			archNumberMap := make(map[string]int)
			for _, serverPlanning := range serverPlannings {
				serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
				archNumberMap[serverBaseline.Arch] += serverPlanning.Number * serverBaseline.CpuNum
			}
			for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
				archNumber, ok := archNumberMap[softwareBomLicenseBaseline.HardwareArch]
				if ok {
					bomMap[softwareBomLicenseBaseline.BomId] = archNumber
				}
			}
		case constant.ProductCodeBMS:
			// BMS节点的CPU数量
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodeBMS]
			if len(serverPlannings) == 0 {
				continue
			}
			var number int
			for _, serverPlanning := range serverPlannings {
				serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
				number += serverPlanning.Number * serverBaseline.CpuNum
			}
			for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
				bomMap[softwareBomLicenseBaseline.BomId] = number
			}
		case constant.ProductCodeCKE:
			// CKE容量vCPU数量/100
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodeCompute]
			if len(serverPlannings) == 0 {
				continue
			}
			archNumberMap := make(map[string]int)
			for _, serverPlanning := range serverPlannings {
				serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
				serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputVCpu)]
				var number int
				if serverCapPlanning != nil {
					number = int(math.Ceil(float64(serverCapPlanning.Number) / 100))
				}
				archNumberMap[serverBaseline.Arch] += number
			}
			for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
				archNumber, ok := archNumberMap[softwareBomLicenseBaseline.HardwareArch]
				if ok {
					if archNumber == 0 {
						archNumber = 1
					}
					bomMap[softwareBomLicenseBaseline.BomId] = archNumber
				}
			}
		case constant.ProductCodeCBR:
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodeCBR]
			if len(serverPlannings) == 0 {
				continue
			}
			var number int
			for _, serverPlanning := range serverPlannings {
				serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputBackupDataCapacity)]
				if serverCapPlanning != nil {
					number += serverCapPlanning.Number
				}
			}
			if number == 0 {
				number = 1
			}
			for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
				bomMap[softwareBomLicenseBaseline.BomId] = number
			}
		case constant.ProductCodeEBS, constant.ProductCodeEFS, constant.ProductCodeOSS:
			// TB，可用容量
			var serverPlannings []*entity.ServerPlanning
			if productCode == constant.ProductCodeEBS {
				serverPlannings = serverPlanningsMap[constant.NodeRoleCodeEBS]
			} else if productCode == constant.ProductCodeEFS {
				serverPlannings = serverPlanningsMap[constant.NodeRoleCodeEFS]
			} else if productCode == constant.ProductCodeOSS {
				serverPlannings = serverPlanningsMap[constant.NodeRoleCodeOSS]
			}
			if len(serverPlannings) == 0 {
				continue
			}
			var number int
			for _, serverPlanning := range serverPlannings {
				// serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
				serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputStorageCapacity)]
				if serverCapPlanning != nil {
					// if serverCapPlanning.Features == constant.FeaturesNameThreeCopies {
					// 	number += int(math.Ceil(float64(serverPlanning.Number*serverBaseline.StorageDiskNum*serverBaseline.StorageDiskCapacity) / 1024 * 0.9 * 0.91 * 1 / 3))
					// }
					// if serverCapPlanning.Features == constant.FeaturesNameEC {
					// 	number += int(math.Ceil(float64(serverPlanning.Number*serverBaseline.StorageDiskNum*serverBaseline.StorageDiskCapacity) / 1024 * 0.9 * 0.91 * 2 / 3))
					// }
					number += int(math.Ceil(float64(serverCapPlanning.Number / 1024)))
				}
			}
			if number == 0 {
				number = 1
			}
			for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
				bomMap[softwareBomLicenseBaseline.BomId] = number
			}
		case constant.ProductCodeVPC:
			// NETWORK、NFV、BMSGW的CPU总数
			networkServerPlannings := serverPlanningsMap[constant.NodeRoleCodeNETWORK]
			nfvServerPlannings := serverPlanningsMap[constant.NodeRoleCodeNFV]
			bmsGWServerPlannings := serverPlanningsMap[constant.NodeRoleCodeBMSGW]
			archCpuNumberMap := make(map[string]int)
			for _, serverPlanning := range networkServerPlannings {
				serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
				archCpuNumberMap[serverBaseline.Arch] += serverPlanning.Number * serverBaseline.CpuNum
			}
			for _, serverPlanning := range nfvServerPlannings {
				serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
				archCpuNumberMap[serverBaseline.Arch] += serverPlanning.Number * serverBaseline.CpuNum
			}
			for _, serverPlanning := range bmsGWServerPlannings {
				serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
				archCpuNumberMap[serverBaseline.Arch] += serverPlanning.Number * serverBaseline.CpuNum
			}
			for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
				cpuNumber := archCpuNumberMap[softwareBomLicenseBaseline.HardwareArch]
				if cpuNumber != 0 {
					bomMap[softwareBomLicenseBaseline.BomId] = cpuNumber
				}
			}
		case constant.ProductCodeCNFW, constant.ProductCodeCWAF:
			// 根据选择的容量数量计算
			var serverPlannings []*entity.ServerPlanning
			if productCode == constant.ProductCodeCNFW {
				serverPlannings = serverPlanningsMap[constant.NodeRoleCodeCompute]
			}
			if productCode == constant.ProductCodeCWAF {
				serverPlannings = serverPlanningsMap[constant.NodeRoleCodeNFV]
			}
			if len(serverPlannings) == 0 {
				continue
			}
			bomMap[SecurityMaintenanceBom] = 1
			bomMap[CloudNativeSecBasicPkgBom] = 1
			var number int
			for _, serverPlanning := range serverPlannings {
				serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputFirewall)]
				if serverCapPlanning != nil {
					number += serverCapPlanning.Number
				}
			}
			if number == 0 {
				number = 1
			}
			for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
				if softwareBomLicenseBaseline.SellType == constant.SoftwareBomLicense {
					bomMap[softwareBomLicenseBaseline.BomId] = number
				}
			}
		case constant.ProductCodeCSOC:
			var assetAccessNumber int
			var logStorageNumber int
			var vulnerabilityScanningNumber int
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodeNFV]
			if len(serverPlannings) == 0 {
				continue
			}
			bomMap[SecurityMaintenanceBom] = 1
			bomMap[CloudNativeSecBasicPkgBom] = 1
			for _, serverPlanning := range serverPlannings {
				assetAccess := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputAssetAccess)]
				if assetAccess != nil {
					assetAccessNumber += assetAccess.Number
				}
				logStorage := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputLogStorageSpace)]
				if logStorage != nil {
					logStorageNumber += logStorage.Number
				}
				vulnerabilityScanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputVulnerabilityScanning)]
				if vulnerabilityScanning != nil {
					vulnerabilityScanningNumber += vulnerabilityScanning.Number
				}
			}
			if assetAccessNumber == 0 {
				assetAccessNumber = 1
			}
			if logStorageNumber == 0 {
				logStorageNumber = 1
			}
			for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
				if softwareBomLicenseBaseline.SellType == constant.SoftwareBomLicense {
					if softwareBomLicenseBaseline.AuthorizedUnit == constant.SoftwareBomAuthorizedUnitAssetAccess {
						bomMap[softwareBomLicenseBaseline.BomId] = assetAccessNumber
					}
					if softwareBomLicenseBaseline.AuthorizedUnit == constant.SoftwareBomAuthorizedUnitLogStorage {
						bomMap[softwareBomLicenseBaseline.BomId] = logStorageNumber
					}
					if softwareBomLicenseBaseline.ValueAddedService == constant.SoftwareBomValueAddedServiceVulnerabilityScanning {
						if vulnerabilityScanningNumber != 0 {
							bomMap[softwareBomLicenseBaseline.BomId] = vulnerabilityScanningNumber
						}
					}
				}
			}
		case constant.ProductCodeDSP:
			// 根据数量匹配1-5，6-30以及30以上，三个阶梯的bom，再根据数据库个数，输出bom的数量，例如有26个数据库实例，则需要输出26个6-30bom
			var number int
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodeNFV]
			if len(serverPlannings) == 0 {
				continue
			}
			bomMap[SecurityMaintenanceBom] = 1
			bomMap[CloudNativeSecBasicPkgBom] = 1
			for _, serverPlanning := range serverPlannings {
				serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputDatabaseAudit)]
				if serverCapPlanning != nil {
					number += serverCapPlanning.Number
				}
			}
			if number == 0 {
				number = 1
			}
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == constant.DSPSoftwareBomCalcMethod1To5Instances && number >= 1 && number <= 5 {
					if softwareBom.SellType == constant.SoftwareBomLicense {
						bomMap[softwareBom.BomId] = number
					}
				}
				if softwareBom.CalcMethod == constant.DSPSoftwareBomCalcMethod6To30Instances && number >= 6 && number <= 30 {
					if softwareBom.SellType == constant.SoftwareBomLicense {
						bomMap[softwareBom.BomId] += number
					}
				}
				if softwareBom.CalcMethod == constant.DSPSoftwareBomCalcMethodOver30Instances && number > 30 {
					if softwareBom.SellType == constant.SoftwareBomLicense {
						bomMap[softwareBom.BomId] += number
					}
				}
			}
		case constant.ProductCodeCNBH:
			bomMap[SecurityMaintenanceBom] = 1
			bomMap[CloudNativeSecBasicPkgBom] = 1
			// 根据包数量直接转换为bom数量
			var oneHundredNumber int
			var fiveHundredNumber int
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodeCompute]
			for _, serverPlanning := range serverPlannings {
				oneHundredServerCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputOneHundred)]
				if oneHundredServerCapPlanning != nil {
					oneHundredNumber += oneHundredServerCapPlanning.Number
				}
				fiveHundredServerCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputFiveHundred)]
				if fiveHundredServerCapPlanning != nil {
					fiveHundredNumber += fiveHundredServerCapPlanning.Number
				}
			}
			for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
				var number int
				if softwareBomLicenseBaseline.SellSpecs == constant.CapPlanningInputOneHundred {
					number = oneHundredNumber
				}
				if softwareBomLicenseBaseline.SellSpecs == constant.CapPlanningInputFiveHundred {
					number = fiveHundredNumber
				}
				if softwareBomLicenseBaseline.SellType == constant.SoftwareBomLicense && number != 0 {
					bomMap[softwareBomLicenseBaseline.BomId] = number
				}
			}
		case constant.ProductCodeCWP:
			var protectiveEcsTerminalNumber int
			var protectiveContainerServiceNumber int
			var protectedWebsiteDirectoryNumber int
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodeNFV]
			if len(serverPlannings) == 0 {
				continue
			}
			bomMap[SecurityMaintenanceBom] = 1
			bomMap[CloudNativeSecBasicPkgBom] = 1
			for _, serverPlanning := range serverPlannings {
				protectiveEcsTerminal := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputProtectiveECSTerminal)]
				if protectiveEcsTerminal != nil {
					protectiveEcsTerminalNumber += protectiveEcsTerminal.Number
				}
				protectiveContainerService := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputProtectiveContainerService)]
				if protectiveContainerService != nil {
					protectiveContainerServiceNumber += protectiveContainerService.Number
				}
				protectedWebsiteDirectory := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputProtectiveWebsiteDirectory)]
				if protectedWebsiteDirectory != nil {
					protectedWebsiteDirectoryNumber += protectedWebsiteDirectory.Number
				}
			}
			if protectiveEcsTerminalNumber == 0 {
				protectiveEcsTerminalNumber = 1
			}
			for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
				if util.IsNotBlank(softwareBomLicenseBaseline.SellSpecs) {
					if softwareBomLicenseBaseline.SellType == constant.SoftwareBomLicense {
						bomMap[softwareBomLicenseBaseline.BomId] = protectiveEcsTerminalNumber
					}
				}
				if softwareBomLicenseBaseline.ValueAddedService == constant.SoftwareBomValueAddedServiceContainerSafety {
					if protectiveContainerServiceNumber != 0 {
						if softwareBomLicenseBaseline.SellType == constant.SoftwareBomLicense {
							bomMap[softwareBomLicenseBaseline.BomId] = protectiveContainerServiceNumber
						}
					}
				}
				if softwareBomLicenseBaseline.ValueAddedService == constant.SoftwareBomValueAddedServiceWebPageTamperPrevention {
					if protectedWebsiteDirectoryNumber != 0 {
						if softwareBomLicenseBaseline.SellType == constant.SoftwareBomLicense {
							bomMap[softwareBomLicenseBaseline.BomId] = protectedWebsiteDirectoryNumber
						}
					}
				}
			}
		case constant.ProductCodeDES:
			bomMap[SecurityMaintenanceBom] = 1
			bomMap[CloudNativeSecBasicPkgBom] = 1
			// 没有输入，输出1
			number := 1
			for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
				if softwareBomLicenseBaseline.SellType == constant.SoftwareBomLicense {
					bomMap[softwareBomLicenseBaseline.BomId] = number
				}
			}
		case constant.ProductCodeCEASQLTX, constant.ProductCodeMYSQL, constant.ProductCodeCEASQLDW, constant.ProductCodeCEASQLCK, constant.ProductCodeREDIS, constant.ProductCodePOSTGRESQL, constant.ProductCodeMONGODB, constant.ProductCodeINFLUXDB:
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodeDATABASE]
			if len(serverPlannings) == 0 {
				continue
			}
			bomMap[DatabaseManagementBom] = 1
			var number int
			for _, serverPlanning := range serverPlannings {
				var vCpuNumber int
				var copyNumber int
				vCpuServerCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputVCpuTotal)]
				copyServerCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputCopy)]
				if vCpuServerCapPlanning != nil {
					vCpuNumber = vCpuServerCapPlanning.Number
				}
				if copyServerCapPlanning != nil {
					copyNumber = copyServerCapPlanning.Number
				}
				if vCpuNumber != 0 && copyNumber != 0 {
					// serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
					// if serverBaseline.Arch == constant.CpuArchX86 {
					// 	number += int(math.Ceil(float64(vCpuNumber*copyNumber) / 2))
					// }
					// if serverBaseline.Arch == constant.CpuArchARM {
					// 	number += vCpuNumber * copyNumber
					// }
					number += vCpuNumber * copyNumber
				}
			}
			if number == 0 || number < 192 {
				number = 192
			}
			for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
				bomMap[softwareBomLicenseBaseline.BomId] = number
			}
		case constant.ProductCodeDTS:
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodeDATABASE]
			if len(serverPlannings) == 0 {
				continue
			}
			bomMap[DatabaseManagementBom] = 1
			var number int
			for _, serverPlanning := range serverPlannings {
				smallServerCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputSmall)]
				middleServerCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputMiddle)]
				largeServerCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputLarge)]
				if smallServerCapPlanning != nil {
					number += smallServerCapPlanning.Number
				}
				if middleServerCapPlanning != nil {
					number += middleServerCapPlanning.Number
				}
				if largeServerCapPlanning != nil {
					number += largeServerCapPlanning.Number
				}
			}
			if number == 0 {
				number = 1
			}
			for _, softwareBom := range softwareBomLicenseBaselineList {
				bomMap[softwareBom.BomId] = number
			}
		case constant.ProductCodeKAFKA, constant.ProductCodeRABBITMQ:
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodePAASData]
			if len(serverPlannings) == 0 {
				continue
			}
			var number int
			for _, serverPlanning := range serverPlannings {
				brokerNumber, standardEditionNumber, professionalEditionNumber, enterpriseEditionNumber, platinumEditionNumber := handlePAASCapPlanningInput(serverPlanning.ResourcePoolId, serverCapPlanningMap, productCode)
				number += (standardEditionNumber*2 + professionalEditionNumber*4 + enterpriseEditionNumber*8 + platinumEditionNumber*16) * brokerNumber
			}
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == constant.KAFKASoftwareBomCalcMethodBasePackage {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.CalcMethod == constant.KAFKASoftwareBomCalcMethodExpansionPackage {
					if number-200 > 0 {
						bomMap[softwareBom.BomId] = int(math.Ceil(float64(number-200) / 100))
					}
				}
			}
		case constant.ProductCodeCSP:
			serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputMicroservice)]
			if serverCapPlanning == nil {
				continue
			}
			number := serverCapPlanning.Number
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == constant.CSPSoftwareBomCalcMethodBasePackage {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.CalcMethod == constant.CSPSoftwareBomCalcMethodExpansionPackage {
					if number-500 > 0 {
						bomMap[softwareBom.BomId] = int(math.Ceil(float64(number-500) / 100))
					}
				}
			}
		case constant.ProductCodeROCKETMQ:
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodePAASData]
			if len(serverPlannings) == 0 {
				continue
			}
			var number int
			for _, serverPlanning := range serverPlannings {
				_, standardEditionNumber, professionalEditionNumber, enterpriseEditionNumber, platinumEditionNumber := handlePAASCapPlanningInput(serverPlanning.ResourcePoolId, serverCapPlanningMap, productCode)
				number += standardEditionNumber*12 + professionalEditionNumber*24 + enterpriseEditionNumber*36 + platinumEditionNumber*48
			}
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == constant.ROCKETMQSoftwareBomCalcMethodBasePackage {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.CalcMethod == constant.ROCKETMQSoftwareBomCalcMethodExpansionPackage {
					if number-200 > 0 {
						bomMap[softwareBom.BomId] = int(math.Ceil(float64(number-200) / 100))
					}
				}
			}
		case constant.ProductCodeAPIM:
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodePAASCompute]
			if len(serverPlannings) == 0 {
				continue
			}
			var number int
			for _, serverPlanning := range serverPlannings {
				_, standardEditionNumber, professionalEditionNumber, enterpriseEditionNumber, _ := handlePAASCapPlanningInput(serverPlanning.ResourcePoolId, serverCapPlanningMap, productCode)
				number += standardEditionNumber*3 + professionalEditionNumber*6 + enterpriseEditionNumber*12
			}
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == constant.APIMSoftwareBomCalcMethodBasePackage {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.CalcMethod == constant.APIMSoftwareBomCalcMethodExpansionPackage {
					if number-200 > 0 {
						bomMap[softwareBom.BomId] = int(math.Ceil(float64(number-200) / 100))
					}
				}
			}
		case constant.ProductCodeCONNECT:
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodePAASCompute]
			if len(serverPlannings) == 0 {
				continue
			}
			var number int
			for _, serverPlanning := range serverPlannings {
				_, standardEditionNumber, professionalEditionNumber, enterpriseEditionNumber, _ := handlePAASCapPlanningInput(serverPlanning.ResourcePoolId, serverCapPlanningMap, productCode)
				number += standardEditionNumber*40 + professionalEditionNumber*80 + enterpriseEditionNumber*120
			}
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == constant.CONNECTSoftwareBomCalcMethodBasePackage {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.CalcMethod == constant.CONNECTSoftwareBomCalcMethodExpansionPackage {
					if number-200 > 0 {
						bomMap[softwareBom.BomId] = int(math.Ceil(float64(number-200) / 50))
					}
				}
			}
		case constant.ProductCodeCLCP:
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodePAASCompute]
			if len(serverPlannings) == 0 {
				continue
			}
			var number int
			for _, serverPlanning := range serverPlannings {
				_, standardEditionNumber, professionalEditionNumber, enterpriseEditionNumber, _ := handlePAASCapPlanningInput(serverPlanning.ResourcePoolId, serverCapPlanningMap, productCode)
				number += standardEditionNumber*16 + professionalEditionNumber*64 + enterpriseEditionNumber*96
			}
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == constant.CLCPSoftwareBomCalcMethodBasePackage {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.CalcMethod == constant.CLCPSoftwareBomCalcMethodExpansionPackage {
					if number-48 > 0 {
						bomMap[softwareBom.BomId] = int(math.Ceil(float64(number-48) / 16))
					}
				}
				if softwareBom.CalcMethod == constant.CLCPSoftwareBomCalcMethodBITool || softwareBom.CalcMethod == constant.CLCPSoftwareBomCalcMethodVisualLargeScreenTool {
					bomMap[softwareBom.BomId] = 1
				}
			}
		case constant.ProductCodeCOS:
			var monitoringNodeNumber int
			monitoringNode := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputMonitoringNode)]
			if monitoringNode != nil {
				monitoringNodeNumber = monitoringNode.Number
			}
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == constant.COSSoftwareBomCalcMethodBasePackage {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.CalcMethod == constant.COSSoftwareBomCalcMethodExpansionPackage {
					if monitoringNodeNumber-1000 > 0 {
						bomMap[softwareBom.BomId] = int(math.Ceil(float64(monitoringNodeNumber-1000) / 200))
					}
				}
			}
		case constant.ProductCodeCLS:
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodePAASData]
			if len(serverPlannings) == 0 {
				continue
			}
			var logStorageNumber int
			for _, serverPlanning := range serverPlannings {
				logStorage := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputLogStorage)]
				if logStorage != nil {
					logStorageNumber += logStorage.Number
				}
			}
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == constant.CLSSoftwareBomCalcMethodBasePackage {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.CalcMethod == constant.CLSSoftwareBomCalcMethodExpansionPackage {
					if logStorageNumber-10 > 0 {
						bomMap[softwareBom.BomId] = int(math.Ceil(float64(logStorageNumber-10) / 5))
					}
				}
			}
		case constant.ProductCodeES:
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodeDATABASE]
			if len(serverPlannings) == 0 {
				continue
			}
			bomMap[BigDataManagementBom] = 1
			bomMap[BigDataMaintenanceBom] = 1
			var number int
			for _, serverPlanning := range serverPlannings {
				var vCpuNumber int
				var copyNumber int
				vCpuServerCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputVCpuTotal)]
				copyServerCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, serverPlanning.ResourcePoolId, constant.CapPlanningInputCopy)]
				if vCpuServerCapPlanning != nil {
					vCpuNumber = vCpuServerCapPlanning.Number
				}
				if copyServerCapPlanning != nil {
					copyNumber = copyServerCapPlanning.Number
				}
				if vCpuNumber != 0 && copyNumber != 0 {
					// serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
					// if serverBaseline.Arch == constant.CpuArchX86 {
					// 	number += int(math.Ceil(float64(vCpuNumber*copyNumber) / 2))
					// }
					// if serverBaseline.Arch == constant.CpuArchARM {
					// 	number += vCpuNumber * copyNumber
					// }
					number += vCpuNumber * copyNumber
				}
			}
			if number == 0 || number < 100 {
				number = 100
			}
			for _, softwareBomLicenseBaseline := range softwareBomLicenseBaselineList {
				bomMap[softwareBomLicenseBaseline.BomId] = number
			}
		case constant.ProductCodeCIK:
			serverPlannings := serverPlanningsMap[constant.NodeRoleCodeBIGDATA]
			if len(serverPlannings) == 0 {
				continue
			}
			bomMap[BigDataManagementBom] = 1
			bomMap[BigDataMaintenanceBom] = 1
			var number int
			for _, serverPlanning := range serverPlannings {
				serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
				number += serverPlanning.Number * serverBaseline.Cpu
			}
			if number == 0 || number < 300 {
				number = 300
			}
			bomMap[BigDataPlatformScaleBom] = number
		}
	}
	return bomMap
}

func handlePAASCapPlanningInput(resourcePoolId int64, serverCapPlanningMap map[string]*entity.ServerCapPlanning, productCode string) (int, int, int, int, int) {
	var brokerNumber, standardEditionNumber, professionalEditionNumber, enterpriseEditionNumber, platinumEditionNumber int
	broker := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, resourcePoolId, constant.CapPlanningInputBroker)]
	if broker != nil {
		brokerNumber = broker.Number
	}
	standardEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, resourcePoolId, constant.CapPlanningInputStandardEdition)]
	if standardEdition != nil {
		standardEditionNumber = standardEdition.Number
	}
	professionalEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, resourcePoolId, constant.CapPlanningInputProfessionalEdition)]
	if professionalEdition != nil {
		professionalEditionNumber = professionalEdition.Number
	}
	enterpriseEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, resourcePoolId, constant.CapPlanningInputEnterpriseEdition)]
	if enterpriseEdition != nil {
		enterpriseEditionNumber = enterpriseEdition.Number
	}
	platinumEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v-%v", productCode, resourcePoolId, constant.CapPlanningInputPlatinumEdition)]
	if platinumEdition != nil {
		platinumEditionNumber = platinumEdition.Number
	}
	return brokerNumber, standardEditionNumber, professionalEditionNumber, enterpriseEditionNumber, platinumEditionNumber
}

const (
	PlatformName               = "平台规模授权"
	PlatformCode               = "Platform"
	PlatformBom                = "0100115148387809"
	SoftwareName               = "软件base"
	SoftwareCode               = "SoftwareBase"
	SoftwareBaseBom            = "0100115150861886"
	ServiceYearName            = "平台升级维保"
	ServiceYearCode            = "ServiceYear"
	DatabaseManagementName     = "数据库管理平台授权"
	DatabaseManagementCode     = "DatabaseManagementPlatform"
	DatabaseManagementBom      = "0100115140403032"
	SecurityMaintenanceName    = "安全产品统一维保"
	SecurityMaintenanceCode    = "SecurityMaintenance"
	SecurityMaintenanceBom     = "0100115099084508"
	CloudNativeSecBasicPkgName = "云原生安全-基础安全包"
	CloudNativeSecBasicPkgCode = "CloudNativeSecurityBasicPkg"
	CloudNativeSecBasicPkgBom  = "0100115808142197"
	BigDataManagementName      = "中电云大数据开发管理平台"
	BigDataManagementCode      = "BigDataManagementPlatform"
	BigDataManagementBom       = "0100115228973107"
	BigDataMaintenanceName     = "中电云大数据平台(维保服务)"
	BigDataMaintenanceCode     = "BigDataMaintenance"
	BigDataMaintenanceBom      = "0100115230260255"
	BigDataPlatformScaleName   = "大数据CeaInsight-平台规模授权"
	BigDataPlatformScaleCode   = "BigDataPlatformScale"
	BigDataPlatformScaleBom    = "0100115139411762"
)

var ServiceYearBom = map[int]string{1: "0100115152958526", 2: "0100115153975617", 3: "0100115154780568", 4: "0100115155303482", 5: "0100115156784743"}
