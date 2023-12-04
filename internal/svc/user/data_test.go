package user

import (
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"testing"
)

func Test_Dn2Department(t *testing.T) {
	dn := "dn:uid=dc_portal_admin4,ou=平台开发组,ou=数字化部（中电云）,ou=职能管理,ou=中电云计算技术有限公司,ou=数字化业务,ou=深圳市桑达实业股份有限公司（集团）,dc=mylitboy,dc=com"
	department := Dn2Department(dn)

	log.Infof("department:%s", department)
}
