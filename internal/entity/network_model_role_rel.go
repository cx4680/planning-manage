package entity

const NetworkModelRoleRelTable = "network_model_role_rel"

type NetworkModelRoleRel struct {
	NetworkDeviceRoleId int64 `gorm:"column:network_device_role_id" json:"networkDeviceRoleId"` // 网络设备角色id
	NetworkModel        int   `gorm:"column:network_model" json:"networkModel"`                 // 网络组网模式，网络组网模式，1：三网合一，2：两网分离，3：三网分离
	AssociatedType      int   `gorm:"column:associated_type" json:"associatedType"`             // 关联类型，0：节点角色，1：网络设备角色
	RoleId              int64 `gorm:"column:role_id" json:"roleId"`                             // 关联的节点角色id或者网络设备角色id
	RoleNum             int   `gorm:"column:role_num" json:"roleNum"`                           // 关联相同角色数量
}

func (entity *NetworkModelRoleRel) TableName() string {
	return NetworkModelRoleRelTable
}
