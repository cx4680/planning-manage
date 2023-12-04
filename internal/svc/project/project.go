package project

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
	Id                int64
	UserId            string
	Name              string `form:"name"`
	CloudPlatformId   int64  `form:"cloudPlatformId"`
	CloudPlatformType string `form:"cloudPlatformType"`
	RegionId          int64  `form:"regionId"`
	AzId              int64  `form:"azId"`
	CellId            int64  `form:"cellId"`
	CustomerId        int64  `form:"customerId"`
	Type              string `form:"type"`
	Stage             string `form:"stage"`
	SortField         string `form:"sortField"`
	Sort              string `form:"sort"`
	Current           int    `json:"current"`
	PageSize          int    `json:"pageSize"`
}

func Page(c *gin.Context) {
	request := &Request{Current: 1, PageSize: 10}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list project bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.CustomerId == 0 {
		result.Failure(c, "customerId参数为空", http.StatusBadRequest)
		return
	}
	list, count, err := PageProject(request)
	if err != nil {
		log.Errorf("list project error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.SuccessPage(c, count, list)
}

func Create(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("create project bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if err := checkRequest(request, true); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	if err := CreateProject(request); err != nil {
		log.Errorf("create project error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func Update(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("update project bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}

	request.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	if err := checkRequest(request, false); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	if err := UpdateProject(request); err != nil {
		log.Errorf("update project error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func Delete(c *gin.Context) {
	request := &Request{}
	request.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	if request.Id == 0 {
		result.Failure(c, "id参数为空", http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	if err := DeleteProject(request); err != nil {
		log.Errorf("delete project error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func checkRequest(request *Request, isCreate bool) error {
	if util.IsBlank(request.Name) {
		return errors.New("name参数为空")
	}
	if isCreate {
		if request.CloudPlatformId == 0 {
			return errors.New("cloudPlatformId参数为空")
		}
		if util.IsBlank(request.Type) {
			return errors.New("type参数为空")
		}
		if request.RegionId == 0 {
			return errors.New("regionId参数为空")
		}
		if request.AzId == 0 {
			return errors.New("azId参数为空")
		}
	} else {
		if request.Id == 0 {
			return errors.New("id参数为空")
		}
	}
	return nil
}
