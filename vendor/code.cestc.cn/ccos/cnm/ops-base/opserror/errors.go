package opserror

import (
	"net/http"
)

var (
	SuccessCode     = "Success"
	ParamsErrorCode = "params error"
	ForbiddenCode   = "Forbidden"
	Success         = genError(SuccessCode, "success", http.StatusOK)
	ServerError     = genError("server error", "internal error", http.StatusInternalServerError)
	ParamsError     = genError(ParamsErrorCode, "params error", http.StatusBadRequest)
	SignError       = genError("sign error", "sign error", http.StatusBadRequest)
)

// ErrorCode 重定义错误码，以便增加新的支持
type ErrorCode struct {
	code     Code
	err      error
	httpCode int
}

func (c ErrorCode) Unwrap() error {
	return c.err
}

func (c ErrorCode) Error() string {
	return c.code.Error()
}

// Code return error code
func (c ErrorCode) Code() string {
	return c.code.Code()
}

// Message return error message
func (c ErrorCode) Message() string {
	return c.code.Message()
}

// Equal for compatible.
func (c ErrorCode) Equal(err error) bool {
	return c.code.Equal(err)
}

func (c ErrorCode) HTTPCode() int {
	return c.httpCode
}

func genError(code string, msg string, httpCode int) ErrorCode {
	return ErrorCode{
		code:     Error(code, msg),
		httpCode: httpCode,
	}
}

// AddError 添加错误码
func AddError(code string, msg string, httpCode int) ErrorCode {
	return genError(code, msg, httpCode)
}

func DMError(err error) Codes {
	if err == nil {
		return Success
	}
	switch err.(type) {
	case ErrorCode:
		return err.(ErrorCode)
	case Code:
		return err.(Code)
	}
	if c, ok := err.(Codes); ok {
		return c
	}
	return ServerError
}
