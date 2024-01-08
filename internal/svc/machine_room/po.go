package machine_room

type PageRequest struct {
	PlanId   int64 `form:"planId"`
	Current  int   `json:"current"`
	PageSize int   `json:"pageSize"`
}

type CabinetExcel struct {
	MachineRoomAbbr       string `excel:"name:机房缩写;" json:"machineRoomAbbr"`           // 机房缩写
	MachineRoomNum        string `excel:"name:房间号;" json:"machineRoomNum"`             // 房间号
	ColumnNum             string `excel:"name:列号;" json:"columnNum"`                   // 列号
	CabinetNum            string `excel:"name:机柜编号;" json:"cabinetNum"`                // 机柜编号
	OriginalNum           string `excel:"name:原始编号;" json:"originalNum"`               // 原始编号
	CabinetType           string `excel:"name:机柜类型;" json:"cabinetType"`               // 机柜类型
	BusinessAttribute     string `excel:"name:业务属性;" json:"businessAttribute"`         // 业务属性
	CabinetAsw            string `excel:"name:机柜ASW组;" json:"cabinetAsw"`              // 机柜ASW组
	TotalPower            int    `excel:"name:总功率（W）;" json:"totalPower"`              // 总功率（W）
	ResidualPower         int    `excel:"name:剩余功率（W）;" json:"residualPower"`          // 剩余功率（W）
	TotalSlotNum          int    `excel:"name:总槽位数（U位）;" json:"totalSlotNum"`          // 总槽位数（U位）
	IdleSlotRange         string `excel:"name:空闲槽位（U位）范围;" json:"idleSlotRange"`       // 空闲槽位（U位）范围
	MaxRackServerNum      int    `excel:"name:最大可上架服务器数;" json:"maxRackServerNum"`     // 最大可上架服务器数
	ResidualRackServerNum int    `excel:"name:剩余上架服务器数;" json:"residualRackServerNum"` // 剩余上架服务器数
	RackServerSlot        string `excel:"name:已上架服务器（U位）;" json:"rackServerSlot"`      // 已上架服务器（U位）
	ResidualRackAswPort   string `excel:"name:剩余可上架ASW端口;" json:"residualRackAswPort"` // 剩余可上架ASW端口
}
