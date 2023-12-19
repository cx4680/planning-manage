package entity

import "time"

const CellManageTable = "cell_manage"

type CellManage struct {
	Id           int64     `gorm:"column:id" json:"id"`                       //cell Id
	Name         string    `gorm:"column:name" json:"name"`                   //cell名称
	AzId         int64     `gorm:"column:az_id" json:"azId"`                  //azId
	Type         string    `gorm:"column:type" json:"type"`                   //cell类型
	CreateUserId string    `gorm:"column:create_user_id" json:"createUserId"` //创建人id
	CreateTime   time.Time `gorm:"column:create_time" json:"createTime"`      //创建时间
	UpdateUserId string    `gorm:"column:update_user_id" json:"updateUserId"` //更新人id
	UpdateTime   time.Time `gorm:"column:update_time" json:"updateTime"`      //更新时间
	DeleteState  int       `gorm:"column:delete_state" json:"-"`              //作废状态：1，作废；0，正常
}

func (entity *CellManage) TableName() string {
	return CellManageTable
}
