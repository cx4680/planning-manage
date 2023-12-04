package user

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"fmt"
	ldap "github.com/go-ldap/ldap/v3"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"os"
	"strings"
)

func SaveUser(ldapUser *ldap.Entry) (*entity.UserManage, error) {

	var userManage entity.UserManage
	userId := ldapUser.GetAttributeValue("uid")
	if err := data.DB.Table(entity.UserManageTable).Where("id=?", userId).Scan(&userManage).Error; err != nil {
		log.Errorf("[saveUser] query db error, %v", err)
		return nil, err
	}
	if userManage.ID == "" {
		//解析ldap用户信息
		department, officeName := Dn2Department(ldapUser.DN)
		userManage = entity.UserManage{
			ID:               userId,
			Username:         ldapUser.GetAttributeValue("displayName"),
			EmployeeNumber:   ldapUser.GetAttributeValue("employeeNumber"),
			TelephoneNumber:  ldapUser.GetAttributeValue("telephoneNumber"),
			Department:       department,
			OfficeName:       officeName,
			DepartmentNumber: ldapUser.GetAttributeValue("departmentNumber"),
			Mail:             ldapUser.GetAttributeValue("mail"),
			DeleteState:      0,
		}
		if err := data.DB.Table(entity.UserManageTable).Create(&userManage).Error; err != nil {
			log.Errorf("[saveUser] query db error, %v", err)
			return nil, err
		}
		return &userManage, nil
	}
	return &userManage, nil
}

func Dn2Department(dn string) (string, string) {
	dnInfoList := strings.Split(dn, ",")
	var departList []string
	var officeName string
	for i := len(dnInfoList) - 1; i >= 0; i-- {
		if strings.Contains(dnInfoList[i], "ou=") {
			if i == len(dnInfoList)-1 {
				officeName = strings.Trim(dnInfoList[i], "ou=")
			}
			departList = append(departList, strings.Trim(dnInfoList[i], "ou="))
		}
	}
	department := strings.Join(departList, "-")
	return department, officeName
}

func SearchUserById(userId string) (*ldap.Entry, error) {
	con, err := ldap.DialURL(os.Getenv("LDAPURLS"))
	if err != nil {
		log.Errorf("connect err:%v", err)
		return nil, err
	}
	defer func() {
		if con != nil {
			con.Close()
		}
	}()
	con.Debug.Enable(false)
	err = con.Bind(os.Getenv("LDAPUSERNAME"), os.Getenv("LDAPPASSWORD"))
	if err != nil {
		log.Errorf("bind err:", err)
		return nil, err
	}

	filter := fmt.Sprintf("(&(objectClass=posixAccount)(uid=%s))", userId)
	searchRequest := ldap.NewSearchRequest(
		os.Getenv("LDAPBASE"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn", "cn", "objectClass", "uid", "sn", "mail", "telephoneNumber", "displayName"},
		nil,
	)
	searchResult, err := con.Search(searchRequest)
	if err != nil {
		log.Errorf("can't search %s", err.Error())
	}

	return searchResult.Entries[0], nil
}
