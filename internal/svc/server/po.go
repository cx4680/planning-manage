package server

type Request struct {
	Id               int64
	UserId           string
	PlanId           int64            `form:"planId"`
	NetworkInterface string           `form:"networkInterface"`
	CpuType          string           `form:"cpuType"`
	ServerList       []*RequestServer `form:"serverList"`
}

type RequestServer struct {
	NodeRoleId       int64 `form:"nodeRoleId"`
	MixedNodeRoleId  int64 `form:"mixedNodeRoleId"`
	ServerBaselineId int64 `form:"serverBaselineId"`
	Number           int   `form:"number"`
	OpenDpdk         int   `form:"openDpdk"`
}

type ResponseDownloadServer struct {
	NodeRole   string `json:"nodeRole" excel:"nodeRole:角色;index:0"`
	ServerType string `json:"serverType" excel:"serverType:设备类型;index:1"`
	BomCode    string `json:"bomCode" excel:"bomCode:机型;index:2"`
	Spec       string `json:"spec" excel:"spec:规格;index:3"`
	Number     string `json:"number" excel:"number:数量;index:4"`
}
