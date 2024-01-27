package customer

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/user"
	"encoding/json"
	"errors"
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

	customerList, count := pageCustomer(customerPageParam, context.GetString(constant.CurrentUserId))

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
	if len(customerParam.LeaderId) < 1 || len(customerParam.LeaderName) < 1 || len(customerParam.MembersId) < 1 || len(customerParam.MembersName) < 1 {
		log.Errorf("[Create] customer membersId or membersName can not be nil", err)
		result.FailureWithMsg(context, errorcodes.InvalidParam, http.StatusBadRequest, "客户接口人及项目成员必选")
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

	currentUserId := context.GetString(constant.CurrentUserId)
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

	if err = updateCustomer(customerParam, context.GetString(constant.CurrentUserId)); err != nil {
		log.Errorf("[Update] customer updateCustomer: ", err)
	}

	result.Success(context, nil)
	return
}

func InnerCreate(c *gin.Context) {
	var request CreateCustomerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("[Create] customer bind param error", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if err := checkRequest(&request, nil, true); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
	}
	ldapUser, err := user.SearchUserById(request.LeaderId)
	if err != nil {
		log.Errorf("[Create] customer search ldap user error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	customer, err := InnerCreateCustomer(request, request.LeaderId, ldapUser, "")
	if err != nil {
		log.Errorf("[Create] customer %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}

	result.Success(c, customer)
	return
}

func InnerUpdate(c *gin.Context) {
	var request = &UpdateCustomerRequest{}
	if err := c.BindJSON(&request); err != nil {
		log.Errorf("[Update] customer bind param error", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if err := checkRequest(nil, request, false); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
	}
	if err := InnerUpdateCustomer(*request); err != nil {
		log.Errorf("[Update] customer updateCustomer: ", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
	return
}

func checkRequest(createRequest *CreateCustomerRequest, updateRequest *UpdateCustomerRequest, isCreate bool) error {
	if isCreate {
		if createRequest.CustomerName == "" {
			log.Error("[Create] customer customerName can not be nil")
			return errors.New("customerName不能为空")
		}
		if len(createRequest.CustomerName) > 30 {
			log.Error("[Create] customer customerName is limited 30 character")
			return errors.New("客户名称不可超过30个字符")
		}
		if len(createRequest.LeaderId) < 1 || len(createRequest.LeaderName) < 1 || len(createRequest.MembersId) < 1 || len(createRequest.MembersName) < 1 {
			log.Error("[Create] customer membersId or membersName can not be nil")
			return errors.New("客户接口人及项目成员必选")
		}
		customerExist, err := searchCustomerByName(createRequest.CustomerName)
		if err != nil {
			log.Errorf("[Create] customer search customer by name error", err)
			return err
		}
		if len(customerExist) > 0 {
			log.Errorf("[Create] customer customerName has exists")
			return errors.New("客户名称重复")
		}
	} else {
		if updateRequest.ID == 0 {
			log.Error("[Update] customer customerId can not be nil")
			return errors.New("id不能为空")
		}
	}
	return nil
}
