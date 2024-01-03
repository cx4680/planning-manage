package userutils

import (
	"context"
	"encoding/base64"
	"net/http"

	"code.cestc.cn/ccos/auth-share/pkg/utils"
	jsoniter "github.com/json-iterator/go"

	"code.cestc.cn/ccos/cnm/ops-base/utils/fileutils"
)

const (
	GatewayOldInfoHeaderKey = "X-CC-AuthData"
	GatewayNewInfoHeaderKey = "X-CC-UserData"
	GatewayInnerHeaderKey   = "X-CC-"
	SystemOps               = "ops" // 运营
	SystemOmp               = "omp" // 运维
)

type userInterface interface {
	getUserId() string
	getUserCode() string
	getDepartmentId() string
	getTenantId() string
	getType() int64
	getExtendField() interface{}
}

type User struct {
	isPublic    bool
	System      string                 `json:"system"`
	UserInfoMap map[string]interface{} `json:"userInfo"`
	userInfo    userInterface
}

func (u *User) SetPublic(isPublic bool) {
	u.isPublic = isPublic
}

type Option func(user *User)

// GetUser 获取用户信息
func GetUser(request *http.Request, secretFile string, opts ...Option) (*User, error) {

	u := &User{}

	for _, o := range opts {
		o(u)
	}

	var (
		decode []byte
		err    error
	)

	headerUserData := request.Header.Get(GatewayNewInfoHeaderKey)
	headerAuthData := request.Header.Get(GatewayOldInfoHeaderKey)
	headerInnerData := request.Header.Get(GatewayInnerHeaderKey)

	switch true {
	case len(headerUserData) > 0:
		decode, err = headerNew(headerUserData, secretFile)
		break
	case len(headerAuthData) > 0:
		decode, err = base64.StdEncoding.DecodeString(headerAuthData)
		break
	case len(headerInnerData) > 0:
		decode, err = headerInner(headerInnerData)
		break
	}

	if err != nil {
		if !u.isPublic {
			return u, err
		}
		decode = []byte{}
	}

	_ = jsoniter.Unmarshal(decode, u)

	u.setUserInfo()

	return u, nil
}

func headerNew(header, secretFile string) ([]byte, error) {
	// 解析base64，加密后的用户信息
	cryptographicHeader, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return nil, err
	}

	// 如果headerSecret有值，获取秘钥
	secret, err := fileutils.GetByFileName(secretFile)
	if err != nil {
		return nil, err
	}

	// sdk 获取解密base64
	return utils.RSADecryptBySecretStr(cryptographicHeader, secret)
}

func headerInner(headerInnerData string) ([]byte, error) {
	// TODO 生成user信息 marshal
	return base64.StdEncoding.DecodeString(headerInnerData)
}

func (u *User) setUserInfo() {

	var info userInterface
	switch u.System {
	case SystemOps:
		info = &UserOps{}
	case SystemOmp:
		info = &UserOmp{}
	default:
		info = &userBase{}
	}

	bytes, _ := jsoniter.Marshal(u.UserInfoMap)
	_ = jsoniter.Unmarshal(bytes, info)
	u.userInfo = info
}

func (u *User) GetSystem() string {
	return u.System
}

func (u *User) GetExtendField() interface{} {
	if u.userInfo == nil {
		return nil
	}
	return u.userInfo.getExtendField()
}

func (u *User) GetUserId() string {
	if u.userInfo == nil {
		return ""
	}
	return u.userInfo.getUserId()
}

func (u *User) GetDepartmentId() string {
	if u.userInfo == nil {
		return ""
	}
	return u.userInfo.getDepartmentId()
}

func (u *User) GetUserCode() string {
	if u.userInfo == nil {
		return ""
	}
	return u.userInfo.getUserCode()
}

func (u *User) GetTenantId() string {
	if u.userInfo == nil {
		return ""
	}
	return u.userInfo.getTenantId()
}

func (u *User) GetType() int64 {
	if u.userInfo == nil {
		return 0
	}
	return u.userInfo.getType()
}

func (u *User) GenOpsUser() UserOps {
	if u.userInfo == nil {
		return UserOps{}
	}
	info, ok := u.userInfo.(*UserOps)
	if !ok {
		return UserOps{}
	}
	return *info
}

func (u *User) GenOmpUser() UserOmp {
	if u.userInfo == nil {
		return UserOmp{}
	}
	info, ok := u.userInfo.(*UserOmp)
	if !ok {
		return UserOmp{}
	}
	return *info
}

func GetUserByContext(ctx context.Context) *User {
	val := ctx.Value(AuthUserKey)
	user := &User{}
	if val != nil {
		user, _ = val.(*User)
	}
	return user
}

func SetUserToContext(ctx context.Context, val *User) context.Context {
	return context.WithValue(ctx, AuthUserKey, val)
}
