package machine_room

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
)

func PageCabinets(context *gin.Context) {
	request := &PageRequest{Current: 1, PageSize: 10}
	if err := context.ShouldBindQuery(&request); err != nil {
		log.Errorf("page cabinets bind param error: ", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.PlanId == 0 {
		result.Failure(context, "方案id参数为空", http.StatusBadRequest)
		return
	}
	list, count, err := QueryCabinetsPage(request)
	if err != nil {
		log.Errorf("page cabinets error: ", err)
		result.Failure(context, err.Error(), http.StatusInternalServerError)
		return
	}
	result.SuccessPage(context, count, list)
}
