package server

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/user"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"net/http"
	"strconv"
)

type Request struct {
	Id             int64
	UserId         string
	PlanId         string    `form:"planId"`
	NetworkVersion string    `form:"networkVersion"`
	CpuType        int64     `form:"cpuType"`
	serverList     []*server `form:"serverList"`
}

type server struct {
	Region   string `form:"region"`
	Role     string `form:"role"`
	ServerId int64  `form:"serverId"`
	Number   string `form:"number"`
	Master   string `form:"master"`
}

func List(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list server bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if util.IsBlank(request.PlanId) {
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

func Create(c *gin.Context) {
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
	if err := CreateServer(request); err != nil {
		log.Errorf("create server error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func Update(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("update server bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}

	request.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	if err := checkRequest(request, false); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	if err := UpdateServer(request); err != nil {
		log.Errorf("update server error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
}

func checkRequest(request *Request, isCreate bool) error {
	if util.IsBlank(request.PlanId) {
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
