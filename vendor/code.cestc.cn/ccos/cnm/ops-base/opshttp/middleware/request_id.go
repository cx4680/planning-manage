package middleware

import (
	"github.com/gin-gonic/gin"

	"code.cestc.cn/ccos/cnm/ops-base/trace"
	"code.cestc.cn/ccos/cnm/ops-base/utils/idutils"
)

func SetRequestId() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := c.Request.Header.Get(trace.HeaderRequestIdKey)
		if len(requestId) <= 0 {
			requestId = idutils.GetUUID()
		}

		c.Set(trace.CtxRequestKey, requestId)
		c.Writer.Header().Set(trace.HeaderRequestIdKey, requestId)

		ctx := trace.SetRequestId(c.Request.Context(), requestId)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
