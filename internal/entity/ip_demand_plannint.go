package entity

import "time"

const IPDemandPlanningTable = "ip_demand_planning"

type IPDemandPlanning struct {
	Id              int64     `gorm:"column:id" json:"id"`                            // 主键id
	PlanId          int64     `gorm:"column:plan_id" json:"planId"`                   // 方案ID
	LogicalGrouping string    `gorm:"column:logical_grouping" json:"logicalGrouping"` // 逻辑分组
	SegmentType     string    `gorm:"column:segment_type" json:"segmentType"`         // 网段类型
	NetworkType     int       `gorm:"column:network_type" json:"networkType"`         // 网络类型，0：ipv4，1：ipv6
	Vlan            string    `gorm:"column:vlan" json:"vlan"`                        // vlan id
	Cnum            string    `gorm:"column:c_num" json:"cNum"`                       // C数量
	Address         string    `gorm:"column:address" json:"address"`                  // 地址段
	Describe        string    `gorm:"column:describe" json:"describe"`                // 描述
	AddressPlanning string    `gorm:"column:address_planning" json:"addressPlanning"` // IP地址规划建议
	CreateTime      time.Time `gorm:"column:create_time" json:"createTime"`           // 创建时间
	UpdateTime      time.Time `gorm:"column:update_time" json:"updateTime"`           // 更新时间
}

func (entity *IPDemandPlanning) TableName() string {
	return IPDemandPlanningTable
}
