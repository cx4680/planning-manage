package main

import (
	"code.cestc.cn/zhangzhi/planning-manage/internal/app"
	"code.cestc.cn/zhangzhi/planning-manage/internal/app/http"
)

func main() {
	app.Run(http.Router)
}
