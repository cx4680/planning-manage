package entity

import "time"

const VlanIdConfigTable = "vlan_id_config"

type VlanIdConfig struct {
	Id                 int64     `gorm:"column:id" json:"id"`                                    // 主键Id
	PlanId             int64     `gorm:"column:plan_id" json:"planId"`                           // 方案Id
	InBandMgtVlanId    string    `gorm:"column:in_band_mgt_vlan_id" json:"inBandMgtVlanId"`      // 带内管理Vlan ID
	LocalStorageVlanId string    `gorm:"column:local_storage_vlan_id" json:"localStorageVlanId"` // 本地存储网Vlan ID
	BizIntranetVlanId  string    `gorm:"column:biz_intranet_vlan_id" json:"bizIntranetVlanId"`   // 业务内网Vlan ID
	CreateUserId       string    `gorm:"column:create_user_id" json:"createUserId"`              // 创建人id
	CreateTime         time.Time `gorm:"column:create_time" json:"createTime"`                   // 创建时间
	UpdateUserId       string    `gorm:"column:update_user_id" json:"updateUserId"`              // 更新人id
	UpdateTime         time.Time `gorm:"column:update_time" json:"updateTime"`                   // 更新时间
}

func (entity *VlanIdConfig) TableName() string {
	return VlanIdConfigTable
}
