package baseline

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"code.cestc.cn/ccos/cnm/ops-base/tools/stringx"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/excel"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
)

func Import(context *gin.Context) {
	file, err := context.FormFile("file")
	if err != nil {
		log.Error(err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if file.Size == 0 {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	cloudPlatformType := context.PostForm("cloudPlatformType")
	baselineVersion := context.PostForm("baselineVersion")
	baselineType := context.PostForm("baselineType")
	releaseTime := context.PostForm("releaseTime")
	if cloudPlatformType == "" || baselineVersion == "" || baselineType == "" || releaseTime == "" {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	cloudPlatformTypes, err := QueryCloudPlatformType()
	if err != nil {
		log.Error(err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if !stringx.Contains(cloudPlatformTypes, cloudPlatformType) {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	softwareVersion, err := QuerySoftwareVersionByVersion(baselineVersion, cloudPlatformType)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}

	now := datetime.GetNow()
	if softwareVersion.Id > 0 {
		// 编辑软件版本
		if err := UpdateSoftwareVersion(softwareVersion); err != nil {
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return
		}
	} else {
		softwareVersion = entity.SoftwareVersion{
			SoftwareVersion:   baselineVersion,
			CloudPlatformType: cloudPlatformType,
			ReleaseTime:       datetime.StrToTime(datetime.FullTimeFmt, releaseTime),
			CreateTime:        now,
		}
		// 新增软件版本
		if err := CreateSoftwareVersion(&softwareVersion); err != nil {
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return
		}
	}
	filePath := fmt.Sprintf("%s/%s-%d-%d.xlsx", "exampledir", "baseline", time.Now().Unix(), rand.Uint32())
	if err := context.SaveUploadedFile(file, filePath); err != nil {
		log.Error(err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Errorf("excelize openFile error: %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		if err := f.Close(); err != nil {
			log.Errorf("excelize close error: %v", err)
		}
		if err := os.Remove(filePath); err != nil {
			log.Errorf("os removeFile error: %v", err)
		}
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("excelize close error: %v", err)
		}
		if err := os.Remove(filePath); err != nil {
			log.Errorf("os removeFile error: %v", err)
		}
	}()
	switch baselineType {
	case NodeRoleBaselineType:
		if ImportNodeRoleBaseline(context, softwareVersion.Id, f) {
			return
		}
		break
	case CloudProductBaselineType:
		if ImportCloudProductBaseline(context, softwareVersion.Id, f) {
			return
		}
		break
	case ServerBaselineType:
		if ImportServerBaseline(context, softwareVersion.Id, f) {
			return
		}
		break
	case NetworkDeviceRoleBaselineType:
		if ImportNetworkDeviceRoleBaseline(context, softwareVersion.Id, f) {
			return
		}
		break
	case NetworkDeviceBaselineType:
		if ImportNetworkDeviceBaseline(context, softwareVersion.Id, f) {
			return
		}
		break
	case IPDemandBaselineType:
		if ImportIPDemandBaseline(context, softwareVersion.Id, f) {
			return
		}
		break
	case CapConvertBaselineType:
		if ImportCapConvertBaseline(context, softwareVersion.Id, f) {
			return
		}
		break
	case CapActualResBaselineType:
		if ImportCapActualResBaseline(context, softwareVersion.Id, f) {
			return
		}
		break
	case CapServerCalcBaselineType:
		if ImportCapServerCalcBaseline(context, softwareVersion.Id, f) {
			return
		}
		break
	default:
		break
	}
	result.Success(context, nil)
}

func ImportCloudProductBaseline(context *gin.Context, versionId int64, f *excelize.File) bool {
	// 先查询节点角色表，导入的版本是否已有数据，如没有，提示先导入节点角色基线
	nodeRoleBaselines, err := QueryNodeRoleBaselineByVersionId(versionId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			result.Failure(context, errorcodes.NodeRoleMustImportFirst, http.StatusBadRequest)
		} else {
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		}
		return true
	}
	var cloudProductBaselineExcelList []CloudProductBaselineExcel
	if err := excel.ImportBySheet(f, &cloudProductBaselineExcelList, CloudProductBaselineSheetName, 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return true
	}
	if len(cloudProductBaselineExcelList) > 0 {
		var cloudProductBaselines []entity.CloudProductBaseline
		for i := range cloudProductBaselineExcelList {
			dependProductCode := cloudProductBaselineExcelList[i].DependProductCode
			if dependProductCode != "" {
				cloudProductBaselineExcelList[i].DependProductCodes = util.SplitString(dependProductCode, constant.SplitLineBreak)
			}
			controlResNodeRoleCode := cloudProductBaselineExcelList[i].ControlResNodeRoleCode
			if controlResNodeRoleCode != "" {
				cloudProductBaselineExcelList[i].ControlResNodeRoleCodes = util.SplitString(controlResNodeRoleCode, constant.SplitLineBreak)
			}
			resNodeRoleCode := cloudProductBaselineExcelList[i].ResNodeRoleCode
			if resNodeRoleCode != "" {
				cloudProductBaselineExcelList[i].ResNodeRoleCodes = util.SplitString(resNodeRoleCode, constant.SplitLineBreak)
			}
			whetherRequired := cloudProductBaselineExcelList[i].WhetherRequired
			whetherRequiredType := constant.WhetherRequiredNo
			if whetherRequired == constant.WhetherRequiredYesChinese {
				whetherRequiredType = constant.WhetherRequiredYes
			}
			cloudProductBaselines = append(cloudProductBaselines, entity.CloudProductBaseline{
				VersionId:       versionId,
				ProductType:     cloudProductBaselineExcelList[i].ProductType,
				ProductName:     cloudProductBaselineExcelList[i].ProductName,
				ProductCode:     cloudProductBaselineExcelList[i].ProductCode,
				SellSpecs:       cloudProductBaselineExcelList[i].SellSpecs,
				AuthorizedUnit:  cloudProductBaselineExcelList[i].AuthorizedUnit,
				WhetherRequired: whetherRequiredType,
				Instructions:    cloudProductBaselineExcelList[i].Instructions,
			})
		}
		originCloudProductBaselines, err := QueryCloudProductBaselineByVersionId(versionId)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return true
		}
		if len(originCloudProductBaselines) > 0 {
			// TODO 先查询云产品基线表，看看相同的版本号是否已存在数据，如果已存在，需要先删除已有数据
		} else {
			if err := BatchCreateCloudProductBaseline(cloudProductBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			cloudProductCodeMap := make(map[string]int64)
			for _, cloudProductBaseline := range cloudProductBaselines {
				cloudProductCodeMap[cloudProductBaseline.ProductCode] = cloudProductBaseline.Id
			}
			nodeRoleCodeMap := make(map[string]int64)
			for _, nodeRoleBaseline := range nodeRoleBaselines {
				nodeRoleCodeMap[nodeRoleBaseline.NodeRoleCode] = nodeRoleBaseline.Id
			}
			for _, cloudProductBaselineExcel := range cloudProductBaselineExcelList {
				productId := cloudProductCodeMap[cloudProductBaselineExcel.ProductCode]
				// 处理依赖服务编码
				dependProductCodes := cloudProductBaselineExcel.DependProductCodes
				if len(dependProductCodes) > 0 {
					var cloudProductDependRels []entity.CloudProductDependRel
					for _, dependProductCode := range dependProductCodes {
						cloudProductDependProductCode, ok := cloudProductCodeMap[dependProductCode]
						if !ok {
							log.Errorf("import cloudProductBaseline invalid dependCode: %s", dependProductCode)
							result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
							return true
						}
						cloudProductDependRels = append(cloudProductDependRels, entity.CloudProductDependRel{
							ProductId:       productId,
							DependProductId: cloudProductDependProductCode,
						})
					}
					if err := BatchCreateCloudProductDependRel(cloudProductDependRels); err != nil {
						result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
						return true
					}
				}
				// 处理管控资源节点角色和资源节点角色
				controlResNodeRoleCodes := cloudProductBaselineExcel.ControlResNodeRoleCodes
				var cloudProductNodeRoleRels []entity.CloudProductNodeRoleRel
				if len(controlResNodeRoleCodes) > 0 {
					for _, controlResNodeRoleCode := range controlResNodeRoleCodes {
						nodeRoleId, ok := nodeRoleCodeMap[controlResNodeRoleCode]
						if !ok {
							log.Errorf("import cloudProductBaseline invalid controlResNodeRoleCode: %s", controlResNodeRoleCode)
							result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
							return true
						}
						cloudProductNodeRoleRels = append(cloudProductNodeRoleRels, entity.CloudProductNodeRoleRel{
							ProductId:    productId,
							NodeRoleId:   nodeRoleId,
							NodeRoleType: constant.ControlNodeRoleType,
						})
					}
				}
				resNodeRoleCodes := cloudProductBaselineExcel.ResNodeRoleCodes
				if len(resNodeRoleCodes) > 0 {
					for _, resNodeRoleCode := range resNodeRoleCodes {
						nodeRoleId, ok := nodeRoleCodeMap[resNodeRoleCode]
						if !ok {
							log.Errorf("import cloudProductBaseline invalid resNodeRoleCode: %s", resNodeRoleCode)
							result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
							return true
						}
						cloudProductNodeRoleRels = append(cloudProductNodeRoleRels, entity.CloudProductNodeRoleRel{
							ProductId:    productId,
							NodeRoleId:   nodeRoleId,
							NodeRoleType: constant.ResNodeRoleType,
						})
					}
				}
				if len(cloudProductNodeRoleRels) > 0 {
					if err := BatchCreateCloudProductNodeRoleRel(cloudProductNodeRoleRels); err != nil {
						result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
						return true
					}
				}
			}
		}
	}
	return false
}

func ImportNodeRoleBaseline(context *gin.Context, versionId int64, f *excelize.File) bool {
	var nodeRoleBaselineExcelList []NodeRoleBaselineExcel
	if err := excel.ImportBySheet(f, &nodeRoleBaselineExcelList, NodeRoleBaselineSheetName, 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return true
	}
	if len(nodeRoleBaselineExcelList) > 0 {
		var nodeRoleBaselines []entity.NodeRoleBaseline
		for i := range nodeRoleBaselineExcelList {
			mixedDeploy := nodeRoleBaselineExcelList[i].MixedDeploy
			if mixedDeploy != "" {
				nodeRoleBaselineExcelList[i].MixedDeploys = util.SplitString(mixedDeploy, constant.SplitLineBreak)
			}
			supportDPDK := constant.NodeRoleNotSupportDPDK
			if nodeRoleBaselineExcelList[i].SupportDPDK == constant.NodeRoleSupportDPDKCn {
				supportDPDK = constant.NodeRoleSupportDPDK
			}
			nodeRoleBaselines = append(nodeRoleBaselines, entity.NodeRoleBaseline{
				VersionId:    versionId,
				NodeRoleCode: nodeRoleBaselineExcelList[i].NodeRoleCode,
				NodeRoleName: nodeRoleBaselineExcelList[i].NodeRoleName,
				MinimumNum:   nodeRoleBaselineExcelList[i].MinimumCount,
				DeployMethod: nodeRoleBaselineExcelList[i].DeployMethod,
				SupportDPDK:  supportDPDK,
				Classify:     nodeRoleBaselineExcelList[i].Classify,
				Annotation:   nodeRoleBaselineExcelList[i].Annotation,
				BusinessType: nodeRoleBaselineExcelList[i].BusinessType,
			})
		}
		originNodeRoleBaselines, err := QueryNodeRoleBaselineByVersionId(versionId)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return true
		}
		if len(originNodeRoleBaselines) > 0 {
			// 识别出新增、修改、删除的数据

		} else {
			if err := BatchCreateNodeRoleBaseline(nodeRoleBaselines); err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			nodeRoleCodeMap := make(map[string]int64)
			for _, nodeRoleBaseline := range nodeRoleBaselines {
				nodeRoleCodeMap[nodeRoleBaseline.NodeRoleCode] = nodeRoleBaseline.Id
			}
			for _, nodeRoleBaselineExcel := range nodeRoleBaselineExcelList {
				nodeRoleCode := nodeRoleBaselineExcel.NodeRoleCode
				mixedDeploys := nodeRoleBaselineExcel.MixedDeploys
				nodeRoleId := nodeRoleCodeMap[nodeRoleCode]
				if len(mixedDeploys) > 0 {
					var mixedNodeRoles []entity.NodeRoleMixedDeploy
					for _, mixedDeploy := range mixedDeploys {
						mixDeployNodeRoleId, ok := nodeRoleCodeMap[mixedDeploy]
						if !ok {
							log.Infof("import nodeRoleBaseline fail, can not find mixDeployNodeRoleCode: %s", mixedDeploy)
							result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
							return true
						}
						mixedNodeRoles = append(mixedNodeRoles, entity.NodeRoleMixedDeploy{
							NodeRoleId:      nodeRoleId,
							MixedNodeRoleId: mixDeployNodeRoleId,
						})
					}
					if len(mixedNodeRoles) > 0 {
						if err := BatchCreateNodeRoleMixedDeploy(mixedNodeRoles); err != nil {
							result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func ImportServerBaseline(context *gin.Context, versionId int64, f *excelize.File) bool {
	// 先查询节点角色表，导入的版本是否已有数据，如没有，提示先导入节点角色基线
	nodeRoleBaselines, err := QueryNodeRoleBaselineByVersionId(versionId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			result.Failure(context, errorcodes.NodeRoleMustImportFirst, http.StatusBadRequest)
		} else {
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		}
		return true
	}
	var serverBaselineExcelList []ServerBaselineExcel
	if err := excel.ImportBySheet(f, &serverBaselineExcelList, ServerBaselineSheetName, 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return true
	}
	if len(serverBaselineExcelList) > 0 {
		var serverBaselines []entity.ServerBaseline
		for i := range serverBaselineExcelList {
			nodeRoleCode := serverBaselineExcelList[i].NodeRoleCode
			if nodeRoleCode != "" {
				serverBaselineExcelList[i].NodeRoleCodes = util.SplitString(nodeRoleCode, constant.SplitLineBreak)
			}
			serverBaselines = append(serverBaselines, entity.ServerBaseline{
				VersionId:           versionId,
				Arch:                serverBaselineExcelList[i].Arch,
				NetworkInterface:    serverBaselineExcelList[i].NetworkInterface,
				BomCode:             serverBaselineExcelList[i].BomCode,
				ConfigurationInfo:   serverBaselineExcelList[i].ConfigurationInfo,
				Spec:                serverBaselineExcelList[i].Spec,
				CpuType:             serverBaselineExcelList[i].CpuType,
				Cpu:                 serverBaselineExcelList[i].Cpu,
				Gpu:                 serverBaselineExcelList[i].Gpu,
				Memory:              serverBaselineExcelList[i].Memory,
				SystemDiskType:      serverBaselineExcelList[i].SystemDiskType,
				SystemDisk:          serverBaselineExcelList[i].SystemDisk,
				StorageDiskType:     serverBaselineExcelList[i].StorageDiskType,
				StorageDiskNum:      serverBaselineExcelList[i].StorageDiskNum,
				StorageDiskCapacity: serverBaselineExcelList[i].StorageDiskCapacity,
				RamDisk:             serverBaselineExcelList[i].RamDisk,
				NetworkCardNum:      serverBaselineExcelList[i].NetworkCardNum,
				Power:               serverBaselineExcelList[i].Power,
			})
		}
		originServerBaselines, err := QueryServerBaselineByVersionId(versionId)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return true
		}
		if len(originServerBaselines) > 0 {
			// TODO 先查询服务器基线表，看看相同的版本号是否已存在数据，如果已存在，需要先删除已有数据
		} else {
			if err := BatchCreateServerBaseline(serverBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			serverModelMap := make(map[string]int64)
			for _, serverBaseline := range serverBaselines {
				serverModelMap[serverBaseline.BomCode] = serverBaseline.Id
			}
			nodeRoleCodeMap := make(map[string]int64)
			for _, nodeRoleBaseline := range nodeRoleBaselines {
				nodeRoleCodeMap[nodeRoleBaseline.NodeRoleCode] = nodeRoleBaseline.Id
			}
			for _, serverBaselineExcel := range serverBaselineExcelList {
				// 处理节点角色
				nodeRoleCodes := serverBaselineExcel.NodeRoleCodes
				if len(nodeRoleCodes) > 0 {
					var serverNodeRoleRels []entity.ServerNodeRoleRel
					serverId := serverModelMap[serverBaselineExcel.BomCode]
					for _, nodeRoleCode := range nodeRoleCodes {
						nodeRoleId, ok := nodeRoleCodeMap[nodeRoleCode]
						if !ok {
							log.Infof("import serverBaseline fail, can not find nodeRoleCode: %s", nodeRoleCode)
							result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
							return true
						}
						serverNodeRoleRels = append(serverNodeRoleRels, entity.ServerNodeRoleRel{
							ServerId:   serverId,
							NodeRoleId: nodeRoleId,
						})
					}
					if err := BatchCreateServerNodeRoleRel(serverNodeRoleRels); err != nil {
						result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
						return true
					}
				}
			}
		}
	}
	return false
}

func ImportNetworkDeviceRoleBaseline(context *gin.Context, versionId int64, f *excelize.File) bool {
	// 先查询节点角色表，导入的版本是否已有数据，如没有，提示先导入节点角色基线
	nodeRoleBaselines, err := QueryNodeRoleBaselineByVersionId(versionId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			result.Failure(context, errorcodes.NodeRoleMustImportFirst, http.StatusBadRequest)
		} else {
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		}
		return true
	}
	var networkDeviceRoleBaselineExcelList []NetworkDeviceRoleBaselineExcel
	if err := excel.ImportBySheet(f, &networkDeviceRoleBaselineExcelList, NetworkDeviceRoleBaselineSheetName, 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return true
	}
	if len(networkDeviceRoleBaselineExcelList) > 0 {
		var networkDeviceRoleBaselines []entity.NetworkDeviceRoleBaseline
		for i := range networkDeviceRoleBaselineExcelList {
			twoNetworkIso := networkDeviceRoleBaselineExcelList[i].TwoNetworkIso
			threeNetworkIso := networkDeviceRoleBaselineExcelList[i].ThreeNetworkIso
			triplePlay := networkDeviceRoleBaselineExcelList[i].TriplePlay
			var twoNetworkIsoEnum int
			var threeNetworkIsoEnum int
			var triplePlayEnum int
			if twoNetworkIso == constant.NetworkModelYesChinese {
				twoNetworkIsoEnum = constant.NetworkModelYes
			} else if twoNetworkIso == "" || twoNetworkIso == constant.NetworkModelNoChinese {
				twoNetworkIsoEnum = constant.NetworkModelNo
			} else {
				twoNetworkIsoEnum = constant.NeedQueryOtherTable
				networkDeviceRoleBaselineExcelList[i].TwoNetworkIsos = util.SplitString(twoNetworkIso, constant.SplitLineBreak)
			}

			if threeNetworkIso == constant.NetworkModelYesChinese {
				threeNetworkIsoEnum = constant.NetworkModelYes
			} else if threeNetworkIso == "" || threeNetworkIso == constant.NetworkModelNoChinese {
				threeNetworkIsoEnum = constant.NetworkModelNo
			} else {
				threeNetworkIsoEnum = constant.NeedQueryOtherTable
				networkDeviceRoleBaselineExcelList[i].ThreeNetworkIsos = util.SplitString(threeNetworkIso, constant.SplitLineBreak)
			}

			if triplePlay == constant.NetworkModelYesChinese {
				triplePlayEnum = constant.NetworkModelYes
			} else if triplePlay == "" || triplePlay == constant.NetworkModelNoChinese {
				triplePlayEnum = constant.NetworkModelNo
			} else {
				triplePlayEnum = constant.NeedQueryOtherTable
				networkDeviceRoleBaselineExcelList[i].TriplePlays = util.SplitString(triplePlay, constant.SplitLineBreak)
			}
			networkDeviceRoleBaselines = append(networkDeviceRoleBaselines, entity.NetworkDeviceRoleBaseline{
				VersionId:       versionId,
				DeviceType:      networkDeviceRoleBaselineExcelList[i].DeviceType,
				FuncType:        networkDeviceRoleBaselineExcelList[i].FuncType,
				FuncCompoName:   networkDeviceRoleBaselineExcelList[i].FuncCompoName,
				FuncCompoCode:   networkDeviceRoleBaselineExcelList[i].FuncCompoCode,
				Description:     networkDeviceRoleBaselineExcelList[i].Description,
				TwoNetworkIso:   twoNetworkIsoEnum,
				ThreeNetworkIso: threeNetworkIsoEnum,
				TriplePlay:      triplePlayEnum,
				MinimumNumUnit:  networkDeviceRoleBaselineExcelList[i].MinimumNumUnit,
				UnitDeviceNum:   networkDeviceRoleBaselineExcelList[i].UnitDeviceNum,
				DesignSpec:      networkDeviceRoleBaselineExcelList[i].DesignSpec,
			})
		}

		originNetworkDeviceRoleBaselines, err := QueryNetworkDeviceRoleBaselineByVersionId(versionId)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return true
		}
		if len(originNetworkDeviceRoleBaselines) > 0 {
			// TODO 该版本之前已导入数据，需删除所有数据，范围巨大。。。必须重新导入其他所有基线
		} else {
			if err := BatchCreateNetworkDeviceRoleBaseline(networkDeviceRoleBaselines); err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			nodeRoleCodeMap := make(map[string]int64)
			for _, nodeRoleBaseline := range nodeRoleBaselines {
				nodeRoleCodeMap[nodeRoleBaseline.NodeRoleCode] = nodeRoleBaseline.Id
			}
			networkDeviceRoleCodeMap := make(map[string]int64)
			for _, networkDeviceRoleBaseline := range networkDeviceRoleBaselines {
				networkDeviceRoleCodeMap[networkDeviceRoleBaseline.FuncCompoCode] = networkDeviceRoleBaseline.Id
			}
			var networkModelRoleRels []entity.NetworkModelRoleRel
			for _, networkDeviceRoleBaselineExcel := range networkDeviceRoleBaselineExcelList {
				networkDeviceRoleId := networkDeviceRoleCodeMap[networkDeviceRoleBaselineExcel.FuncCompoCode]
				twoNetworkIsos := networkDeviceRoleBaselineExcel.TwoNetworkIsos
				threeNetworkIsos := networkDeviceRoleBaselineExcel.ThreeNetworkIsos
				triplePlays := networkDeviceRoleBaselineExcel.TriplePlays
				networkModelRoleRels, err = HandleNetworkModelRoleRels(networkDeviceRoleId, twoNetworkIsos, nodeRoleCodeMap, networkDeviceRoleCodeMap, networkModelRoleRels, 2)
				if err != nil {
					result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
					return true
				}
				networkModelRoleRels, err = HandleNetworkModelRoleRels(networkDeviceRoleId, threeNetworkIsos, nodeRoleCodeMap, networkDeviceRoleCodeMap, networkModelRoleRels, 3)
				if err != nil {
					result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
					return true
				}
				networkModelRoleRels, err = HandleNetworkModelRoleRels(networkDeviceRoleId, triplePlays, nodeRoleCodeMap, networkDeviceRoleCodeMap, networkModelRoleRels, 1)
				if err != nil {
					result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
					return true
				}
			}
			if len(networkModelRoleRels) > 0 {
				if err := BatchCreateNetworkModelRoleRel(networkModelRoleRels); err != nil {
					log.Error(err)
					result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
					return true
				}
			}
		}
	}
	return false
}

func HandleNetworkModelRoleRels(networkDeviceRoleId int64, networkModelRoles []string, nodeRoleCodeMap map[string]int64, networkDeviceRoleCodeMap map[string]int64, networkModelRoleRels []entity.NetworkModelRoleRel, networkModel int) ([]entity.NetworkModelRoleRel, error) {
	for _, networkModelRole := range networkModelRoles {
		var associatedType int
		roleNum, roleCode := GetRoleNameAndNum(networkModelRole)
		roleId, ok := nodeRoleCodeMap[roleCode]
		if !ok {
			roleId, ok = networkDeviceRoleCodeMap[roleCode]
			if !ok {
				errorString := fmt.Sprintf("import networkDeviceRoleBaseline fail, can not find nodeRoleCode or networkDeviceRoleCode: %s", roleCode)
				log.Info(errorString)
				return networkModelRoleRels, errors.New(errorString)
			}
			associatedType = constant.NetworkDeviceRoleType
		} else {
			associatedType = constant.NodeRoleType
		}
		networkModelRoleRels = append(networkModelRoleRels, entity.NetworkModelRoleRel{
			NetworkDeviceRoleId: networkDeviceRoleId,
			NetworkModel:        networkModel,
			AssociatedType:      associatedType,
			RoleId:              roleId,
			RoleNum:             roleNum,
		})
	}
	return networkModelRoleRels, nil
}

func GetRoleNameAndNum(role string) (int, string) {
	if role != "" {
		if strings.Contains(role, constant.SplitLineAsterisk) {
			roles := util.SplitString(role, constant.SplitLineAsterisk)
			num, err := strconv.Atoi(roles[len(roles)-1])
			if err != nil {
				log.Error("get roleNum error: ", err)
			}
			return num, roles[0]
		} else {
			return 1, role
		}
	}
	return 0, role
}

func ImportNetworkDeviceBaseline(context *gin.Context, versionId int64, f *excelize.File) bool {
	// 先查询网络设备角色表，导入的版本是否已有数据，如没有，提示先导入网络设备角色基线
	networkDeviceRoleBaselines, err := QueryNetworkDeviceRoleBaselineByVersionId(versionId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			result.Failure(context, errorcodes.NetworkDeviceRoleMustImportFirst, http.StatusBadRequest)
		} else {
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		}
		return true
	}
	var networkDeviceBaselineExcelList []NetworkDeviceBaselineExcel
	if err := excel.ImportBySheet(f, &networkDeviceBaselineExcelList, NetworkDeviceBaselineSheetName, 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return true
	}
	if len(networkDeviceBaselineExcelList) > 0 {
		var networkDeviceBaselines []entity.NetworkDeviceBaseline
		for i := range networkDeviceBaselineExcelList {
			networkDeviceRole := networkDeviceBaselineExcelList[i].NetworkDeviceRoleCode
			if networkDeviceRole != "" {
				networkDeviceBaselineExcelList[i].NetworkDeviceRoleCodes = util.SplitString(networkDeviceRole, constant.SplitLineBreak)
			}
			var deviceType int
			if networkDeviceBaselineExcelList[i].DeviceType == constant.NetworkDeviceTypeXinchuangCn {
				deviceType = constant.NetworkDeviceTypeXinchuang
			} else {
				deviceType = constant.NetworkDeviceTypeCommercial
			}
			networkDeviceBaselines = append(networkDeviceBaselines, entity.NetworkDeviceBaseline{
				VersionId:    versionId,
				DeviceModel:  networkDeviceBaselineExcelList[i].DeviceModel,
				Manufacturer: networkDeviceBaselineExcelList[i].Manufacturer,
				DeviceType:   deviceType,
				NetworkModel: networkDeviceBaselineExcelList[i].NetworkModel,
				ConfOverview: networkDeviceBaselineExcelList[i].ConfOverview,
				Purpose:      networkDeviceBaselineExcelList[i].Purpose,
			})
		}
		originNetworkDeviceBaselines, err := QueryNetworkDeviceBaselineByVersionId(versionId)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return true
		}
		if len(originNetworkDeviceBaselines) > 0 {
			// TODO 该版本之前已导入数据，需删除所有数据，范围巨大。。。必须重新导入其他所有基线
		} else {
			if err := BatchCreateNetworkDeviceBaseline(networkDeviceBaselines); err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			networkDeviceBaselineMap := make(map[string]int64)
			for _, networkDeviceBaseline := range networkDeviceBaselines {
				networkDeviceBaselineMap[networkDeviceBaseline.DeviceModel] = networkDeviceBaseline.Id
			}
			networkDeviceRoleCodeMap := make(map[string]int64)
			for _, networkDeviceRoleBaseline := range networkDeviceRoleBaselines {
				networkDeviceRoleCodeMap[networkDeviceRoleBaseline.FuncCompoCode] = networkDeviceRoleBaseline.Id
			}
			var networkDeviceRoleRels []entity.NetworkDeviceRoleRel
			for _, networkDeviceBaselineExcel := range networkDeviceBaselineExcelList {
				networkDeviceId := networkDeviceBaselineMap[networkDeviceBaselineExcel.DeviceModel]
				for _, networkDeviceRoleCode := range networkDeviceBaselineExcel.NetworkDeviceRoleCodes {
					networkDeviceRoleId, ok := networkDeviceRoleCodeMap[networkDeviceRoleCode]
					if !ok {
						log.Infof("import networkDeviceBaseline fail, can not find networkDeviceRoleCode: %s", networkDeviceRoleCode)
						result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
						return true
					}
					networkDeviceRoleRels = append(networkDeviceRoleRels, entity.NetworkDeviceRoleRel{
						DeviceId:     networkDeviceId,
						DeviceRoleId: networkDeviceRoleId,
					})
				}
			}
			if len(networkDeviceRoleRels) > 0 {
				if err := BatchCreateNetworkDeviceRoleRel(networkDeviceRoleRels); err != nil {
					log.Error(err)
					result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
					return true
				}
			}
		}
	}
	return false
}

func ImportIPDemandBaseline(context *gin.Context, versionId int64, f *excelize.File) bool {
	// 先查询网络设备角色表，导入的版本是否已有数据，如没有，提示先导入网络设备角色基线
	networkDeviceRoleBaselines, err := QueryNetworkDeviceRoleBaselineByVersionId(versionId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			result.Failure(context, errorcodes.NetworkDeviceRoleMustImportFirst, http.StatusBadRequest)
		} else {
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		}
		return true
	}
	var ipDemandBaselineExcelList []IPDemandBaselineExcel
	if err := excel.ImportBySheet(f, &ipDemandBaselineExcelList, IPDemandBaselineSheetName, 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return true
	}
	if len(ipDemandBaselineExcelList) > 0 {
		var ipDemandBaselines []entity.IPDemandBaseline
		for i := range ipDemandBaselineExcelList {
			networkDeviceRole := ipDemandBaselineExcelList[i].NetworkDeviceRoleCode
			if networkDeviceRole != "" {
				ipDemandBaselineExcelList[i].NetworkDeviceRoleCodes = util.SplitString(networkDeviceRole, constant.SplitLineBreak)
			}
			ipDemandBaselines = append(ipDemandBaselines, entity.IPDemandBaseline{
				VersionId:    versionId,
				Vlan:         ipDemandBaselineExcelList[i].Vlan,
				Explain:      ipDemandBaselineExcelList[i].Explain,
				Description:  ipDemandBaselineExcelList[i].Description,
				IPSuggestion: ipDemandBaselineExcelList[i].IPSuggestion,
				AssignNum:    ipDemandBaselineExcelList[i].AssignNum,
				Remark:       ipDemandBaselineExcelList[i].Remark,
			})
		}
		originIPDemandBaselines, err := QueryIPDemandBaselineByVersionId(versionId)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return true
		}
		if len(originIPDemandBaselines) > 0 {
			// TODO 该版本之前已导入数据，需删除所有数据，范围巨大。。。必须重新导入其他所有基线
		} else {
			if err := BatchCreateIPDemandBaseline(ipDemandBaselines); err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			ipDemandBaselineMap := make(map[string]int64)
			for _, ipDemandBaseline := range ipDemandBaselines {
				ipDemandBaselineMap[ipDemandBaseline.Vlan] = ipDemandBaseline.Id
			}
			networkDeviceRoleCodeMap := make(map[string]int64)
			for _, networkDeviceRoleBaseline := range networkDeviceRoleBaselines {
				networkDeviceRoleCodeMap[networkDeviceRoleBaseline.FuncCompoCode] = networkDeviceRoleBaseline.Id
			}
			var ipDemandDeviceRoleRels []entity.IPDemandDeviceRoleRel
			for _, ipDemandBaselineExcel := range ipDemandBaselineExcelList {
				ipDemandId := ipDemandBaselineMap[ipDemandBaselineExcel.Vlan]
				for _, networkDeviceRoleCode := range ipDemandBaselineExcel.NetworkDeviceRoleCodes {
					deviceRoleId, ok := networkDeviceRoleCodeMap[networkDeviceRoleCode]
					if !ok {
						log.Infof("import IPDemandBaseline fail, can not find networkDeviceRoleCode: %s", networkDeviceRoleCode)
						result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
						return true
					}
					ipDemandDeviceRoleRels = append(ipDemandDeviceRoleRels, entity.IPDemandDeviceRoleRel{
						IPDemandId:   ipDemandId,
						DeviceRoleId: deviceRoleId,
					})
				}
			}
			if len(ipDemandDeviceRoleRels) > 0 {
				if err := BatchCreateIPDemandDeviceRoleRel(ipDemandDeviceRoleRels); err != nil {
					log.Error(err)
					result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
					return true
				}
			}
		}
	}
	return false
}

func ImportCapConvertBaseline(context *gin.Context, versionId int64, f *excelize.File) bool {
	var capConvertBaselineExcelList []CapConvertBaselineExcel
	if err := excel.ImportBySheet(f, &capConvertBaselineExcelList, CapConvertBaselineSheetName, 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return true
	}
	if len(capConvertBaselineExcelList) > 0 {
		var capConvertBaselines []entity.CapConvertBaseline
		for _, capConvertBaselineExcel := range capConvertBaselineExcelList {
			capConvertBaselines = append(capConvertBaselines, entity.CapConvertBaseline{
				VersionId:        versionId,
				ProductName:      capConvertBaselineExcel.ProductName,
				ProductCode:      capConvertBaselineExcel.ProductCode,
				SellSpecs:        capConvertBaselineExcel.SellSpecs,
				CapPlanningInput: capConvertBaselineExcel.CapPlanningInput,
				Unit:             capConvertBaselineExcel.Unit,
				Features:         capConvertBaselineExcel.Features,
				Description:      capConvertBaselineExcel.Description,
			})
		}
		originCapConvertBaselines, err := QueryCapConvertBaselineByVersionId(versionId)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return true
		}
		if len(originCapConvertBaselines) > 0 {
			// TODO 该版本之前已导入数据，需删除所有数据，范围巨大。。。必须重新导入其他所有基线
		} else {
			if err := BatchCreateCapConvertBaseline(capConvertBaselines); err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
		}
	}
	return false
}

func ImportCapActualResBaseline(context *gin.Context, versionId int64, f *excelize.File) bool {
	var capActualBaselineExcelList []CapActualResBaselineExcel
	if err := excel.ImportBySheet(f, &capActualBaselineExcelList, CapActualResBaselineSheetName, 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return true
	}
	if len(capActualBaselineExcelList) > 0 {
		var capActualResBaselines []entity.CapActualResBaseline
		for _, capActualResBaselineExcel := range capActualBaselineExcelList {
			occRatio := capActualResBaselineExcel.OccRatio
			var occRatioNumerator string
			var occRatioDenominator string
			if occRatio != "" {
				occRatios := strings.Split(occRatio, constant.SplitLineColon)
				if len(occRatios) != 2 {
					log.Infof("import capActualResBaseline fail, occRatio length: ", len(occRatios))
					result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
					return true
				}
				occRatioNumerator = occRatios[0]
				occRatioDenominator = occRatios[1]
			}
			capActualResBaselines = append(capActualResBaselines, entity.CapActualResBaseline{
				VersionId:           versionId,
				ProductCode:         capActualResBaselineExcel.ProductCode,
				SellSpecs:           capActualResBaselineExcel.SellSpecs,
				SellUnit:            capActualResBaselineExcel.SellUnit,
				ExpendRes:           capActualResBaselineExcel.ExpendRes,
				ExpendResCode:       capActualResBaselineExcel.ExpendResCode,
				Features:            capActualResBaselineExcel.Features,
				OccRatioNumerator:   occRatioNumerator,
				OccRatioDenominator: occRatioDenominator,
				Remarks:             capActualResBaselineExcel.Remarks,
			})
		}
		originCapActualResBaselines, err := QueryCapActualResBaselineByVersionId(versionId)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return true
		}
		if len(originCapActualResBaselines) > 0 {
			// TODO 该版本之前已导入数据，需删除所有数据，范围巨大。。。必须重新导入其他所有基线
		} else {
			if err := BatchCreateCapActualResBaseline(capActualResBaselines); err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
		}
	}
	return false
}

func ImportCapServerCalcBaseline(context *gin.Context, versionId int64, f *excelize.File) bool {
	var capServerCalcBaselineExcelList []CapServerCalcBaselineExcel
	if err := excel.ImportBySheet(f, &capServerCalcBaselineExcelList, CapServerCalcBaselineSheetName, 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return true
	}
	if len(capServerCalcBaselineExcelList) > 0 {
		var capServerCalcBaselines []entity.CapServerCalcBaseline
		for _, capServerCalcBaselineExcel := range capServerCalcBaselineExcelList {
			nodeWastageCalcTypeStr := capServerCalcBaselineExcel.NodeWastageCalcType
			var nodeWastageCalcType int
			if nodeWastageCalcTypeStr == constant.NodeWastageCalcTypeNumCn {
				nodeWastageCalcType = constant.NodeWastageCalcTypeNum
			} else if nodeWastageCalcTypeStr == constant.NodeWastageCalcTypePercentCn {
				nodeWastageCalcType = constant.NodeWastageCalcTypePercent
			} else {
				nodeWastageCalcType = 0
			}
			capServerCalcBaselines = append(capServerCalcBaselines, entity.CapServerCalcBaseline{
				VersionId:           versionId,
				ExpendRes:           capServerCalcBaselineExcel.ExpendRes,
				ExpendResCode:       capServerCalcBaselineExcel.ExpendResCode,
				ExpendNodeRoleCode:  capServerCalcBaselineExcel.ExpendNodeRoleCode,
				OccNodeRes:          capServerCalcBaselineExcel.OccNodeRes,
				OccNodeResCode:      capServerCalcBaselineExcel.OccNodeResCode,
				NodeWastage:         capServerCalcBaselineExcel.NodeWastage,
				NodeWastageCalcType: nodeWastageCalcType,
				WaterLevel:          capServerCalcBaselineExcel.WaterLevel,
			})
		}
		originCapServerCalcBaselines, err := QueryCapServerCalcBaselineByVersionId(versionId)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return true
		}
		if len(originCapServerCalcBaselines) > 0 {
			// TODO 该版本之前已导入数据，需删除所有数据，范围巨大。。。必须重新导入其他所有基线
		} else {
			if err := BatchCreateCapServerCalcBaseline(capServerCalcBaselines); err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
		}
	}
	return false
}
