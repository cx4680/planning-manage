package middleware

import (
	"bytes"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"code.cestc.cn/ccos/cnm/ops-base/logging"
	"code.cestc.cn/ccos/cnm/ops-base/trace"
	"code.cestc.cn/ccos/cnm/ops-base/utils"
	"code.cestc.cn/ccos/cnm/ops-base/utils/timeutils"

	"github.com/gin-gonic/gin"
)

const (
	respKey = "response_data"
	reqKey  = "req_data"
)

type responseWriter struct {
	gin.ResponseWriter
	b *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.b.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w responseWriter) WriteString(s string) (int, error) {
	w.b.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func LoggingAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions ||
			c.Request.Method == http.MethodHead ||
			c.Request.RequestURI == "/favicon.ico" ||
			strings.Contains(c.Request.RequestURI, "swagger") ||
			c.Request.RequestURI == "/health" {
			c.Next()
		} else {
			// 当前时间
			nowTime := time.Now()

			// response
			writer := &responseWriter{
				c.Writer,
				bytes.NewBuffer([]byte{}),
			}
			c.Writer = writer

			// request
			var reqBody string
			if c.Request.Method == http.MethodPost ||
				c.Request.Method == http.MethodPut ||
				c.Request.Method == http.MethodPatch ||
				c.Request.Method == http.MethodDelete {

				requestBody, _ := c.GetRawData()
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
				reqBody = string(requestBody)
			}

			c.Next()

			// 获取trace_id
			traceId := trace.ExtraTraceID(c)
			// 服务名称
			hostname, _ := os.Hostname()

			replyBody := writer.b.String()

			if len(reqBody) >= 512 {
				reqBody = reqBody[0:512]
			}

			if len(replyBody) >= 512 {
				replyBody = replyBody[0:512]
			}

			logItems := []interface{}{
				"start", nowTime.Format(timeutils.TimeFormatYYYYMMDDHHmmSS),
				"cost", math.Ceil(float64(time.Since(nowTime).Nanoseconds()) / 1e6),
				"trace_id", traceId,
				"host_ip", utils.GetHost(),
				"host_name", hostname,
				"req_method", c.Request.Method,
				"req_uri", c.Request.RequestURI,
				"real_ip", c.ClientIP(),
				"http_code", c.Writer.Status(),
				"req_body", reqBody,
				"resp_body", replyBody,
			}
			logging.DefaultKit.A().Debugw("http_server", logItems...)
		}
	}
}
