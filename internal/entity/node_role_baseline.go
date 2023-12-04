package entity

const NodeRoleBaselineTable = "node_role_baseline"

type NodeRoleBaseline struct {
	Id           int64  `gorm:"column:id" json:"id"`                       // 主键id
	VersionId    int64  `gorm:"column:version_id" json:"versionId"`        // 版本id
	NodeRoleCode string `gorm:"column:node_role_code" json:"nodeRoleCode"` // 节点角色编码
	RoleName     string `gorm:"column:role_name" json:"roleName"`          // 角色
	MinimumNum   int    `gorm:"column:minimum_num" json:"minimumNum"`      // 单独部署最小数量
	Annotation   string `gorm:"column:annotation" json:"annotation"`       // 节点说明
	BusinessType string `gorm:"column:business_type" json:"businessType"`  // 业务类型
}

func (entity *NodeRoleBaseline) TableName() string {
	return NodeRoleBaselineTable
}
