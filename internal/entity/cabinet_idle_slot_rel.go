package entity

const CabinetIdleSlotRelTable = "cabinet_idle_slot_rel"

type CabinetIdleSlotRel struct {
	CabinetId      int64 `gorm:"column:cabinet_id" json:"cabinetId"`            // 机柜id
	IdleSlotNumber int   `gorm:"column:idle_slot_number" json:"idleSlotNumber"` // 空闲槽位（U位）号
}

func (entity *CabinetIdleSlotRel) TableName() string {
	return CabinetIdleSlotRelTable
}
