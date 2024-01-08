package entity

const ServerIpTable = "server_ip"

type ServerIp struct {
	ServerId    int64  `gorm:"column:server_id" json:"serverId"`       // 服务器Id
	NetworkType int    `gorm:"column:network_type" json:"networkType"` // 网络类型，1：管理网IP，2：业务内网IP，2：业务外网IP，3：存储前端网IP，4：BMC网IP
	Vlan        string `gorm:"column:vlan" json:"vlan"`                // vlan id
	Ip          string `gorm:"column:ip" json:"ip"`                    // IP地址
}

func (entity *ServerIp) TableName() string {
	return ServerIpTable
}
