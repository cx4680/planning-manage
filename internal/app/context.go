package app

import (
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/dataid"
)

func buildContext(context *gin.Context) {
	requestId := context.GetHeader(constant.XRequestID)
	if requestId == "" {
		requestId = dataid.DataID()
	}
	context.Set(constant.XRequestID, requestId)

	type pageParameter struct {
		Current int `form:"current"`
		Size    int `form:"size"`
	}
	p := pageParameter{}
	err := context.BindQuery(&p)
	if err != nil {
		log.Error(err, "page parameter parser error", "request query", context.Request.RequestURI)
		p.Current = constant.CurrentValue
		p.Size = constant.SizeValue
	}

	if p.Current == 0 {
		p.Current = constant.CurrentValue
	}

	switch {
	case p.Size > 100:
		p.Size = 100
	case p.Size <= 0:
		p.Size = constant.SizeValue
	}
	context.Set(constant.Current, p.Current)
	context.Set(constant.Size, p.Size)
}
