package indentify

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type identityResponse struct {
	body []byte
	err  error
}

type identityRequest struct {
	url      string
	body     interface{}
	userInfo string
}

func PostMenu(url string, params *IdentityUserMenuRequest, userInfo string) (*IdentifyPostMenuResponse, error) {
	resp := postIdentityRequest(identityRequest{url, params, userInfo})
	rsp := new(IdentifyPostMenuResponse)
	err := json.Unmarshal(resp.body, &rsp)
	return rsp, err
}

func PostList(url string, params *IdentityUserListRequest, userInfo string) (*IdentifyPostListResponse, error) {
	resp := postIdentityRequest(identityRequest{url, params, userInfo})
	rsp := new(IdentifyPostListResponse)
	err := json.Unmarshal(resp.body, &rsp)
	return rsp, err
}

func PostOperation(url string, params *IdentityUserOperationRequest, userInfo string) (*IdentifyPostOperationResponse, error) {
	resp := postIdentityRequest(identityRequest{url, params, userInfo})
	rsp := new(IdentifyPostOperationResponse)
	err := json.Unmarshal(resp.body, &rsp)
	return rsp, err
}

func PostDelegateOperation(url string, params *IdentityDelegateOperationRequest) (*IdentifyPostOperationResponse, error) {
	resp := postIdentityRequest(identityRequest{url, params, ""})
	rsp := new(IdentifyPostOperationResponse)
	err := json.Unmarshal(resp.body, &rsp)
	return rsp, err
}

func BatchPostOperation(url string, params *BatchIdentityUserOperationRequest, userInfo string) (*IdentifyPostOperationResponse, error) {
	resp := postIdentityRequest(identityRequest{url, params, userInfo})
	rsp := new(IdentifyPostOperationResponse)
	err := json.Unmarshal(resp.body, &rsp)
	return rsp, err
}

func postIdentityRequest(req identityRequest) identityResponse {
	body, err := json.Marshal(req.body)
	if err != nil {
		return identityResponse{nil, err}
	}

	log.Printf("[SDK] 鉴权请求参数： url=%s, 请求参数%s\n", req.url, string(body))

	httpReq, err := http.NewRequest(http.MethodPost, req.url, bytes.NewReader(body))
	if err != nil {
		log.Printf("[SDK] 鉴权请求服务异常[请求参数处理]： url=%s, 请求参数%s\n", req.url, err)
		return identityResponse{nil, err}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-CC-AuthData", req.userInfo)

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[SDK] 鉴权请求服务异常[请求结果]： url=%s, 异常信息%s\n", req.url, err)
		return identityResponse{nil, err}
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[SDK] 鉴权请求服务异常[respBody读取]： url=%s, 异常信息%s\n", req.url, err)
		return identityResponse{nil, err}
	}

	log.Printf("[SDK] 鉴权结果：%s\n", string(respBody))

	if resp.StatusCode != http.StatusOK {
		log.Printf("[SDK] 鉴权请求服务异常[状态码]： url=%s, StatusCode%s\n", req.url, resp.StatusCode)
		return identityResponse{nil, errors.New("鉴权请求失败")}
	}

	return identityResponse{respBody, nil}
}
