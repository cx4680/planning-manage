package entity

const CapActualResBaselineTable = "cap_actual_res_baseline"

type CapActualResBaseline struct {
	Id                  int64  `gorm:"column:id" json:"id"`                                     // 主键id
	VersionId           int64  `gorm:"column:version_id" json:"versionId"`                      // 版本id
	ProductCode         string `gorm:"column:product_code" json:"productCode"`                  // 产品编码
	SellSpecs           string `gorm:"column:sell_specs" json:"sellSpecs"`                      // 售卖规格
	ValueAddedService   string `gorm:"column:value_added_service" json:"valueAddedService"`     // 增值服务
	SellUnit            string `gorm:"column:sell_unit" json:"sellUnit"`                        // 售卖单元
	ExpendRes           string `gorm:"column:expend_res" json:"expendRes"`                      // 消耗资源
	ExpendResCode       string `gorm:"column:expend_res_code" json:"expendResCode"`             // 消耗资源编码
	Features            string `gorm:"column:features" json:"features"`                         // 特性
	OccRatioNumerator   string `gorm:"column:occ_ratio_numerator" json:"occRatioNumerator"`     // 占用比例分子，为N的时候需要根据用户实际填写来做计算
	OccRatioDenominator string `gorm:"column:occ_ratio_denominator" json:"occRatioDenominator"` // 占用比例分母
	ActualConsume       string `gorm:"column:actual_consume" json:"actualConsume"`              // 实际消耗
	Remarks             string `gorm:"column:remarks" json:"remarks"`                           // 备注
}

func (entity *CapActualResBaseline) TableName() string {
	return CapActualResBaselineTable
}
