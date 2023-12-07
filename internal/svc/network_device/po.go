package network_device

type NetworkDeviceListExportResponse struct {
	NetworkDeviceRoleName string `gorm:"column:network_device_role_name" json:"networkDeviceRoleName" excel:"name:设备类型;index:0"`
	NetworkDeviceRole     string `gorm:"column:network_device_role" json:"networkDeviceRole" excel:"name:设备角色;index:1"`
	Brand                 string `gorm:"column:brand" json:"brand" excel:"name:厂商;index:2"`
	DeviceModel           string `gorm:"column:device_model" json:"deviceModel" excel:"name:机型;index:3"`
	ConfOverview          string `gorm:"column:conf_overview" json:"confOverview" excel:"name:规格参数;index:4"`
	Num                   int    `gorm:"column:num" json:"num" excel:"name:数量;index:5"`
}

type NetworkDeviceRoleIdNum struct {
	NetworkDeviceRoleId int64 `gorm:"column:network_device_role_id" json:"networkDeviceRoleId"`
	Num                 int   `gorm:"column:num" json:"num" excel:"数量"`
}

type BoxTotalResponse struct {
	Count int64 `json:"count"`
}

type DeviceRoleGroupNum struct {
	DeviceRoleId int64 `form:"deviceRoleId"`
	GroupNum     int   `form:"groupNum"`
}

type Request struct {
	PlanId                int64            `form:"planId"`
	Brand                 string           `form:"brand"`
	ApplicationDispersion string           `form:"applicationDispersion"`
	AwsServerNum          int              `form:"awsServerNum"`
	AwsBoxNum             int              `form:"awsBoxNum"`
	TotalBoxNum           int              `form:"totalBoxNum"`
	Ipv6                  string           `form:"ipv6"`
	NetworkModel          int              `form:"networkModel"`
	DeviceType            int              `form:"deviceType"`
	CloudPlatformType     string           `form:"cloudPlatformType"`
	BaselineVersion       string           `form:"baselineVersion"`
	Devices               []NetworkDevices `form:"devices"`
}

type NetworkDevices struct {
	PlanId                int64                `form:"planId"`
	NetworkDeviceRoleId   int64                `form:"networkDeviceRoleId"`
	NetworkDeviceRole     string               `form:"networkDeviceRole"`
	NetworkDeviceRoleName string               `form:"networkDeviceRoleName"`
	LogicalGrouping       string               `form:"logicalGrouping"`
	DeviceId              string               `form:"deviceId"`
	Brand                 string               `form:"brand"`
	DeviceModel           string               `form:"deviceModel"`
	ConfOverview          string               `form:"confOverview"`
	DeviceModels          []NetworkDeviceModel `form:"deviceModels"`
}

type NetworkDeviceModel struct {
	ConfOverview string `form:"configurationOverview"`
	DeviceModel  string `form:"deviceModel"`
}
