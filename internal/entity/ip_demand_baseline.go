package entity

const IPDemandBaselineTable = "ip_demand_baseline"

type IPDemandBaseline struct {
	Id           int64  `gorm:"column:id" json:"id"`                      // 主键id
	VersionId    int64  `gorm:"column:version_id" json:"versionId"`       // 版本id
	Vlan         string `gorm:"column:vlan" json:"vlan"`                  // vlan id
	Explain      string `gorm:"column:explain" json:"explain"`            // 说明
	Description  string `gorm:"column:description" json:"description"`    // 描述
	IPSuggestion string `gorm:"column:ip_suggestion" json:"IPSuggestion"` // IP地址规划建议
	AssignNum    string `gorm:"column:assign_num" json:"assignNum"`       // 分配数量
	Remark       string `gorm:"column:remark" json:"remark"`              // 备注
}

func (entity *IPDemandBaseline) TableName() string {
	return IPDemandBaselineTable
}
