package plan

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/user"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
)

func Page(c *gin.Context) {
	request := &Request{Current: 1, PageSize: 10}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list plan bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.ProjectId == 0 && request.Id == 0 {
		result.Failure(c, "projectId参数为空", http.StatusBadRequest)
		return
	}
	list, count, err := PagePlan(request)
	if err != nil {
		log.Errorf("list plan error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.SuccessPage(c, count, list)
}

func Create(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("create plan bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if err := checkRequest(request, true); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	if err := CreatePlan(request); err != nil {
		log.Errorf("create plan error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func Update(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("update plan bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}

	request.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	if err := checkRequest(request, false); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	if err := UpdatePlan(request); err != nil {
		log.Errorf("update plan error: ", err)
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
	if err := DeletePlan(request); err != nil {
		log.Errorf("delete plan error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func checkRequest(request *Request, isCreate bool) error {
	if isCreate {
		if request.ProjectId == 0 {
			return errors.New("projectId参数为空")
		}
		if util.IsBlank(request.Name) {
			return errors.New("name参数为空")
		}
	} else {
		if request.Id == 0 {
			return errors.New("id参数为空")
		}
	}
	return nil
}

func Send(c *gin.Context) {
	Id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	data, err := SendPlan(Id)
	if err != nil {
		message := fmt.Sprintf("创建bom请求体错误：%s", err.Error())
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, message)
		return
	}
	if data.Success {
		result.Success(c, data.Data)
	} else {
		message := fmt.Sprintf("请求bom错误：%s, %s", data.Desc, data.Message)
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, message)
	}
}

func Copy(c *gin.Context) {
	request := &Request{}
	request.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	request.UserId = user.GetUserId(c)
	if err := checkRequest(request, false); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	if err := CopyPlan(request); err != nil {
		log.Errorf("copy plan error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}
