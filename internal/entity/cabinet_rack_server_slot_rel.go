package entity

const CabinetRackServerSlotRelTable = "cabinet_rack_server_slot_rel"

type CabinetRackServerSlotRel struct {
	CabinetId         int64 `gorm:"column:cabinet_id" json:"cabinetId"`                   // 机柜id
	RackServerSlotNum int   `gorm:"column:rack_server_slot_num" json:"rackServerSlotNum"` // 已上架服务器槽位（U位）号
}

func (entity *CabinetRackServerSlotRel) TableName() string {
	return CabinetRackServerSlotRelTable
}
