package capacity_planning

type Request struct {
	Id                 int64
	PlanId             int64                    `form:"planId"`
	NetworkInterface   string                   `form:"networkInterface"`
	CpuType            string                   `form:"cpuType"`
	ServerList         []*RequestServer         `form:"serverList"`
	ServerCapacityList []*RequestServerCapacity `form:"serverCapacityList"`
	EcsCapacity        *EcsCapacity             `form:"ecsCapacity"`
	UserId             string
}

type RequestServer struct {
	NodeRoleId         int64  `form:"nodeRoleId"`
	MixedNodeRoleId    int64  `form:"mixedNodeRoleId"`
	ServerBaselineId   int64  `form:"serverBaselineId"`
	Number             int    `form:"number"`
	OpenDpdk           int    `form:"openDpdk"`
	BusinessAttributes string `form:"businessAttributes"` // 业务属性
	ShelveMode         string `form:"shelveMode"`         // 上架模式
	ShelvePriority     int    `form:"shelvePriority"`     // 上架优先级
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

type ResponseCapClassification struct {
	Classification string                `json:"classification"` // 分类
	ProductName    string                `json:"productName"`    // 产品名称
	ProductCode    string                `json:"productCode"`    // 产品编码
	CapConvert     []*ResponseCapConvert `json:"capConvert"`
	Special        *EcsCapacity          `json:"special"`
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

type EcsCapacity struct {
	CapacityIdList []int64 `form:"capacityIdList" json:"capacityIdList"`
	FeatureNumber  int     `form:"featureNumber" json:"featureNumber"`
	List           []*struct {
		CpuNumber    int `form:"cpuNumber" json:"cpuNumber"`
		MemoryNumber int `form:"memoryNumber" json:"memoryNumber"`
		Count        int `form:"count" json:"count"`
	} `json:"list"`
}
