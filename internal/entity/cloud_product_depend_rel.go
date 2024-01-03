package entity

const CloudProductDependRelTable = "cloud_product_depend_rel"

type CloudProductDependRel struct {
	ProductId       int64 `gorm:"column:product_id" json:"productId"`              // 云产品id
	DependProductId int64 `gorm:"column:depend_product_id" json:"dependProductId"` // 依赖的云产品id
}

func (entity *CloudProductDependRel) TableName() string {
	return CloudProductDependRelTable
}
