package cloud_product

type CloudProductPlanningRequest struct {
	PlanId int64 `json:"planId"`
	//VersionId int64 `json:"versionId"`
	ServiceYear int `json:"serviceYear"`
	ProductList []struct {
		ProductId int64  `json:"productId"`
		SellSpec  string "sellSpec"
	} `json:"productList"`
}

type CloudProductBaselineResponse struct {
	ID                int64    `json:"id"`
	VersionId         int64    `json:"versionId"`
	ProductType       string   `json:"productType"`
	ProductName       string   `json:"productName"`
	ProductCode       string   `json:"productCode"`
	SellSpecs         []string `json:"sellSpecs"`
	AuthorizedUnit    string   `json:"authorizedUnit"`
	WhetherRequired   int      `json:"whetherRequired"`
	Instructions      string   `json:"instructions"`
	DependProductId   int64    `json:"dependProductId"`
	DependProductName string   `json:"dependProductName"`
}

type CloudProductBaselineDependResponse struct {
	ID                int64  `gorm:"column:id" json:"id"`
	DependId          int64  `gorm:"column:dependId" json:"dependId"`
	DependProductName string `gorm:"column:dependProductName" json:"dependProductName"`
	DependProductCode string `gorm:"column:dependProductCode" json:"dependProductCode"`
}

type CloudProductPlanningExportResponse struct {
	ProductType  string `gorm:"column:product_type" json:"productType"`
	ProductName  string `gorm:"column:product_name" json:"productName"`
	Instructions string `gorm:"column:instructions" json:"instructions"`
	SellSpec     string `gorm:"column:sell_spec" json:"sellSpec"`
}
