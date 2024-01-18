package ip_demand

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/excel"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"net/http"
	"strconv"
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
	_, list, err := exportIpDemandPlanningByPlanId(request.PlanId)
	if err != nil {
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
	return
}
