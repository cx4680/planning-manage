package entity

const SoftwareBomPlanningTable = "software_bom_planning"

type SoftwareBomPlanning struct {
	Id                 int64  `gorm:"column:id" json:"id"`                                   // 主键id
	PlanId             int64  `gorm:"column:plan_id" json:"planId"`                          // 方案id
	SoftwareBaselineId int64  `gorm:"column:software_baseline_id" json:"softwareBaselineId"` // 软件基线id
	ServiceYearBom     string `gorm:"column:service_year_bom" json:"serviceYearBom"`         // 维保年限bom
	PlatformBom        string `gorm:"column:platform_bom" json:"platformBom"`                // 平台规模授权bom
	SoftwareBaseBom    string `gorm:"column:software_base_bom" json:"softwareBaseBom"`       // 软件base bom
}

func (entity *SoftwareBomPlanning) TableName() string {
	return SoftwareBomPlanningTable
}
