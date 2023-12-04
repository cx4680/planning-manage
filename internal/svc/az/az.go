package az

import (
	"code.cestc.cn/zhangzhi/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/zhangzhi/planning-manage/internal/pkg/result"
	"code.cestc.cn/zhangzhi/planning-manage/internal/pkg/user"
	"code.cestc.cn/zhangzhi/planning-manage/internal/pkg/util"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"net/http"
	"strconv"
)

type Request struct {
	Id              int64
	UserId          string
	Name            string `form:"name"`
	Code            string `form:"code"`
	RegionId        int64  `form:"regionId"`
	MachineRoomName string `json:"machineRoomName"`
	MachineRoomCode string `json:"machineRoomCode"`
	Province        string `json:"province"`
	City            string `json:"city"`
	Address         string `json:"address"`
	SortField       string `form:"sortField"`
	Sort            string `form:"sort"`
}

func List(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list az bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.RegionId == 0 {
		result.Failure(c, "regionId参数为空", http.StatusBadRequest)
		return
	}
	list, err := ListAz(request)
	if err != nil {
		log.Errorf("list az error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
}

func Create(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("create az bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if err := checkRequest(request, true); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	if err := CreateAz(request); err != nil {
		log.Errorf("create az error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func Update(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("update az bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	request.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	if err := checkRequest(request, false); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	if err := UpdateAz(request); err != nil {
		log.Errorf("update az error: ", err)
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
	if err := DeleteAz(request); err != nil {
		log.Errorf("delete az error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
}

func checkRequest(request *Request, isCreate bool) error {
	if request.RegionId == 0 {
		return errors.New("regionId参数为空")
	}
	if util.IsBlank(request.Name) {
		return errors.New("name参数为空")
	}
	if util.IsBlank(request.Code) {
		return errors.New("code参数为空")
	}
	if !isCreate {
		if request.Id == 0 {
			return errors.New("id参数为空")
		}
	}
	return nil
}
