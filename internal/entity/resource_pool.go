package entity

const ResourcePoolTable = "resource_pool"

type ResourcePool struct {
	Id                  int64  `gorm:"column:id" json:"id"`                                     // 主键Id
	PlanId              int64  `gorm:"column:plan_id" json:"planId"`                            // 方案Id
	NodeRoleId          int64  `gorm:"column:node_role_id" json:"nodeRoleId"`                   // 节点角色Id
	ResourcePoolName    string `gorm:"column:resource_pool_name" json:"resourcePoolName"`       // 资源池名称
	OpenDpdk            int    `gorm:"column:open_dpdk" json:"openDpdk"`                        // 是否开启DPDK，0：未开启，1：开启
	DefaultResourcePool int    `gorm:"column:default_resource_pool" json:"defaultResourcePool"` // 是否为默认资源池，0：否，1：是
}

func (entity *ResourcePool) TableName() string {
	return ResourcePoolTable
}
