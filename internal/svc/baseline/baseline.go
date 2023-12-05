package baseline

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"code.cestc.cn/ccos/cnm/ops-base/tools/stringx"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"github.com/xuri/excelize/v2"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/excel"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
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
	if err != nil {
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
			ReleaseTime:       datetime.StrToTime(releaseTime, datetime.FullTimeFmt),
			CreateTime:        now,
		}
		// 新增软件版本
		if err := CreateSoftwareVersion(softwareVersion); err != nil {
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
		log.Error(err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Error(err)
		}
		if err := os.Remove(filePath); err != nil {
			log.Error(err)
		}
	}()
	switch baselineType {
	case CloudProductBaselineType:
		if ImportCloudProductBaseline(context, softwareVersion, f) {
			return
		}
		break
	case ServerBaselineType:
		if ImportServerBaseline(context, softwareVersion, f) {
			return
		}
		break
	case NetworkDeviceBaselineType:
		break
	case NodeRoleBaselineType:
		if ImportNodeRoleBaseline(context, softwareVersion, f) {
			return
		}
		break
	default:
		break
	}
	result.Success(context, nil)
}

func ImportCloudProductBaseline(context *gin.Context, softwareVersion entity.SoftwareVersion, f *excelize.File) bool {
	// 先查询节点角色表，导入的版本是否已有数据，如没有，提示先导入节点角色基线
	nodeRoleBaselines, err := QueryNodeRoleBaselineByVersionId(softwareVersion.Id)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return true
	}
	if len(nodeRoleBaselines) == 0 {
		result.Failure(context, errorcodes.NodeRoleMustImportFirst, http.StatusBadRequest)
		return true
	}
	var cloudProductBaselineExcelList []CloudProductBaselineExcel
	if err := excel.ImportBySheet(f, &cloudProductBaselineExcelList, CloudProductBaselineSheetName, 0, 1); err != nil {
		log.Error(err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return true
	}
	if len(cloudProductBaselineExcelList) > 0 {
		var cloudProductBaselines []entity.CloudProductBaseline
		for i := range cloudProductBaselineExcelList {
			dependProductCode := cloudProductBaselineExcelList[i].DependProductCode
			if dependProductCode != "" {
				cloudProductBaselineExcelList[i].DependProductCodes = strings.Split(dependProductCode, constant.SplitLineBreak)
			}
			controlResNodeRole := cloudProductBaselineExcelList[i].ControlResNodeRole
			if controlResNodeRole != "" {
				cloudProductBaselineExcelList[i].ControlResNodeRoles = strings.Split(controlResNodeRole, constant.SplitLineBreak)
			}
			resNodeRole := cloudProductBaselineExcelList[i].ResNodeRole
			if resNodeRole != "" {
				cloudProductBaselineExcelList[i].ResNodeRoles = strings.Split(resNodeRole, constant.SplitLineBreak)
			}
			whetherRequired := cloudProductBaselineExcelList[i].WhetherRequired
			var whetherRequiredType int
			if whetherRequired == constant.WhetherRequiredNoChinese {
				whetherRequiredType = constant.WhetherRequiredNo
			} else if whetherRequired == constant.WhetherRequiredYesChinese {
				whetherRequiredType = constant.WhetherRequiredYes
			} else {
				result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
				return true
			}
			cloudProductBaselines = append(cloudProductBaselines, entity.CloudProductBaseline{
				VersionId:       softwareVersion.Id,
				ProductType:     cloudProductBaselineExcelList[i].ProductType,
				ProductName:     cloudProductBaselineExcelList[i].ProductName,
				ProductCode:     cloudProductBaselineExcelList[i].ProductCode,
				SellSpec:        cloudProductBaselineExcelList[i].SellSpecs,
				AuthorizedUnit:  cloudProductBaselineExcelList[i].AuthorizedUnit,
				WhetherRequired: whetherRequiredType,
				Instructions:    cloudProductBaselineExcelList[i].Instructions,
			})
		}
		cloudProductBaselines, err := QueryCloudProductBaselineByVersionId(softwareVersion.Id)
		if err != nil {
			log.Error(err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return true
		}
		if len(cloudProductBaselines) > 0 {
			// TODO 先查询云产品基线表，看看相同的版本号是否已存在数据，如果已存在，需要先删除已有数据
		} else {
			if err := BatchCreateCloudProductBaseline(cloudProductBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			cloudProductBaselines, err = QueryCloudProductBaselineByVersionId(softwareVersion.Id)
			if err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			cloudProductCodeMap := make(map[string]int64)
			for _, cloudProductBaseline := range cloudProductBaselines {
				cloudProductCodeMap[cloudProductBaseline.ProductCode] = cloudProductBaseline.Id
			}
			nodeRoleBaselines, err := QueryNodeRoleBaselineByVersionId(softwareVersion.Id)
			if err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			nodeRoleNameMap := make(map[string]int64)
			for _, nodeRoleBaseline := range nodeRoleBaselines {
				nodeRoleNameMap[nodeRoleBaseline.NodeRoleName] = nodeRoleBaseline.Id
			}
			for _, cloudProductBaselineExcel := range cloudProductBaselineExcelList {
				// 处理依赖服务编码
				dependProductCodes := cloudProductBaselineExcel.DependProductCodes
				if len(dependProductCodes) > 0 {
					var cloudProductDependRels []entity.CloudProductDependRel
					for _, dependProductCode := range dependProductCodes {
						cloudProductDependRels = append(cloudProductDependRels, entity.CloudProductDependRel{
							ProductId:       cloudProductCodeMap[cloudProductBaselineExcel.ProductCode],
							DependProductId: cloudProductCodeMap[dependProductCode],
						})
					}
					if err := BatchCreateCloudProductDependRel(cloudProductDependRels); err != nil {
						result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
						return true
					}
				}
				// 处理管控资源节点角色和资源节点角色
				controlResNodeRoles := cloudProductBaselineExcel.ControlResNodeRoles
				var cloudProductNodeRoleRels []entity.CloudProductNodeRoleRel
				if len(controlResNodeRoles) > 0 {
					for _, controlResNodeRole := range controlResNodeRoles {
						cloudProductNodeRoleRels = append(cloudProductNodeRoleRels, entity.CloudProductNodeRoleRel{
							ProductId:    cloudProductCodeMap[cloudProductBaselineExcel.ProductCode],
							NodeRoleId:   nodeRoleNameMap[controlResNodeRole],
							NodeRoleType: constant.ControlNodeRoleType,
						})
					}
				}
				resNodeRoles := cloudProductBaselineExcel.ResNodeRoles
				if len(resNodeRoles) > 0 {
					for _, resNodeRole := range resNodeRoles {
						cloudProductNodeRoleRels = append(cloudProductNodeRoleRels, entity.CloudProductNodeRoleRel{
							ProductId:    cloudProductCodeMap[cloudProductBaselineExcel.ProductCode],
							NodeRoleId:   nodeRoleNameMap[resNodeRole],
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

func ImportNodeRoleBaseline(context *gin.Context, softwareVersion entity.SoftwareVersion, f *excelize.File) bool {
	var nodeRoleBaselineExcelList []NodeRoleBaselineExcel
	if err := excel.ImportBySheet(f, &nodeRoleBaselineExcelList, NodeRoleBaselineSheetName, 0, 1); err != nil {
		log.Error(err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return true
	}
	if len(nodeRoleBaselineExcelList) > 0 {
		var nodeRoleBaselineList []entity.NodeRoleBaseline
		for i := range nodeRoleBaselineExcelList {
			mixedDeploy := nodeRoleBaselineExcelList[i].MixedDeploy
			if nodeRoleBaselineExcelList[i].MixedDeploy != "" {
				nodeRoleBaselineExcelList[i].MixedDeploys = strings.Split(mixedDeploy, constant.SplitLineBreak)
			}
			nodeRoleBaselineList = append(nodeRoleBaselineList, entity.NodeRoleBaseline{
				VersionId:    softwareVersion.Id,
				NodeRoleCode: nodeRoleBaselineExcelList[i].NodeRoleCode,
				NodeRoleName: nodeRoleBaselineExcelList[i].NodeRoleName,
				MinimumNum:   nodeRoleBaselineExcelList[i].MinimumCount,
				DeployMethod: nodeRoleBaselineExcelList[i].DeployMethod,
				Classify:     nodeRoleBaselineExcelList[i].Classify,
				Annotation:   nodeRoleBaselineExcelList[i].Annotation,
				BusinessType: nodeRoleBaselineExcelList[i].BusinessType,
			})
		}
		nodeRoleBaselines, err := QueryNodeRoleBaselineByVersionId(softwareVersion.Id)
		if err != nil {
			log.Error(err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return true
		}
		if len(nodeRoleBaselines) > 0 {
			// TODO 该版本之前已导入数据，需删除所有数据，范围巨大。。。必须重新导入其他所有基线

		} else {
			if err := BatchCreateNodeRoleBaseline(nodeRoleBaselineList); err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			nodeRoleBaselines, err = QueryNodeRoleBaselineByVersionId(softwareVersion.Id)
			if err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			nodeRoleMap := make(map[string]int64)
			for _, nodeRoleBaseline := range nodeRoleBaselines {
				nodeRoleMap[nodeRoleBaseline.NodeRoleName] = nodeRoleBaseline.Id
			}
			for _, nodeRoleBaselineExcel := range nodeRoleBaselineExcelList {
				nodeRoleName := nodeRoleBaselineExcel.NodeRoleName
				mixedDeploys := nodeRoleBaselineExcel.MixedDeploys
				if len(mixedDeploys) > 0 {
					var mixedNodeRoles []entity.NodeRoleMixedDeploy
					for _, mixedDeploy := range mixedDeploys {
						mixedNodeRoles = append(mixedNodeRoles, entity.NodeRoleMixedDeploy{
							NodeRoleId:      nodeRoleMap[nodeRoleName],
							MixedNodeRoleId: nodeRoleMap[mixedDeploy],
						})
					}
					if err := BatchCreateNodeRoleMixedDeploy(mixedNodeRoles); err != nil {
						result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
						return true
					}
				}
			}
		}
	}
	return false
}

func ImportServerBaseline(context *gin.Context, softwareVersion entity.SoftwareVersion, f *excelize.File) bool {
	// 先查询节点角色表，导入的版本是否已有数据，如没有，提示先导入节点角色基线
	nodeRoleBaselines, err := QueryNodeRoleBaselineByVersionId(softwareVersion.Id)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return true
	}
	if len(nodeRoleBaselines) == 0 {
		result.Failure(context, errorcodes.NodeRoleMustImportFirst, http.StatusBadRequest)
		return true
	}
	var serverBaselineExcelList []ServerBaselineExcel
	if err := excel.ImportBySheet(f, &serverBaselineExcelList, ServerBaselineSheetName, 0, 1); err != nil {
		log.Error(err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return true
	}
	if len(serverBaselineExcelList) > 0 {
		var serverBaselines []entity.ServerBaseline
		for i := range serverBaselineExcelList {
			nodeRole := serverBaselineExcelList[i].NodeRole
			if nodeRole != "" {
				serverBaselineExcelList[i].NodeRoles = strings.Split(nodeRole, constant.SplitLineBreak)
			}
			serverBaselines = append(serverBaselines, entity.ServerBaseline{
				VersionId:           softwareVersion.Id,
				Arch:                serverBaselineExcelList[i].Arch,
				NetworkInterface:    serverBaselineExcelList[i].NetworkInterface,
				ServerModel:         serverBaselineExcelList[i].ServerModel,
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
		serverBaselines, err := QueryServerBaselineByVersionId(softwareVersion.Id)
		if err != nil {
			log.Error(err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return true
		}
		if len(serverBaselines) > 0 {
			// TODO 先查询服务器基线表，看看相同的版本号是否已存在数据，如果已存在，需要先删除已有数据
		} else {
			if err := BatchCreateServerBaseline(serverBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			serverBaselines, err = QueryServerBaselineByVersionId(softwareVersion.Id)
			if err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			serverModelMap := make(map[string]int64)
			for _, serverBaseline := range serverBaselines {
				serverModelMap[serverBaseline.ServerModel] = serverBaseline.Id
			}
			nodeRoleBaselines, err := QueryNodeRoleBaselineByVersionId(softwareVersion.Id)
			if err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			nodeRoleNameMap := make(map[string]int64)
			for _, nodeRoleBaseline := range nodeRoleBaselines {
				nodeRoleNameMap[nodeRoleBaseline.NodeRoleName] = nodeRoleBaseline.Id
			}
			for _, serverBaselineExcel := range serverBaselineExcelList {
				// 处理节点角色
				nodeRoles := serverBaselineExcel.NodeRoles
				if len(nodeRoles) > 0 {
					var serverNodeRoleRels []entity.ServerNodeRoleRel
					for _, nodeRole := range nodeRoles {
						serverNodeRoleRels = append(serverNodeRoleRels, entity.ServerNodeRoleRel{
							ServerId:   serverModelMap[serverBaselineExcel.ServerModel],
							NodeRoleId: nodeRoleNameMap[nodeRole],
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
