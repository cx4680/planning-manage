package server

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"time"
)

type Request struct {
	Id                 int64
	UserId             string
	PlanId             int64                    `form:"planId"`
	NetworkInterface   string                   `form:"networkInterface"`
	CpuType            string                   `form:"cpuType"`
	ServerList         []*RequestServer         `form:"serverList"`
	ServerCapacityList []*RequestServerCapacity `form:"serverCapacityList"`
}

type Server struct {
	Id                 int64            `gorm:"column:id" json:"id"`                               // 服务器规划id
	PlanId             int64            `gorm:"column:plan_id" json:"planId"`                      // 方案id
	NodeRoleId         int64            `gorm:"column:node_role_id" json:"nodeRoleId"`             // 节点角色id
	ServerBaselineId   int64            `gorm:"column:server_baseline_id" json:"serverBaselineId"` // 服务器基线表id
	MixedNodeRoleId    int64            `gorm:"column:mixed_node_role_id" json:"mixedNodeRoleId"`  // 混合部署节点角色id
	Number             int              `gorm:"column:number" json:"number"`                       // 数量
	OpenDpdk           int              `gorm:"column:open_dpdk" json:"openDpdk"`                  // 是否开启DPDK，1：开启，0：关闭
	NetworkInterface   string           `gorm:"column:network_interface" form:"networkInterface"`  // 网络类型
	CpuType            string           `gorm:"column:cpu_type" form:"cpuType"`                    // cpu类型
	CreateUserId       string           `gorm:"column:create_user_id" json:"createUserId"`         // 创建人id
	CreateTime         time.Time        `gorm:"column:create_time" json:"createTime"`              // 创建时间
	UpdateUserId       string           `gorm:"column:update_user_id" json:"updateUserId"`         // 更新人id
	UpdateTime         time.Time        `gorm:"column:update_time" json:"updateTime"`              // 更新时间
	DeleteState        int              `gorm:"column:delete_state" json:"-"`                      // 作废状态：1，作废；0，正常
	NodeRoleName       string           `gorm:"-" json:"nodeRoleName"`                             // 节点角色名称
	NodeRoleClassify   string           `gorm:"-" json:"nodeRoleClassify"`                         // 节点角色分类
	NodeRoleAnnotation string           `gorm:"-" json:"nodeRoleAnnotation"`                       // 节点说明
	SupportDpdk        int              `gorm:"-" json:"supportDpdk"`                              // 是否支持DPDK, 0:否，1：是
	ServerBomCode      string           `gorm:"-" json:"serverBomCode"`                            // BOM编码
	ServerArch         string           `gorm:"-" json:"ServerArch"`                               // 架构
	ServerBaselineList []*Baseline      `gorm:"-" json:"serverBaselineList"`                       // 可选择机型列表
	MixedNodeRoleList  []*MixedNodeRole `gorm:"-" json:"mixedNodeRoleList"`                        // 可混合部署角色列表
}

type MixedNodeRole struct {
	Id   int64  `gorm:"-" json:"id"`   // 混合节点角色id
	Name string `gorm:"-" json:"name"` // 混合节点角色名称

}

type Baseline struct {
	Id                int64  `gorm:"-" json:"id"`                // 服务器id
	BomCode           string `gorm:"-" json:"bomCode"`           // BOM编码
	NetworkInterface  string `gorm:"-" json:"networkInterface"`  // 网络类型
	Cpu               int    `gorm:"-" json:"cpu"`               // cpu损耗
	CpuType           string `gorm:"-" json:"cpuType"`           // cpu类型
	Arch              string `gorm:"-" json:"arch"`              // 硬件架构
	ConfigurationInfo string `gorm:"-" json:"configurationInfo"` // 配置概要
}

type RequestServer struct {
	NodeRoleId       int64 `form:"nodeRoleId"`
	MixedNodeRoleId  int64 `form:"mixedNodeRoleId"`
	ServerBaselineId int64 `form:"serverBaselineId"`
	Number           int   `form:"number"`
	OpenDpdk         int   `form:"openDpdk"`
}

type RequestServerCapacity struct {
	Id            int64 `form:"id"`
	Number        int   `form:"number"`
	FeatureNumber int   `form:"featureNumber"`
}

type ResponseDownloadServer struct {
	NodeRole   string `json:"nodeRole" excel:"name:角色;"`
	ServerType string `json:"serverType" excel:"name:设备类型;"`
	BomCode    string `json:"bomCode" excel:"name:机型;"`
	Spec       string `json:"spec" excel:"name:规格;"`
	Number     string `json:"number" excel:"name:数量;"`
}

type ResponseCapClassification struct {
	Classification string                `json:"classification"` // 分类
	CapConvert     []*ResponseCapConvert `json:"capConvert"`
}

type ResponseCapConvert struct {
	VersionId        int64               `json:"versionId"`        // 版本id
	ProductName      string              `json:"productName"`      // 产品名称
	ProductCode      string              `json:"productCode"`      // 产品编码
	SellSpecs        string              `json:"sellSpecs"`        // 售卖规格
	CapPlanningInput string              `json:"capPlanningInput"` // 容量规划输入
	Number           int                 `json:"number"`           // 数量
	Unit             string              `json:"unit"`             // 单位
	FeatureId        int64               `json:"featureId"`        // 特性id
	FeatureType      string              `json:"featureType"`      // 特性类型
	FeatureNumber    int                 `json:"featureNumber"`    // 特性数量
	Features         []*ResponseFeatures `json:"features"`         // 特性选项
	Description      string              `json:"description"`      // 说明
}

type ResponseFeatures struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

var FeatureMap = map[string]string{"超分": "超分比", "三副本": "副本模式", "EC纠删码": "副本模式"}

type ResponseServerShelve struct {
	*entity.ServerShelve
	NodeRoleName string `gorm:"-" json:"nodeRoleName"` // 节点角色名称
}

type ShelveDownload struct {
	SortNumber            int    `excel:"name:序号;" json:"sortNumber"`
	NodeRoleName          string `excel:"name:节点角色;" json:"nodeRoleName"`
	NodeIp                string `excel:"name:节点IP;" json:"nodeIp"`
	Sn                    string `excel:"name:SN;" json:"sn"`
	Model                 string `excel:"name:机型;" json:"model"`
	MachineRoomAbbr       string `excel:"name:机房缩写;" json:"machineRoomAbbr"`
	MachineRoomNumber     string `excel:"name:房间号;" form:"machineRoomNumber"`
	ColumnNumber          string `excel:"name:列号;" form:"columnNumber"`
	CabinetAsw            string `excel:"name:机柜ASW组;" json:"cabinetAsw"`
	CabinetNumber         string `excel:"name:机柜编号;" json:"cabinetNumber"`
	CabinetOriginalNumber string `excel:"name:机柜原始编号;" json:"cabinetOriginalNumber"`
	CabinetLocation       string `excel:"name:机柜位置;" json:"cabinetLocation"`
	SlotPosition          string `excel:"name:槽位（U位）;" json:"slotPosition"`
	NetworkInterface      string `excel:"name:区域;" json:"networkInterface"`
	BmcUserName           string `excel:"name:bmc用户名;" json:"bmcUserName"`
	BmcPassword           string `excel:"name:bmc密码;" json:"bmcPassword"`
	BmcIp                 string `excel:"name:bmc IP地址;" json:"bmcIp"`
	BmcMac                string `excel:"name:bmc mac地址;" json:"bmcMac"`
	Mask                  string `excel:"name:掩码;" json:"mask"`
	Gateway               string `excel:"name:网关;" json:"gateway"`
}
