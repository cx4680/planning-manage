package customer

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/user"
	"encoding/json"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"net/http"
	"strconv"
)

func Page(context *gin.Context) {
	var customerPageParam PageCustomerRequest
	err := context.BindJSON(&customerPageParam)
	if err != nil {
		log.Errorf("[Page] customer bind param error", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}

	session := sessions.Default(context)
	currentUserId := session.Get("userId").(string)
	customerList, count := pageCustomer(customerPageParam, currentUserId)

	var customerResponseList []CustomerResponse
	marshal, err := json.Marshal(customerList)
	err = json.Unmarshal(marshal, &customerResponseList)
	if err != nil {
		log.Errorf("[page] customer marshal customer err,%v", err)
		result.Failure(context, errorcodes.InvalidData, http.StatusBadRequest)
		return
	}
	//封装权限成员
	var responseList []CustomerResponse
	for _, customer := range customerResponseList {
		members, err := searchMembersByCustomerId(customer.ID)
		if err != nil {
			log.Errorf("[page] customer search members error, %v", err)
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return
		}
		var memberIds []string
		var memberNames []string
		for _, member := range members {
			memberIds = append(memberIds, member.UserId)
			memberNames = append(memberNames, member.UserName)
		}
		customer.MembersId = memberIds
		customer.MembersName = memberNames
		responseList = append(responseList, customer)
	}
	result.SuccessPage(context, count, responseList)
	return
}

func GetById(context *gin.Context) {
	id := context.Param("id")
	idInt, _ := strconv.ParseInt(id, 10, 64)
	customer, err := searchCustomerById(idInt)
	if err != nil {
		log.Errorf("[GetById] search customer by id error, %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	//members, err := searchMembersByCustomerId(customer.ID)
	result.Success(context, customer)
	return
}

func Create(context *gin.Context) {
	var customerParam CreateCustomerRequest
	err := context.BindJSON(&customerParam)
	if err != nil {
		log.Errorf("[Create] customer bind param error", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}

	if customerParam.CustomerName == "" {
		log.Errorf("[Create] customer customerName can not be nil", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if len(customerParam.CustomerName) > 30 {
		log.Errorf("[Create] customer customerName is limited 30 character", err)
		result.FailureWithMsg(context, errorcodes.InvalidParam, http.StatusBadRequest, "客户名称不可超过30个字符")
		return
	}
	customerExist, err := searchCustomerByName(customerParam.CustomerName)
	if err != nil {
		log.Errorf("[Create] customer search customer by name error", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if len(customerExist) > 0 {
		log.Errorf("[Create] customer customerName has exists")
		result.FailureWithMsg(context, errorcodes.CustomerNameExistsError, http.StatusBadRequest, "客户名称重复")
		return
	}

	session := sessions.Default(context)
	currentUserId := session.Get("userId").(string)
	leaderId := customerParam.LeaderId
	if leaderId == "" {
		leaderId = currentUserId
	}
	ldapUser, err := user.SearchUserById(leaderId)
	if err != nil {
		log.Errorf("[Create] customer search ldap user error, %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}

	customer, err := createCustomer(customerParam, leaderId, ldapUser, currentUserId)
	if err != nil {
		log.Errorf("[Create] customer %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}

	result.Success(context, customer)
	return
}

func Update(context *gin.Context) {
	var customerParam UpdateCustomerRequest
	err := context.BindJSON(&customerParam)
	if err != nil {
		log.Errorf("[Update] customer bind param error", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if customerParam.ID == 0 {
		log.Errorf("[Update] customer customerId can not be nil", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}

	if customerParam.CustomerName == "" {
		log.Errorf("[Update] customer customerName can not be nil", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if len(customerParam.CustomerName) > 30 {
		log.Errorf("[Update] customer customerName is limited 30 character", err)
		result.FailureWithMsg(context, errorcodes.InvalidParam, http.StatusBadRequest, "客户名称不可超过30个字符")
		return
	}

	session := sessions.Default(context)
	currentUserId := session.Get("userId").(string)
	updateCustomer(customerParam, currentUserId)

	result.Success(context, nil)
	return
}
