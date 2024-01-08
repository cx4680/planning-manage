package entity

import "time"

const AzManageTable = "az_manage"

type AzManage struct {
	Id              int64          `gorm:"column:id" json:"id"`                       //azId
	Code            string         `gorm:"column:code" json:"code"`                   //az编码
	RegionId        int64          `gorm:"column:region_id" json:"regionId"`          //regionId
	CreateUserId    string         `gorm:"column:create_user_id" json:"createUserId"` //创建人id
	CreateTime      time.Time      `gorm:"column:create_time" json:"createTime"`      //创建时间
	UpdateUserId    string         `gorm:"column:update_user_id" json:"updateUserId"` //更新人id
	UpdateTime      time.Time      `gorm:"column:update_time" json:"updateTime"`      //更新时间
	DeleteState     int            `gorm:"column:delete_state" json:"-"`              //作废状态：1，作废；0，正常
	MachineRoomList []*MachineRoom `gorm:"-" json:"machineRoomList"`
	CellList        []*CellManage  `gorm:"-" json:"cellList"`
}

func (entity *AzManage) TableName() string {
	return AzManageTable
}
