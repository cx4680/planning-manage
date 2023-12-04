package middleware

import (
	"github.com/gin-gonic/gin"
)

func GetOpts() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		LoggingAccess(), // 生成access_log
		SetTrace(),      // 设置trace
		RecoverSysMW(),  // recover
		SetUser(),       // 设置用户
		SetHeaders(),    // 设置统一header
	}
}
