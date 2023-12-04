package entity

const NetworkDeviceRoleRelTable = "network_device_role_rel"

type NetworkDeviceRoleRel struct {
	DeviceId     int64 `gorm:"column:device_id" json:"deviceId"`          // 设备id
	DeviceRoleId int64 `gorm:"column:device_role_id" json:"deviceRoleId"` // 设备角色id
}

func (entity *NetworkDeviceRoleRel) TableName() string {
	return NetworkDeviceRoleRelTable
}
