package entity

import "time"

const (
	ServerPlanningTable    = "server_planning"
	ServerCapPlanningTable = "server_cap_planning"
	ServerShelveTable      = "server_shelve"
)

type ServerPlanning struct {
	Id                 int64     `gorm:"column:id" json:"id"`                                  // 服务器规划id
	PlanId             int64     `gorm:"column:plan_id" json:"planId"`                         // 方案id
	NodeRoleId         int64     `gorm:"column:node_role_id" json:"nodeRoleId"`                // 节点角色id
	ServerBaselineId   int64     `gorm:"column:server_baseline_id" json:"serverBaselineId"`    // 服务器基线表id
	MixedNodeRoleId    int64     `gorm:"column:mixed_node_role_id" json:"mixedNodeRoleId"`     // 混合部署节点角色id
	Number             int       `gorm:"column:number" json:"number"`                          // 数量
	OpenDpdk           int       `gorm:"column:open_dpdk" json:"openDpdk"`                     // 是否开启DPDK，1：开启，0：关闭
	NetworkInterface   string    `gorm:"column:network_interface" json:"networkInterface"`     // 网络类型
	CpuType            string    `gorm:"column:cpu_type" json:"cpuType"`                       // CPU类型
	BusinessAttributes string    `gorm:"column:business_attributes" json:"businessAttributes"` // 业务属性
	ShelveMode         string    `gorm:"column:shelve_mode" json:"shelveMode"`                 // 上架模式
	ShelvePriority     int       `gorm:"column:shelve_priority" json:"shelvePriority"`         // 上架优先级
	CreateUserId       string    `gorm:"column:create_user_id" json:"createUserId"`            // 创建人id
	CreateTime         time.Time `gorm:"column:create_time" json:"createTime"`                 // 创建时间
	UpdateUserId       string    `gorm:"column:update_user_id" json:"updateUserId"`            // 更新人id
	UpdateTime         time.Time `gorm:"column:update_time" json:"updateTime"`                 // 更新时间
	DeleteState        int       `gorm:"column:delete_state" json:"-"`                         // 作废状态：1，作废；0，正常
}

func (entity *ServerPlanning) TableName() string {
	return ServerPlanningTable
}

type ServerCapPlanning struct {
	Id                 int64  `gorm:"column:id" json:"id"`                                   // 容量规划id
	PlanId             int64  `gorm:"column:plan_id" json:"planId"`                          // 方案id
	NodeRoleId         int64  `gorm:"column:node_role_id" json:"nodeRoleId"`                 // 节点角色id
	CapacityBaselineId int64  `gorm:"column:capacity_baseline_id" json:"capacityBaselineId"` // 容量指标id
	Number             int    `gorm:"column:number" json:"number"`                           // 数量
	FeatureNumber      int    `gorm:"column:feature_number" json:"featureNumber"`            // 特性数量
	CapacityNumber     string `gorm:"column:capacity_number" json:"capacityNumber"`          // 容量总消耗
	ExpendResCode      string `gorm:"column:expend_res_code" json:"expendResCode"`           // 消耗资源编码
}

func (entity *ServerCapPlanning) TableName() string {
	return ServerCapPlanningTable
}

type ServerShelve struct {
	Id                    int64     `gorm:"column:id" json:"id"`                                         // 主键id
	PlanId                int64     `gorm:"column:plan_id" json:"planId"`                                // 方案id
	SortNumber            int       `gorm:"column:sort_number" json:"sortNumber"`                        // 序号
	NodeRoleId            int64     `gorm:"column:node_role_id" json:"nodeRoleId"`                       // 节点角色id
	NodeIp                string    `gorm:"column:node_ip" json:"nodeIp"`                                // 节点IP
	Sn                    string    `gorm:"column:sn" json:"sn"`                                         // SN
	Model                 string    `gorm:"column:model" json:"model"`                                   // 机型
	CabinetId             int64     `gorm:"column:cabinet_id" json:"cabinetId"`                          // 机柜id
	MachineRoomAbbr       string    `gorm:"column:machine_room_abbr" json:"machineRoomAbbr"`             // 机房缩写
	MachineRoomNumber     string    `gorm:"column:machine_room_number" form:"machineRoomNumber"`         // 房间号
	ColumnNumber          string    `gorm:"column:column_number" form:"columnNumber"`                    // 列号
	CabinetAsw            string    `gorm:"column:cabinet_asw" json:"cabinetAsw"`                        // 机柜ASW组
	CabinetNumber         string    `gorm:"column:cabinet_number" json:"cabinetNumber"`                  // 机柜编号
	CabinetOriginalNumber string    `gorm:"column:cabinet_original_number" json:"cabinetOriginalNumber"` // 机柜原始编号
	CabinetLocation       string    `gorm:"column:cabinet_location" json:"cabinetLocation"`              // 机柜位置
	SlotPosition          string    `gorm:"column:slot_position" json:"slotPosition"`                    // 槽位（U位）
	NetworkInterface      string    `gorm:"column:network_interface" json:"networkInterface"`            // 网络接口
	BmcUserName           string    `gorm:"column:bmc_user_name" json:"bmcUserName"`                     // bmc用户名
	BmcPassword           string    `gorm:"column:bmc_password" json:"bmcPassword"`                      // bmc密码
	BmcIp                 string    `gorm:"column:bmc_ip" json:"bmcIp"`                                  // bmc IP地址
	BmcMac                string    `gorm:"column:bmc_mac" json:"bmcMac"`                                // bmc mac地址
	Mask                  string    `gorm:"column:mask" json:"mask"`                                     // 掩码
	Gateway               string    `gorm:"column:gateway" json:"gateway"`                               // 网关
	CreateUserId          string    `gorm:"column:create_user_id" json:"createUserId"`                   // 创建人id
	CreateTime            time.Time `gorm:"column:create_time" json:"createTime"`                        // 创建时间
}

func (entity *ServerShelve) TableName() string {
	return ServerShelveTable
}
