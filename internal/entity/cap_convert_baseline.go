package entity

const CapConvertBaselineTable = "cap_convert_baseline"

type CapConvertBaseline struct {
	Id               int64  `gorm:"column:id" json:"id"`                               // 主键id
	VersionId        int64  `gorm:"column:version_id" json:"versionId"`                // 版本id
	ProductName      string `gorm:"column:product_name" json:"productName"`            // 产品名称
	ProductCode      string `gorm:"column:product_code" json:"productCode"`            // 产品编码
	SellSpecs        string `gorm:"column:sell_specs" json:"sellSpecs"`                // 售卖规格
	CapPlanningInput string `gorm:"column:cap_planning_input" json:"capPlanningInput"` // 容量规划输入
	Unit             string `gorm:"column:unit" json:"unit"`                           // 单位
	FeaturesMode     string `gorm:"column:features_mode" json:"featuresMode"`          // 特性模式
	Features         string `gorm:"column:features" json:"features"`                   // 特性
	Description      string `gorm:"column:description" json:"description"`             // 说明
}

func (entity *CapConvertBaseline) TableName() string {
	return CapConvertBaselineTable
}
