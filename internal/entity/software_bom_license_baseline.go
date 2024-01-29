package entity

const SoftwareBomLicenseBaselineTable = "software_bom_license_baseline"

type SoftwareBomLicenseBaseline struct {
	Id             int64  `gorm:"column:id" json:"id"`                          // 主键id
	VersionId      int64  `gorm:"column:version_id" json:"versionId"`           // 版本id
	CloudService   string `gorm:"column:cloud_service" json:"cloudService"`     // 云服务
	ServiceCode    string `gorm:"column:service_code" json:"serviceCode"`       // 服务编码
	SellSpecs      string `gorm:"column:sell_specs" json:"sellSpecs"`           // 售卖规格
	AuthorizedUnit string `gorm:"column:authorized_unit" json:"authorizedUnit"` // 授权单元
	SellType       string `gorm:"column:sell_type" json:"sellType"`             // 售卖类型
	HardwareArch   string `gorm:"column:hardware_arch" json:"hardwareArch"`     // 硬件架构
	BomId          string `gorm:"column:bom_id" json:"bomId"`                   // bom id
	CalcMethod     string `gorm:"column:calc_method" json:"calcMethod"`         // 计算方式
}

func (entity *SoftwareBomLicenseBaseline) TableName() string {
	return SoftwareBomLicenseBaselineTable
}
