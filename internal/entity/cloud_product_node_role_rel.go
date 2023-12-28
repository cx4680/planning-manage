package entity

const CloudProductNodeRoleRelTable = "cloud_product_node_role_rel"

type CloudProductNodeRoleRel struct {
	ProductId    int64 `gorm:"column:product_id" json:"productId"`        // 云产品id
	NodeRoleId   int64 `gorm:"column:node_role_id" json:"nodeRoleId"`     // 节点角色id
	NodeRoleType int   `gorm:"column:node_role_type" json:"nodeRoleType"` // 节点角色类型，1：管控资源节点角色，0：资源节点角色
}

func (entity *CloudProductNodeRoleRel) TableName() string {
	return CloudProductNodeRoleRelTable
}
