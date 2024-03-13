package entity

import "time"

const (
	CloudPlatformTable = "cloud_platform_manage"
	RegionManageTable  = "region_manage"
	AzManageTable      = "az_manage"
	CellManageTable    = "cell_manage"
)

type CloudPlatformManage struct {
	Id           int64     `gorm:"column:id" json:"id"`                       //云平台id
	Name         string    `gorm:"column:name" json:"name"`                   //云平台名称
	Type         string    `gorm:"column:type" json:"type"`                   //云平台类型（运营云、交付云）
	CustomerId   int64     `gorm:"column:customer_id" json:"customerId"`      //客户id
	CreateUserId string    `gorm:"column:create_user_id" json:"createUserId"` //创建人id
	CreateTime   time.Time `gorm:"column:create_time" json:"createTime"`      //创建时间
	UpdateUserId string    `gorm:"column:update_user_id" json:"updateUserId"` //更新人id
	UpdateTime   time.Time `gorm:"column:update_time" json:"updateTime"`      //更新时间
	DeleteState  int       `gorm:"column:delete_state" json:"-"`              //作废状态：1，作废；0，正常
}

func (entity *CloudPlatformManage) TableName() string {
	return CloudPlatformTable
}

type RegionManage struct {
	Id              int64     `gorm:"column:id" json:"id"`                             //regionId
	Code            string    `gorm:"column:code" json:"code"`                         //region编码
	Name            string    `gorm:"column:name" json:"name"`                         //region名称
	Type            string    `gorm:"column:type" json:"type"`                         //region类型
	CloudPlatformId int64     `gorm:"column:cloud_platform_id" json:"cloudPlatformId"` //云平台id
	CreateUserId    string    `gorm:"column:create_user_id" json:"createUserId"`       //创建人id
	CreateTime      time.Time `gorm:"column:create_time" json:"createTime"`            //创建时间
	UpdateUserId    string    `gorm:"column:update_user_id" json:"updateUserId"`       //更新人id
	UpdateTime      time.Time `gorm:"column:update_time" json:"updateTime"`            //更新时间
	DeleteState     int       `gorm:"column:delete_state" json:"-"`                    //作废状态：1，作废；0，正常
}

func (entity *RegionManage) TableName() string {
	return RegionManageTable
}

type AzManage struct {
	Id           int64     `gorm:"column:id" json:"id"`                       //azId
	Code         string    `gorm:"column:code" json:"code"`                   //az编码
	RegionId     int64     `gorm:"column:region_id" json:"regionId"`          //regionId
	CreateUserId string    `gorm:"column:create_user_id" json:"createUserId"` //创建人id
	CreateTime   time.Time `gorm:"column:create_time" json:"createTime"`      //创建时间
	UpdateUserId string    `gorm:"column:update_user_id" json:"updateUserId"` //更新人id
	UpdateTime   time.Time `gorm:"column:update_time" json:"updateTime"`      //更新时间
	DeleteState  int       `gorm:"column:delete_state" json:"-"`              //作废状态：1，作废；0，正常
}

func (entity *AzManage) TableName() string {
	return AzManageTable
}

type CellManage struct {
	Id           int64     `gorm:"column:id" json:"id"`                       //cell Id
	Name         string    `gorm:"column:name" json:"name"`                   //cell名称
	AzId         int64     `gorm:"column:az_id" json:"azId"`                  //azId
	Type         string    `gorm:"column:type" json:"type"`                   //cell类型
	CreateUserId string    `gorm:"column:create_user_id" json:"createUserId"` //创建人id
	CreateTime   time.Time `gorm:"column:create_time" json:"createTime"`      //创建时间
	UpdateUserId string    `gorm:"column:update_user_id" json:"updateUserId"` //更新人id
	UpdateTime   time.Time `gorm:"column:update_time" json:"updateTime"`      //更新时间
	DeleteState  int       `gorm:"column:delete_state" json:"-"`              //作废状态：1，作废；0，正常
}

func (entity *CellManage) TableName() string {
	return CellManageTable
}
