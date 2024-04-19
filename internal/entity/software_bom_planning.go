package entity

const SoftwareBomPlanningTable = "software_bom_planning"

type SoftwareBomPlanning struct {
	Id                 int64   `gorm:"column:id" json:"id"`                                   // 主键id
	PlanId             int64   `gorm:"column:plan_id" json:"planId"`                          // 方案id
	SoftwareBaselineId int64   `gorm:"column:software_baseline_id" json:"softwareBaselineId"` // 软件基线id
	BomId              string  `gorm:"column:bom_id" json:"bomId"`                            // bom id
	Number             float64 `gorm:"column:number" json:"number"`                           // 数量
	CloudService       string  `gorm:"column:cloud_service" json:"cloudService"`              // 云服务
	ServiceCode        string  `gorm:"column:service_code" json:"serviceCode"`                // 服务编码
	SellSpecs          string  `gorm:"column:sell_specs" json:"sellSpecs"`                    // 售卖规格
	AuthorizedUnit     string  `gorm:"column:authorized_unit" json:"authorizedUnit"`          // 授权单元
	SellType           string  `gorm:"column:sell_type" json:"sellType"`                      // 售卖类型
	HardwareArch       string  `gorm:"column:hardware_arch" json:"hardwareArch"`              // 硬件架构
	ValueAddedService  string  `gorm:"column:value_added_service" json:"valueAddedService"`   // 增值服务
}

func (entity *SoftwareBomPlanning) TableName() string {
	return SoftwareBomPlanningTable
}
