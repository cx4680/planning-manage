package server

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
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

type RequestServerCapacityCount struct {
	PlanId           int64 `form:"planId"`
	NodeRoleId       int64 `form:"nodeRoleId"`
	ServerBaselineId int64 `form:"serverBaselineId"`
}

type Server struct {
	entity.ServerPlanning
	NodeRoleName       string           `gorm:"-" json:"nodeRoleName"`       // 节点角色名称
	NodeRoleClassify   string           `gorm:"-" json:"nodeRoleClassify"`   // 节点角色分类
	NodeRoleAnnotation string           `gorm:"-" json:"nodeRoleAnnotation"` // 节点说明
	SupportDpdk        int              `gorm:"-" json:"supportDpdk"`        // 是否支持DPDK, 0:否，1：是
	ServerBomCode      string           `gorm:"-" json:"serverBomCode"`      // BOM编码
	ServerArch         string           `gorm:"-" json:"serverArch"`         // 架构
	ServerBaselineList []*Baseline      `gorm:"-" json:"serverBaselineList"` // 可选择机型列表
	MixedNodeRoleList  []*MixedNodeRole `gorm:"-" json:"mixedNodeRoleList"`  // 可混合部署角色列表
	Upload             int              `gorm:"-" json:"upload"`             // 是否已上传
}

type MixedNodeRole struct {
	Id   int64  `gorm:"-" json:"id"`   // 混合节点角色id
	Name string `gorm:"-" json:"name"` // 混合节点角色名称

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
	Classification string                `json:"classification"` // 分类
	CapConvert     []*ResponseCapConvert `json:"capConvert"`
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
