package cell

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

func List(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list cell bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.AzId == 0 {
		result.Failure(c, "azId参数为空", http.StatusBadRequest)
		return
	}
	list, err := ListCell(request)
	if err != nil {
		log.Errorf("list cell error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
}

func Create(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("create cell bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if err := checkRequest(request, true); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	if err := CreateCell(request); err != nil {
		log.Errorf("create cell error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func Update(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("update cell bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	request.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	if err := checkRequest(request, false); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	if err := UpdateCell(request); err != nil {
		log.Errorf("update cell error: ", err)
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
	if err := DeleteCell(request); err != nil {
		log.Errorf("delete cell error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func checkRequest(request *Request, isCreate bool) error {
	if util.IsBlank(request.Name) {
		return errors.New("name参数为空")
	}
	if request.AzId == 0 {
		return errors.New("azId参数为空")
	}
	if !isCreate {
		if request.Id == 0 {
			return errors.New("id参数为空")
		}
	}
	return nil
}
