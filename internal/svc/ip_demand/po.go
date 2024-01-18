package ip_demand

type Request struct {
	PlanId int64 `form:"planId"`
}

type IpDemandPlanningExportResponse struct {
	LogicalGrouping string `json:"logicalGrouping" excel:"name:网络设备逻辑分组;index:0"` // 逻辑分组
	SegmentType     string `json:"segmentType" excel:"name:网段类型;index:1"`         // 网段类型
	NetworkType     string `json:"networkType" excel:"name:网络类型;index:2"`         // 网络类型，ipv4或者ipv6
	Describe        string `json:"describe" excel:"name:描述;index:3"`              // 描述
	Vlan            string `json:"vlan" excel:"name:VLAN ID;index:4"`             // vlan id
	CNum            string `json:"cNum" excel:"name:C数量(C);index:5"`              // C数量
	Address         string `json:"address" excel:"name:地址段;index:6"`              // 地址段
	AddressPlanning string `json:"addressPlanning" excel:"name:IP地址规划建议;index:7"` // IP地址规划建议
}

type IpDemandBaselineDto struct {
	ID           int64  `form:"id"`
	VersionId    int64  `form:"versionId"`
	Vlan         string `form:"vlan"`
	Explain      string `form:"explain"`
	NetworkType  int    `form:"networkType"`
	Description  string `form:"description"`
	IpSuggestion string `form:"ipSuggestion"`
	AssignNum    string `form:"assignNum"`
	Remark       string `form:"remark"`
	DeviceRoleId int64  `form:"deviceRoleId"`
}
