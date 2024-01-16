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
		if err = UpdateSoftwareVersion(softwareVersion); err != nil {
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
		if err = CreateSoftwareVersion(&softwareVersion); err != nil {
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return
		}
	}
	filePath := fmt.Sprintf("%s/%s-%d-%d.xlsx", "exampledir", "baseline", time.Now().Unix(), rand.Uint32())
	if err = context.SaveUploadedFile(file, filePath); err != nil {
		log.Error(err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Errorf("excelize openFile error: %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		if err = f.Close(); err != nil {
			log.Errorf("excelize close error: %v", err)
		}
		if err = os.Remove(filePath); err != nil {
			log.Errorf("os removeFile error: %v", err)
		}
		return
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Errorf("excelize close error: %v", err)
		}
		if err = os.Remove(filePath); err != nil {
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
			originCloudProductMap := make(map[string]entity.CloudProductBaseline)
			var insertCloudProductBaselines []entity.CloudProductBaseline
			var updateCloudProductBaselines []entity.CloudProductBaseline
			for _, originCloudProductBaseline := range originCloudProductBaselines {
				originCloudProductMap[originCloudProductBaseline.ProductCode] = originCloudProductBaseline
			}
			if err := DeleteCloudProductDependRel(); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if err := DeleteCloudProductNodeRoleRel(); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			for _, cloudProductBaseline := range cloudProductBaselines {
				originCloudProductBaseline, ok := originCloudProductMap[cloudProductBaseline.ProductCode]
				if ok {
					cloudProductBaseline.Id = originCloudProductBaseline.Id
					updateCloudProductBaselines = append(updateCloudProductBaselines, cloudProductBaseline)
					delete(originCloudProductMap, cloudProductBaseline.ProductCode)
				} else {
					insertCloudProductBaselines = append(insertCloudProductBaselines, cloudProductBaseline)
				}
			}
			if err := BatchCreateCloudProductBaseline(insertCloudProductBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if err := UpdateCloudProductBaseline(updateCloudProductBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if len(originCloudProductMap) > 0 {
				var deleteCloudProductBaselines []entity.CloudProductBaseline
				for _, cloudProductBaseline := range originCloudProductMap {
					deleteCloudProductBaselines = append(deleteCloudProductBaselines, cloudProductBaseline)
				}
				if err := DeleteCloudProductBaseline(deleteCloudProductBaselines); err != nil {
					result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
					return true
				}
			}
			cloudProductBaselines = append(insertCloudProductBaselines, updateCloudProductBaselines...)
			return HandleCloudProductDependAndNodeRole(context, cloudProductBaselines, nodeRoleBaselines, cloudProductBaselineExcelList)
		} else {
			if err := BatchCreateCloudProductBaseline(cloudProductBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			return HandleCloudProductDependAndNodeRole(context, cloudProductBaselines, nodeRoleBaselines, cloudProductBaselineExcelList)
		}
	}
	return false
}

func HandleCloudProductDependAndNodeRole(context *gin.Context, cloudProductBaselines []entity.CloudProductBaseline, nodeRoleBaselines []entity.NodeRoleBaseline, cloudProductBaselineExcelList []CloudProductBaselineExcel) bool {
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
		if err := BatchCreateCloudProductNodeRoleRel(cloudProductNodeRoleRels); err != nil {
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return true
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
			originNodeRoleMap := make(map[string]entity.NodeRoleBaseline)
			var updateNodeRoleBaselines []entity.NodeRoleBaseline
			var insertNodeRoleBaselines []entity.NodeRoleBaseline
			for _, originNodeRoleBaseline := range originNodeRoleBaselines {
				originNodeRoleMap[originNodeRoleBaseline.NodeRoleCode] = originNodeRoleBaseline
			}
			if err := DeleteNodeRoleMixedDeploy(); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			for _, nodeRoleBaseline := range nodeRoleBaselines {
				originNodeRoleBaseline, ok := originNodeRoleMap[nodeRoleBaseline.NodeRoleCode]
				if ok {
					nodeRoleBaseline.Id = originNodeRoleBaseline.Id
					updateNodeRoleBaselines = append(updateNodeRoleBaselines, nodeRoleBaseline)
					delete(originNodeRoleMap, nodeRoleBaseline.NodeRoleCode)
				} else {
					insertNodeRoleBaselines = append(insertNodeRoleBaselines, nodeRoleBaseline)
				}
			}
			if err := BatchCreateNodeRoleBaseline(insertNodeRoleBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if err := UpdateNodeRoleBaseline(updateNodeRoleBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if len(originNodeRoleMap) > 0 {
				// 删除数据
				var deleteNodeRoleBaselines []entity.NodeRoleBaseline
				for _, nodeRoleBaseline := range originNodeRoleMap {
					deleteNodeRoleBaselines = append(deleteNodeRoleBaselines, nodeRoleBaseline)
				}
				if err := DeleteNodeRoleBaseline(deleteNodeRoleBaselines); err != nil {
					result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
					return true
				}
			}
			nodeRoleBaselines = append(insertNodeRoleBaselines, updateNodeRoleBaselines...)
			return HandleNodeRoleMixedDeploy(context, nodeRoleBaselines, nodeRoleBaselineExcelList)
		} else {
			if err := BatchCreateNodeRoleBaseline(nodeRoleBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			return HandleNodeRoleMixedDeploy(context, nodeRoleBaselines, nodeRoleBaselineExcelList)
		}
	}
	return false
}

func HandleNodeRoleMixedDeploy(context *gin.Context, nodeRoleBaselines []entity.NodeRoleBaseline, nodeRoleBaselineExcelList []NodeRoleBaselineExcel) bool {
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
			if err := BatchCreateNodeRoleMixedDeploy(mixedNodeRoles); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
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
			originServerMap := make(map[string]entity.ServerBaseline)
			var insertServerBaselines []entity.ServerBaseline
			var updateServerBaselines []entity.ServerBaseline
			if err := DeleteServerNodeRoleRel(); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			for _, originServerBaseline := range originServerBaselines {
				originServerMap[originServerBaseline.BomCode] = originServerBaseline
			}
			for _, serverBaseline := range serverBaselines {
				originServerBaseline, ok := originServerMap[serverBaseline.BomCode]
				if ok {
					serverBaseline.Id = originServerBaseline.Id
					updateServerBaselines = append(updateServerBaselines, serverBaseline)
					delete(originServerMap, serverBaseline.BomCode)
				} else {
					insertServerBaselines = append(insertServerBaselines, serverBaseline)
				}
			}
			if err := BatchCreateServerBaseline(insertServerBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if err := UpdateServerBaseline(updateServerBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if len(originServerMap) > 0 {
				var deleteServerBaselines []entity.ServerBaseline
				for _, serverBaseline := range originServerMap {
					deleteServerBaselines = append(deleteServerBaselines, serverBaseline)
				}
				if err := DeleteServerBaseline(deleteServerBaselines); err != nil {
					result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
					return true
				}
			}
			serverBaselines = append(insertServerBaselines, updateServerBaselines...)
			return HandleServerNodeRole(context, serverBaselines, nodeRoleBaselines, serverBaselineExcelList)
		} else {
			if err := BatchCreateServerBaseline(serverBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			return HandleServerNodeRole(context, serverBaselines, nodeRoleBaselines, serverBaselineExcelList)
		}
	}
	return false
}

func HandleServerNodeRole(context *gin.Context, serverBaselines []entity.ServerBaseline, nodeRoleBaselines []entity.NodeRoleBaseline, serverBaselineExcelList []ServerBaselineExcel) bool {
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
			originNetworkDeviceRoleMap := make(map[string]entity.NetworkDeviceRoleBaseline)
			var insertNetworkDeviceRoleBaselines []entity.NetworkDeviceRoleBaseline
			var updateNetworkDeviceRoleBaselines []entity.NetworkDeviceRoleBaseline
			if err := DeleteNetworkModelRoleRel(); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			for _, originNetworkDeviceRoleBaseline := range originNetworkDeviceRoleBaselines {
				originNetworkDeviceRoleMap[originNetworkDeviceRoleBaseline.FuncCompoCode] = originNetworkDeviceRoleBaseline
			}
			for _, networkDeviceRoleBaseline := range networkDeviceRoleBaselines {
				originNetworkDeviceRoleBaseline, ok := originNetworkDeviceRoleMap[networkDeviceRoleBaseline.FuncCompoCode]
				if ok {
					networkDeviceRoleBaseline.Id = originNetworkDeviceRoleBaseline.Id
					updateNetworkDeviceRoleBaselines = append(updateNetworkDeviceRoleBaselines, networkDeviceRoleBaseline)
					delete(originNetworkDeviceRoleMap, originNetworkDeviceRoleBaseline.FuncCompoCode)
				} else {
					insertNetworkDeviceRoleBaselines = append(insertNetworkDeviceRoleBaselines, networkDeviceRoleBaseline)
				}
			}
			if err := BatchCreateNetworkDeviceRoleBaseline(insertNetworkDeviceRoleBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if err := UpdateNetworkDeviceRoleBaseline(updateNetworkDeviceRoleBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if len(originNetworkDeviceRoleMap) > 0 {
				var deleteNetworkDeviceRoleBaselines []entity.NetworkDeviceRoleBaseline
				for _, networkDeviceRoleBaseline := range originNetworkDeviceRoleMap {
					deleteNetworkDeviceRoleBaselines = append(deleteNetworkDeviceRoleBaselines, networkDeviceRoleBaseline)
				}
				if err := DeleteNetworkDeviceRoleBaseline(deleteNetworkDeviceRoleBaselines); err != nil {
					result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
					return true
				}
			}
			networkDeviceRoleBaselines = append(insertNetworkDeviceRoleBaselines, updateNetworkDeviceRoleBaselines...)
			return HandleNetworkModelRole(context, nodeRoleBaselines, networkDeviceRoleBaselines, networkDeviceRoleBaselineExcelList)
		} else {
			if err := BatchCreateNetworkDeviceRoleBaseline(networkDeviceRoleBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			return HandleNetworkModelRole(context, nodeRoleBaselines, networkDeviceRoleBaselines, networkDeviceRoleBaselineExcelList)
		}
	}
	return false
}

func HandleNetworkModelRole(context *gin.Context, nodeRoleBaselines []entity.NodeRoleBaseline, networkDeviceRoleBaselines []entity.NetworkDeviceRoleBaseline, networkDeviceRoleBaselineExcelList []NetworkDeviceRoleBaselineExcel) bool {
	nodeRoleCodeMap := make(map[string]int64)
	for _, nodeRoleBaseline := range nodeRoleBaselines {
		nodeRoleCodeMap[nodeRoleBaseline.NodeRoleCode] = nodeRoleBaseline.Id
	}
	networkDeviceRoleCodeMap := make(map[string]int64)
	for _, networkDeviceRoleBaseline := range networkDeviceRoleBaselines {
		networkDeviceRoleCodeMap[networkDeviceRoleBaseline.FuncCompoCode] = networkDeviceRoleBaseline.Id
	}
	var networkModelRoleRels []entity.NetworkModelRoleRel
	var err error
	for _, networkDeviceRoleBaselineExcel := range networkDeviceRoleBaselineExcelList {
		networkDeviceRoleId := networkDeviceRoleCodeMap[networkDeviceRoleBaselineExcel.FuncCompoCode]
		twoNetworkIsos := networkDeviceRoleBaselineExcel.TwoNetworkIsos
		threeNetworkIsos := networkDeviceRoleBaselineExcel.ThreeNetworkIsos
		triplePlays := networkDeviceRoleBaselineExcel.TriplePlays
		networkModelRoleRels, err = HandleNetworkModelRoleRels(networkDeviceRoleId, twoNetworkIsos, nodeRoleCodeMap, networkDeviceRoleCodeMap, networkModelRoleRels, constant.SeparationOfTwoNetworks)
		if err != nil {
			result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
			return true
		}
		networkModelRoleRels, err = HandleNetworkModelRoleRels(networkDeviceRoleId, threeNetworkIsos, nodeRoleCodeMap, networkDeviceRoleCodeMap, networkModelRoleRels, constant.TripleNetworkSeparation)
		if err != nil {
			result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
			return true
		}
		networkModelRoleRels, err = HandleNetworkModelRoleRels(networkDeviceRoleId, triplePlays, nodeRoleCodeMap, networkDeviceRoleCodeMap, networkModelRoleRels, constant.TriplePlay)
		if err != nil {
			result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
			return true
		}
	}
	if err = BatchCreateNetworkModelRoleRel(networkModelRoleRels); err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return true
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
			originNetworkDeviceMap := make(map[string]entity.NetworkDeviceBaseline)
			var insertNetworkDeviceBaselines []entity.NetworkDeviceBaseline
			var updateNetworkDeviceBaselines []entity.NetworkDeviceBaseline
			if err := DeleteNetworkDeviceRoleRel(); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			for _, originNetworkDeviceBaseline := range originNetworkDeviceBaselines {
				originNetworkDeviceMap[originNetworkDeviceBaseline.DeviceModel] = originNetworkDeviceBaseline
			}
			for _, networkDeviceBaseline := range networkDeviceBaselines {
				originNetworkDeviceBaseline, ok := originNetworkDeviceMap[networkDeviceBaseline.DeviceModel]
				if ok {
					networkDeviceBaseline.Id = originNetworkDeviceBaseline.Id
					updateNetworkDeviceBaselines = append(updateNetworkDeviceBaselines, networkDeviceBaseline)
					delete(originNetworkDeviceMap, networkDeviceBaseline.DeviceModel)
				} else {
					insertNetworkDeviceBaselines = append(insertNetworkDeviceBaselines, networkDeviceBaseline)
				}
			}
			if err := BatchCreateNetworkDeviceBaseline(insertNetworkDeviceBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if err := UpdateNetworkDeviceBaseline(updateNetworkDeviceBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if len(originNetworkDeviceMap) > 0 {
				var deleteNetworkDeviceBaselines []entity.NetworkDeviceBaseline
				for _, networkDeviceBaseline := range originNetworkDeviceMap {
					deleteNetworkDeviceBaselines = append(deleteNetworkDeviceBaselines, networkDeviceBaseline)
				}
				if err := DeleteNetworkDeviceBaseline(deleteNetworkDeviceBaselines); err != nil {
					result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
					return true
				}
			}
			networkDeviceBaselines = append(insertNetworkDeviceBaselines, updateNetworkDeviceBaselines...)
			return HandleNetworkDeviceRoleRel(context, networkDeviceBaselines, networkDeviceRoleBaselines, networkDeviceBaselineExcelList)
		} else {
			if err := BatchCreateNetworkDeviceBaseline(networkDeviceBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			return HandleNetworkDeviceRoleRel(context, networkDeviceBaselines, networkDeviceRoleBaselines, networkDeviceBaselineExcelList)
		}
	}
	return false
}

func HandleNetworkDeviceRoleRel(context *gin.Context, networkDeviceBaselines []entity.NetworkDeviceBaseline, networkDeviceRoleBaselines []entity.NetworkDeviceRoleBaseline, networkDeviceBaselineExcelList []NetworkDeviceBaselineExcel) bool {
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
	if err := BatchCreateNetworkDeviceRoleRel(networkDeviceRoleRels); err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return true
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
			originIPDemandMap := make(map[string]entity.IPDemandBaseline)
			var insertIPDemandBaselines []entity.IPDemandBaseline
			var updateIPDemandBaselines []entity.IPDemandBaseline
			if err := DeleteIPDemandDeviceRoleRel(); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			for _, originIPDemandBaseline := range originIPDemandBaselines {
				originIPDemandMap[originIPDemandBaseline.Vlan] = originIPDemandBaseline
			}
			for _, ipDemandBaseline := range ipDemandBaselines {
				originIPDemandBaseline, ok := originIPDemandMap[ipDemandBaseline.Vlan]
				if ok {
					ipDemandBaseline.Id = originIPDemandBaseline.Id
					updateIPDemandBaselines = append(updateIPDemandBaselines, ipDemandBaseline)
					delete(originIPDemandMap, ipDemandBaseline.Vlan)
				} else {
					insertIPDemandBaselines = append(insertIPDemandBaselines, ipDemandBaseline)
				}
			}
			if err := BatchCreateIPDemandBaseline(insertIPDemandBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if err := UpdateIPDemandBaseline(updateIPDemandBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if len(originIPDemandMap) > 0 {
				var deleteIPDemandBaselines []entity.IPDemandBaseline
				for _, originIPDemandBaseline := range originIPDemandMap {
					deleteIPDemandBaselines = append(deleteIPDemandBaselines, originIPDemandBaseline)
				}
				if err := DeleteIPDemandBaseline(deleteIPDemandBaselines); err != nil {
					result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
					return true
				}
			}
		} else {
			if err := BatchCreateIPDemandBaseline(ipDemandBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			return HandleIPDemandDeviceRoleRel(context, ipDemandBaselines, networkDeviceRoleBaselines, ipDemandBaselineExcelList)
		}
	}
	return false
}

func HandleIPDemandDeviceRoleRel(context *gin.Context, ipDemandBaselines []entity.IPDemandBaseline, networkDeviceRoleBaselines []entity.NetworkDeviceRoleBaseline, ipDemandBaselineExcelList []IPDemandBaselineExcel) bool {
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
	if err := BatchCreateIPDemandDeviceRoleRel(ipDemandDeviceRoleRels); err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return true
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
			originCapConvertMap := make(map[string]entity.CapConvertBaseline)
			var insertCapConvertBaselines []entity.CapConvertBaseline
			var updateCapConvertBaselines []entity.CapConvertBaseline
			for _, originCapConvertBaseline := range originCapConvertBaselines {
				key := originCapConvertBaseline.ProductCode + originCapConvertBaseline.SellSpecs + originCapConvertBaseline.CapPlanningInput + originCapConvertBaseline.Features
				originCapConvertMap[key] = originCapConvertBaseline
			}
			for _, capConvertBaseline := range capConvertBaselines {
				key := capConvertBaseline.ProductCode + capConvertBaseline.SellSpecs + capConvertBaseline.CapPlanningInput + capConvertBaseline.Features
				originCapConvertBaseline, ok := originCapConvertMap[key]
				if ok {
					capConvertBaseline.Id = originCapConvertBaseline.Id
					updateCapConvertBaselines = append(updateCapConvertBaselines, capConvertBaseline)
					delete(originCapConvertMap, key)
				} else {
					insertCapConvertBaselines = append(insertCapConvertBaselines, capConvertBaseline)
				}
			}
			if err := BatchCreateCapConvertBaseline(insertCapConvertBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if err := UpdateCapConvertBaseline(updateCapConvertBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if len(originCapConvertMap) > 0 {
				var deleteCapConvertBaselines []entity.CapConvertBaseline
				for _, originCapConvertBaseline := range originCapConvertMap {
					deleteCapConvertBaselines = append(deleteCapConvertBaselines, originCapConvertBaseline)
				}
				if err := DeleteCapConvertBaseline(deleteCapConvertBaselines); err != nil {
					result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
					return true
				}
			}
		} else {
			if err := BatchCreateCapConvertBaseline(capConvertBaselines); err != nil {
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
			originCapActualResMap := make(map[string]entity.CapActualResBaseline)
			var insertCapActualResBaselines []entity.CapActualResBaseline
			var updateCapActualResBaselines []entity.CapActualResBaseline
			for _, originCapActualResBaseline := range originCapActualResBaselines {
				key := originCapActualResBaseline.ProductCode + originCapActualResBaseline.SellSpecs + originCapActualResBaseline.SellUnit + originCapActualResBaseline.Features
				originCapActualResMap[key] = originCapActualResBaseline
			}
			for _, capActualResBaseline := range capActualResBaselines {
				key := capActualResBaseline.ProductCode + capActualResBaseline.SellSpecs + capActualResBaseline.SellUnit + capActualResBaseline.Features
				originCapActualResBaseline, ok := originCapActualResMap[key]
				if ok {
					capActualResBaseline.Id = originCapActualResBaseline.Id
					updateCapActualResBaselines = append(updateCapActualResBaselines, capActualResBaseline)
					delete(originCapActualResMap, key)
				} else {
					insertCapActualResBaselines = append(insertCapActualResBaselines, capActualResBaseline)
				}
			}
			if err := BatchCreateCapActualResBaseline(insertCapActualResBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if err := UpdateCapActualResBaseline(updateCapActualResBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if len(originCapActualResMap) > 0 {
				var deleteCapActualResBaselines []entity.CapActualResBaseline
				for _, originCapActualResBaseline := range originCapActualResMap {
					deleteCapActualResBaselines = append(deleteCapActualResBaselines, originCapActualResBaseline)
				}
				if err := DeleteCapActualResBaseline(deleteCapActualResBaselines); err != nil {
					result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
					return true
				}
			}
		} else {
			if err := BatchCreateCapActualResBaseline(capActualResBaselines); err != nil {
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
			originCapServerCalcMap := make(map[string]entity.CapServerCalcBaseline)
			var insertCapServerCalcBaselines []entity.CapServerCalcBaseline
			var updateCapServerCalcBaselines []entity.CapServerCalcBaseline
			for _, originCapServerCalcBaseline := range originCapServerCalcBaselines {
				key := originCapServerCalcBaseline.ExpendResCode + originCapServerCalcBaseline.ExpendNodeRoleCode
				originCapServerCalcMap[key] = originCapServerCalcBaseline
			}
			for _, capServerCalcBaseline := range capServerCalcBaselines {
				key := capServerCalcBaseline.ExpendResCode + capServerCalcBaseline.ExpendNodeRoleCode
				originCapServerCalcBaseline, ok := originCapServerCalcMap[key]
				if ok {
					capServerCalcBaseline.Id = originCapServerCalcBaseline.Id
					updateCapServerCalcBaselines = append(updateCapServerCalcBaselines, capServerCalcBaseline)
					delete(originCapServerCalcMap, key)
				} else {
					insertCapServerCalcBaselines = append(insertCapServerCalcBaselines, capServerCalcBaseline)
				}
			}
			if err := BatchCreateCapServerCalcBaseline(insertCapServerCalcBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if err := UpdateCapServerCalcBaseline(updateCapServerCalcBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
			if len(originCapServerCalcMap) > 0 {
				var deleteCapServerCalcBaselines []entity.CapServerCalcBaseline
				for _, originCapServerCalcBaseline := range originCapServerCalcMap {
					deleteCapServerCalcBaselines = append(deleteCapServerCalcBaselines, originCapServerCalcBaseline)
				}
				if err := DeleteCapServerCalcBaseline(deleteCapServerCalcBaselines); err != nil {
					result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
					return true
				}
			}
		} else {
			if err := BatchCreateCapServerCalcBaseline(capServerCalcBaselines); err != nil {
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return true
			}
		}
	}
	return false
}
