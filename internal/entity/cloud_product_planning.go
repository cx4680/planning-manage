package entity

import "time"

const CloudProductPlanningTable = "cloud_product_planning"

type CloudProductPlanning struct {
	Id          int64     `gorm:"column:id" json:"id"`                    // 主键id
	PlanId      int64     `gorm:"column:plan_id" json:"planId"`           // 方案id
	ProductId   int64     `gorm:"column:product_id" json:"productId"`     // 产品id
	SellSpec    string    `gorm:"column:sell_spec" json:"sellSpec"`       // 售卖规格
	ServiceYear int       `gorm:"column:service_year" json:"serviceYear"` // 维保年限
	CreateTime  time.Time `gorm:"column:create_time" json:"createTime"`   // 创建时间
	UpdateTime  time.Time `gorm:"column:update_time" json:"updateTime"`   // 更新时间
}

func (entity *CloudProductPlanning) TableName() string {
	return CloudProductPlanningTable
}
