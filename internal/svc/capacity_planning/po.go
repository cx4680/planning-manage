package capacity_planning

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
)

type Request struct {
	Id                 int64
	PlanId             int64                         `form:"planId"`
	NetworkInterface   string                        `form:"networkInterface"`
	CpuType            string                        `form:"cpuType"`
	ServerList         []*RequestServer              `form:"serverList"`
	ServerCapacityList []*ResourcePoolServerCapacity `form:"serverCapacityList"`
	UserId             string
}

type RequestServer struct {
	NodeRoleId         int64  `form:"nodeRoleId"`
	MixedNodeRoleId    int64  `form:"mixedNodeRoleId"`
	ServerBaselineId   int64  `form:"serverBaselineId"`
	Number             int    `form:"number"`
	OpenDpdk           int    `form:"openDpdk"`
	ResourcePoolId     int64  `form:"resourcePoolId"`
	ResourcePoolName   string `form:"resourcePoolName"`
	EditDpdk           int    `form:"editDpdk"`
	BusinessAttributes string `form:"businessAttributes"` // 业务属性
	ShelveMode         string `form:"shelveMode"`         // 上架模式
	ShelvePriority     int    `form:"shelvePriority"`     // 上架优先级
}

type ResourcePoolServerCapacity struct {
	ResourcePoolId           int64                    `form:"resourcePoolId"`
	CommonServerCapacityList []*RequestServerCapacity `form:"commonServerCapacityList"`
	EcsCapacity              *EcsCapacity             `form:"ecsCapacity"`
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
	ResourcePoolId   int64 `form:"resourcePoolId"`
}

type ResponseCapClassification struct {
	Classification          string                    `json:"classification"` // 分类
	ProductName             string                    `json:"productName"`    // 产品名称
	ProductCode             string                    `json:"productCode"`    // 产品编码
	ProductType             string                    `json:"productType"`    // 产品类型
	ResourcePoolCapConverts []*ResourcePoolCapConvert `json:"resourcePoolCapConverts"`
	ResourcePoolList        []*entity.ResourcePool    `json:"resourcePoolList"`
}

type ResponseCapCount struct {
	Number int `form:"number"`
}

type ResourcePoolCapConvert struct {
	ResourcePoolId      int64                 `json:"resourcePoolId"`   // 资源池id
	ResourcePoolName    string                `json:"resourcePoolName"` // 资源池名称
	ResponseCapConverts []*ResponseCapConvert `json:"capConverts"`      // 特性选项
	Specials            *EcsCapacity          `json:"specials"`
}

type ResponseCapConvert struct {
	VersionId        int64               `json:"versionId"`        // 版本id
	ProductName      string              `json:"productName"`      // 产品名称
	ProductCode      string              `json:"productCode"`      // 产品编码
	ProductType      string              `json:"productType"`      // 产品分类
	SellSpecs        string              `json:"sellSpecs"`        // 售卖规格
	CapPlanningInput string              `json:"capPlanningInput"` // 容量规划输入
	Number           int                 `json:"number"`           // 数量
	Unit             string              `json:"unit"`             // 单位
	FeatureId        int64               `json:"featureId"`        // 特性id
	FeatureMode      string              `json:"featureMode"`      // 特性模式
	FeatureNumber    int                 `json:"featureNumber"`    // 特性数量
	Features         []*ResponseFeatures `json:"features"`         // 特性选项
	Description      string              `json:"description"`      // 说明
	ResourcePoolId   int64               `json:"resourcePoolId"`   // 资源池id
	ResourcePoolName string              `json:"resourcePoolName"` // 资源池名称
}

type ResponseFeatures struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type EcsCapacity struct {
	CapacityIdList []int64     `form:"capacityIdList" json:"capacityIdList"`
	FeatureNumber  int         `form:"featureNumber" json:"featureNumber"`
	List           []*EcsSpecs `json:"list"`
}

type EcsSpecs struct {
	CpuNumber    int `form:"cpuNumber" json:"cpuNumber"`
	MemoryNumber int `form:"memoryNumber" json:"memoryNumber"`
	Count        int `form:"count" json:"count"`
}

type ExpendResFeature struct {
	CapActualResBaseline entity.CapActualResBaseline `json:"capActualResBaseline"`
	FeatureNumber        int                         `json:"featureNumber"`
}
