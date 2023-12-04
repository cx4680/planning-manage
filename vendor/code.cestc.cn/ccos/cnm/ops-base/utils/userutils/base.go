package userutils

type userBase struct {
	UserId      string      `json:"userId"`
	UserCode    string      `json:"userCode"`
	ExtendField interface{} `json:"extendField"`
}

func (u *userBase) getUserId() string {
	return u.UserId
}

func (u *userBase) getUserCode() string {
	return u.UserCode
}

func (u *userBase) getExtendField() interface{} {
	return u.ExtendField
}

func (u *userBase) getDepartmentId() string {
	return ""
}

func (u *userBase) getTenantId() string {
	return ""
}

func (u *userBase) getType() int64 {
	return 0
}
