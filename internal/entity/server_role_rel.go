package entity

const ServerRoleRelTable = "server_role_rel"

type ServerRoleRel struct {
	ServerId   int64 `gorm:"column:server_id" json:"serverId"`      // 服务器id
	NodeRoleId int64 `gorm:"column:node_role_id" json:"nodeRoleId"` // 节点角色id
}

func (entity *ServerRoleRel) TableName() string {
	return ServerRoleRelTable
}
