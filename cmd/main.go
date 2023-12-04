package main

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/app"
	"code.cestc.cn/ccos/common/planning-manage/internal/app/http"
)

func main() {
	app.Run(http.Router)
}
