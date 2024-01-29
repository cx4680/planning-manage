package customer

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"
	"os"
	"time"
)

func createCustomer(db *gorm.DB, customerParam CreateCustomerRequest, leaderId string, ldapUser *ldap.Entry, currentUserId string) (*entity.CustomerManage, *CreateCloudPlatform, error) {
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
	if err := db.Create(&customerManage).Find(&customer).Error; err != nil {
		log.Errorf("[createCustomer] query db error", err)
		return nil, nil, err
	}
	// 创建云平台、创建默认region、az、cell
	createCloudPlatform, err := CreateCloudPlatformByCustomerId(db, customer.ID, currentUserId)
	if err != nil {
		log.Errorf("[createCustomer] CreateCloudPlatformByCustomerId error, %v", err)
		return nil, nil, err
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
		if err = db.CreateInBatches(&memberList, len(memberList)).Error; err != nil {
			log.Errorf("[createCustomer] create members error", err)
			return nil, nil, err
		}
	}

	return &customer, createCloudPlatform, nil
}

func pageCustomer(customerPageParam PageCustomerRequest, currentUserId string) ([]entity.CustomerManage, int64) {
	log.Infof("current user id:%s", currentUserId)
	var roleManage entity.RoleManage
	if err := data.DB.Where("user_id=?", currentUserId).Find(&roleManage).Error; err != nil {
		log.Errorf("[pageCustomer] query role manage from db error")
		return nil, 0
	}
	var customerList []entity.CustomerManage
	var count int64

	db := data.DB.Table("customer_manage cm").Select("DISTINCT cm.*")

	db.Where("cm.delete_state=0")
	db.Where("pm.delete_state = 0")
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
	if err := data.DB.Where("id=?", customerParam.ID).Find(&customerManage).Error; err != nil {
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
			if err := tx.Updates(&customerManageUpdate).Error; err != nil {
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
	if err := data.DB.Where("customer_id=? and delete_state=0", customerId).Find(&members).Error; err != nil {
		log.Errorf("[updateCustomer] query customer members error, %v", err)
		return nil, err
	}
	return members, nil
}

func searchCustomerById(customerId int64) (*entity.CustomerManage, error) {
	var customer entity.CustomerManage
	if err := data.DB.Where("id=? and delete_state=0", customerId).Find(&customer).Error; err != nil {
		log.Errorf("[updateCustomer] query customer members error, %v", err)
		return nil, err
	}
	return &customer, nil
}

func searchCustomerByName(customerName string) ([]entity.CustomerManage, error) {
	var customerList []entity.CustomerManage
	if err := data.DB.Where("customer_name=? and delete_state=0", customerName).Find(&customerList).Error; err != nil {
		log.Errorf("[searchCustomerByName] query customer by name error, %v", err)
		return nil, err
	}
	return customerList, nil
}

func CreateCloudPlatformByCustomerId(db *gorm.DB, customerId int64, userId string) (*CreateCloudPlatform, error) {
	now := datetime.GetNow()
	cloudPlatformEntity := &entity.CloudPlatformManage{
		Name:         "云平台1",
		Type:         "operational",
		CustomerId:   customerId,
		CreateUserId: userId,
		CreateTime:   now,
		UpdateUserId: userId,
		UpdateTime:   now,
		DeleteState:  0,
	}
	regionEntity := &entity.RegionManage{
		Name:         "region1",
		Code:         "region1",
		Type:         "merge",
		CreateUserId: userId,
		CreateTime:   now,
		UpdateUserId: userId,
		UpdateTime:   now,
		DeleteState:  0,
	}
	azEntity := &entity.AzManage{
		Code:         "zone1",
		CreateUserId: userId,
		CreateTime:   now,
		UpdateUserId: userId,
		UpdateTime:   now,
		DeleteState:  0,
	}
	cellEntity := &entity.CellManage{
		Name:         "cell1",
		Type:         constant.CellTypeControl,
		CreateUserId: userId,
		CreateTime:   now,
		UpdateUserId: userId,
		UpdateTime:   now,
		DeleteState:  0,
	}
	if err := db.Create(&cloudPlatformEntity).Error; err != nil {
		return nil, err
	}
	regionEntity.CloudPlatformId = cloudPlatformEntity.Id
	if err := db.Create(&regionEntity).Error; err != nil {
		return nil, err
	}
	azEntity.RegionId = regionEntity.Id
	if err := db.Create(&azEntity).Error; err != nil {
		return nil, err
	}
	cellEntity.AzId = azEntity.Id
	if err := db.Create(&cellEntity).Error; err != nil {
		return nil, err
	}
	return &CreateCloudPlatform{CloudPlatformManage: cloudPlatformEntity, RegionManage: regionEntity, AzManage: azEntity, CellManage: cellEntity}, nil
}

func InnerCreateCustomer(customerParam CreateCustomerRequest, leaderId string, ldapUser *ldap.Entry, currentUserId string) (*InnerCreateCustomerResponse, error) {
	var customer *entity.CustomerManage
	var createCloudPlatform *CreateCloudPlatform
	var projectEntity *entity.ProjectManage
	var now = datetime.GetNow()
	var err error
	if err = data.DB.Transaction(func(tx *gorm.DB) error {
		customer, createCloudPlatform, err = createCustomer(tx, customerParam, leaderId, ldapUser, currentUserId)
		if err != nil {
			return err
		}
		//默认创建项目
		projectEntity = &entity.ProjectManage{
			Name:            "默认项目",
			CloudPlatformId: createCloudPlatform.CloudPlatformManage.Id,
			RegionId:        createCloudPlatform.RegionManage.Id,
			AzId:            createCloudPlatform.AzManage.Id,
			CellId:          createCloudPlatform.CellManage.Id,
			CustomerId:      createCloudPlatform.CloudPlatformManage.CustomerId,
			Type:            "create",
			Stage:           constant.ProjectStagePlanning,
			DeleteState:     0,
			CreateUserId:    currentUserId,
			CreateTime:      now,
			UpdateUserId:    currentUserId,
			UpdateTime:      now,
		}
		if err = tx.Create(&projectEntity).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &InnerCreateCustomerResponse{CustomerManage: customer, FrontUrl: fmt.Sprintf("%v/projectInfo?projectid=%v", os.Getenv(constant.FrontUrl), projectEntity.Id)}, nil
}

func InnerUpdateCustomer(customerParam UpdateCustomerRequest) error {
	if err := updateCustomer(customerParam, customerParam.LeaderId); err != nil {
		return err
	}
	return nil
}
