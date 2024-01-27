package result

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
)

const (
	requestId   = "requestId"
	codeKey     = "code"
	dataKey     = "data"
	successCode = "Success"
	msg         = "errorMsg"
)

type ResData struct {
	ResourceId string `json:"resourceId"` // 资源ID
	Code       string `json:"code"`       //
	Message    string `json:"message"`    //
}

// SuccessPage response
//
//	@param context
//	@param count data count
//	@param data
func SuccessPage(context *gin.Context, count int64, data interface{}) {
	type pageData struct {
		Size      int64       `json:"size"`
		Current   int         `json:"current"`
		Total     int64       `json:"total"`
		TotalPage int64       `json:"totalPage"`
		List      interface{} `json:"list"`
	}

	size := int64(context.GetInt(constant.Size))
	if size == 0 {
		size = constant.SizeValue
	}

	if IsNil(data) {
		data = []interface{}{}
		count = 0
	}

	Success(context, pageData{
		Size:      size,
		Current:   context.GetInt(constant.Current),
		Total:     count,
		TotalPage: (count + size - 1) / size,
		List:      data,
	})
}

// Success response
//
//	@param context
//	@param data
func Success(context *gin.Context, data interface{}) {
	output := map[string]interface{}{
		requestId: context.GetString(constant.XRequestID),
		codeKey:   successCode,
	}
	if !IsNil(data) {
		output[dataKey] = data
	}
	context.JSON(http.StatusOK, output)
}

// SuccessCode response
//
//	@param context
//	@param code
//	@param data
func SuccessCode(context *gin.Context, code string) {
	context.JSON(http.StatusOK, map[string]interface{}{
		requestId: context.GetString(constant.XRequestID),
		codeKey:   code,
	})
}

// InternalServerFailure response
//
//	@param context
//	@param errorCode
//	@param httpStatusCode
func InternalServerFailure(context *gin.Context, errorCode string) {
	context.JSON(http.StatusInternalServerError, map[string]interface{}{
		requestId: context.GetString(constant.XRequestID),
		codeKey:   errorCode,
	})
}

// Failure response
//
//	@param context
//	@param errorCode
//	@param httpStatusCode
func Failure(context *gin.Context, errorCode string, httpStatusCode int) {
	context.JSON(httpStatusCode, map[string]interface{}{
		requestId: context.GetString(constant.XRequestID),
		codeKey:   errorCode,
	})
}

func FailureWithMsg(context *gin.Context, errorCode string, httpStatusCode int, errorMsg string) {
	context.JSON(httpStatusCode, map[string]interface{}{
		requestId: context.GetString(constant.XRequestID),
		codeKey:   errorCode,
		msg:       errorMsg,
	})
}

func FailureWithData(context *gin.Context, errorCode string, httpStatusCode int, data interface{}) {
	output := map[string]interface{}{
		requestId: context.GetString(constant.XRequestID),
		codeKey:   errorCode,
	}
	if !IsNil(data) {
		output[dataKey] = data
	}
	context.JSON(httpStatusCode, output)
}
