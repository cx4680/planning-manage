package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	oteltrace "go.opentelemetry.io/otel/trace"

	"code.cestc.cn/ccos/cnm/ops-base/logging"
	"code.cestc.cn/ccos/cnm/ops-base/opserror"
	"code.cestc.cn/ccos/cnm/ops-base/opshttp"
	"code.cestc.cn/ccos/cnm/ops-base/tools/jsonx"
	"code.cestc.cn/ccos/cnm/ops-base/trace"
	"code.cestc.cn/ccos/cnm/ops-base/utils/commonutils"
	"code.cestc.cn/ccos/cnm/ops-base/utils/timeutils"
	"code.cestc.cn/ccos/cnm/ops-base/utils/userutils"

	"go.uber.org/zap"
)

const (
	defaultTimeout = 30
)

var clientOnce = sync.Once{}

var _defaultClient *http.Client

func DefaultClient() *http.Client {
	clientOnce.Do(func() {
		_defaultClient = NewHttpClient(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}, 30)
	})
	return _defaultClient
}

// NewHttpClient 创建http client
// 可以根据配置自动创建client，生成map，通过需要调用的service名获取client
func NewHttpClient(transport http.RoundTripper, timeout int64) *http.Client {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if transport == nil {
		transport = http.DefaultTransport
	}
	trace.Init()
	return &http.Client{
		Transport: otelhttp.NewTransport(transport),
		Timeout:   time.Duration(timeout) * time.Second,
	}
}

type request struct {
	ctx         context.Context
	httpClient  *http.Client
	url         string
	header      http.Header
	body        io.Reader
	reqBody     []byte
	method      string
	err         error
	respBody    []byte
	statusCode  int
	status      string
	traceId     string
	spanContext context.Context
}

func NewReq(ctx context.Context, httpClient *http.Client) *request {
	r := &request{
		ctx:        ctx,
		header:     http.Header{},
		httpClient: httpClient,
	}
	r.genTrace()
	r.setPublicHeaders()
	return r
}

func (r *request) genTrace() {
	// trace相关
	spanCtx := context.Background()
	var span oteltrace.Span
	ginCtx, ok := r.ctx.(*gin.Context)
	if ok {
		span = oteltrace.SpanFromContext(ginCtx.Request.Context())
	} else {
		span = oteltrace.SpanFromContext(r.ctx)
	}
	sc := span.SpanContext()
	if !sc.HasTraceID() {
		tracer := trace.GetTracer()
		spanName := os.Getenv(trace.EnvServiceName)
		spanCtx, span = tracer.Start(spanCtx, spanName)
	}

	r.spanContext = oteltrace.ContextWithSpan(spanCtx, span)
	sc = span.SpanContext()
	r.traceId = sc.TraceID().String()
}

//func NewReq(ctx context.Context, httpClient *http.Client) *request {
//	r := &request{
//		ctx:        ctx,
//		header:     http.Header{},
//		httpClient: httpClient,
//	}
//	r.header.Set("Content-Type", "application/json")
//
//	// trace相关
//	var span oteltrace.Span
//	ginCtx, ok := r.ctx.(*gin.Context)
//	if ok {
//		r.ctx = ginCtx.Request.Context()
//	}
//	span = oteltrace.SpanFromContext(r.ctx)
//
//	sc := span.SpanContext()
//	if !sc.HasTraceID() {
//		newCtx := r.ctx
//		tracer := trace.GetTracer()
//		spanName := os.Getenv(trace.EnvServiceName)
//		newCtx, span = tracer.Start(r.ctx, spanName)
//		r.ctx = oteltrace.ContextWithSpan(newCtx, span)
//	}
//
//	sc = span.SpanContext()
//	r.traceId = sc.TraceID().String()
//
//	r.setPublicHeaders()
//	return r
//}

func (r *request) setPublicHeaders() {
	r.header.Set("Content-Type", "application/json")
	r.header.Set(userutils.GatewayInnerHeaderKey, commonutils.GetInnerData(r.ctx))
	r.header.Set(trace.HeaderRequestIdKey, trace.GetRequestId(r.ctx))
}

func (r *request) Get(url string) *request {
	r.url = url
	r.method = http.MethodGet
	return r
}

func (r *request) Post(url string) *request {
	r.url = url
	r.method = http.MethodPost
	return r
}

func (r *request) WithHeader(k string, v interface{}) *request {
	if r.header == nil {
		r.header = http.Header{}
	}
	r.header.Set(k, fmt.Sprint(v))
	return r
}

func (r *request) WithHeaderMap(header map[string]interface{}) *request {
	for k, v := range header {
		r.header.Set(k, fmt.Sprint(v))
	}
	return r
}

func (r *request) WithHeaders(keyAndValues ...interface{}) *request {
	l := len(keyAndValues) - 1
	for i := 0; i < l; i += 2 {
		k := fmt.Sprint(keyAndValues[i])
		r.header.Set(k, fmt.Sprint(keyAndValues[i+1]))
	}
	if (l+1)%2 == 1 {
		logging.For(r.ctx, zap.String("func", "client.NewReq().XXX().WithHeaders")).Warnw("the keys are not aligned")
		k := fmt.Sprint(keyAndValues[l])
		r.header.Set(k, "")
	}
	return r
}

func (r *request) WithUser() *request {
	r.header.Set(userutils.GatewayOldInfoHeaderKey, commonutils.GetAuthData(r.ctx))
	r.header.Set(userutils.GatewayNewInfoHeaderKey, commonutils.GetUserData(r.ctx))
	return r
}

func (r *request) WithBody(body interface{}) *request {
	switch v := body.(type) {
	case io.Reader:
		buf, err := ioutil.ReadAll(v)
		if err != nil {
			r.err = err
			return r
		}
		r.body = bytes.NewReader(buf)
		r.reqBody = buf
	case []byte:
		r.body = bytes.NewReader(v)
		r.reqBody = body.([]byte)
	case string:
		r.body = strings.NewReader(v)
		r.reqBody = []byte(body.(string))
	default:
		buf, err := jsonx.Marshal(body)
		if err != nil {
			r.err = err
			return r
		}
		r.body = bytes.NewReader(buf)
		r.reqBody = buf
	}
	return r
}

func (r *request) Response() *request {
	if r.method == http.MethodGet {
		r.body = nil
	}

	nowTime := time.Now()
	req, err := http.NewRequestWithContext(r.spanContext, r.method, r.url, r.body)
	if err != nil {
		r.err = err
		return r
	}
	req.Header = r.header
	resp, err := r.httpClient.Do(req)

	var (
		body       []byte
		status     = fmt.Sprint(http.StatusInternalServerError)
		statusCode = http.StatusInternalServerError
	)
	if err != nil {
		r.err = err
		status = err.Error()
	} else {
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			r.err = err
		}
		status = resp.Status
		statusCode = resp.StatusCode
		defer func() {
			_ = resp.Body.Close()
		}()
	}

	// 日志打印
	var logReplyBody, logReqBody []byte
	logReplyBody = append(logReplyBody, body...)
	if len(logReplyBody) > 512 {
		logReplyBody = logReplyBody[0:512]
	}

	logReqBody = append(logReqBody, r.reqBody...)
	if len(logReqBody) > 512 {
		logReqBody = logReqBody[0:512]
	}

	logItems := []interface{}{
		"start", nowTime.Format(timeutils.TimeFormatYYYYMMDDHHmmSS),
		"cost", math.Ceil(float64(time.Since(nowTime).Nanoseconds()) / 1e6),
		"trace_id", r.traceId,
		"req_method", r.method,
		"req_uri", r.url,
		"http_code", statusCode,
		"status", status,
		"req_body", string(logReqBody),
		"resp_body", string(logReplyBody),
	}
	logging.DefaultKit.R().Infow("http_client", logItems...)

	r.respBody = body
	r.statusCode = statusCode
	r.status = status
	return r
}

func (r *request) ParseJson(data interface{}) error {
	return r.ParseDataJson(data)
}

func (r *request) ParseString(str *string) error {
	if r.err != nil {
		return r.err
	}

	if r.statusCode != http.StatusOK {
		return errors.New(r.status)
	}

	*str = string(r.respBody)
	return nil
}

func (r *request) ParseEmpty() error {
	return r.ParseDataJson(nil)
}

func (r *request) ParseDataJson(data interface{}) error {
	if r.err != nil {
		return r.err
	}

	// 空解析
	if data == nil {
		return nil
	}

	var resp opshttp.WrapResp

	err := jsonx.Unmarshal(r.respBody, &resp)
	if err != nil && data != nil {
		return jsonx.Unmarshal(r.respBody, data)
	}

	if len(resp.Code) > 0 && strings.ToLower(resp.Code) != strings.ToLower(opserror.SuccessCode) {
		return opserror.AddError(resp.Code, resp.Msg, r.statusCode)
	}

	marshal, _ := jsonx.Marshal(resp.Data)

	return jsonx.Unmarshal(marshal, data)
}

func (r *request) ParseAllJson(data interface{}) error {
	if r.err != nil {
		return r.err
	}

	return jsonx.Unmarshal(r.respBody, data)
}
