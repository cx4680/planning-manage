package entity

const CapServerCalcBaselineTable = "cap_server_calc_baseline"

type CapServerCalcBaseline struct {
	Id                  int64  `gorm:"column:id" json:"id"`                                      // 主键id
	VersionId           int64  `gorm:"column:version_id" json:"versionId"`                       // 版本id
	ExpendRes           string `gorm:"column:expend_res" json:"expendRes"`                       // 消耗资源
	ExpendResCode       string `gorm:"column:expend_res_code" json:"expendResCode"`              // 消耗资源编码
	ExpendNodeRoleCode  string `gorm:"column:expend_node_role_code" json:"expendNodeRoleCode"`   // 消耗节点角色编码
	OccNodeRes          string `gorm:"column:occ_node_res" json:"occNodeRes"`                    // 占用节点资源
	OccNodeResCode      string `gorm:"column:occ_node_res_code" json:"occNodeResCode"`           // 占用节点资源编码
	NodeWastage         string `gorm:"column:node_wastage" json:"nodeWastage"`                   // 节点损耗
	NodeWastageCalcType int    `gorm:"column:node_wastage_calc_type" json:"nodeWastageCalcType"` // 节点损耗计算类型，1：数量，2：百分比，3：数据盘数量
	WaterLevel          string `gorm:"column:water_level" json:"waterLevel"`                     // 水位
}

func (entity *CapServerCalcBaseline) TableName() string {
	return CapServerCalcBaselineTable
}
