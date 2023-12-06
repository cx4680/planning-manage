package ip_demand

type IpDemandPlanningExportResponse struct {
	SegmentType     string `gorm:"column:segment_type" json:"segmentType" excel:"name:网段类型;index:0"`             // 网段类型
	Describe        string `gorm:"column:describe" json:"describe" excel:"name:描述;index:1"`                      // 描述
	Vlan            string `gorm:"column:vlan" json:"vlan" excel:"name:VLAN ID;index:2"`                         // vlan id
	Cnum            string `gorm:"column:c_num" json:"cNum" excel:"name:C数量(C);index:3"`                         // C数量
	Address         string `gorm:"column:address" json:"address" excel:"name:地址段;index:4"`                       // 地址段
	AddressPlanning string `gorm:"column:address_planning" json:"addressPlanning" excel:"name:IP地址规划建议;index:5"` // IP地址规划建议
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
