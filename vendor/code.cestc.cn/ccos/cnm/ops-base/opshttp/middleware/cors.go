package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CrossDomain 跨域
func CrossDomain(corsDomains []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") //请求头部
		//判断origin是否允许跨域请求
		var isExist bool
		for _, val := range corsDomains {
			if val == origin {
				isExist = true
			}
		}
		if isExist {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token,Authorization,Token,Sign,Sec-WebSocket-Protocol")
			c.Header("Access-Control-Allow-Methods", "PUT,DELETE,POST,GET,OPTIONS")
			c.Header("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
		}
		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
