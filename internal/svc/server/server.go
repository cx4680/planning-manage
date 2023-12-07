package server

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/user"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"net/http"
)

func List(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list server bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
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
		log.Errorf("create server bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if err := checkRequest(request, true); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	if err := SaveServer(request); err != nil {
		log.Errorf("create server error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func CpuTypeList(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list server arch bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
	}
	list, err := ListServerCpuType(request)
	if err != nil {
		log.Errorf("list server arch error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
}

func CapacityList(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list server arch bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
	}
	list, err := ListServerCapacity(request)
	if err != nil {
		log.Errorf("list server arch error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
}

//func ModelList(c *gin.Context) {
//	request := &Request{}
//	if err := c.ShouldBindQuery(&request); err != nil {
//		log.Errorf("list server arch bind param error: ", err)
//		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
//		return
//	}
//	if request.PlanId == 0 {
//		result.Failure(c, "planId参数为空", http.StatusBadRequest)
//	}
//	list, err := ListServerModel(request)
//	if err != nil {
//		log.Errorf("list server arch error: ", err)
//		result.Failure(c, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	result.Success(c, list)
//}

func checkRequest(request *Request, isCreate bool) error {
	if request.PlanId == 0 {
		return errors.New("planId参数为空")
	}
	if isCreate {
	} else {
		if request.Id == 0 {
			return errors.New("id参数为空")
		}
	}
	return nil
}
