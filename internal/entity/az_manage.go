package entity

import "time"

const AzManageTable = "az_manage"

type AzManage struct {
	Id              int64     `gorm:"column:id" json:"id"`                             //azId
	Code            string    `gorm:"column:code" json:"code"`                         //az编码
	Name            string    `gorm:"column:name" json:"name"`                         //az名称
	RegionId        int64     `gorm:"column:region_id" json:"regionId"`                //regionId
	MachineRoomName string    `gorm:"column:machine_room_name" json:"machineRoomName"` //机房名称
	MachineRoomCode string    `gorm:"column:machine_room_code" json:"machineRoomCode"` //机房缩写
	Province        string    `gorm:"column:province" json:"province"`                 //省
	City            string    `gorm:"column:city" json:"city"`                         //市
	Address         string    `gorm:"column:address" json:"address"`                   //地址
	CreateUserId    string    `gorm:"column:create_user_id" json:"createUserId"`       //创建人id
	CreateTime      time.Time `gorm:"column:create_time" json:"createTime"`            //创建时间
	UpdateUserId    string    `gorm:"column:update_user_id" json:"updateUserId"`       //更新人id
	UpdateTime      time.Time `gorm:"column:update_time" json:"updateTime"`            //更新时间
	DeleteState     int       `gorm:"column:delete_state" json:"-"`                    //作废状态：1，作废；0，正常
}

func (entity *AzManage) TableName() string {
	return AzManageTable
}
