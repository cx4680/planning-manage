package entity

const NodeRoleMixedDeployTable = "node_role_mixed_deploy"

type NodeRoleMixedDeploy struct {
	NodeRoleId      int64 `gorm:"column:node_role_id" json:"nodeRoleId"`            // 节点角色id
	MixedNodeRoleId int64 `gorm:"column:mixed_node_role_id" json:"mixedNodeRoleId"` // 混部的节点角色id
}

func (entity *NodeRoleMixedDeploy) TableName() string {
	return NodeRoleMixedDeployTable
}
