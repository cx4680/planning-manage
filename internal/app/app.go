package app

import (
	"code.cestc.cn/zhangzhi/planning-manage/internal/app/settings"
	"code.cestc.cn/zhangzhi/planning-manage/internal/data"
	"code.cestc.cn/zhangzhi/planning-manage/internal/pkg/dataid"
	"code.cestc.cn/zhangzhi/planning-manage/internal/pkg/httpcall"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"go.uber.org/zap/zapcore"
)


func Run(routerFunc GinEngineRouterFunc) {
	setting := settings.NewSetting()

	level, _ := zapcore.ParseLevel(setting.LogLevel)
	log.Init(setting.LogPath, log.Level(level))
	// init id
	dataid.InitIDWorker()
	// init http call
	httpcall.Init(setting)
	// init database
	data.InitDatabase(setting)
	// init validate
	data.IntiValidate()

	httpEngine := newEngine(setting)
	ginEngine := httpEngine.initGinEngine(routerFunc)
	httpEngine.serverRun(ginEngine)
}
