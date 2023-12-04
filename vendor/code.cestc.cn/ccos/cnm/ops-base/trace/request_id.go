package trace

import (
	"context"
	"net/http"
)

const (
	CtxRequestKey = "ctx-x-request-id"
)

var HeaderRequestIdKey = http.CanonicalHeaderKey("x-request-id")

func GetRequestId(ctx context.Context) string {
	requestId, ok := ctx.Value(CtxRequestKey).(string)
	if ok {
		return requestId
	}
	return ""
}

func SetRequestId(ctx context.Context, val string) context.Context {
	ctx = context.WithValue(ctx, CtxRequestKey, val)
	return ctx
}
