package software_bom

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"math"
	"net/http"
	"strconv"
	"strings"
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
	for k, v := range bomMap {
		if k == DatabaseManagementBom {
			//默认输出数据库管理平台授权，BOM iD：0100115140403032，单位：套
			softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: DatabaseManagementBom, CloudService: DatabaseManagementName, ServiceCode: DatabaseManagementCode, Number: v})
			continue
		}
		softwareBomLicenseBaseline := softwareData.SoftwareBomLicenseBaselineMap[k]
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
	//平台规模授权：0100115148387809，按云平台下服务器数量计算，N=整网所有服务器的物理CPU数量之和-管理减免（10）；N大于等于0
	var cpuNumber int
	for _, serverPlanning := range softwareData.ServerPlanningMap {
		serverBaseline := softwareData.ServerBaselineMap[serverPlanning.ServerBaselineId]
		cpuNumber += serverPlanning.Number * serverBaseline.CpuNum
	}
	cpuNumber = cpuNumber - 10
	if cpuNumber < 0 {
		cpuNumber = 0
	}
	softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: PlatformBom, CloudService: PlatformName, ServiceCode: PlatformCode, Number: cpuNumber})
	//软件base：0100115150861886，默认1套
	softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: SoftwareBaseBom, CloudService: SoftwareName, ServiceCode: SoftwareCode, Number: 1})
	//平台升级维保：根据选择年限对应不同BOM
	softwareBomPlanningList = append(softwareBomPlanningList, &entity.SoftwareBomPlanning{PlanId: planId, BomId: ServiceYearBom[softwareData.ServiceYear], CloudService: ServiceYearName, ServiceCode: ServiceYearCode, Number: 1})
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
	//产品id为key
	var cloudProductPlanningMap = make(map[int64]*entity.CloudProductPlanning)
	for _, v := range cloudProductPlanningList {
		productIdList = append(productIdList, v.ProductId)
		cloudProductPlanningMap[v.ProductId] = v
	}
	//查询云产品和角色关联表
	var cloudProductNodeRoleRelList []*entity.CloudProductNodeRoleRel
	if err := db.Where("product_id IN (?)", productIdList).Find(&cloudProductNodeRoleRelList).Error; err != nil {
		return nil, err
	}
	var nodeRoleIdList []int64
	for _, v := range cloudProductNodeRoleRelList {
		nodeRoleIdList = append(nodeRoleIdList, v.NodeRoleId)
	}
	//查询角色节点基线
	var nodeRoleBaselineList []*entity.NodeRoleBaseline
	if err := db.Where("id IN (?)", nodeRoleIdList).Find(&nodeRoleBaselineList).Error; err != nil {
		return nil, err
	}
	//角色id为key
	var nodeRoleCodeMap = make(map[int64]string)
	for _, v := range nodeRoleBaselineList {
		nodeRoleCodeMap[v.Id] = v.NodeRoleCode
	}
	//查询服务器规划
	var serverPlanningList []*entity.ServerPlanning
	if err := db.Where("plan_id = ?", planId).Find(&serverPlanningList).Error; err != nil {
		return nil, err
	}
	//角色节点code为key
	var serverPlanningMap = make(map[string]*entity.ServerPlanning)
	var serverBaselineIdList []int64
	for _, v := range serverPlanningList {
		nodeRoleCode := nodeRoleCodeMap[v.NodeRoleId]
		serverPlanningMap[nodeRoleCode] = v
		serverBaselineIdList = append(serverBaselineIdList, v.ServerBaselineId)
	}
	//查询服务器基线表
	var serverBaselineList []*entity.ServerBaseline
	if err := db.Where("id IN (?)", serverBaselineIdList).Find(&serverBaselineList).Error; err != nil {
		return nil, err
	}
	//服务器基线id为key
	var serverBaselineMap = make(map[int64]*entity.ServerBaseline)
	for _, v := range serverBaselineList {
		if v.Arch == constant.CpuArchARM {
			v.Arch = constant.CpuArchXC
		}
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
	//查询云产品基线
	var cloudProductBaselineList []*entity.CloudProductBaseline
	if err := db.Where("id IN (?)", productIdList).Find(&cloudProductBaselineList).Error; err != nil {
		return nil, err
	}
	//产品id为key
	var cloudProductBaselineMap = make(map[int64]*entity.CloudProductBaseline)
	var productCodeList []string
	for _, v := range cloudProductBaselineList {
		productCodeList = append(productCodeList, v.ProductCode)
		cloudProductBaselineMap[v.Id] = v
	}
	//查询软件bom表
	var softwareBomLicenseBaselineList []*entity.SoftwareBomLicenseBaseline
	if err := db.Where("service_code IN (?) AND version_id = ?", productCodeList, cloudProductPlanningList[0].VersionId).Find(&softwareBomLicenseBaselineList).Error; err != nil {
		return nil, err
	}
	//根据产品编码-售卖规格、产品编码-增值服务、产品编码-硬件架构 筛选容量输入列表
	var screenCloudProductSellSpecMap = make(map[string]interface{})
	var screenCloudProductValueAddedServiceMap = make(map[string]interface{})
	for _, v := range cloudProductPlanningList {
		//根据产品编码-售卖规格
		if util.IsNotBlank(v.SellSpec) {
			screenCloudProductSellSpecMap[fmt.Sprintf("%s-%s", cloudProductBaselineMap[v.ProductId].ProductCode, v.SellSpec)] = nil
		}
		//产品编码-增值服务
		if util.IsNotBlank(v.ValueAddedService) {
			for _, valueAddedService := range strings.Split(v.ValueAddedService, ",") {
				screenCloudProductValueAddedServiceMap[fmt.Sprintf("%s-%s", cloudProductBaselineMap[v.ProductId].ProductCode, valueAddedService)] = nil
			}
		}
	}
	//产品编码-售卖规格-增值服务-硬件架构为key
	var softwareBomLicenseBaselineListMap = make(map[string][]*entity.SoftwareBomLicenseBaseline)
	//软件bom为key
	var softwareBomLicenseBaselineMap = make(map[string]*entity.SoftwareBomLicenseBaseline)
	for _, v := range softwareBomLicenseBaselineList {
		//根据产品编码-售卖规格、产品编码-增值服务筛选容量输入列表
		if util.IsNotBlank(v.SellSpecs) {
			if _, ok := screenCloudProductSellSpecMap[fmt.Sprintf("%s-%s", v.ServiceCode, v.SellSpecs)]; !ok {
				continue
			}
		}
		if util.IsNotBlank(v.ValueAddedService) {
			if _, ok := screenCloudProductValueAddedServiceMap[fmt.Sprintf("%s-%s", v.ServiceCode, v.ValueAddedService)]; !ok {
				continue
			}
		}
		softwareBomLicenseBaselineListMap[v.ServiceCode] = append(softwareBomLicenseBaselineListMap[v.ServiceCode], v)
		softwareBomLicenseBaselineMap[v.BomId] = v
	}
	return &SoftwareData{
		ServiceYear:                       cloudProductPlanningList[0].ServiceYear,
		CloudProductBaselineList:          cloudProductBaselineList,
		ServerPlanningMap:                 serverPlanningMap,
		ServerBaselineMap:                 serverBaselineMap,
		ServerCapPlanningMap:              serverCapPlanningMap,
		SoftwareBomLicenseBaselineMap:     softwareBomLicenseBaselineMap,
		SoftwareBomLicenseBaselineListMap: softwareBomLicenseBaselineListMap,
	}, nil
}

func ComputingSoftwareBom(softwareData *SoftwareData) map[string]int {
	//bomId为key，数量为value
	var bomMap = make(map[string]int)
	serviceYear := softwareData.ServiceYear
	serverPlanningMap := softwareData.ServerPlanningMap
	serverBaselineMap := softwareData.ServerBaselineMap
	serverCapPlanningMap := softwareData.ServerCapPlanningMap
	for _, v := range softwareData.CloudProductBaselineList {
		productCode := v.ProductCode
		softwareBomLicenseBaselineList := softwareData.SoftwareBomLicenseBaselineListMap[productCode]
		if len(softwareBomLicenseBaselineList) == 0 {
			continue
		}
		switch productCode {
		case constant.ProductCodeECS, constant.ProductCodeBMS:
			//COMPUTE节点的CPU数量，BMS节点的CPU数量
			serverPlanning := serverPlanningMap[constant.NodeRoleCodeCompute]
			serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
			number := serverPlanning.Number * serverBaseline.CpuNum
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.HardwareArch == serverBaseline.Arch {
					bomMap[softwareBom.BomId] += number
				}
			}
		case constant.ProductCodeCKE:
			//CKE容量vCPU数量/100
			serverPlanning := serverPlanningMap[constant.NodeRoleCodeCompute]
			serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
			serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputVCpu)]
			if serverCapPlanning == nil {
				continue
			}
			number := int(math.Ceil(float64(serverCapPlanning.Number) / 100))
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.HardwareArch == serverBaseline.Arch {
					bomMap[softwareBom.BomId] += number
				}
			}
		case constant.ProductCodeCBR:
			//TB（还没给数据，默认先输出1）
			for _, softwareBom := range softwareBomLicenseBaselineList {
				bomMap[softwareBom.BomId] = 1
			}
		case constant.ProductCodeEBS, constant.ProductCodeEFS, constant.ProductCodeOSS:
			//TB，可用容量
			serverPlanning := serverPlanningMap[constant.NodeRoleCodeCompute]
			serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
			serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputStorageCapacity)]
			if serverCapPlanning == nil {
				continue
			}
			var number int
			if serverCapPlanning.Features == constant.FeaturesNameThreeCopies {
				number = int(math.Ceil(float64(serverPlanning.Number*serverBaseline.StorageDiskNum*serverBaseline.StorageDiskCapacity) / 1024 * 0.9 * 0.91 * 1 / 3))
			}
			if serverCapPlanning.Features == constant.FeaturesNameEC {
				number = int(math.Ceil(float64(serverPlanning.Number*serverBaseline.StorageDiskNum*serverBaseline.StorageDiskCapacity) / 1024 * 0.9 * 0.91 * 2 / 3))
			}
			for _, softwareBom := range softwareBomLicenseBaselineList {
				bomMap[softwareBom.BomId] += number
			}
		case constant.ProductCodeVPC:
			//NETWORK、NFV、BMSGW的CPU总数
			serverPlanningNETWORK := serverPlanningMap[constant.NodeRoleCodeNETWORK]
			serverBaselineNETWORK := serverBaselineMap[serverPlanningNETWORK.ServerBaselineId]
			serverPlanningNFV := serverPlanningMap[constant.NodeRoleCodeNFV]
			serverBaselineNFV := serverBaselineMap[serverPlanningNFV.ServerBaselineId]
			serverPlanningBMSGW := serverPlanningMap[constant.NodeRoleCodeBMSGW]
			serverBaselineBMSGW := serverBaselineMap[serverPlanningBMSGW.ServerBaselineId]
			number := serverPlanningNETWORK.Number*serverBaselineNETWORK.CpuNum + serverPlanningNFV.Number*serverBaselineNFV.CpuNum + serverPlanningBMSGW.Number*serverBaselineBMSGW.CpuNum
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.HardwareArch == serverBaselineNETWORK.Arch {
					bomMap[softwareBom.BomId] += number
				}
				if softwareBom.HardwareArch == serverBaselineNETWORK.Arch {
					bomMap[softwareBom.BomId] += number
				}
				if softwareBom.HardwareArch == serverBaselineNETWORK.Arch {
					bomMap[softwareBom.BomId] += number
				}
			}
		case constant.ProductCodeCNFW, constant.ProductCodeCWAF:
			//根据选择的容量数量计算；
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.SellType == constant.SoftwareBomLicense {
					serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputFirewall)]
					if serverCapPlanning != nil {
						bomMap[softwareBom.BomId] += serverCapPlanning.Number
					}
				}
				if softwareBom.SellType == constant.SoftwareBomMaintenance {
					bomMap[softwareBom.BomId] = serviceYear
				}
			}
		case constant.ProductCodeCSOC:
			//是3个容量输入对应3个bom，日志存储空间不足500G的为第一档，每G一个bom，超过500G的为第二档，每500G一个bom
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.SellType == constant.SoftwareBomLicense {
					if softwareBom.AuthorizedUnit == constant.SoftwareBomAuthorizedUnitAssetAccess {
						serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputAssetAccess)]
						if serverCapPlanning != nil {
							bomMap[softwareBom.BomId] += serverCapPlanning.Number
						}
					}
					if softwareBom.AuthorizedUnit == constant.SoftwareBomAuthorizedUnitLogStorage {
						serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputLogStorage)]
						if serverCapPlanning != nil {
							if softwareBom.CalcMethod == constant.SoftwareBomAuthorizedUnit500G {
								number := serverCapPlanning.Number / 500
								if number > 0 {
									bomMap[softwareBom.BomId] += number
								}
							} else {
								number := serverCapPlanning.Number % 500
								if number > 0 {
									bomMap[softwareBom.BomId] += number
								}
							}
						}
					}
					if softwareBom.ValueAddedService == constant.SoftwareBomValueAddedServiceVulnerabilityScanning {
						serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputVulnerabilityScanning)]
						if serverCapPlanning != nil {
							bomMap[softwareBom.BomId] += serverCapPlanning.Number
						}
					}
				}
				if softwareBom.SellType == constant.SoftwareBomMaintenance {
					bomMap[softwareBom.BomId] = serviceYear
				}
			}
		case constant.ProductCodeDSP:
			//根据数量匹配1-5，6-30以及30以上，三个阶梯的bom，再根据数据库个数，输出bom的数量，例如有26个数据库实例，则需要输出26个6-30bom
			serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputDatabaseAudit)]
			if serverCapPlanning == nil {
				continue
			}
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == "1-5实例" && serverCapPlanning.Number >= 1 && serverCapPlanning.Number <= 5 {
					if softwareBom.SellType == constant.SoftwareBomLicense {
						bomMap[softwareBom.BomId] += serverCapPlanning.Number
					}
					if softwareBom.SellType == constant.SoftwareBomMaintenance {
						bomMap[softwareBom.BomId] = serviceYear
					}
				}
				if softwareBom.CalcMethod == "6-30实例" && serverCapPlanning.Number >= 6 && serverCapPlanning.Number <= 30 {
					if softwareBom.SellType == constant.SoftwareBomLicense {
						bomMap[softwareBom.BomId] += serverCapPlanning.Number
					}
					if softwareBom.SellType == constant.SoftwareBomMaintenance {
						bomMap[softwareBom.BomId] = serviceYear
					}
				}
				if softwareBom.CalcMethod == "30实例以上" && serverCapPlanning.Number > 30 {
					if softwareBom.SellType == constant.SoftwareBomLicense {
						bomMap[softwareBom.BomId] += serverCapPlanning.Number
					}
					if softwareBom.SellType == constant.SoftwareBomMaintenance {
						bomMap[softwareBom.BomId] = serviceYear
					}
				}
			}
		case constant.ProductCodeCNBH:
			//根据包数量直接转换为bom数量
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.SellSpecs == constant.CapPlanningInputOneHundred {
					serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputOneHundred)]
					if serverCapPlanning != nil && softwareBom.SellType == constant.SoftwareBomLicense {
						bomMap[softwareBom.BomId] += serverCapPlanning.Number
					}
					if softwareBom.SellType == constant.SoftwareBomMaintenance {
						bomMap[softwareBom.BomId] = serviceYear
					}
				}
				if softwareBom.SellSpecs == constant.CapPlanningInputFiveHundred {
					serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputFiveHundred)]
					if serverCapPlanning != nil && softwareBom.SellType == constant.SoftwareBomLicense {
						bomMap[softwareBom.BomId] += serverCapPlanning.Number
					}
					if softwareBom.SellType == constant.SoftwareBomMaintenance {
						bomMap[softwareBom.BomId] = serviceYear
					}
				}
				if softwareBom.SellSpecs == constant.CapPlanningInputOneThousand {
					serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputOneThousand)]
					if serverCapPlanning != nil && softwareBom.SellType == constant.SoftwareBomLicense {
						bomMap[softwareBom.BomId] += serverCapPlanning.Number
					}
					if softwareBom.SellType == constant.SoftwareBomMaintenance {
						bomMap[softwareBom.BomId] = serviceYear
					}
				}
			}
		case constant.ProductCodeCWP:
			//还没给数据，默认先输出1
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.SellType == constant.SoftwareBomLicense {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.SellType == constant.SoftwareBomMaintenance {
					bomMap[softwareBom.BomId] = serviceYear
				}
			}
		case constant.ProductCodeDES:
			//没有输入，输出1
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.SellType == constant.SoftwareBomLicense {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.SellType == constant.SoftwareBomMaintenance {
					bomMap[softwareBom.BomId] = serviceYear
				}
			}
		case constant.ProductCodeCEASQLTX, constant.ProductCodeMYSQL, constant.ProductCodeCEASQLDW, constant.ProductCodeCEASQLCK, constant.ProductCodeREDIS, constant.ProductCodePOSTGRESQL:
			serverPlanning := serverPlanningMap[constant.NodeRoleCodeDATABASE]
			serverBaseline := serverBaselineMap[serverPlanning.ServerBaselineId]
			serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputBusinessDataVolume)]
			for _, softwareBom := range softwareBomLicenseBaselineList {
				bomMap[DatabaseManagementBom] = 1
				if serverCapPlanning != nil {
					if serverBaseline.Arch == constant.CpuArchX86 {
						bomMap[softwareBom.BomId] += serverCapPlanning.Number / 2
					}
					if serverBaseline.Arch == constant.CpuArchXC {
						bomMap[softwareBom.BomId] += serverCapPlanning.Number
					}
				}
			}
		case constant.ProductCodeDTS:
			serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputLinks)]
			for _, softwareBom := range softwareBomLicenseBaselineList {
				bomMap[DatabaseManagementBom] = 1
				if serverCapPlanning != nil {
					bomMap[softwareBom.BomId] += serverCapPlanning.Number
				}
			}
		case constant.ProductCodeKAFKA, constant.ProductCodeRABBITMQ:
			broker := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputBroker)]
			if broker == nil {
				continue
			}
			var standardEditionNumber, professionalEditionNumber, enterpriseEditionNumber, platinumEditionNumber int
			standardEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputStandardEdition)]
			if standardEdition != nil {
				standardEditionNumber = standardEdition.Number
			}
			professionalEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputProfessionalEdition)]
			if professionalEdition != nil {
				professionalEditionNumber = professionalEdition.Number
			}
			enterpriseEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputEnterpriseEdition)]
			if enterpriseEdition != nil {
				enterpriseEditionNumber = enterpriseEdition.Number
			}
			platinumEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputPlatinumEdition)]
			if enterpriseEdition != nil {
				platinumEditionNumber = platinumEdition.Number
			}
			number := (standardEditionNumber*2 + professionalEditionNumber*4 + enterpriseEditionNumber*8 + platinumEditionNumber*16) * broker.Number
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == "基础包，200vCPU" {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.CalcMethod == "扩展包，100vCPU" {
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
				if softwareBom.CalcMethod == "基础包，500个微服务实例" {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.CalcMethod == "扩展包，100个微服务实例" {
					if number-500 > 0 {
						bomMap[softwareBom.BomId] = int(math.Ceil(float64(number-200) / 100))
					}
				}
			}
		case constant.ProductCodeROCKETMQ:
			var standardEditionNumber, professionalEditionNumber, enterpriseEditionNumber, platinumEditionNumber int
			standardEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputStandardEdition)]
			if standardEdition != nil {
				standardEditionNumber = standardEdition.Number
			}
			professionalEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputProfessionalEdition)]
			if professionalEdition != nil {
				professionalEditionNumber = professionalEdition.Number
			}
			enterpriseEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputEnterpriseEdition)]
			if enterpriseEdition != nil {
				enterpriseEditionNumber = enterpriseEdition.Number
			}
			platinumEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputPlatinumEdition)]
			if enterpriseEdition != nil {
				platinumEditionNumber = platinumEdition.Number
			}
			number := standardEditionNumber*12 + professionalEditionNumber*24 + enterpriseEditionNumber*36 + platinumEditionNumber*48
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == "基础包，200vCPU" {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.CalcMethod == "扩展包，100vCPU" {
					if number-200 > 0 {
						bomMap[softwareBom.BomId] = int(math.Ceil(float64(number-200) / 100))
					}
				}
			}
		case constant.ProductCodeAPIM:
			var standardEditionNumber, professionalEditionNumber, enterpriseEditionNumber int
			standardEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputStandardEdition)]
			if standardEdition != nil {
				standardEditionNumber = standardEdition.Number
			}
			professionalEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputProfessionalEdition)]
			if professionalEdition != nil {
				professionalEditionNumber = professionalEdition.Number
			}
			enterpriseEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputEnterpriseEdition)]
			if enterpriseEdition != nil {
				enterpriseEditionNumber = enterpriseEdition.Number
			}
			number := standardEditionNumber*3 + professionalEditionNumber*6 + enterpriseEditionNumber*12
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == "基础包，200vCPU" {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.CalcMethod == "扩展包，100vCPU" {
					if number-200 > 0 {
						bomMap[softwareBom.BomId] = int(math.Ceil(float64(number-200) / 100))
					}
				}
			}
		case constant.ProductCodeCONNECT:
			var standardEditionNumber, professionalEditionNumber, enterpriseEditionNumber int
			standardEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputStandardEdition)]
			if standardEdition != nil {
				standardEditionNumber = standardEdition.Number
			}
			professionalEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputProfessionalEdition)]
			if professionalEdition != nil {
				professionalEditionNumber = professionalEdition.Number
			}
			enterpriseEdition := serverCapPlanningMap[fmt.Sprintf("%v-%v", productCode, constant.CapPlanningInputEnterpriseEdition)]
			if enterpriseEdition != nil {
				enterpriseEditionNumber = enterpriseEdition.Number
			}
			number := standardEditionNumber*4 + professionalEditionNumber*8 + enterpriseEditionNumber*24
			for _, softwareBom := range softwareBomLicenseBaselineList {
				if softwareBom.CalcMethod == "基础包，200vCPU" {
					bomMap[softwareBom.BomId] = 1
				}
				if softwareBom.CalcMethod == "扩展包，100vCPU" {
					if number-200 > 0 {
						bomMap[softwareBom.BomId] = int(math.Ceil(float64(number-200) / 100))
					}
				}
			}
		case constant.ProductCodeCLCP:
		case constant.ProductCodeCOS:
		case constant.ProductCodeCLS:
		}
	}
	return bomMap
}

const (
	PlatformName           = "平台规模授权"
	PlatformCode           = "Platform"
	PlatformBom            = "0100115148387809"
	SoftwareName           = "软件base"
	SoftwareCode           = "SoftwareBase"
	SoftwareBaseBom        = "0100115150861886"
	ServiceYearName        = "平台升级维保"
	ServiceYearCode        = "ServiceYear"
	DatabaseManagementName = "数据库管理平台授权"
	DatabaseManagementCode = "DatabaseManagementPlatform"
	DatabaseManagementBom  = "0100115140403032"
)

var ServiceYearBom = map[int]string{1: "0100115152958526", 2: "0100115153975617", 3: "0100115154780568", 4: "0100115155303482", 5: "0100115156784743"}
