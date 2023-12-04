package entity

const IPDemandDeviceRoleRelTable = "ip_demand_device_role_rel"

type IPDemandDeviceRoleRel struct {
	IPDemandId   int64 `gorm:"column:ip_demand_id" json:"IPDemandId"`     // ip需求规划id
	DeviceRoleId int64 `gorm:"column:device_role_id" json:"deviceRoleId"` // 设备角色id
}

func (entity *IPDemandDeviceRoleRel) TableName() string {
	return IPDemandDeviceRoleRelTable
}
