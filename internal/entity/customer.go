package entity

import "time"

const (
	UserManageTable        = "user_manage"
	RoleManageTable        = "role_manage"
	CustomerManageTable    = "customer_manage"
	PermissionsManageTable = "permissions_manage"
)

type UserManage struct {
	ID               string `json:"id" gorm:"primaryKey;type:varchar(50) NOT NULL;column:id;comment:用户id(ldap uid)"`
	Username         string `json:"username" gorm:"type:varchar(255);DEFAULT NULL;column:user_name;comment:用户名称"`
	EmployeeNumber   string `json:"employeeNumber" gorm:"type:varchar(255);DEFAULT NULL;column:employee_number;comment:员工编号"`
	TelephoneNumber  string `json:"telephoneNumber" gorm:"type:varchar(255);DEFAULT NULL;column:telephone_number;comment:电话"`
	Department       string `json:"department" gorm:"type:varchar(255);DEFAULT NULL;column:department;comment:所属部门"`
	OfficeName       string `json:"officeName" gorm:"type:varchar(255);DEFAULT NULL;column:office_name;"`
	DepartmentNumber string `json:"departmentNumber" gorm:"type:varchar(255);DEFAULT NULL;column:department_number;comment:部门编号"`
	Mail             string `json:"mail" gorm:"type:varchar(255);DEFAULT NULL;column:mail;comment:邮箱"`
	DeleteState      int    `json:"deleteState" gorm:"type:tinyint(1) unsigned DEFAULT 0;column:delete_state;comment:作废状态：0，正常；1，作废"`
}

func (m *UserManage) TableName() string {
	return UserManageTable
}

type RoleManage struct {
	UserId string `json:"userId" gorm:"type:varchar(50) NOT NULL;column:user_id;comment:客户id"`
	Role   string `json:"role" gorm:"type:varchar(50) NOT NULL;column:role;comment:用户角色：admin-管理员，normal-普通用户"`
}

func (m *RoleManage) TableName() string {
	return RoleManageTable
}

type CustomerManage struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement;type:bigint(20) NOT NULL;column:id;comment:客户id"`
	CustomerName string    `json:"customerName" gorm:"type:varchar(50);DEFAULT NULL;column:customer_name;comment:客户名称"`
	LeaderId     string    `json:"leaderId" gorm:"type:varchar(50);DEFAULT NULL;column:leader_id;comment:客户接口人id"`
	LeaderName   string    `json:"leaderName" gorm:"type:varchar(255);DEFAULT NULL;column:leader_name;comment:客户接口人名称"`
	CreateUserId string    `json:"createUserId" gorm:"type:varchar(50);DEFAULT NULL;column:create_user_id;comment:创建人"`
	CreateTime   time.Time `json:"createTime" gorm:"type:datetime;autoCreateTime:milli;column:create_time;comment:创建时间"`
	UpdateUserId string    `json:"updateUserId" gorm:"type:varchar(50);DEFAULT NULL;column:update_user_id;comment:更新人"`
	UpdateTime   time.Time `json:"updateTime" gorm:"type:datetime;autoUpdateTime:milli;column:update_time;comment:更新时间"`
	DeleteState  int       `json:"deleteState" gorm:"type:tinyint(1) unsigned DEFAULT 0;column:delete_state;comment:作废状态：0，正常；1，作废"`
}

func (m *CustomerManage) TableName() string {
	return CustomerManageTable
}

type PermissionsManage struct {
	UserId      string `json:"userId" gorm:"type:varchar(50) NOT NULL;column:user_id;comment:用户id(ldap uid)"`
	UserName    string `json:"userName" gorm:"type:varchar(255) NOT NULL;column:user_name;comment:用户名称"`
	CustomerId  int64  `json:"customerId" gorm:"type:bigint(20) NOT NULL;column:customer_id;comment:客户id"`
	DeleteState int    `json:"deleteState" gorm:"type:tinyint(1) unsigned DEFAULT 0;column:delete_state;comment:作废状态：0，正常；1，作废"`
}

func (m *PermissionsManage) TableName() string {
	return PermissionsManageTable
}
