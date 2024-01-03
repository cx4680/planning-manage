package entity

import "time"

const CloudPlatformTable = "cloud_platform_manage"

type CloudPlatformManage struct {
	Id           int64           `gorm:"column:id" json:"id"`                       //云平台id
	Name         string          `gorm:"column:name" json:"name"`                   //云平台名称
	Type         string          `gorm:"column:type" json:"type"`                   //云平台类型（运营云、交付云）
	CustomerId   int64           `gorm:"column:customer_id" json:"customerId"`      //客户id
	CreateUserId string          `gorm:"column:create_user_id" json:"createUserId"` //创建人id
	CreateTime   time.Time       `gorm:"column:create_time" json:"createTime"`      //创建时间
	UpdateUserId string          `gorm:"column:update_user_id" json:"updateUserId"` //更新人id
	UpdateTime   time.Time       `gorm:"column:update_time" json:"updateTime"`      //更新时间
	DeleteState  int             `gorm:"column:delete_state" json:"-"`              //作废状态：1，作废；0，正常
	RegionList   []*RegionManage `gorm:"-" json:"regionList"`
	LeaderId     string          `gorm:"-" json:"leaderId"`
	LeaderName   string          `gorm:"-" json:"leaderName" `
}

func (entity *CloudPlatformManage) TableName() string {
	return CloudPlatformTable
}
