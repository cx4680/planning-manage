package entity

const ServerIpTable = "server_ip"

type ServerIp struct {
	Id                int64  `gorm:"column:id" json:"id"`                                 // 主键Id
	PlanId            int64  `gorm:"column:plan_id" json:"planId"`                        // 计划Id
	Sn                string `gorm:"column:sn" json:"sn"`                                 // SN
	HostName          string `gorm:"column:host_name" json:"hostName"`                    // 主机名
	ManageNetworkIp   string `gorm:"column:manage_network_ip" json:"manageNetworkIp"`     // 管理网IP
	ManageNetworkIpv6 string `gorm:"column:manage_network_ipv6" json:"manageNetworkIpv6"` // 管理网ipv6
	BizIntranetIp     string `gorm:"column:biz_intranet_ip" json:"bizIntranetIp"`         // 业务内网IP
}

func (entity *ServerIp) TableName() string {
	return ServerIpTable
}
