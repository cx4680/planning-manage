package entity

import "time"

const (
	NetworkDevicePlanningTable = "network_device_planning"
	NetworkDeviceListTable     = "network_device_list"
)

type NetworkDevicePlanning struct {
	Id                    int64     `gorm:"column:id" json:"id"`                                        // 主键
	PlanId                int64     `gorm:"column:plan_id" json:"planId"`                               // 方案ID
	Brand                 string    `gorm:"column:brand" json:"brand"`                                  // 厂商
	ApplicationDispersion string    `gorm:"column:application_dispersion" json:"applicationDispersion"` // 应用分散度: 1-分散在不同服务器
	AwsServerNum          int       `gorm:"column:aws_server_num" json:"awsServerNum"`                  // AWS下连服务器数44/45
	AwsBoxNum             int       `gorm:"column:aws_box_num" json:"awsBoxNum"`                        // 每组AWS几个机柜4/3
	TotalBoxNum           int       `gorm:"column:total_box_num" json:"totalBoxNum"`                    // 机柜估算数量
	CreateTime            time.Time `gorm:"column:create_time" json:"createTime"`                       // 创建时间
	UpdateTime            time.Time `gorm:"column:update_time" json:"updateTime"`                       // 更新时间
	Ipv6                  string    `gorm:"column:ipv6" json:"ipv6"`                                    // 是否为ipv4/ipv6双栈交付 0：ipv4交付 1:ipv4/ipv6双栈交付
	NetworkModel          int       `gorm:"column:network_model" json:"networkModel"`                   // 组网模型: 1-三网合一  2-两网分离  3-三网分离
	DeviceType            int       `gorm:"column:device_type" json:"deviceType"`                       // 设备类型，0：信创，1：商用
}

func (entity *NetworkDevicePlanning) TableName() string {
	return NetworkDevicePlanningTable
}

type NetworkDeviceList struct {
	Id                    int64     `gorm:"column:id" json:"id"`                                          // 主键
	PlanId                int64     `gorm:"column:plan_id" json:"planId"`                                 // 方案ID
	NetworkDeviceRoleId   int64     `gorm:"column:network_device_role_id" json:"networkDeviceRoleId"`     // 设备类型->网络设备角色ID
	NetworkDeviceRoleName string    `gorm:"column:network_device_role_name" json:"networkDeviceRoleName"` // 设备类型->设备角色名称
	NetworkDeviceRole     string    `gorm:"column:network_device_role" json:"networkDeviceRole"`          // 设备类型->网络设备角色编码
	LogicalGrouping       string    `gorm:"column:logical_grouping" json:"logicalGrouping"`               // 逻辑分组
	DeviceId              string    `gorm:"column:device_id" json:"deviceId"`                             // 设备ID
	Brand                 string    `gorm:"column:brand" json:"brand"`                                    // 厂商
	ConfOverview          string    `gorm:"column:conf_overview" json:"confOverview"`                     // 配置概述
	DeviceModel           string    `gorm:"column:device_model" json:"deviceModel"`                       // 设备型号
	BomId                 string    `gorm:"column:bom_id" json:"bomId"`                                   // bom id
	CreateTime            time.Time `gorm:"column:create_time" json:"createTime"`                         // 创建时间
	UpdateTime            time.Time `gorm:"column:update_time" json:"updateTime"`                         // 更新时间
	DeleteState           int       `gorm:"column:delete_state" json:"deleteState"`                       // 删除状态0：未删除；1：已删除
}

func (entity *NetworkDeviceList) TableName() string {
	return NetworkDeviceListTable
}
