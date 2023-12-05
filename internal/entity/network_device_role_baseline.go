package entity

const NetworkDeviceRoleBaselineTable = "network_device_role_baseline"

type NetworkDeviceRoleBaseline struct {
	Id              int64  `gorm:"column:id" json:"id"`                             // 主键id
	VersionId       int64  `gorm:"column:version_id" json:"versionId"`              // 版本id
	DeviceType      string `gorm:"column:device_type" json:"deviceType"`            // 设备类型
	FuncType        string `gorm:"column:func_type" json:"funcType"`                // 类型
	FuncCompo       string `gorm:"column:func_compo" json:"funcCompo"`              // 功能组件
	FuncCompoName   string `gorm:"column:func_compo_name" json:"funcCompoName"`     // 功能组件名称
	Description     string `gorm:"column:description" json:"description"`           // 描述
	TwoNetworkIso   int    `gorm:"column: two_network_iso" json:"twoNetworkIso"`    // 两网分离
	ThreeNetworkIso int    `gorm:"column:three_network_iso" json:"threeNetworkIso"` // 三网分离
	TriplePlay      int    `gorm:"column:triple_play" json:"triplePlay"`            // 三网合一
	MinimumNumUnit  int    `gorm:"column:minimum_num_unit" json:"minimumNumUnit"`   // 最小单元数
	UnitDeviceNum   int    `gorm:"column:unit_device_num" json:"unitDeviceNum"`     // 单元设备数量
	DesignSpec      string `gorm:"column:design_spec" json:"designSpec"`            // 设计规格
}

func (entity *NetworkDeviceRoleBaseline) TableName() string {
	return NetworkDeviceRoleBaselineTable
}
