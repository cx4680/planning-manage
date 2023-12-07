package cloud_platform

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
	Id         int64
	UserId     string
	Name       string `form:"name"`
	Type       string `form:"type"`
	CustomerId int64  `form:"customerId"`
	SortField  string `form:"sortField"`
	Sort       string `form:"sort"`
}

func List(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list platform bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.CustomerId == 0 {
		result.Failure(c, "customerId参数为空", http.StatusBadRequest)
		return
	}
	list, err := ListCloudPlatform(request)
	if err != nil {
		log.Errorf("list platform error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
}

func Create(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("create platform bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if err := checkRequest(request, true); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	if err := CreateCloudPlatform(request); err != nil {
		log.Errorf("create platform error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func Update(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("update platform bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}

	request.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	if err := checkRequest(request, false); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	if err := UpdateCloudPlatform(request); err != nil {
		log.Errorf("update platform error: ", err)
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
		if request.CustomerId == 0 {
			return errors.New("customerId参数为空")
		}
	} else {
		if request.Id == 0 {
			return errors.New("id参数为空")
		}
	}
	return nil
}
