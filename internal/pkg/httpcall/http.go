package httpcall

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"github.com/pkg/errors"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/app/settings"
)

const (
	CODE    = "code"
	SUCCESS = "success"
)

var client *http.Client

type HttpRequest struct {
	Context *gin.Context
	URI     string
	Headers map[string]string
	Body    io.Reader
	method  string
}

func Init(setting *settings.Setting) {
	tr := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 50,
		MaxConnsPerHost:     50,
	}

	client = &http.Client{
		Transport: tr,
		Timeout:   setting.HttpCallTimeoutMinute,
	}
}

func GET(httpRequest HttpRequest) (timeout bool, response *http.Response, err error) {
	httpRequest.method = http.MethodGet
	return doHttpCall(httpRequest)
}

func POST(httpRequest HttpRequest) (timeout bool, response *http.Response, err error) {
	httpRequest.method = http.MethodPost
	return doHttpCall(httpRequest)
}

func HttpCall(method string, httpRequest HttpRequest) (timeout bool, response *http.Response, err error) {
	httpRequest.method = method
	return doHttpCall(httpRequest)
}

func GetResponse(httpRequest HttpRequest) (map[string]interface{}, error) {
	httpRequest.method = http.MethodGet
	return Response(httpRequest)
}

func POSTResponse(httpRequest HttpRequest) (map[string]interface{}, error) {
	httpRequest.method = http.MethodPost
	return Response(httpRequest)
}

func Response(httpRequest HttpRequest) (map[string]interface{}, error) {
	var mapResult map[string]interface{}

	timeout, response, err := doHttpCall(httpRequest)
	if timeout {
		return mapResult, errors.New("httpRequest Timeout")
	}

	if err != nil {
		if response == nil {
			return mapResult, errors.Wrap(err, "response is nil")
		}

		return mapResult, errors.Wrap(err, "response err")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return mapResult, errors.Wrap(err, "response data read err")
	}

	err = json.Unmarshal(body, &mapResult)
	if err != nil {
		return mapResult, errors.Wrap(err, fmt.Sprintf("invalid response data %s with err", string(body)))
	}

	// 如果返回信息是有code字段说明是中心region返回的数据，其他调用次方法的地方进行判断code
	if data, ok := mapResult[CODE]; ok {
		log.Info("result", "code", data)
		return mapResult, nil
	}

	return mapResult, errors.Errorf("invalid response data: does not contain 'code'. The data was :%v", mapResult)
}

func doHttpCall(httpRequest HttpRequest) (timeout bool, response *http.Response, err error) {
	request, err := http.NewRequest(httpRequest.method, httpRequest.URI, httpRequest.Body)
	if err != nil {
		return
	}

	for header, value := range httpRequest.Headers {
		request.Header.Add(header, value)
	}

	if httpRequest.Context != nil && requestID(httpRequest.Context) != "" {
		request.Header.Set(constant.XRequestID, requestID(httpRequest.Context))
	}

	response, err = client.Do(request)
	if err != nil {
		urlError := err.(*url.Error)
		timeout = urlError != nil && urlError.Timeout()
		return
	}

	return
}

func requestID(context *gin.Context) string {
	return context.GetString(constant.XRequestID)
}
