package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"code.cestc.cn/ccos/cnm/ops-base/opserror"
	"code.cestc.cn/ccos/cnm/ops-base/opshttp"
	"code.cestc.cn/ccos/cnm/ops-base/utils/userutils"
)

const (
	defaultSecretPath = "/app/secret/userCellSecret/userSecretPrivateKey"
)

func SetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := userutils.GetUser(c.Request, defaultSecretPath, func(user *userutils.User) {
			user.SetPublic(true)
		})
		if err != nil {
			opshttp.WriteJson(c, nil, opserror.AddSpecialError("GetUserError", err.Error(), http.StatusBadRequest))
			return
		} else {
			c.Set(userutils.AuthUserKey, user)
			ctx := userutils.SetUserToContext(c.Request.Context(), user)
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		}
	}
}
