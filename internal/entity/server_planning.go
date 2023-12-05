package entity

import "time"

const ServerPlanningTable = "server_planning"

type ServerPlanningManage struct {
	Id               int64     `gorm:"column:id" json:"id"`                               //服务器规划id
	PlanId           int64     `gorm:"column:plan_id" json:"planId"`                      //方案id
	ServerBaselineId int64     `gorm:"column:server_baseline_id" json:"serverBaselineId"` //服务器基线表id
	Number           int64     `gorm:"column:number" json:"number"`                       //数量
	CreateUserId     string    `gorm:"column:create_user_id" json:"createUserId"`         //创建人id
	CreateTime       time.Time `gorm:"column:create_time" json:"createTime"`              //创建时间
	UpdateUserId     string    `gorm:"column:update_user_id" json:"updateUserId"`         //更新人id
	UpdateTime       time.Time `gorm:"column:update_time" json:"updateTime"`              //更新时间
	DeleteState      int       `gorm:"column:delete_state" json:"-"`                      //作废状态：1，作废；0，正常
}

func (entity *ServerPlanningManage) TableName() string {
	return ServerPlanningTable
}
