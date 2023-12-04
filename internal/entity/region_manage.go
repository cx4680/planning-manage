package entity

import "time"

const RegionManageTable = "region_manage"

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
