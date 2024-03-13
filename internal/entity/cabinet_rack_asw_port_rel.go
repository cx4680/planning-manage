package entity

const CabinetRackAswPortRelTable = "cabinet_rack_asw_port_rel"

type CabinetRackAswPortRel struct {
	CabinetId              int64 `gorm:"column:cabinet_id" json:"cabinetId"`                              // 机柜id
	ResidualRackAswPortNum int   `gorm:"column:residual_rack_asw_port_num" json:"residualRackAswPortNum"` // 剩余可上架ASW端口号
}

func (entity *CabinetRackAswPortRel) TableName() string {
	return CabinetRackAswPortRelTable
}
