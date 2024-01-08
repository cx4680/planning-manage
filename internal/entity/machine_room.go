package entity

const MachineRoomTable = "machine_room"

type MachineRoom struct {
	Id       int64  `gorm:"column:id" json:"id"`             //id
	AzId     int64  `gorm:"column:az_id" json:"regionId"`    //azId
	Name     string `gorm:"column:name" json:"name"`         //机房名称
	Abbr     string `gorm:"column:abbr" json:"abbr"`         //机房缩写
	Province string `gorm:"column:province" json:"province"` //省
	City     string `gorm:"column:city" json:"city"`         //市
	Address  string `gorm:"column:address" json:"address"`   //地址
}

func (entity *MachineRoom) TableName() string {
	return MachineRoomTable
}
