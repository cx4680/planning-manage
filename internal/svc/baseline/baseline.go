package baseline

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"github.com/xuri/excelize/v2"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
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
	var importBaselineRequest ImportBaselineRequest
	importBaselineRequest.CloudPlatformType = context.PostForm("cloudPlatformType")
	importBaselineRequest.BaselineVersion = context.PostForm("baselineVersion")
	importBaselineRequest.BaselineType = context.PostForm("baselineType")
	importBaselineRequest.ReleaseTime = context.PostForm("releaseTime")
	if importBaselineRequest.CloudPlatformType == "" || importBaselineRequest.BaselineVersion == "" || importBaselineRequest.BaselineType == "" {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
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
	switch importBaselineRequest.BaselineType {
	case CloudProductBaselineType:
		// 先查询节点角色表，导入的版本是否已有数据，如没有，提示先导入节点角色基线
		var cloudProductBaselineExcelList []CloudProductBaselineExcel
		if err := excel.ImportBySheet(f, &cloudProductBaselineExcelList, CloudProductBaselineSheetName, 0, 1); err != nil {
			log.Error(err)
			result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
			return
		}
		if len(cloudProductBaselineExcelList) > 0 {
			for _, cloudProductBaselineExcel := range cloudProductBaselineExcelList {
				controlResNodeRole := cloudProductBaselineExcel.ControlResNodeRole
				if controlResNodeRole != "" {
					cloudProductBaselineExcel.ControlResNodeRoles = strings.Split(controlResNodeRole, "\n")
				}
				resNodeRole := cloudProductBaselineExcel.ResNodeRole
				if resNodeRole != "" {
					cloudProductBaselineExcel.ResNodeRoles = strings.Split(resNodeRole, "\n")
				}
			}
		}
		break
	case ServerBaselineType:
		break
	case NetworkDeviceBaselineType:
		break
	case NodeRoleBaselineType:
		var nodeRoleBaselineExcelList []NodeRoleBaselineExcel
		if err := excel.ImportBySheet(f, &nodeRoleBaselineExcelList, NodeRoleBaselineSheetName, 0, 1); err != nil {
			log.Error(err)
			result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
			return
		}
		break
	default:
		break
	}
	result.Success(context, nil)
}
