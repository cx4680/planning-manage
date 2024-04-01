package entity

const NodeRoleBaselineTable = "node_role_baseline"

type NodeRoleBaseline struct {
	Id                       int64  `gorm:"column:id" json:"id"`                                                // 主键id
	VersionId                int64  `gorm:"column:version_id" json:"versionId"`                                 // 版本id
	NodeRoleCode             string `gorm:"column:node_role_code" json:"nodeRoleCode"`                          // 节点角色编码
	NodeRoleName             string `gorm:"column:node_role_name" json:"nodeRoleName"`                          // 节点角色名称
	MinimumNum               int    `gorm:"column:minimum_num" json:"minimumNum"`                               // 单独部署最小数量
	DeployMethod             string `gorm:"column:deploy_method" json:"deployMethod"`                           // 部署方式
	SupportDPDK              int    `gorm:"column:support_dpdk" json:"supportDPDK"`                             // 是否支持DPDK, 0:否，1：是
	Classify                 string `gorm:"column:classify" json:"classify"`                                    // 分类
	Annotation               string `gorm:"column:annotation" json:"annotation"`                                // 节点说明
	BusinessType             string `gorm:"column:business_type" json:"businessType"`                           // 业务类型
	SupportMultiResourcePool int    `gorm:"column:support_multi_resource_pool" json:"supportMultiResourcePool"` // 是否支持多资源池, 0:否，1：是
}

func (entity *NodeRoleBaseline) TableName() string {
	return NodeRoleBaselineTable
}
