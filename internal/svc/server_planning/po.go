package server_planning

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/capacity_planning"
)

type Request struct {
	Id                 int64
	PlanId             int64                          `form:"planId"`
	NetworkInterface   string                         `form:"networkInterface"`
	CpuType            string                         `form:"cpuType"`
	ServerList         []*RequestServer               `form:"serverList"`
	ServerCapacityList []*RequestServerCapacity       `form:"serverCapacityList"`
	EcsCapacity        *capacity_planning.EcsCapacity `form:"ecsCapacity"`
	UserId             string
}

type RequestServer struct {
	NodeRoleId         int64  `form:"nodeRoleId"`
	MixedNodeRoleId    int64  `form:"mixedNodeRoleId"`
	ServerBaselineId   int64  `form:"serverBaselineId"`
	Number             int    `form:"number"`
	OpenDpdk           int    `form:"openDpdk"`
	ResourcePoolId     int64  `form:"resourcePoolId"`
	BusinessAttributes string `form:"businessAttributes"` // 业务属性
	ShelveMode         string `form:"shelveMode"`         // 上架模式
	ShelvePriority     int    `form:"shelvePriority"`     // 上架优先级
}

type RequestServerCapacity struct {
	Id             int64 `form:"id"`
	Number         int   `form:"number"`
	FeatureNumber  int   `form:"featureNumber"`
	ResourcePoolId int64 `form:"resourcePoolId"`
}

type RequestServerCapacityCount struct {
	PlanId           int64 `form:"planId"`
	NodeRoleId       int64 `form:"nodeRoleId"`
	ServerBaselineId int64 `form:"serverBaselineId"`
}

type Server struct {
	entity.ServerPlanning
	NodeRoleName             string           `gorm:"-" json:"nodeRoleName"`             // 节点角色名称
	NodeRoleClassify         string           `gorm:"-" json:"nodeRoleClassify"`         // 节点角色分类
	NodeRoleAnnotation       string           `gorm:"-" json:"nodeRoleAnnotation"`       // 节点说明
	SupportDpdk              int              `gorm:"-" json:"supportDpdk"`              // 是否支持DPDK, 0:否，1：是
	ServerBomCode            string           `gorm:"-" json:"serverBomCode"`            // BOM编码
	ServerArch               string           `gorm:"-" json:"serverArch"`               // 架构
	ServerBaselineList       []*Baseline      `gorm:"-" json:"serverBaselineList"`       // 可选择机型列表
	MixedNodeRoleList        []*MixedNodeRole `gorm:"-" json:"mixedNodeRoleList"`        // 可混合部署角色列表
	Upload                   int              `gorm:"-" json:"upload"`                   // 是否已上传
	ResourcePoolName         string           `gorm:"-" json:"resourcePoolName"`         // 资源池名称
	SupportMultiResourcePool int              `gorm:"-" json:"supportMultiResourcePool"` // 是否支持多资源池
	DefaultResourcePool      int              `gorm:"-" json:"defaultResourcePool"`      // 是否为默认资源池
	// EditDpdk           int              `gorm:"-" json:"editDpdk"`           // 是否可编辑DPDK, 0:可编辑，1：不可编辑
}

type MixedNodeRole struct {
	Id                  int64                `gorm:"-" json:"id"`                  // 混合节点角色id
	Name                string               `gorm:"-" json:"name"`                // 混合节点角色名称
	MixResourcePoolList []*MixedResourcePool `gorm:"-" json:"mixResourcePoolList"` // 混合部署资源池列表
}

type MixedResourcePool struct {
	ResourcePoolId   int64  `gorm:"-" json:"resourcePoolId"`   // 混合部署资源池id
	ResourcePoolName string `gorm:"-" json:"resourcePoolName"` // 混合部署资源池名称
}

type Baseline struct {
	Id                  int64  `gorm:"-" json:"id"`                  // 服务器id
	BomCode             string `gorm:"-" json:"bomCode"`             // BOM编码
	NetworkInterface    string `gorm:"-" json:"networkInterface"`    // 网络类型
	CpuType             string `gorm:"-" json:"cpuType"`             // cpu类型
	Cpu                 int    `gorm:"-" json:"cpu"`                 // cpu损耗
	Memory              int    `gorm:"-" json:"memory"`              // 内存损耗
	StorageDiskNum      int    `gorm:"-" json:"storageDiskNum"`      // 存储盘数量
	StorageDiskCapacity int    `gorm:"-" json:"storageDiskCapacity"` // 存储盘单盘容量（G）
	Arch                string `gorm:"-" json:"arch"`                // 硬件架构
	ConfigurationInfo   string `gorm:"-" json:"configurationInfo"`   // 配置概要
}

type ResponseDownloadServer struct {
	NodeRole   string `json:"nodeRole" excel:"name:角色;"`
	ServerType string `json:"serverType" excel:"name:设备类型;"`
	BomCode    string `json:"bomCode" excel:"name:机型;"`
	Spec       string `json:"spec" excel:"name:规格;"`
	Number     string `json:"number" excel:"name:数量;"`
}

type ResponseCapClassification struct {
	Classification string                         `json:"classification"` // 分类
	ProductName    string                         `json:"productName"`    // 产品名称
	ProductCode    string                         `json:"productCode"`    // 产品编码
	CapConvert     []*ResponseCapConvert          `json:"capConvert"`
	Special        *capacity_planning.EcsCapacity `json:"special"`
}

type ResponseCapCount struct {
	Number int `form:"number"`
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
	FeatureMode      string              `json:"featureMode"`      // 特性模式
	FeatureNumber    int                 `json:"featureNumber"`    // 特性数量
	Features         []*ResponseFeatures `json:"features"`         // 特性选项
	Description      string              `json:"description"`      // 说明
}

type ResponseFeatures struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type ResponseServerShelve struct {
	*entity.ServerShelve
	NodeRoleName string `gorm:"-" json:"nodeRoleName"` // 节点角色名称
}

type ShelveDownload struct {
	SortNumber            int    `excel:"name:序号;" json:"sortNumber"`
	NodeRoleName          string `excel:"name:节点角色;" json:"nodeRoleName"`
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

type Cabinet struct {
	*entity.CabinetInfo
	CabinetLocation string `json:"cabinetLocation"`
	IdleSlot        string `json:"idleSlot"`
}
