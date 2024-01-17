package server

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/excel"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/user"
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

func List(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list server bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	log.Info(request)
	if request.PlanId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
		return
	}
	list, err := ListServer(request)
	if err != nil {
		log.Errorf("list server error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
}

func Save(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("save server bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	if err := SaveServer(request); err != nil {
		log.Errorf("save server error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}
func NetworkTypeList(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list server network bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
		return
	}
	list, err := ListServerNetworkType(request)
	if err != nil {
		log.Errorf("list server network error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
}

func CpuTypeList(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list server cpu bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
		return
	}
	list, err := ListServerCpuType(request)
	if err != nil {
		log.Errorf("list server cpu error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
}

func CapacityList(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list server capacity bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
		return
	}
	list, err := ListServerCapacity(request)
	if err != nil {
		log.Errorf("list server capacity error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
}

func SaveCapacity(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("save server capacity bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	err := SaveServerCapacity(request)
	if err != nil {
		log.Errorf("save server capacity error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func Download(c *gin.Context) {
	planId, _ := strconv.ParseInt(c.Param("planId"), 10, 64)
	if planId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
		return
	}
	response, fileName, err := DownloadServer(planId)
	if err != nil {
		log.Errorf("download server error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = excel.NormalDownLoad(fileName, "服务器规划清单", "", false, response, c.Writer); err != nil {
		log.Errorf("导出错误：", err)
	}
	return
}

func ListServerShelvePlanning(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list server bind param error: ", err)
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
		return
	}
	list, err := getServerShelvePlanningList(request.PlanId)
	if err != nil {
		log.Errorf("ListServerShelve error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
	return
}

func DownloadServerShelveTemplate(c *gin.Context) {
	planId, _ := strconv.ParseInt(c.Param("planId"), 10, 64)
	if planId == 0 {
		result.Failure(c, "planId不能为空", http.StatusBadRequest)
		return
	}
	response, fileName, err := getServerShelveDownloadTemplate(planId)
	if err != nil {
		log.Errorf("ListNetworkShelve error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if err = excel.NormalDownLoad(fileName, "服务器上架表", "", false, response, c.Writer); err != nil {
		log.Errorf("下载错误：", err)
	}
	return
}

func UploadServerShelve(c *gin.Context) {
	planId, _ := strconv.ParseInt(c.Param("planId"), 10, 64)
	if planId == 0 {
		result.Failure(c, "planId不能为空", http.StatusBadRequest)
		return
	}
	//上传文件处理
	file, err := c.FormFile("file")
	if err != nil {
		log.Error(err)
		result.Failure(c, "文件错误", http.StatusBadRequest)
		return
	}
	filePath := fmt.Sprintf("%s/%s-%d-%d.xlsx", "exampledir", "serverShelve", time.Now().Unix(), rand.Uint32())
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
	var serverShelveDownload []ShelveDownload
	if err = excel.ImportBySheet(f, &serverShelveDownload, "服务器上架表", 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(c, "解析文件错误", http.StatusInternalServerError)
		return
	}
	userId := user.GetUserId(c)
	if err = uploadServerShelve(planId, serverShelveDownload, userId); err != nil {
		log.Errorf("ListNetworkShelve error, %v", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
	return
}

func SaveServerPlanning(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error(err)
	}
	request.UserId = user.GetUserId(c)
	if err := saveServerPlanning(request); err != nil {
		log.Errorf("SaveNetworkShelve error, %v", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
	return
}

func SaveServerShelve(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error(err)
	}
	request.UserId = user.GetUserId(c)
	if err := saveServerShelve(request); err != nil {
		log.Errorf("SaveNetworkShelve error, %v", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
	return
}

func DownloadServerShelve(c *gin.Context) {
	planId, _ := strconv.ParseInt(c.Param("planId"), 10, 64)
	if planId == 0 {
		result.Failure(c, "planId不能为空", http.StatusBadRequest)
		return
	}
	response, fileName, err := getServerShelveDownload(planId)
	if err != nil {
		log.Errorf("ListNetworkShelve error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if err = excel.NormalDownLoad(fileName, "服务器上架表", "", false, response, c.Writer); err != nil {
		log.Errorf("下载错误：", err)
	}
	return
}
