package user

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"net/http"
	"os"
)

func Login(context *gin.Context) {
	userInfo := struct {
		UserId   string `json:"userId"`
		Password string `json:"password"`
	}{}
	err := context.BindJSON(&userInfo)
	if err != nil {
		log.Errorf("[Login] bind param error", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}

	ldapUser, err := SearchUserById(userInfo.UserId)
	if err != nil {
		log.Errorf("[Login] search ldap user by uid error,%v", err)
		result.Failure(context, errorcodes.InvalidUserError, http.StatusInternalServerError)
		return
	}
	userdn := ldapUser.DN
	con, err := ldap.DialURL(os.Getenv("LDAPURLS"))
	if err != nil {
		log.Errorf("[Login] connect err:%v", err)
		result.Failure(context, errorcodes.InvalidUserError, http.StatusInternalServerError)
		return
	}
	defer con.Close()
	con.Debug.Enable(false)
	err = con.Bind(userdn, userInfo.Password)
	if err != nil {
		log.Errorf("[Login] bind user password error,%v", err)
		result.Failure(context, errorcodes.InvalidUserError, http.StatusInternalServerError)
		return
	}
	session := sessions.Default(context)
	log.Debugf("session userId:%s", session.Get("userId"))
	session.Set("userId", userInfo.UserId)
	session.Save()
	result.Success(context, ldapUser)
	log.Debugf("session userId:%s", session.Get("userId"))
	return
}

func Logout(context *gin.Context) {
	session := sessions.Default(context)
	log.Debugf("session userId:%s", session.Get("userId"))
	session.Delete("userId")
	session.Save()
	log.Debugf("session userId:%s", session.Get("userId"))
	result.Success(context, nil)
	return
}

func ListByName(context *gin.Context) {
	queryName := context.Query("name")

	con, err := ldap.DialURL(os.Getenv("LDAPURLS"))
	if err != nil {
		log.Errorf("connect err:%v", err)
	}
	defer func() {
		if con != nil {
			con.Close()
		}
	}()
	con.Debug.Enable(false)
	err = con.Bind(os.Getenv("LDAPUSERNAME"), os.Getenv("LDAPPASSWORD"))
	if err != nil {
		log.Fatal("bind err:", err)
	}

	filter := fmt.Sprintf("(|(uid=*%s*)(displayName=*%s*)(cn=*%s*))", queryName, queryName, queryName)
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
	//log.Infof("Referrals:%v, Controls:%v", searchResult.Referrals, searchResult.Controls)
	var ldapUserList []entity.UserManage
	for _, item := range searchResult.Entries {
		log.Infof("%v", item)
		for _, attribute := range item.Attributes {
			log.Infof("%v:%v", attribute.Name, attribute.Values)
		}
		//item.PrettyPrint(4)
		department, officeName := Dn2Department(item.DN)
		userManage := entity.UserManage{
			ID:               item.GetAttributeValue("uid"),
			Username:         fmt.Sprintf("%s(%s)", item.GetAttributeValue("displayName"), item.GetAttributeValue("uid")),
			EmployeeNumber:   item.GetAttributeValue("employeeNumber"),
			TelephoneNumber:  item.GetAttributeValue("telephoneNumber"),
			Department:       department,
			OfficeName:       officeName,
			DepartmentNumber: item.GetAttributeValue("departmentNumber"),
			Mail:             item.GetAttributeValue("mail"),
			DeleteState:      0,
		}
		ldapUserList = append(ldapUserList, userManage)
	}

	result.Success(context, ldapUserList)
	return
}
