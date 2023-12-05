package cloud_product

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"net/http"
	"strconv"
)

func ListVersion(context *gin.Context) {
	param := context.Query("projectId")
	projectId, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		log.Errorf("[ListVersion] invalid param error, %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	//根据项目id查询云平台类型
	versionList, err := getVersionListByProjectId(projectId)
	if err != nil {
		log.Errorf("[ListVersion] getVersionListByProjectId error", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, versionList)
	return
}

func ListCloudProductBaseline(context *gin.Context) {
	param := context.Query("versionId")
	versionId, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		log.Errorf("[ListVersion] invalid param error, %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	baselineList, err := getCloudProductBaseListByVersionId(versionId)
	if err != nil {
		log.Errorf("[ListCloudProductBaseline] getCloudProductBaseListByVersionId error", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, baselineList)
	return
}

func Save(context *gin.Context) {
	var request CloudProductPlanningRequest
	err := context.BindJSON(&request)
	if err != nil {
		log.Errorf("[Save] invalid param error, %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.PlanId == 0 || len(request.ProductList) < 1 {
		log.Errorf("[Save] invalid param error, request:%v, %v", request, err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	session := sessions.Default(context)
	currentUserId := session.Get("userId").(string)
	err = saveCloudProductPlanning(request, currentUserId)
	if err != nil {
		log.Errorf("[Save] cloudProductPlanning error, %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, nil)
	return
}

func List(context *gin.Context) {
	param := context.Param("planId")
	planId, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		log.Errorf("[List] invalid param error, %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	cloudProductPlannings, err := listCloudProductPlanningByPlanId(planId)
	if err != nil {
		log.Errorf("[List] cloudProductPlanning error, %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, cloudProductPlannings)
	return
}
