package indentify

type IdentityUserMenuRequest struct {
	Url    string `json:"url"  binding:"required"`    //动作标识
	Method string `json:"method"`                     //运营单位id
	UserId string `json:"userId"  binding:"required"` //用户id
}

type IdentityUserMenuResponse struct {
	IsAuthorized   bool   `json:"isAuthorized"`
	IdentityReason string `json:"identityReason"` // IsAuthorized =fasle 有值，返回鉴权不通过的原因
}

type IdentityUserListRequest struct {
	ActionCode string `json:"actionCode"  binding:"required"` //动作标识
	TenantId   string `json:"tenantId"`                       //运营单位id
	UserId     string `json:"userId"  binding:"required"`     //用户id
}

type IdentityUserListResponse struct {
	IsAuthorized        bool     `json:"isAuthorized"`
	Filter              bool     `json:"filter"`              //动作标识
	ResourceIdsForAllow []string `json:"resourceIdsForAllow"` //有权限的资源id
	ResourceIdsForDeny  []string `json:"resourceIdsForDeny"`  //没有权限的资源id
	DepartmentIds       []string `json:"departmentIds"`       //允许范围的部门id
	ResourceGroupIds    []string `json:"resourceGroupIds"`    //允许范围的资源集id
	IdentityReason      string   `json:"identityReason"`      // IsAuthorized =fasle 有值，返回鉴权不通过的原因
}

type IdentityUserOperationRequest struct {
	ActionCode      string `json:"actionCode"  binding:"required"` //动作标识
	TenantId        string `json:"tenantId"`                       //运营单位id
	UserId          string `json:"userId"  binding:"required"`     //用户id
	ResourceId      string `json:"resourceId"`                     //资源id 没有传0
	DepartmentId    string `json:"departmentId"`                   //部门id 没有传0
	ResourceGroupId string `json:"resourceGroupId"`                //资源组id 没有传0

}

type IdentityUserOperationResponse struct {
	IsAuthorized   bool   `json:"isAuthorized"`
	IdentityReason string `json:"identityReason"` // IsAuthorized =fasle 有值，返回鉴权不通过的原因
}

type IdentifyPostMenuResponse struct {
	RequestId string                    `json:"requestId"`
	Code      string                    `json:"code"`
	Message   string                    `json:"message"`
	Data      *IdentityUserMenuResponse `json:"data"`
}

type IdentifyPostListResponse struct {
	RequestId string                    `json:"requestId"`
	Code      string                    `json:"code"`
	Message   string                    `json:"message"`
	Data      *IdentityUserListResponse `json:"data"`
}

type IdentifyPostOperationResponse struct {
	RequestId string                         `json:"requestId"`
	Code      string                         `json:"code"`
	Message   string                         `json:"message"`
	Data      *IdentityUserOperationResponse `json:"data"`
}

type BatchIdentityUserOperationRequest struct {
	ActionCode       string   `json:"actionCode"  binding:"required"` //动作标识
	TenantId         string   `json:"tenantId"`                       //运营单位id
	UserId           string   `json:"userId"  binding:"required"`     //用户id
	ResourceIds      []string `json:"resourceIds"`                    //资源ids 没有传
	DepartmentIds    []string `json:"departmentIds"`                  //部门ids 没有传
	ResourceGroupIds []string `json:"resourceGroupIds"`               //资源组ids 没有传

}
