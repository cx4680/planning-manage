package userutils

import (
	"context"
	"net/http"

	"code.cestc.cn/ccos/account-tenant-sdk/pkg/indentify"

	"code.cestc.cn/ccos/cnm/ops-base/logging"
	"code.cestc.cn/ccos/cnm/ops-base/opserror"
)

func Auth(request *http.Request, secretFile, action, tenantId, resourceId, departmentId, resourceGroupId string) (bool, error) {
	user, err := GetUser(request, secretFile)
	if err != nil {
		return false, err
	}

	// 用户为空时返回鉴权成功
	if len(user.GetUserId()) <= 0 {
		return true, nil
	}

	// 运维测用户不用鉴权
	if user.GetSystem() == SystemOmp {
		return true, nil
	}

	// 调用sdk判断
	resp := indentify.IdentifyService.IdentifyAuth(&indentify.IdentityUserOperationRequest{
		ActionCode:      action,
		TenantId:        tenantId,
		UserId:          user.GetUserId(),
		ResourceId:      resourceId,
		DepartmentId:    departmentId,
		ResourceGroupId: resourceGroupId,
	}, request)

	if !resp.IsAuthorized {
		err = opserror.AddSpecialError(opserror.ForbiddenCode, resp.IdentityReason, http.StatusBadRequest)
		logging.Errorw("error", err)
		return false, err
	}
	return true, nil
}

func AuthWithRequest(request *http.Request, secretFile string, operation indentify.IdentityUserOperationRequest) (bool, error) {
	user, err := GetUser(request, secretFile)
	if err != nil {
		return false, err
	}

	// 用户为空时返回鉴权成功
	if len(user.GetUserId()) <= 0 {
		return true, nil
	}

	// 运维测用户不用鉴权
	if user.GetSystem() == SystemOmp {
		return true, nil
	}

	// 调用sdk判断
	resp := indentify.IdentifyService.IdentifyAuth(&operation, request)

	if !resp.IsAuthorized {
		err = opserror.AddSpecialError(opserror.ForbiddenCode, resp.IdentityReason, http.StatusBadRequest)
		logging.Errorw("error", err)
		return false, err
	}
	return true, nil
}

func AuthWithUser(request *http.Request, user *User, action, tenantId, resourceId, departmentId, resourceGroupId string) (bool, error) {

	// 用户为空时返回鉴权成功
	if user == nil || len(user.GetUserId()) <= 0 {
		return true, nil
	}

	// 运维侧用户不用鉴权
	if user.GetSystem() == SystemOmp {
		return true, nil
	}

	// 运营侧用户鉴权：调用sdk判断
	resp := indentify.IdentifyService.IdentifyAuth(&indentify.IdentityUserOperationRequest{
		ActionCode:      action,
		TenantId:        tenantId,
		UserId:          user.GetUserId(),
		ResourceId:      resourceId,
		DepartmentId:    departmentId,
		ResourceGroupId: resourceGroupId,
	}, request)

	if !resp.IsAuthorized {
		err := opserror.AddSpecialError(opserror.ForbiddenCode, resp.IdentityReason, http.StatusBadRequest)
		logging.Errorw("error", err)
		return false, err
	}
	return true, nil
}

func AuthOnlyInner(request *http.Request) bool {
	header := request.Header.Get(GatewayNewInfoHeaderKey)
	if len(header) <= 0 {
		header = request.Header.Get(GatewayOldInfoHeaderKey)
	}
	if len(header) <= 0 {
		return true
	}
	return false
}

func AuthOnlyOmp(request *http.Request, secretFile string) (bool, error) {
	user, err := GetUser(request, secretFile)
	if err != nil || len(user.GetUserId()) <= 0 {
		return false, err
	}
	if user.GetSystem() == SystemOmp {
		return true, nil
	}
	return false, nil
}

func AuthOnlyOmpUser(user *User) (bool, error) {
	if user == nil || len(user.GetUserId()) <= 0 {
		return false, nil
	}
	if user.GetSystem() == SystemOmp {
		return true, nil
	}
	return false, nil
}

func AuthOnlyOps(request *http.Request, secretFile, action, tenantId, resourceId, departmentId, resourceGroupId string) (bool, error) {
	user, err := GetUser(request, secretFile)
	if err != nil || len(user.GetUserId()) <= 0 {
		return false, err
	}
	if user.GetSystem() != SystemOps {
		return false, nil
	}

	// 调用sdk判断
	resp := indentify.IdentifyService.IdentifyAuth(&indentify.IdentityUserOperationRequest{
		ActionCode:      action,
		TenantId:        tenantId,
		UserId:          user.GetUserId(),
		ResourceId:      resourceId,
		DepartmentId:    departmentId,
		ResourceGroupId: resourceGroupId,
	}, request)

	if !resp.IsAuthorized {
		err = opserror.AddSpecialError(opserror.ForbiddenCode, resp.IdentityReason, http.StatusBadRequest)
		logging.Errorw("error", err)
		return false, err
	}
	return true, nil
}

func AuthOnlyOpsUser(request *http.Request, user *User, action, tenantId, resourceId, departmentId, resourceGroupId string) (bool, error) {
	if user == nil || len(user.GetUserId()) <= 0 {
		return false, nil
	}
	if user.GetSystem() != SystemOps {
		return false, nil
	}

	// 调用sdk判断
	resp := indentify.IdentifyService.IdentifyAuth(&indentify.IdentityUserOperationRequest{
		ActionCode:      action,
		TenantId:        tenantId,
		UserId:          user.GetUserId(),
		ResourceId:      resourceId,
		DepartmentId:    departmentId,
		ResourceGroupId: resourceGroupId,
	}, request)

	if !resp.IsAuthorized {
		err := opserror.AddSpecialError(opserror.ForbiddenCode, resp.IdentityReason, http.StatusBadRequest)
		logging.Errorw("error", err)
		return false, err
	}
	return true, nil
}

func ListAuth(request *http.Request, secretFile, action string) (*indentify.IdentityUserListResponse, error) {
	resp := &indentify.IdentityUserListResponse{
		IsAuthorized:        false,
		Filter:              false,
		ResourceIdsForAllow: make([]string, 0),
		ResourceIdsForDeny:  make([]string, 0),
		DepartmentIds:       make([]string, 0),
		ResourceGroupIds:    make([]string, 0),
	}

	user, err := GetUser(request, secretFile)
	if err != nil {
		return resp, err
	}

	// 用户为空时返回鉴权成功
	if len(user.GetUserId()) <= 0 {
		return resp, nil
	}

	// 运维测用户不用鉴权
	if user.GetSystem() == SystemOmp {
		return resp, nil
	}

	// 调用sdk
	resp = indentify.IdentifyService.IdentifyListAuth(&indentify.IdentityUserListRequest{
		ActionCode: action,
		TenantId:   user.GetTenantId(),
		UserId:     user.GetUserId(),
	}, request)
	if !resp.IsAuthorized {
		err = opserror.AddSpecialError(opserror.ForbiddenCode, resp.IdentityReason, http.StatusBadRequest)
		return resp, err
	}
	return resp, nil
}

func ListAuthWithUser(request *http.Request, user *User, action string) (*indentify.IdentityUserListResponse, error) {
	resp := &indentify.IdentityUserListResponse{
		IsAuthorized:        false,
		Filter:              false,
		ResourceIdsForAllow: make([]string, 0),
		ResourceIdsForDeny:  make([]string, 0),
		DepartmentIds:       make([]string, 0),
		ResourceGroupIds:    make([]string, 0),
	}

	// 用户为空时返回鉴权成功
	if user == nil || len(user.GetUserId()) <= 0 {
		resp.IsAuthorized = true
		return resp, nil
	}

	// 运维测用户不用鉴权
	if user.GetSystem() == SystemOmp {
		resp.IsAuthorized = true
		return resp, nil
	}

	// 调用sdk
	resp = indentify.IdentifyService.IdentifyListAuth(&indentify.IdentityUserListRequest{
		ActionCode: action,
		TenantId:   user.GetTenantId(),
		UserId:     user.GetUserId(),
	}, request)
	if !resp.IsAuthorized {
		err := opserror.AddSpecialError(opserror.ForbiddenCode, resp.IdentityReason, http.StatusBadRequest)
		return resp, err
	}
	return resp, nil
}

func AuthOmp(ctx context.Context) bool {
	user := GetUserByContext(ctx)
	return user != nil && user.GetSystem() == SystemOmp
}

func AuthOps(ctx context.Context) bool {
	user := GetUserByContext(ctx)
	return user != nil && user.GetSystem() == SystemOps
}

func AuthInner(ctx context.Context) bool {
	user := GetUserByContext(ctx)
	return user == nil || len(user.GetUserId()) <= 0
}

func AuthOmpOps(ctx context.Context) bool {
	user := GetUserByContext(ctx)
	return user != nil && (user.GetSystem() == SystemOmp || user.GetSystem() == SystemOps)
}

func AuthOmpInner(ctx context.Context) bool {
	user := GetUserByContext(ctx)
	return user == nil || len(user.GetUserId()) <= 0 || user.GetSystem() == SystemOmp
}

func AuthOpsInner(ctx context.Context) bool {
	user := GetUserByContext(ctx)
	return user == nil || len(user.GetUserId()) <= 0 || user.GetSystem() == SystemOps
}

func AuthAll(ctx context.Context) bool {
	return true
}
