package opshttp

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"code.cestc.cn/ccos/cnm/ops-base/opserror"
	"code.cestc.cn/ccos/cnm/ops-base/trace"
)

type WrapResp struct {
	Code      string      `json:"code"`
	Msg       string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestId string      `json:"requestId"`
}

func newWrapResp(data interface{}, err error, traceId string) (WrapResp, int) {
	var e = opserror.Cause(err)
	return WrapResp{
		Code:      e.Code(),
		Msg:       e.Message(),
		Data:      data,
		RequestId: traceId,
	}, e.HTTPCode()
}

func WriteJson(c *gin.Context, data interface{}, err error) {
	w, httpCode := newWrapResp(data, err, trace.ExtraTraceID(c))
	if httpCode <= 0 {
		if strings.ToLower(w.Code) == "success" {
			httpCode = http.StatusOK
		} else {
			httpCode = http.StatusInternalServerError
		}
	}
	c.JSON(httpCode, w)
	c.Abort()
}

func WriteParamsError(c *gin.Context, err error, data interface{}) {
	var isParamsError bool
	var msg string
	obj := reflect.TypeOf(data)
	if validErr, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validErr {
			if f, exist := obj.Elem().FieldByName(e.Field()); exist && len(f.Tag.Get("msg")) > 0 {
				isParamsError = true
				msg = fmt.Sprintf("参数错误，%s", f.Tag.Get("msg"))
			}
		}
	}
	if isParamsError {
		WriteJson(c, nil, opserror.AddSpecialError(opserror.ParamsErrorCode, msg, http.StatusBadRequest))
	} else {
		WriteJson(c, nil, opserror.ParamsError)
	}
}
