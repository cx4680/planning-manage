package user

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"golang.org/x/crypto/sha3"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/httpcall"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
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
	// 设置session有效期为1天
	sessionAgeStr := os.Getenv("SESSION_AGE")
	sessionAge, _ := strconv.Atoi(sessionAgeStr)
	session.Options(sessions.Options{MaxAge: sessionAge, Path: "/"})
	log.Debugf("session userId:%s", session.Get("userId"))
	session.Set("userId", userInfo.UserId)
	session.Save()
	result.Success(context, ldapUser)
	log.Debugf("session userId:%s", session.Get("userId"))
	return
}

func Logout(context *gin.Context) {
	cookie, err := context.Request.Cookie("cestcToken")
	if err != nil {
		return
	}
	userCenterUrl := os.Getenv(constant.UserCenterUrl)
	productCode := os.Getenv(constant.ProductCode)
	redirectUrlMap := make(map[string]string)
	redirectUrlMap["redirectUrl"] = fmt.Sprintf("%s/auth/sso/ssoLogin?productCode=%s&redirect=", userCenterUrl, productCode)
	token := cookie.Value
	if token == "" {
		log.Errorf("[Auth] invalid authorized")
		result.FailureWithData(context, errorcodes.InvalidAuthorized, http.StatusUnauthorized, redirectUrlMap)
		return
	}
	url := fmt.Sprintf("%s/auth/sso/logout", userCenterUrl)
	body := TokenCheckRequest{
		ProductCode: productCode,
		CestcToken:  token,
	}
	reqJson, err := json.Marshal(body)
	if err != nil {
		log.Errorf("Body json marshal error: %v", err)
	}
	response, err := httpcall.POSTResponse(httpcall.HttpRequest{
		Context: context,
		URI:     url,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: bytes.NewBuffer(reqJson),
	})
	if err != nil {
		log.Errorf("call sso logout error: %v", err)
	}
	log.Infof("call sso logout: %v", response)
	resByte, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Marshal json error: %v", err)
	}
	var responseData LogoutResponse
	if err = json.Unmarshal(resByte, &responseData); err != nil {
		log.Errorf("Unmarshal json error: %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if responseData.Code != RequestSuccessCode {
		responseJson, _ := json.Marshal(response)
		log.Error("call sso logout failure: %s", string(responseJson))
		result.FailureWithData(context, errorcodes.InvalidAuthorized, http.StatusUnauthorized, redirectUrlMap)
		return
	}
	result.Success(context, redirectUrlMap)
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
		[]string{"dn", "cn", "objectClass", "uid", "sn", "mail", "telephoneNumber", "displayName", "employeeNumber"},
		nil,
	)
	searchResult, err := con.Search(searchRequest)
	if err != nil {
		log.Errorf("can't search %s", err.Error())
	}
	// log.Infof("Referrals:%v, Controls:%v", searchResult.Referrals, searchResult.Controls)
	var ldapUserList []entity.UserManage
	for _, item := range searchResult.Entries {
		log.Debugf("dn:%v", item.DN)
		for _, attribute := range item.Attributes {
			log.Infof("%v:%v", attribute.Name, attribute.Values)
		}
		// item.PrettyPrint(4)
		department, officeName := Dn2Department(item.DN)
		userManage := entity.UserManage{
			ID:               item.GetAttributeValue("uid"),
			Username:         item.GetAttributeValue("displayName"),
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

func ListByUid(context *gin.Context) {
	queryName := context.Query("name")

	timestamp := datetime.CurrentUnixMilli()
	productCode := os.Getenv(constant.ProductCode)
	userCenterSecretKey := os.Getenv(constant.UserCenterSecretKey)
	sign := Sha3224(fmt.Sprintf("%s%s%d", productCode, userCenterSecretKey, timestamp))
	url := fmt.Sprintf("%s/auth/apihandler/getUsersByUids", os.Getenv(constant.UserCenterUrl))
	body := QueryUserByUidRequest{
		ProductCode: productCode,
		Sign:        sign,
		Timestamp:   strconv.FormatInt(timestamp, 10),
	}
	body.Data = append(body.Data, queryName)
	reqJson, err := json.Marshal(body)
	if err != nil {
		log.Errorf("Body json marshal error: %v", err)
	}
	response, err := httpcall.POSTResponse(httpcall.HttpRequest{
		Context: context,
		URI:     url,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: bytes.NewBuffer(reqJson),
	})
	if err != nil {
		log.Errorf("call sso getUsersByUids error: %v", err)
	}
	log.Infof("call sso getUsersByUids: %v", response)
	resByte, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Marshal json error: %v", err)
	}
	var responseData QueryUserByUidResponse
	if err = json.Unmarshal(resByte, &responseData); err != nil {
		log.Errorf("Unmarshal json error: %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	var ldapUserList []entity.UserManage
	for _, item := range responseData.Data {
		userManage := entity.UserManage{
			ID:               item.Uid,
			Username:         item.DisplayName,
			EmployeeNumber:   item.EmployeeNumber,
			TelephoneNumber:  item.TelephoneNumber,
			Department:       item.PhysicalDeliveryOfficeName,
			OfficeName:       item.PhysicalDeliveryOfficeName,
			DepartmentNumber: item.DepartmentNumber,
			Mail:             item.Mail,
			DeleteState:      0,
		}
		ldapUserList = append(ldapUserList, userManage)
	}
	result.Success(context, ldapUserList)
	return
}

func ListByEmployeeNumber(context *gin.Context) {
	var empNumList []string
	queryEmpNumList := context.Query("empNumList")
	if queryEmpNumList != "" {
		empNumList = strings.Split(queryEmpNumList, constant.Comma)
	}
	ldapUserList, err := GetEmployeeListByNumberList(empNumList)
	if err != nil {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, ldapUserList)
	return
}

func GetEmployeeListByNumberList(empNumList []string) ([]entity.UserManage, error) {
	var ldapUserList []entity.UserManage
	timestamp := datetime.CurrentUnixMilli()
	productCode := os.Getenv(constant.ProductCode)
	userCenterSecretKey := os.Getenv(constant.UserCenterSecretKey)
	sign := Sha3224(fmt.Sprintf("%s%s%d", productCode, userCenterSecretKey, timestamp))
	url := fmt.Sprintf("%s/auth/apihandler/getStaffListByEmployeeNumList", os.Getenv(constant.UserCenterUrl))
	body := QueryUserByEmployeeNumRequest{
		ProductCode: productCode,
		Sign:        sign,
		Timestamp:   strconv.FormatInt(timestamp, 10),
	}
	body.Data.EmpNumList = empNumList
	reqJson, err := json.Marshal(body)
	if err != nil {
		log.Errorf("Body json marshal error: %v", err)
	}
	response, err := httpcall.POSTResponse(httpcall.HttpRequest{
		Context: &gin.Context{},
		URI:     url,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: bytes.NewBuffer(reqJson),
	})
	if err != nil {
		log.Errorf("call sso getUsersByEmployeeNumber error: %v", err)
	}
	log.Infof("call sso getUsersByEmployeeNumber: %v", response)
	resByte, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Marshal json error: %v", err)
	}
	var responseData QueryUserByUidResponse
	if err = json.Unmarshal(resByte, &responseData); err != nil {
		log.Errorf("Unmarshal json error: %v", err)
		return ldapUserList, err
	}
	for _, item := range responseData.Data {
		userManage := entity.UserManage{
			ID:               item.Uid,
			Username:         item.DisplayName,
			EmployeeNumber:   item.EmployeeNumber,
			TelephoneNumber:  item.TelephoneNumber,
			Department:       item.PhysicalDeliveryOfficeName,
			OfficeName:       item.PhysicalDeliveryOfficeName,
			DepartmentNumber: item.DepartmentNumber,
			Mail:             item.Mail,
			DeleteState:      0,
		}
		ldapUserList = append(ldapUserList, userManage)
	}
	return ldapUserList, nil
}

func Sha3224(str string) string {
	hash := sha3.New224()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}

func CurrentUserInfo(context *gin.Context) {
	userCenterUrl := os.Getenv(constant.UserCenterUrl)
	productCode := os.Getenv(constant.ProductCode)
	redirectUrlMap := make(map[string]string)
	redirectUrlMap["redirectUrl"] = fmt.Sprintf("%s/auth/sso/ssoLogin?productCode=%s&redirect=", userCenterUrl, productCode)
	cookie, err := context.Request.Cookie("cestcToken")
	if err != nil {
		log.Errorf("[Auth] invalid authorized")
		result.FailureWithData(context, errorcodes.InvalidAuthorized, http.StatusUnauthorized, redirectUrlMap)
		return
	}
	token := cookie.Value
	if token == "" {
		log.Errorf("[Auth] invalid authorized")
		result.FailureWithData(context, errorcodes.InvalidAuthorized, http.StatusUnauthorized, redirectUrlMap)
		return
	}
	url := fmt.Sprintf("%s/auth/sso/tokenCheck", userCenterUrl)
	body := TokenCheckRequest{
		ProductCode: productCode,
		CestcToken:  token,
	}
	reqJson, err := json.Marshal(body)
	if err != nil {
		log.Errorf("Body json marshal error: %v", err)
		result.FailureWithData(context, errorcodes.InvalidAuthorized, http.StatusUnauthorized, redirectUrlMap)
		return
	}

	response, err := httpcall.POSTResponse(httpcall.HttpRequest{
		Context: context,
		URI:     url,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: bytes.NewBuffer(reqJson),
	})
	if err != nil {
		log.Errorf("call sso tokenCheck error: %v", err)
		result.FailureWithData(context, errorcodes.InvalidAuthorized, http.StatusUnauthorized, redirectUrlMap)
		return
	}
	log.Infof("call sso tokenCheck: %v", response)
	resByte, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Marshal json error: %v", err)
		result.FailureWithData(context, errorcodes.InvalidAuthorized, http.StatusUnauthorized, redirectUrlMap)
		return
	}
	var responseData TokenCheckResponse
	if err = json.Unmarshal(resByte, &responseData); err != nil {
		log.Errorf("Unmarshal json error: %v", err)
		result.FailureWithData(context, errorcodes.InvalidAuthorized, http.StatusUnauthorized, redirectUrlMap)
		return
	}
	if responseData.Code != RequestSuccessCode {
		responseJson, _ := json.Marshal(response)
		log.Error("call sso tokenCheck failure: %s", string(responseJson))
		result.FailureWithData(context, errorcodes.InvalidAuthorized, http.StatusUnauthorized, redirectUrlMap)
		return
	}
	result.Success(context, map[string]string{
		"deptNum":     responseData.Data.DeptNum,
		"displayName": responseData.Data.DisplayName,
		"workNum":     responseData.Data.WorkNum,
		"l":           responseData.Data.L,
		"email":       responseData.Data.Email,
		"deptName":    responseData.Data.DeptName,
		"mobile":      responseData.Data.Mobile,
		"userName":    responseData.Data.UserName,
		"deptRoot":    responseData.Data.DeptRoot,
	})
	return
}
