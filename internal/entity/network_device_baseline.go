package entity

const NetworkDeviceBaselineTable = "network_device_baseline"

type NetworkDeviceBaseline struct {
	Id           int64  `gorm:"column:id" json:"id"`                      // 主键id
	VersionId    int64  `gorm:"column:version_id" json:"versionId"`       // 版本id
	DeviceModel  string `gorm:"column:device_model" json:"deviceModel"`   // 设备型号
	Manufacturer string `gorm:"column:manufacturer" json:"manufacturer"`  // 厂商
	DeviceType   int    `gorm:"column:device_type" json:"deviceType"`     // 信创/商用，0：信创，1：商用
	NetworkModel string `gorm:"column:network_model" json:"networkModel"` // 网络模型
	ConfOverview string `gorm:"column:conf_overview" json:"confOverview"` // 配置概述
	Purpose      string `gorm:"column:purpose" json:"purpose"`            // 用途
	BomId        string `gorm:"column:bom_id" json:"bomId"`               // bom id
}

func (entity *NetworkDeviceBaseline) TableName() string {
	return NetworkDeviceBaselineTable
}
