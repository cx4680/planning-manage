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
		// 先查询节点角色表，导入的版本是否已有数据，如没有，提示先导入节点角色基线
		nodeRoleBaselines, err := QueryNodeRoleBaselineByVersionId(softwareVersion.Id)
		if err != nil {
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return
		}
		if len(nodeRoleBaselines) == 0 {
			result.Failure(context, errorcodes.NodeRoleMustImportFirst, http.StatusBadRequest)
			return
		}
		var cloudProductBaselineExcelList []CloudProductBaselineExcel
		if err := excel.ImportBySheet(f, &cloudProductBaselineExcelList, CloudProductBaselineSheetName, 0, 1); err != nil {
			log.Error(err)
			result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
			return
		}
		if len(cloudProductBaselineExcelList) > 0 {
			for i := range cloudProductBaselineExcelList {
				controlResNodeRole := cloudProductBaselineExcelList[i].ControlResNodeRole
				if controlResNodeRole != "" {
					cloudProductBaselineExcelList[i].ControlResNodeRoles = strings.Split(controlResNodeRole, constant.SplitLineBreak)
				}
				resNodeRole := cloudProductBaselineExcelList[i].ResNodeRole
				if resNodeRole != "" {
					cloudProductBaselineExcelList[i].ResNodeRoles = strings.Split(resNodeRole, constant.SplitLineBreak)
				}
			}
		}
		break
	case ServerBaselineType:
		break
	case NetworkDeviceBaselineType:
		break
	case NodeRoleBaselineType:
		ImportNodeRoleBaseline(context, f, softwareVersion)
		break
	default:
		break
	}
	result.Success(context, nil)
}

func ImportNodeRoleBaseline(context *gin.Context, f *excelize.File, softwareVersion entity.SoftwareVersion) {
	var nodeRoleBaselineExcelList []NodeRoleBaselineExcel
	if err := excel.ImportBySheet(f, &nodeRoleBaselineExcelList, NodeRoleBaselineSheetName, 0, 1); err != nil {
		log.Error(err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
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
				Annotation:   nodeRoleBaselineExcelList[i].Annotation,
				BusinessType: nodeRoleBaselineExcelList[i].BusinessType,
			})
		}
		nodeRoleBaselines, err := QueryNodeRoleBaselineByVersionId(softwareVersion.Id)
		if err != nil {
			log.Error(err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return
		}
		if len(nodeRoleBaselines) > 0 {
			// TODO 该版本之前已导入数据，需删除所有数据，范围巨大。。。必须重新导入其他所有基线

		} else {
			if err := BatchCreateNodeRoleBaseline(nodeRoleBaselineList); err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return
			}
			nodeRoleBaselines, err = QueryNodeRoleBaselineByVersionId(softwareVersion.Id)
			if err != nil {
				log.Error(err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return
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
						return
					}
				}
			}
		}
	}
}
