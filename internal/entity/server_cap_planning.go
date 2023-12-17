package entity

const ServerCapPlanningTable = "server_cap_planning"

type ServerCapPlanning struct {
	Id                 int64  `gorm:"column:id" json:"id"`                                   // 容量规划id
	PlanId             int64  `gorm:"column:plan_id" json:"planId"`                          // 方案id
	NodeRoleId         int64  `gorm:"column:node_role_id" json:"nodeRoleId"`                 // 节点角色id
	CapacityBaselineId int64  `gorm:"column:capacity_baseline_id" json:"capacityBaselineId"` // 容量指标id
	Number             int    `gorm:"column:number" json:"number"`                           // 数量
	FeatureNumber      int    `gorm:"column:feature_number" json:"featureNumber"`            // 特性数量
	CapacityNumber     string `gorm:"column:capacity_number" json:"capacityNumber"`          // 容量总消耗
	ExpendResCode      string `gorm:"column:expend_res_code" json:"expendResCode"`           // 消耗资源编码
}

func (entity *ServerCapPlanning) TableName() string {
	return ServerCapPlanningTable
}
