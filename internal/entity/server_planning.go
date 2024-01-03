package entity

import "time"

const ServerPlanningTable = "server_planning"

type ServerPlanning struct {
	Id                 int64                     `gorm:"column:id" json:"id"`                               // 服务器规划id
	PlanId             int64                     `gorm:"column:plan_id" json:"planId"`                      // 方案id
	NodeRoleId         int64                     `gorm:"column:node_role_id" json:"nodeRoleId"`             // 节点角色id
	ServerBaselineId   int64                     `gorm:"column:server_baseline_id" json:"serverBaselineId"` // 服务器基线表id
	MixedNodeRoleId    int64                     `gorm:"column:mixed_node_role_id" json:"mixedNodeRoleId"`  // 混合部署节点角色id
	Number             int                       `gorm:"column:number" json:"number"`                       // 数量
	OpenDpdk           int                       `gorm:"column:open_dpdk" json:"openDpdk"`                  // 是否开启DPDK，1：开启，0：关闭
	NetworkInterface   string                    `gorm:"column:network_interface" form:"networkInterface"`  // 网络类型
	CpuType            string                    `gorm:"column:cpu_type" form:"cpuType"`                    // cpu类型
	CreateUserId       string                    `gorm:"column:create_user_id" json:"createUserId"`         // 创建人id
	CreateTime         time.Time                 `gorm:"column:create_time" json:"createTime"`              // 创建时间
	UpdateUserId       string                    `gorm:"column:update_user_id" json:"updateUserId"`         // 更新人id
	UpdateTime         time.Time                 `gorm:"column:update_time" json:"updateTime"`              // 更新时间
	DeleteState        int                       `gorm:"column:delete_state" json:"-"`                      // 作废状态：1，作废；0，正常
	NodeRoleName       string                    `gorm:"-" json:"nodeRoleName"`                             // 节点角色名称
	NodeRoleClassify   string                    `gorm:"-" json:"nodeRoleClassify"`                         // 节点角色分类
	NodeRoleAnnotation string                    `gorm:"-" json:"nodeRoleAnnotation"`                       // 节点说明
	SupportDpdk        int                       `gorm:"-" json:"supportDpdk"`                              // 是否支持DPDK, 0:否，1：是
	ServerBomCode      string                    `gorm:"-" json:"serverBomCode"`                            // BOM编码
	ServerArch         string                    `gorm:"-" json:"ServerArch"`                               // 架构
	ServerBaselineList []*ServerPlanningBaseline `gorm:"-" json:"serverBaselineList"`                       // 可选择机型列表
	MixedNodeRoleList  []*MixedNodeRole          `gorm:"-" json:"mixedNodeRoleList"`                        // 可混合部署角色列表
}

func (entity *ServerPlanning) TableName() string {
	return ServerPlanningTable
}

type MixedNodeRole struct {
	Id   int64  `gorm:"-" json:"id"`   // 混合节点角色id
	Name string `gorm:"-" json:"name"` // 混合节点角色名称

}

type ServerPlanningBaseline struct {
	Id                int64  `gorm:"-" json:"id"`                // 服务器id
	BomCode           string `gorm:"-" json:"bomCode"`           // BOM编码
	NetworkInterface  string `gorm:"-" json:"networkInterface"`  // 网络类型
	Cpu               int    `gorm:"-" json:"cpu"`               // cpu损耗
	CpuType           string `gorm:"-" json:"cpuType"`           // cpu类型
	Arch              string `gorm:"-" json:"arch"`              // 硬件架构
	ConfigurationInfo string `gorm:"-" json:"configurationInfo"` // 配置概要
}
