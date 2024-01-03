package indentify

import (
	"net/http"
	"os"
	"reflect"
	"strconv"
)

var (
	IdentifyService IIdentifyService = IdentifyServiceImpl{}
)

var checkMenuUrl = "/identity/account/user/menu"

var checkListUrl = "/identity/account/user/list"

var checkActionUrl = "/identity/account/user/operation"

var checkDelegateActionUrl = "/identity/account/delegate/operation"

var batchCheckActionUrl = "/identity/account/user/batch-operation"

var baseUrl = "http://account-operation-console-svc.ccos-ops-user"

var baseUrlKey = "account.operation.identity.url"

type IIdentifyService interface {
	IdentifyMenu(identify *IdentityUserMenuRequest, re *http.Request) *IdentityUserMenuResponse
	IdentifyListAuth(identify *IdentityUserListRequest, re *http.Request) *IdentityUserListResponse
	IdentifyAuth(identify *IdentityUserOperationRequest, re *http.Request) *IdentityUserOperationResponse
	BatchIdentifyAuth(identify *BatchIdentityUserOperationRequest, re *http.Request) *IdentityUserOperationResponse
}

type IdentifyServiceImpl struct {
}

func (s IdentifyServiceImpl) IdentifyMenu(identify *IdentityUserMenuRequest, request *http.Request) *IdentityUserMenuResponse {
	getUserInfo := request.Header.Get("X-CC-AuthData")
	var requestId = request.Header.Get("requestId")
	var response = new(IdentityUserMenuResponse)
	postResponse, err := PostMenu(getIdentityUrl(MENU), identify, getUserInfo)
	if err != nil {
		println(err)
		response.IsAuthorized = false
		response.IdentityReason = "调用鉴权服务链接异常"
		return response
	}
	if !reflect.DeepEqual(postResponse.Code, "Success") {
		var info = "requestId=" + requestId + "调用鉴权服务发送异常" + postResponse.Message
		println(info)
		response.IsAuthorized = false
		return response
	}
	if postResponse.Data.IdentityReason != "" {
		var info = "requestId=" + requestId + "-调用鉴权服务错误" + postResponse.Data.IdentityReason
		println(info)
	}
	return postResponse.Data

}

func (s IdentifyServiceImpl) IdentifyListAuth(identify *IdentityUserListRequest, request *http.Request) *IdentityUserListResponse {
	getUserInfo := request.Header.Get("X-CC-AuthData")
	var requestId = request.Header.Get("requestId")
	var response = new(IdentityUserListResponse)
	postResponse, err := PostList(getIdentityUrl(LIST), identify, getUserInfo)
	if err != nil {
		println(err)
		response.IsAuthorized = false
		response.IdentityReason = "调用鉴权服务链接异常"
		return response
	}
	if !reflect.DeepEqual(postResponse.Code, "Success") {
		var info = "requestId=" + requestId + "调用鉴权服务发送异常" + postResponse.Message
		println(info)
		response.IsAuthorized = false
		return response
	}
	if postResponse.Data.IdentityReason != "" {
		var info = "requestId=" + requestId + "-调用鉴权服务错误" + postResponse.Data.IdentityReason + "-返回的标志" + strconv.FormatBool(postResponse.Data.IsAuthorized)
		println(info)
	}
	return postResponse.Data
}

func (s IdentifyServiceImpl) IdentifyAuth(identify *IdentityUserOperationRequest, request *http.Request) *IdentityUserOperationResponse {
	getUserInfo := request.Header.Get("X-CC-AuthData")
	delegateToken := request.Header.Get("X-CC-STSToken")

	var requestId = request.Header.Get("requestId")

	var response = new(IdentityUserOperationResponse)

	var err error
	var identifyResponse *IdentifyPostOperationResponse

	if delegateToken != "" {
		identifyDelegate := IdentityDelegateOperationRequest{
			ActionCode:      identify.ActionCode,
			TenantId:        identify.TenantId,
			Token:           delegateToken,
			ResourceId:      identify.ResourceId,
			DepartmentId:    identify.DepartmentId,
			ResourceGroupId: identify.ResourceGroupId,
		}

		identifyResponse, err = PostDelegateOperation(getIdentityUrl(DelegateOption), &identifyDelegate)
	}

	if delegateToken == "" {
		identifyResponse, err = PostOperation(getIdentityUrl(OPTION), identify, getUserInfo)
	}

	if err != nil {
		println(err)
		response.IsAuthorized = false
		response.IdentityReason = "调用鉴权服务链接异常"
		return response
	}

	if !reflect.DeepEqual(identifyResponse.Code, "Success") {
		var info = "requestId=" + requestId + "调用鉴权服务发送异常" + identifyResponse.Message
		println(info)
		response.IsAuthorized = false
		return response
	}

	if identifyResponse.Data.IdentityReason != "" {
		var info = "requestId=" + requestId + "-调用鉴权服务错误" + identifyResponse.Data.IdentityReason + "-返回的标志" + strconv.FormatBool(identifyResponse.Data.IsAuthorized)
		println(info)
	}

	return identifyResponse.Data
}

func (s IdentifyServiceImpl) BatchIdentifyAuth(identify *BatchIdentityUserOperationRequest, request *http.Request) *IdentityUserOperationResponse {
	getUserInfo := request.Header.Get("X-CC-AuthData")
	var requestId = request.Header.Get("requestId")
	var response = new(IdentityUserOperationResponse)
	postResponse, err := BatchPostOperation(getIdentityUrl(BatchOption), identify, getUserInfo)
	if err != nil {
		println(err)
		response.IsAuthorized = false
		response.IdentityReason = "调用鉴权服务链接异常"
		return response
	}
	if !reflect.DeepEqual(postResponse.Code, "Success") {
		var info = "requestId=" + requestId + "调用鉴权服务发送异常" + postResponse.Message
		println(info)
		response.IsAuthorized = false
		return response
	}
	if postResponse.Data.IdentityReason != "" {
		var info = "requestId=" + requestId + "-调用鉴权服务错误" + postResponse.Data.IdentityReason + "-返回的标志" + strconv.FormatBool(postResponse.Data.IsAuthorized)
		println(info)
	}
	return postResponse.Data
}

func getIdentityUrl(option int) string {
	baseEnvUrl := os.Getenv(baseUrlKey)
	if baseEnvUrl != "" && len(baseEnvUrl) > 0 {
		baseUrl = baseEnvUrl
	}

	var urlPath string
	switch option {
	case 1:
		urlPath = checkMenuUrl
	case 2:
		urlPath = checkListUrl
	case 3:
		urlPath = checkActionUrl
	case 4:
		urlPath = batchCheckActionUrl
	case 5:
		urlPath = checkDelegateActionUrl
	default:
		return ""
	}

	return baseUrl + urlPath
}
