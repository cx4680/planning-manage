package entity

import "time"

const PlanManageTable = "plan_manage"

type PlanManage struct {
	Id               int64     `gorm:"column:id" json:"id"`                                //方案id
	Name             string    `gorm:"column:name" json:"name"`                            //方案名称
	Stage            string    `gorm:"column:stage" json:"stage"`                          //方案阶段
	Type             string    `gorm:"column:type" json:"type"`                            //方案类型
	ProjectId        int64     `gorm:"column:project_id" json:"project_id"`                //项目id
	CreateUserId     string    `gorm:"column:create_user_id" json:"createUserId"`          //创建人id
	CreateTime       time.Time `gorm:"column:create_time" json:"createTime"`               //创建时间
	UpdateUserId     string    `gorm:"column:update_user_id" json:"updateUserId"`          //更新人id
	UpdateTime       time.Time `gorm:"column:update_time" json:"updateTime"`               //更新时间
	DeleteState      int       `gorm:"column:delete_state" json:"deleteState"`             //作废状态：1，作废；0，正常
	BusinessPanStage int       `gorm:"column:business_plan_stage" json:"businessPanStage"` //业务规划阶段：0，业务规划开始阶段；1，云产品配置阶段；2，服务器规划阶段；3，网络设备规划阶段； 4，业务规划结束
	DeliverPlanStage int       `gorm:"column:deliver_plan_stage" json:"deliverPlanStage"`  //交付规划阶段：0，交付规划开始阶段
	Alternative      int       `gorm:"-" json:"alternative"`                               //是否有备选方案
}

func (entity *PlanManage) TableName() string {
	return PlanManageTable
}
