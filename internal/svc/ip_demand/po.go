package ip_demand

type Request struct {
	PlanId int64 `form:"planId"`
}

type IpDemandPlanningExportResponse struct {
	SegmentType     string `json:"segmentType" excel:"name:网段类型;index:0"`         // 网段类型
	Describe        string `json:"describe" excel:"name:描述;index:1"`              // 描述
	Vlan            string `json:"vlan" excel:"name:VLAN ID;index:2"`             // vlan id
	CNum            string `json:"cNum" excel:"name:C数量(C);index:3"`              // C数量
	Address         string `json:"address" excel:"name:地址段;index:4"`              // 地址段
	AddressPlanning string `json:"addressPlanning" excel:"name:IP地址规划建议;index:5"` // IP地址规划建议
}

type IpDemandBaselineDto struct {
	ID           int64  `form:"id"`
	VersionId    int64  `form:"versionId"`
	Vlan         string `form:"vlan"`
	Explain      string `form:"explain"`
	Description  string `form:"description"`
	IpSuggestion string `form:"ipSuggestion"`
	AssignNum    string `form:"assignNum"`
	Remark       string `form:"remark"`
	DeviceRoleId int64  `form:"deviceRoleId"`
}
