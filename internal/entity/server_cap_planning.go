package entity

const ServerCapPlanningTable = "server_cap_planning"

type ServerCapPlanning struct {
	PlanId             int64 `gorm:"column:plan_id" json:"planId"`                          // 方案id
	CapacityBaselineId int64 `gorm:"column:capacity_baseline_id" json:"capacityBaselineId"` // 容量指标id
	Number             int   `gorm:"column:number" json:"number"`                           // 数量
}

func (entity *ServerCapPlanning) TableName() string {
	return ServerCapPlanningTable
}
