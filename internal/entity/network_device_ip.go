package entity

const NetworkDeviceIpTable = "network_device_ip"

type NetworkDeviceIp struct {
	NetworkDeviceId int64  `gorm:"column:network_device_id" json:"networkDeviceId"` // 网络设备id
	NetworkType     int    `gorm:"column:network_type" json:"networkType"`          // 网络类型，1：管理网子网，2：管理网网关，3：业务网子网，4：业务网网关，5：存储前端网，6：存储前端网网关，7：业务外网子网，8：业务外网网关，9：BMC子网，10：BMC网关，11：业务外网ipv6子网，12：业务外网ipv6网关
	Vlan            string `gorm:"column:vlan" json:"vlan"`                         // vlan id
	Ip              string `gorm:"column:ip" json:"ip"`                             // IP地址
}

func (entity *NetworkDeviceIp) TableName() string {
	return NetworkDeviceIpTable
}
