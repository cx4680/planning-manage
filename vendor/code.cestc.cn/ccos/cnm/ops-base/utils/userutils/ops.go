package userutils

type UserOps struct {
	userBase
	DepartmentId string `json:"departmentId"`
	TenantId     string `json:"tenantId"`
	Type         int64  `json:"type"`
}

func (u *UserOps) getDepartmentId() string {
	return u.DepartmentId
}

func (u *UserOps) getTenantId() string {
	return u.TenantId
}

func (u *UserOps) getType() int64 {
	return u.Type
}
