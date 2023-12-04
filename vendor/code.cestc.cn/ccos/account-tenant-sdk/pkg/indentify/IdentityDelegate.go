package indentify

type IdentityDelegateOperationResponse struct {
	*IdentityUserOperationResponse
}

type IdentityDelegateOperationRequest struct {
	ActionCode      string `json:"actionCode"  binding:"required"` //动作标识
	TenantId        string `json:"tenantId"`                       //运营单位id
	Token           string `json:"token"  binding:"required"`      //token
	ResourceId      string `json:"resourceId"`                     //资源id 没有传0
	DepartmentId    string `json:"departmentId"`                   //部门id 没有传0
	ResourceGroupId string `json:"resourceGroupId"`                //资源组id 没有传0
}
