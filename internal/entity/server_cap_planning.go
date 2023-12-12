package entity

const ServerCapPlanningTable = "server_cap_planning"

type ServerCapPlanning struct {
	Id                 int64 `gorm:"column:id" json:"id"`                                   // 容量规划id
	PlanId             int64 `gorm:"column:plan_id" json:"planId"`                          // 方案id
	CapacityBaselineId int64 `gorm:"column:capacity_baseline_dd" json:"capacityBaselineId"` // 容量指标id
}

func (entity *ServerCapPlanning) TableName() string {
	return ServerCapPlanningTable
}
