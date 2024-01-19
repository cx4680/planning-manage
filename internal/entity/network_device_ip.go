package entity

const NetworkDeviceIpTable = "network_device_ip"

type NetworkDeviceIp struct {
	Id                         int64  `gorm:"column:id" json:"id"`                                                    // 主键id
	PlanId                     int64  `gorm:"column:plan_id" json:"planId"`                                           // 方案id
	LogicalGrouping            string `gorm:"column:logical_grouping" json:"logicalGrouping"`                         // 网络设备逻辑分组
	PxeSubnet                  string `gorm:"column:pxe_subnet" json:"pxeSubnet"`                                     // PXE子网
	PxeSubnetRange             string `gorm:"column:pxe_subnet_range" json:"pxeSubnetRange"`                          // PXE子网范围
	PxeNetworkGateway          string `gorm:"column:pxe_network_gateway" json:"pxeNetworkGateway"`                    // PXE网网关
	ManageSubnet               string `gorm:"column:manage_subnet" json:"manageSubnet"`                               // 管理网子网
	ManageNetworkGateway       string `gorm:"column:manage_network_gateway" json:"manageNetworkGateway"`              // 管理网网关
	ManageIpv6Subnet           string `gorm:"column:manage_ipv6_subnet" json:"manageIpv6Subnet"`                      // 管理网IPV6子网
	ManageIpv6NetworkGateway   string `gorm:"column:manage_ipv6_network_gateway" json:"manageIpv6NetworkGateway"`     // 管理网IPV6网关
	BizSubnet                  string `gorm:"column:biz_subnet" json:"bizSubnet"`                                     // 业务网子网
	BizNetworkGateway          string `gorm:"column:biz_network_gateway" json:"bizNetworkGateway"`                    // 业务网网关
	StorageFrontNetwork        string `gorm:"column:storage_front_network" json:"storageFrontNetwork"`                // 存储前端网
	StorageFrontNetworkGateway string `gorm:"column:storage_front_network_gateway" json:"storageFrontNetworkGateway"` // 存储前端网网关
}

func (entity *NetworkDeviceIp) TableName() string {
	return NetworkDeviceIpTable
}
