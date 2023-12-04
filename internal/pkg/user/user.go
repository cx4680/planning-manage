package user

import (
	"code.cestc.cn/ccos/cnm/ops-base/utils/userutils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
)

func GetUserInfo(context *gin.Context) (owner string) {
	owner = userutils.GetUserByContext(context).GetUserCode()
	if owner == "" {
		log.Info("用户信息获取失败")
		return "--"
	}
	log.Info("userInfo ", owner)

	return owner
}

func GetUserId(context *gin.Context) (userId string) {
	return sessions.Default(context).Get("userId").(string)
}
