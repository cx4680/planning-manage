package app

import (
	"fmt"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/app/settings"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

type engine struct {
	setting *settings.Setting
}

func newEngine(setting *settings.Setting) *engine {
	return &engine{setting: setting}
}

func (e *engine) initGinEngine(routerFunc GinEngineRouterFunc) *gin.Engine {
	gin.SetMode(e.setting.EnvGinMode)

	router := gin.Default()

	router.Use(buildContext)

	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			// custom format
			return fmt.Sprintf("%s - [%s] \"%s %s %d %s %s \"%s\" %s\"\n",
				param.ClientIP,
				param.TimeStamp.Format(time.RFC1123),
				param.Method,
				// param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Keys[constant.XRequestID],
				param.Latency,
				param.Request.UserAgent(),
				param.ErrorMessage,
			)
		},
		Output:    e.setting.Writer,
		SkipPaths: nil,
	}))

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	router.Use(gin.CustomRecovery(func(context *gin.Context, recovered interface{}) {
		log.Error(fmt.Errorf("%v", recovered), "")
		e.setting.CustomRecoveryGinError(context)
	}))
	// 主入口文件 注册 session 中间件
	store := cookie.NewStore([]byte("secret")) // 设置 Session 密钥
	router.Use(sessions.Sessions("mysession", store))

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	if routerFunc != nil {
		routerFunc(router)
	}

	return router
}
