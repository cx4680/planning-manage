package config_item

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"net/http"
)

func List(c *gin.Context) {
	code := c.Param("code")
	if util.IsBlank(code) {
		log.Error("list config item bind param error")
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	list, err := ListConfigItem(code)
	if err != nil {
		log.Errorf("list config item error: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
}
