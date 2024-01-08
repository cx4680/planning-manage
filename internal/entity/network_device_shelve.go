package entity

import "time"

const NetworkDeviceShelveTable = "network_device_shelve"

type NetworkDeviceShelve struct {
	Id                int64     `gorm:"column:id" json:"id"`                                 //主键
	PlanId            int64     `gorm:"column:plan_id" json:"planId"`                        //方案ID
	DeviceLogicalId   string    `gorm:"column:device_logical_id" json:"deviceLogicalId"`     //网络设备逻辑ID
	DeviceId          string    `gorm:"column:device_id" json:"deviceId"`                    //网络设备ID
	Sn                string    `gorm:"column:sn" json:"sn"`                                 //SN
	MachineRoomAbbr   string    `gorm:"column:machine_room_abbr" json:"machineRoomAbbr"`     //机房缩写
	MachineRoomNumber string    `gorm:"column:machine_room_number" json:"machineRoomNumber"` //机房编号
	CabinetNumber     time.Time `gorm:"column:cabinet_number" json:"cabinetNumber"`          //机柜编号
	SlotPosition      time.Time `gorm:"column:slot_position" json:"slotPosition"`            //槽位
	UNumber           int       `gorm:"column:u_number" json:"uNumber"`                      //U数
	CreateUserId      string    `gorm:"column:create_user_id" json:"createUserId"`           //创建人id
	CreateTime        time.Time `gorm:"column:create_time" json:"createTime"`                //创建时间
}

func (entity *NetworkDeviceShelve) TableName() string {
	return NetworkDeviceShelveTable
}
