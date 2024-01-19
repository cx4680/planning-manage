package ip_demand

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/excel"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"github.com/xuri/excelize/v2"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

func IpDemandListDownload(context *gin.Context) {
	param := context.Param("planId")
	planId, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		log.Errorf("[IpDemandListDownload] invalid param error, %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	fileName, exportResponseDataList, err := exportIpDemandPlanningByPlanId(planId)
	if err != nil {
		log.Errorf("[exportIpDemandPlanningByPlanId] error, %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	_ = excel.NormalDownLoad(fileName, "IP需求清单", "", false, exportResponseDataList, context.Writer)
	return
}

func List(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("getIpDemandList bind param error: ", err)
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
		return
	}
	list, err := getIpDemandPlanningList(request.PlanId)
	if err != nil {
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
	return
}

func UploadIpDemand(c *gin.Context) {
	planId, err := strconv.ParseInt(c.Param("planId"), 10, 64)
	if planId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
		return
	}
	// 上传文件处理
	file, err := c.FormFile("file")
	if err != nil {
		log.Error(err)
		result.Failure(c, "文件错误", http.StatusBadRequest)
		return
	}
	filePath := fmt.Sprintf("%s/%s-%d-%d.xlsx", "exampledir", "ipDemand", time.Now().Unix(), rand.Uint32())
	if err = c.SaveUploadedFile(file, filePath); err != nil {
		log.Error(err)
		result.Failure(c, "保存文件错误", http.StatusInternalServerError)
		return
	}
	f, err := excelize.OpenFile(filePath)
	defer func() {
		if err = f.Close(); err != nil {
			log.Errorf("excelize close error: %v", err)
		}
		if err = os.Remove(filePath); err != nil {
			log.Errorf("os removeFile error: %v", err)
		}
	}()
	if err != nil {
		log.Error(err)
		result.Failure(c, "打开文件错误", http.StatusInternalServerError)
		return
	}
	var ipDemandPlanningExportResponse []IpDemandPlanningExportResponse
	if err = excel.ImportBySheet(f, &ipDemandPlanningExportResponse, "IP需求清单", 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(c, "解析文件错误", http.StatusInternalServerError)
		return
	}
	if err = uploadIpDemand(planId, ipDemandPlanningExportResponse); err != nil {
		log.Errorf("UploadNetworkShelve error, %v", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
	return
}

func SaveIpDemand(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error(err)
	}
	if err := saveIpDemand(request); err != nil {
		log.Errorf("SaveNetworkShelve error, %v", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
	return
}
