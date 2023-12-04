package entity

import "time"

const CloudProductBaselineTable = "cloud_product_baseline"

type CloudProductBaseline struct {
	Id              int64     `gorm:"column:id" json:"id"`                            // 主键id
	VersionId       int64     `gorm:"column:version_id" json:"versionId"`             // 版本id
	ProductType     string    `gorm:"column:product_type" json:"productType"`         // 产品类型
	ProductName     string    `gorm:"column:product_name" json:"productName"`         // 产品名称
	ProductCode     string    `gorm:"column:product_code" json:"productCode"`         // 产品编码
	SellSpec        string    `gorm:"column:sell_specs" json:"sellSpecs"`             // 售卖规格
	AuthorizedUnit  string    `gorm:"column:authorized_unit" json:"authorizedUnit"`   // 授权单元
	WhetherRequired int       `gorm:"column:whether_required" json:"whetherRequired"` // 是否必选，0：否，1：是
	Instructions    string    `gorm:"column:instructions" json:"instructions"`        // 说明
	CreateTime      time.Time `gorm:"column:create_time" json:"createTime"`           // 创建时间
}

func (entity *CloudProductBaseline) TableName() string {
	return CloudProductBaselineTable
}
