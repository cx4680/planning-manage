package commonutils

import (
	"context"
)

const (
	CtxAuthDataKey  = "Ctx-X-CC-AuthData"
	CtxUserDataKey  = "Ctx-X-CC-UserData"
	CtxInnerDataKey = "Ctx-X-CC-InnerData"
)

func SetAuthData(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, CtxAuthDataKey, val)
}

func SetUserData(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, CtxUserDataKey, val)
}

func SetInnerData(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, CtxInnerDataKey, val)
}

func GetAuthData(ctx context.Context) string {
	data, ok := ctx.Value(CtxAuthDataKey).(string)
	if ok {
		return data
	}
	return ""
}

func GetUserData(ctx context.Context) string {
	data, ok := ctx.Value(CtxUserDataKey).(string)
	if ok {
		return data
	}
	return ""
}

func GetInnerData(ctx context.Context) string {
	data, ok := ctx.Value(CtxInnerDataKey).(string)
	if ok {
		return data
	}
	return ""
}
