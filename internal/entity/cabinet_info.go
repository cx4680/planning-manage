package entity

import "time"

const CabinetInfoTable = "cabinet_info"

type CabinetInfo struct {
	Id                    int64     `gorm:"column:id" json:"id"`                                          // 主键Id
	PlanId                int64     `gorm:"column:plan_id" json:"planId"`                                 // 方案Id
	MachineRoomAbbr       string    `gorm:"column:machine_room_abbr" json:"machineRoomAbbr"`              // 机房缩写
	MachineRoomNum        string    `gorm:"column:machine_room_num" json:"machineRoomNum"`                // 房间号
	ColumnNum             string    `gorm:"column:column_num" json:"columnNum"`                           // 列号
	CabinetNum            string    `gorm:"column:cabinet_num" json:"cabinetNum"`                         // 机柜编号
	OriginalNum           string    `gorm:"column:original_num" json:"originalNum"`                       // 原始编号
	CabinetType           int       `gorm:"column:cabinet_type" json:"cabinetType"`                       // 机柜类型，1：网络机柜，2：服务机柜，3：存储机柜
	ServiceAttribute      string    `gorm:"column:service_attribute" json:"serviceAttribute"`             // 业务属性
	CabinetAsw            string    `gorm:"column:cabinet_asw" json:"cabinetAsw"`                         // 机柜ASW组
	TotalPower            int       `gorm:"column:total_power" json:"totalPower"`                         // 总功率（W）
	ResidualPower         int       `gorm:"column:residual_power" json:"residualPower"`                   // 剩余功率（W）
	TotalSlotNum          int       `gorm:"column:total_slot_num" json:"totalSlotNum"`                    // 总槽位数（U位）
	IdleSlotRange         string    `gorm:"column:idle_slot_range" json:"idleSlotRange"`                  // 空闲槽位（U位）范围
	MaxRackServerNum      int       `gorm:"column:max_rack_server_num" json:"maxRackServerNum"`           // 最大可上架服务器数
	ResidualRackServerNum int       `gorm:"column:residual_rack_server_num" json:"residualRackServerNum"` // 剩余上架服务器数
	RackServerSlot        string    `gorm:"column:rack_server_slot" json:"rackServerSlot"`                // 已上架服务器（U位）
	ResidualRackAswPort   string    `gorm:"column:residual_rack_asw_port" json:"residualRackAswPort"`     // 剩余可上架ASW端口
	CreateTime            time.Time `gorm:"column:create_time" json:"createTime"`                         // 创建时间
	UpdateTime            time.Time `gorm:"column:update_time" json:"updateTime"`                         // 更新时间
}

func (entity *CabinetInfo) TableName() string {
	return CabinetInfoTable
}
