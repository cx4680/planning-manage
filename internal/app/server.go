package app

import (
	"fmt"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GinEngineRouterFunc func(engine *gin.Engine)

func (e *engine) serverRun(engine *gin.Engine) {
	var err error

	httpServer := initServer(e.setting.Port, engine)
	if e.setting.Https == true {
		log.Info("start https server listening", "addr", httpServer.Addr)
		err = httpServer.ListenAndServeTLS("./ssl/tls.crt", "./ssl/tls.key")
	} else {
		log.Info("start http server listening", "addr", httpServer.Addr)
		err = httpServer.ListenAndServe()
	}

	if err != nil {
		log.Error(err, "start https server error")
	}
}

func initServer(port string, engine *gin.Engine) *http.Server {
	return &http.Server{
		Addr:     fmt.Sprintf(":%s", port),
		Handler:  engine,
	}
}
