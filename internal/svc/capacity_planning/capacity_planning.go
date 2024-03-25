package capacity_planning

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/user"
)

func List(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Error("list server_planning capacity bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
		return
	}
	list, err := ListServerCapacity(request)
	if err != nil {
		log.Error("list server_planning capacity error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
}

func Computing(c *gin.Context) {
	request := &RequestServerCapacityCount{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Error("CapacityCount bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	data, err := SingleComputing(request)
	if err != nil {
		log.Error("list server_planning capacity error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, data)
}

func Save(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error("save server_planning capacity bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if err := checkRequest(request); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	err := SaveServerCapacity(request)
	if err != nil {
		log.Error("save server_planning capacity error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func checkRequest(request *Request) error {
	if request.PlanId == 0 {
		return errors.New("planId参数为空")
	}
	if len(request.ServerList) == 0 {
		return errors.New("服务器规划为空")
	}
	for _, requestServer := range request.ServerList {
		if requestServer.ResourcePoolId == 0 || requestServer.ServerBaselineId == 0 {
			return errors.New("必传参数为空")
		}
	}
	return nil
}
