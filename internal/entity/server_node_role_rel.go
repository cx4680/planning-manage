package entity

const ServerNodeRoleRelTable = "server_node_role_rel"

type ServerNodeRoleRel struct {
	ServerId   int64 `gorm:"column:server_id" json:"serverId"`      // 服务器id
	NodeRoleId int64 `gorm:"column:node_role_id" json:"nodeRoleId"` // 节点角色id
}

func (entity *ServerNodeRoleRel) TableName() string {
	return ServerNodeRoleRelTable
}
