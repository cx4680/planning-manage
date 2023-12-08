package customer

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/cloud_platform"
	"github.com/go-ldap/ldap/v3"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"
	"time"
)

func createCustomer(customerParam CreateCustomerRequest, leaderId string, ldapUser *ldap.Entry, currentUserId string) (*entity.CustomerManage, error) {
	customerManage := entity.CustomerManage{
		CustomerName: customerParam.CustomerName,
		LeaderId:     leaderId,
		LeaderName:   ldapUser.GetAttributeValue("displayName"),
		CreateUserId: currentUserId,
		UpdateUserId: currentUserId,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
		DeleteState:  0,
	}
	var customer entity.CustomerManage
	if err := data.DB.Table(entity.CustomerManageTable).Create(&customerManage).Scan(&customer).Error; err != nil {
		log.Errorf("[createCustomer] query db error", err)
		return nil, err
	}

	request := cloud_platform.Request{
		UserId:     currentUserId,
		CustomerId: customer.ID,
	}
	// 创建云平台、创建默认region、az、cell
	if err := cloud_platform.CreateCloudPlatformByCustomerId(&request); err != nil {
		log.Errorf("[createCustomer] CreateCloudPlatformByCustomerId error, %v", err)
		return nil, err
	}
	//保存用户到数据库
	//user.SaveUser(ldapUser)
	// 如果membersId不为空，创建成员
	if len(customerParam.MembersId) > 0 {
		var memberList []entity.PermissionsManage
		for i, memberId := range customerParam.MembersId {
			member := entity.PermissionsManage{
				UserId:      memberId,
				UserName:    customerParam.MembersName[i],
				CustomerId:  customer.ID,
				DeleteState: 0,
			}
			memberList = append(memberList, member)
		}
		if err := data.DB.Table(entity.PermissionsManageTable).CreateInBatches(&memberList, len(memberList)).Error; err != nil {
			log.Errorf("[createCustomer] create members error", err)
			return nil, err
		}
	}

	return &customer, nil
}

func pageCustomer(customerPageParam PageCustomerRequest, currentUserId string) ([]entity.CustomerManage, int64) {
	log.Infof("current user id:%s", currentUserId)
	var roleManage entity.RoleManage
	if err := data.DB.Table(entity.RoleManageTable).Where("user_id=?", currentUserId).Scan(&roleManage).Error; err != nil {
		log.Errorf("[pageCustomer] query role manage from db error")
		return nil, 0
	}
	var customerList []entity.CustomerManage
	var count int64

	db := data.DB.Table("customer_manage cm").Select("DISTINCT cm.*")

	db.Where("cm.delete_state=0")
	if len(customerPageParam.CustomerName) > 0 {
		db.Where("cm.customer_name like ?", `%`+customerPageParam.CustomerName+`%`)
	}
	if len(customerPageParam.LeaderName) > 0 {
		db.Where("cm.leader_name like ?", `%`+customerPageParam.LeaderName+`%`)
	}
	if roleManage.Role != "admin" {
		db.Where("cm.leader_id = ? OR pm.user_id = ?", currentUserId, currentUserId)
	}
	if err := db.Joins("LEFT JOIN permissions_manage pm ON pm.customer_id = cm.id").
		Order(customerPageParam.OrderBy).
		Limit(customerPageParam.Size).
		Offset((customerPageParam.Current - 1) * customerPageParam.Size).
		Find(&customerList).
		Select("count(DISTINCT cm.id)").
		Limit(-1).
		Offset(-1).
		Count(&count).Error; err != nil {
		log.Errorf("[pageCustomer] query db error")
		return nil, 0
	}

	return customerList, count
}

func updateCustomer(customerParam UpdateCustomerRequest, currentUserId string) error {
	var customerManage entity.CustomerManage
	if err := data.DB.Table(entity.CustomerManageTable).Where("id=?", customerParam.ID).Scan(&customerManage).Error; err != nil {
		log.Errorf("[updateCustomer] query customer by id error,%v", err)
		return err
	}

	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		if customerManage.ID != 0 {
			customerManageUpdate := entity.CustomerManage{
				ID:           customerParam.ID,
				CustomerName: customerParam.CustomerName,
				LeaderId:     customerParam.LeaderId,
				LeaderName:   customerParam.LeaderName,
				UpdateTime:   time.Now(),
				UpdateUserId: currentUserId,
			}
			if err := tx.Table(entity.CustomerManageTable).Updates(&customerManageUpdate).Error; err != nil {
				log.Errorf("[updateCustomer] update customer error, %v", err)
				return err
			}
			//如果修改了接口人，同时保存接口人信息
			/*if customerManage.LeaderId != customerParam.LeaderId {
				ldapUser, err := user.SearchUserById(customerParam.LeaderId)
				if err != nil {
					log.Errorf("[updateCustomer] search user by id error,%v", err)
					return err
				}
				user.SaveUser(ldapUser)
			}*/
		}
		// 更改成员信息
		members, err := searchMembersByCustomerId(customerParam.ID)
		if err != nil {
			log.Errorf("[updateCustomer] search customer members error, %v", err)
			return err
		}
		var deleteIdList []string
		var addIdList []string
		var addNameList []string
		for i, paramMemberId := range customerParam.MembersId {
			found := false
			for _, member := range members {
				if paramMemberId == member.UserId {
					found = true
					break
				}
			}
			if !found {
				addIdList = append(addIdList, paramMemberId)
				addNameList = append(addNameList, customerParam.MembersName[i])
			}
		}

		for _, member := range members {
			found := false
			for _, paramMemberId := range customerParam.MembersId {
				if paramMemberId == member.UserId {
					found = true
					break
				}
			}
			if !found {
				deleteIdList = append(deleteIdList, member.UserId)
			}
		}

		if len(addIdList) > 0 {
			var permissionManageList []entity.PermissionsManage
			for i, id := range addIdList {
				permissionManage := entity.PermissionsManage{
					UserId:      id,
					UserName:    addNameList[i],
					CustomerId:  customerParam.ID,
					DeleteState: 0,
				}
				permissionManageList = append(permissionManageList, permissionManage)
			}
			if err = tx.Table(entity.PermissionsManageTable).CreateInBatches(&permissionManageList, len(permissionManageList)).Error; err != nil {
				log.Errorf("[updateCustomer] batch create customer members error, %v", err)
				return err
			}
		}
		if len(deleteIdList) > 0 {
			if err = tx.Table(entity.PermissionsManageTable).Where("user_id in ?", deleteIdList).UpdateColumn("delete_state", 1).Error; err != nil {
				log.Errorf("[updateCustomer] batch delete customer members error, %v", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func searchMembersByCustomerId(customerId int64) ([]entity.PermissionsManage, error) {
	var members []entity.PermissionsManage
	if err := data.DB.Table(entity.PermissionsManageTable).Where("customer_id=? and delete_state=0", customerId).Scan(&members).Error; err != nil {
		log.Errorf("[updateCustomer] query customer members error, %v", err)
		return nil, err
	}
	return members, nil
}

func searchCustomerById(customerId int64) (*entity.CustomerManage, error) {
	var customer entity.CustomerManage
	if err := data.DB.Table(entity.CustomerManageTable).Where("id=? and delete_state=0", customerId).Scan(&customer).Error; err != nil {
		log.Errorf("[updateCustomer] query customer members error, %v", err)
		return nil, err
	}
	return &customer, nil
}
