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
