package logging

import (
	"golang.org/x/net/context"
)

type contextFunc func(ctx context.Context) (string, string)

var contextList []contextFunc

func RegisterCtx(cb contextFunc) {
	contextList = append(contextList, cb)
}
