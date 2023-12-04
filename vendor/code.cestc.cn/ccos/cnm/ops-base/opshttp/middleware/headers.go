package middleware

import (
	"code.cestc.cn/ccos/cnm/ops-base/utils/idutils"

	"github.com/gin-gonic/gin"

	"code.cestc.cn/ccos/cnm/ops-base/trace"
	"code.cestc.cn/ccos/cnm/ops-base/utils/commonutils"
	"code.cestc.cn/ccos/cnm/ops-base/utils/userutils"
)

func SetHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(commonutils.CtxAuthDataKey, c.Request.Header.Get(userutils.GatewayOldInfoHeaderKey))
		c.Set(commonutils.CtxUserDataKey, c.Request.Header.Get(userutils.GatewayNewInfoHeaderKey))
		c.Set(commonutils.CtxInnerDataKey, c.Request.Header.Get(userutils.GatewayInnerHeaderKey))

		ctx := commonutils.SetAuthData(c.Request.Context(), c.Request.Header.Get(userutils.GatewayOldInfoHeaderKey))
		ctx = commonutils.SetUserData(ctx, c.Request.Header.Get(userutils.GatewayNewInfoHeaderKey))
		ctx = commonutils.SetInnerData(ctx, c.Request.Header.Get(userutils.GatewayInnerHeaderKey))

		requestId := c.Request.Header.Get(trace.HeaderRequestIdKey)
		if len(requestId) <= 0 {
			requestId = idutils.GetUUID()
		}

		c.Set(trace.CtxRequestKey, requestId)
		c.Writer.Header().Set(trace.HeaderRequestIdKey, requestId)

		ctx = trace.SetRequestId(ctx, requestId)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
