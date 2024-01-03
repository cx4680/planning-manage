package e2e

import (
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/data"
)

const (
	api = "/api/jobm"
)

func InitDB() {
	statement := gorm.Statement{}
	config := gorm.Config{}
	db := gorm.DB{Statement: &statement, Config: &config}
	data.DB = &db
}

func Get(uri string, router *gin.Engine) *httptest.ResponseRecorder {
	// 构造get请求
	req := httptest.NewRequest("GET", uri+"?current=1&size=10", nil)
	// 初始化响应
	w := httptest.NewRecorder()
	// 调用相应的handler接口
	router.ServeHTTP(w, req)
	return w
}

func Post(uri string, router *gin.Engine, body string) *httptest.ResponseRecorder {
	// 构造post请求
	req := httptest.NewRequest("POST", uri, strings.NewReader(body))
	// 初始化响应
	w := httptest.NewRecorder()
	// 调用相应的handler接口
	router.ServeHTTP(w, req)
	return w
}

func Request(method, uri string, router *gin.Engine) *httptest.ResponseRecorder {
	// 构造get请求
	req := httptest.NewRequest(method, uri, nil)
	// 初始化响应
	w := httptest.NewRecorder()
	// 调用相应的handler接口
	router.ServeHTTP(w, req)
	return w
}
