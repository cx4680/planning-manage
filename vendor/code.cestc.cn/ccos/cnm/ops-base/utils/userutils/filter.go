package userutils

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"code.cestc.cn/ccos/cnm/ops-base/logging"
	"code.cestc.cn/ccos/cnm/ops-base/opserror"
	"code.cestc.cn/ccos/cnm/ops-base/opshttp"
)

const (
	AuthUserKey = "ctx-auth-user"
)

const (
	AuthTypeOmp      = iota + 1 // 运维
	AuthTypeOps                 // 运营
	AuthTypeInner               // 内部
	AuthTypeOmpOps              // 运维&运营
	AuthTypeOmpInner            // 运维&内部
	AuthTypeOpsInner            // 运营&内部
	AuthTypeAll                 // 全部
)

var (
	authError = opserror.AddError("Forbidden", "The operation is forbidden", http.StatusForbidden)
)

type authHandler = func(ctx context.Context) bool

func FilterUser(secretPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := GetUser(c.Request, secretPath)
		if err != nil {
			opshttp.WriteJson(c, nil, opserror.AddSpecialError("GetUserError", err.Error(), http.StatusBadRequest))
			return
		} else {
			c.Set(AuthUserKey, user)
			ctx := SetUserToContext(c.Request.Context(), user)
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		}
	}
}

func FilterAuth(authType int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logging.For(c, "func", "FilterAuth",
			zap.Int64("authType", authType),
		)

		// 获取对应的函数
		fun := getAuthHandler(authType)
		if fun == nil {
			log.Errorw("auth type undefined")
			opshttp.WriteJson(c, nil, authError)
			return
		}

		// 调用对应鉴权
		ok := fun(c)
		if !ok {
			log.Errorw("has no authority",
				zap.Any("user", GetUserByContext(c)),
			)
			opshttp.WriteJson(c, nil, authError)
			return
		}

		c.Next()
	}
}

func FilterAuthOmpOps(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logging.For(c, "func", "FilterAuthOmpOps",
			zap.String("action", action),
		)

		user := GetUserByContext(c)
		if user == nil {
			opshttp.WriteJson(c, nil, authError)
			return
		}
		if user.GetSystem() == SystemOps {
			_, err := AuthWithUser(c.Request, user, action, user.GetTenantId(), "", user.GetDepartmentId(), "")
			if err != nil {
				log.Errorw("AuthWithUser error", zap.Error(err))
				opshttp.WriteJson(c, nil, err)
				return
			}
		}

		c.Next()
	}
}

func getAuthHandler(authType int64) authHandler {
	authHandlerMap := map[int64]authHandler{
		AuthTypeOmp:      AuthOmp,
		AuthTypeOps:      AuthOps,
		AuthTypeInner:    AuthInner,
		AuthTypeOmpOps:   AuthOmpOps,
		AuthTypeOmpInner: AuthOmpInner,
		AuthTypeOpsInner: AuthOpsInner,
		AuthTypeAll:      AuthAll,
	}
	return authHandlerMap[authType]
}
