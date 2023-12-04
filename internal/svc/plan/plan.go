package plan

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
	Id        int64
	UserId    string
	Name      string `form:"name"`
	ProjectId int64  `form:"ProjectId"`
	Type      string `form:"type"`
	Stage     string `form:"stage"`
	SortField string `form:"sortField"`
	Sort      string `form:"sort"`
	Current   int    `json:"current"`
	PageSize  int    `json:"pageSize"`
}

func Page(c *gin.Context) {
	request := &Request{Current: 1, PageSize: 10}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list plan bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.ProjectId == 0 {
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
