package entity

import "time"

const ServerPlanningTable = "server_planning"

type ServerPlanning struct {
	Id                 int64     `gorm:"column:id" json:"id"`                               // 服务器规划id
	PlanId             int64     `gorm:"column:plan_id" json:"planId"`                      // 方案id
	NodeRoleId         int64     `gorm:"column:node_role_id" json:"nodeRoleId"`             // 节点角色id
	ServerBaselineId   int64     `gorm:"column:server_baseline_id" json:"serverBaselineId"` // 服务器基线表id
	Number             int       `gorm:"column:number" json:"number"`                       // 数量
	CreateUserId       string    `gorm:"column:create_user_id" json:"createUserId"`         // 创建人id
	CreateTime         time.Time `gorm:"column:create_time" json:"createTime"`              // 创建时间
	UpdateUserId       string    `gorm:"column:update_user_id" json:"updateUserId"`         // 更新人id
	UpdateTime         time.Time `gorm:"column:update_time" json:"updateTime"`              // 更新时间
	DeleteState        int       `gorm:"column:delete_state" json:"-"`                      // 作废状态：1，作废；0，正常
	NodeRoleName       string    `gorm:"column:-" json:"nodeRoleName"`                      // 节点角色名称
	NodeRoleAnnotation string    `gorm:"column:-" json:"nodeRoleAnnotation"`                // 节点说明
	ServerModel        string    `gorm:"column:-" json:"serverModel"`                       // 机型
	ServerArch         string    `gorm:"column:-" json:"ServerArch"`                        // 架构
}

func (entity *ServerPlanning) TableName() string {
	return ServerPlanningTable
}
