package httpcall

import (
	"crypto/tls"
	"encoding/json"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"code.cestc.cn/zhangzhi/planning-manage/internal/app/settings"
)

const (
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	CODE    = "code"
	DATA    = "data"
	SUCCESS = "success"
)

var client *http.Client

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

func GETCall(uri string) (timeout bool, response *http.Response, err error) {
	return doHttpCall(GET, uri, nil)
}

func POSTCall(uri string, body io.Reader) (timeout bool, response *http.Response, err error) {
	return doHttpCall(POST, uri, body)
}

func HttpCall(uri, method string, body io.Reader) (timeout bool, response *http.Response, err error) {
	return doHttpCall(method, uri, body)
}

func GetResponse(uri, method string, bodyQ io.Reader) (map[string]interface{}, error) {
	var mapResult map[string]interface{}

	bTimeout, response, err := doHttpCall(uri, method, bodyQ)
	if bTimeout {
		err = errors.New("Request Timeout")
		log.Error(err, "")
		return mapResult, err
	}

	if err != nil {
		err = errors.Wrap(err, "Response err")
		log.Error(err, "Response err")
		if response == nil {
			return mapResult, errors.Wrap(err, "Response is nil")
		}

		return mapResult, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error(err, "Response data read err")
		return mapResult, errors.Wrap(err, "Response data read err")
	}

	err = json.Unmarshal(body, &mapResult)
	if err != nil {
		log.Error(err, "Invalid response data", "body", string(body))
		return mapResult, errors.Wrap(err, "Invalid response data with err")
	}

	// 如果返回信息是有code字段说明是中心region返回的数据，其他调用次方法的地方进行判断code
	if data, ok := mapResult[CODE]; ok {
		log.Info("result", "code", data)
		return mapResult, nil
	}

	return mapResult, errors.Errorf("Invalid response data: does not contain 'code'. The data was :%v", mapResult)
}

func doHttpCall(uri string, method string, body io.Reader) (timeout bool, response *http.Response, err error) {
	request, err := http.NewRequest(method, uri, body)
	if err != nil {
		return
	}

	response, err = client.Do(request)
	if err != nil {
		urlError := err.(*url.Error)
		timeout = urlError != nil && urlError.Timeout()
		return
	}

	return
}
