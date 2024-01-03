package user

import (
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"testing"
)

func Test_Dn2Department(t *testing.T) {
	dn := "uid=mengxiangyu,ou=安全品质部,ou=中电洲际环保科技发展有限公司,ou=供热环保业务,ou=深圳市桑达实业股份有限公司（集团）,dc=mylitboy,dc=com"
	department, officeName := Dn2Department(dn)

	log.Infof("dn:%s\ndepartment:%s, officeName:%s", dn, department, officeName)
}
