package utils

type UserContext struct {
	System   string
	UserInfo interface{}
}

func (ctx *UserContext) GetSystem() string {
	return ctx.System
}

func (ctx *UserContext) GetUserInfo() interface{} {
	return ctx.UserInfo
}
