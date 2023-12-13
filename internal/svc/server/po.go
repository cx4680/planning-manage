package server

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
	Id     int64 `form:"id"`
	Number int   `form:"number"`
}

type ResponseDownloadServer struct {
	NodeRole   string `json:"nodeRole" excel:"nodeRole:角色;index:0"`
	ServerType string `json:"serverType" excel:"serverType:设备类型;index:1"`
	BomCode    string `json:"bomCode" excel:"bomCode:机型;index:2"`
	Spec       string `json:"spec" excel:"spec:规格;index:3"`
	Number     string `json:"number" excel:"number:数量;index:4"`
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
	Unit             string              `json:"unit"`             // 单位
	FeatureType      string              `json:"featureType"`      // 特性类型
	Features         []*ResponseFeatures `json:"features"`         // 特性
	Description      string              `json:"description"`      // 说明
}

type ResponseFeatures struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

var FeatureMap = map[string]string{"超分": "超分比", "三副本": "副本模式", "EC纠删码": "副本模式"}
