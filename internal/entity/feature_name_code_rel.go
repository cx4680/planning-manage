package entity

type FeatureNameCodeRel struct {
	Id          int64  `gorm:"column:id" json:"id"` // 主键id
	FeatureName string `gorm:"column:feature_name" json:"featureName"`
	FeatureCode string `gorm:"column:feature_code" json:"featureCode"`
	FeatureType string `gorm:"column:feature_type" json:"featureType"`
}

func (entity *FeatureNameCodeRel) TableName() string {
	return "feature_name_code_rel"
}
