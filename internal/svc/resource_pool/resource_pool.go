package resource_pool

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/user"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
)

func Update(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("update resource pool bind param error: %v", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	request.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	if err := checkRequest(request); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	if err := UpdateResourcePool(request); err != nil {
		log.Errorf("update resource pool error: %v", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func Create(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("create resource pool bind param error: %v", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	if err := CreateResourcePool(request); err != nil {
		log.Errorf("create resource pool error: %v", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func Delete(c *gin.Context) {
	request := &Request{}
	request.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	if err := DeleteResourcePool(request); err != nil {
		log.Errorf("delete resource pool error: %v", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func checkRequest(request *Request) error {
	if util.IsBlank(request.ResourcePoolName) {
		return errors.New("resourcePoolName参数为空")
	}
	return nil
}
