package entity

const AzCellRelTable = "az_cell_rel"

type AzCellRel struct {
	AzId   int64 `gorm:"column:az_id" json:"azId"`     //azId
	CellId int64 `gorm:"column:cell_id" json:"cellId"` //cell Id
}

func (entity *AzCellRel) TableName() string {
	return AzCellRelTable
}
