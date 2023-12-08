package network_device

type NetworkDeviceListExportResponse struct {
	NetworkDeviceRoleName string `gorm:"column:network_device_role_name" json:"networkDeviceRoleName" excel:"name:设备类型;index:0"`
	NetworkDeviceRole     string `gorm:"column:network_device_role" json:"networkDeviceRole" excel:"name:设备角色;index:1"`
	Brand                 string `gorm:"column:brand" json:"brand" excel:"name:厂商;index:2"`
	DeviceModel           string `gorm:"column:device_model" json:"deviceModel" excel:"name:机型;index:3"`
	ConfOverview          string `gorm:"column:conf_overview" json:"confOverview" excel:"name:规格参数;index:4"`
	Num                   string `gorm:"column:num" json:"num" excel:"name:数量;index:5"`
}

type NetworkDeviceRoleIdNum struct {
	NetworkDeviceRoleId int64 `gorm:"column:network_device_role_id" json:"networkDeviceRoleId"`
	Num                 int   `gorm:"column:num" json:"num"`
}

type BoxTotalResponse struct {
	Count int64 `json:"count"`
}

type DeviceRoleGroupNum struct {
	DeviceRoleId int64 `gorm:"column:network_device_role_id" form:"deviceRoleId"`
	GroupNum     int   `gorm:"column:groupNum" form:"groupNum"`
}

type Request struct {
	PlanId                int64            `json:"planId" form:"planId"`
	Brand                 string           `json:"brand" form:"brand"`
	ApplicationDispersion string           `json:"applicationDispersion" form:"applicationDispersion"`
	AwsServerNum          int              `json:"awsServerNum" form:"awsServerNum"`
	AwsBoxNum             int              `json:"awsBoxNum" form:"awsBoxNum"`
	TotalBoxNum           int              `json:"totalBoxNum" form:"totalBoxNum"`
	Ipv6                  string           `json:"ipv6" form:"ipv6"`
	NetworkModel          int              `json:"networkModel" form:"networkModel"`
	DeviceType            int              `json:"deviceType" form:"deviceType"`
	CloudPlatformType     string           `json:"cloudPlatformType" form:"cloudPlatformType"`
	BaselineVersion       string           `json:"baselineVersion" form:"baselineVersion"`
	Devices               []NetworkDevices `json:"devices" form:"devices"`
}

type NetworkDevices struct {
	PlanId                int64                `json:"devices" form:"planId"`
	NetworkDeviceRoleId   int64                `json:"networkDeviceRoleId" form:"networkDeviceRoleId"`
	NetworkDeviceRole     string               `json:"networkDeviceRole" form:"networkDeviceRole"`
	NetworkDeviceRoleName string               `json:"networkDeviceRoleName" form:"networkDeviceRoleName"`
	LogicalGrouping       string               `json:"logicalGrouping" form:"logicalGrouping"`
	DeviceId              string               `json:"deviceId" form:"deviceId"`
	Brand                 string               `json:"brand" form:"brand"`
	DeviceModel           string               `json:"deviceModel" form:"deviceModel"`
	ConfOverview          string               `json:"confOverview" form:"confOverview"`
	DeviceModels          []NetworkDeviceModel `json:"deviceModels" form:"deviceModels"`
}

type NetworkDeviceModel struct {
	ConfOverview string `json:"confOverview" form:"confOverview"`
	DeviceModel  string `json:"deviceModel" form:"deviceModel"`
}
