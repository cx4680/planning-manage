package server

type Request struct {
	Id             int64
	UserId         string
	PlanId         int64            `form:"planId"`
	NetworkVersion string           `form:"networkVersion"`
	CpuType        string           `form:"cpuType"`
	serverList     []*RequestServer `form:"serverList"`
}

type RequestServer struct {
	NodeRoleId       int64 `form:"nodeRoleId"`
	MixedNodeRoleId  int64 `form:"mixedNodeRoleId"`
	ServerBaselineId int64 `form:"serverBaselineId"`
	Number           int   `form:"number"`
	OpenDpdk         int   `form:"openDpdk"`
}

type ResponseCapacity struct {
	Id              int64  `json:"id"`
	ProductId       int64  `json:"productId"`
	ProductName     string `json:"productName"`
	CapacitySpecs   string `json:"capacitySpecs"`
	SalesSpecs      string `json:"salesSpecs"`
	OverbookingRate string `json:"overbookingRate"`
	Number          string `json:"number"`
	Unit            string `json:"unit"`
}

type ResponseDownloadServer struct {
	NodeRole   string `form:"nodeRole"`
	ServerType string `form:"serverType"`
	BomCode    string `form:"bomCode"`
	Spec       string `form:"Spec"`
	Number     int    `form:"number"`
}
