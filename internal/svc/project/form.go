package project

type Request struct {
	Id                int64
	UserId            string
	Name              string `form:"name"`
	CustomerName      string `form:"CustomerName"`
	Type              string `form:"type"`
	Stage             string `form:"stage"`
	CloudPlatformId   int64  `form:"cloudPlatformId"`
	CloudPlatformType string `form:"cloudPlatformType"`
	RegionId          int64  `form:"regionId"`
	AzId              int64  `form:"azId"`
	CellId            int64  `form:"cellId"`
	CustomerId        int64  `form:"customerId"`
	SortField         string `form:"sortField"`
	Sort              string `form:"sort"`
	Current           int    `json:"current"`
	PageSize          int    `json:"pageSize"`
}
