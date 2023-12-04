package app

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"

	"code.cestc.cn/zhangzhi/planning-manage/internal/api/constant"
)

func Test_buildContext(t *testing.T) {
	type args struct {
		context *gin.Context
	}
	context := &gin.Context{
		Request: &http.Request{RequestURI: "/avc?Current=1&Size=10", URL: &url.URL{
			Scheme:      "http",
			Opaque:      "",
			User:        nil,
			Host:        "localhost",
			Path:        "/acc",
			RawPath:     "",
			ForceQuery:  false,
			RawQuery:    "",
			Fragment:    "",
			RawFragment: "",
		}},
	}
	context.Keys = map[string]interface{}{}
	tests := []struct {
		name string
		args args
	}{
		{
			args: args{context: context},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildContext(tt.args.context)
			value, exists := tt.args.context.Get(constant.Current)
			if exists && value != 1 {
				t.Errorf("not eq %v", value)
			}
		})
	}
}
