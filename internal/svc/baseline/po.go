package baseline

const (
	CloudProductBaselineSheetName      = "云产品售卖清单"
	ServerBaselineSheetName            = "服务器基线"
	NetworkDeviceBaselineSheetName     = "网络设备基线"
	NetworkDeviceRoleBaselineSheetName = "网络设备角色基线"
	NodeRoleBaselineSheetName          = "节点角色基线"
	IPDemandBaselineSheetName          = "IP需求规划"
	CapConvertBaselineSheetName        = "云服务容量换算"
	CapActualResBaselineSheetName      = "云服务容量实际资源消耗"
	CapServerCalcBaselineSheetName     = "服务器数量计算"
)

const (
	CloudProductBaselineType      = "cloudProductListBaseline"
	ServerBaselineType            = "serverBaseline"
	NetworkDeviceBaselineType     = "networkDeviceBaseline"
	NetworkDeviceRoleBaselineType = "networkDeviceRoleBaseline"
	NodeRoleBaselineType          = "nodeRoleBaseline"
	IPDemandBaselineType          = "ipDemandBaseline"
	CapConvertBaselineType        = "capConvertBaseline"
	CapActualResBaselineType      = "capActualResBaseline"
	CapServerCalcBaselineType     = "capServerCalcBaseline"
)

type CloudProductBaselineExcel struct {
	ProductType             string   `excel:"name:云服务类型;" json:"productType"`             // 产品类型
	ProductName             string   `excel:"name:云服务;" json:"productName"`               // 产品名称
	ProductCode             string   `excel:"name:服务编码;" json:"productCode"`              // 产品编码
	SellSpecs               string   `excel:"name:售卖规格;" json:"sellSpecs"`                // 售卖规格
	AuthorizedUnit          string   `excel:"name:授权单元;" json:"authorizedUnit"`           // 授权单元
	WhetherRequired         string   `excel:"name:是否必选;" json:"whetherRequired"`          // 是否必选，0：否，1：是
	Instructions            string   `excel:"name:说明;" json:"instructions"`               // 说明
	DependProductCode       string   `excel:"name:依赖服务编码;" json:"dependProductCode"`      // 依赖产品Code
	ControlResNodeRole      string   `excel:"name:管控资源节点角色;" json:"controlResNodeRole"`   // 管控资源节点角色
	ResNodeRole             string   `excel:"name:资源节点角色;" json:"resNodeRole"`            // 资源节点角色
	ControlResNodeRoleCode  string   `excel:"name:管控角色编码;" json:"controlResNodeRoleCode"` // 管控资源节点角色编码
	ResNodeRoleCode         string   `excel:"name:资源角色编码;" json:"resNodeRoleCode"`        // 资源节点角色编码
	DependProductCodes      []string `json:"dependProductCodes"`                          // 依赖产品Code数组
	ControlResNodeRoleCodes []string `json:"controlResNodeRoleCodes"`                     // 管控资源节点角色编码数组
	ResNodeRoleCodes        []string `json:"resNodeRoleCodes"`                            // 资源节点角色编码数组
}

type NodeRoleBaselineExcel struct {
	NodeRoleCode string   `excel:"name:角色code;" json:"nodeRoleCode"`  // 角色code
	NodeRoleName string   `excel:"name:角色名称;" json:"nodeRoleName"`    // 角色名称
	MinimumCount int      `excel:"name:最小部署数量;" json:"minimumCount"`  // 单独部署最小数量
	DeployMethod string   `excel:"name:部署方式;" json:"deployMethod"`    // 部署方式
	SupportDPDK  string   `excel:"name:是否支持DPDK;" json:"supportDPDK"` // 是否支持DPDK，0：否，1：是
	Classify     string   `excel:"name:分类;" json:"classify"`          // 分类
	MixedDeploy  string   `excel:"name:节点混部;" json:"mixedDeploy"`     // 节点混部
	Annotation   string   `excel:"name:说明;" json:"annotation"`        // 说明
	BusinessType string   `excel:"name:业务类型;" json:"businessType"`    // 业务类型
	MixedDeploys []string `json:"mixedDeploys"`                       // 节点混部数组
}

type ServerBaselineExcel struct {
	Arch                string   `excel:"name:硬件架构;" json:"Arch"`                      // 硬件架构
	NetworkInterface    string   `excel:"name:网络接口;" json:"networkInterface"`          // 网络接口
	NodeRoleName        string   `excel:"name:节点角色;" json:"nodeRole"`                  // 节点角色名称
	NodeRoleCode        string   `excel:"name:节点角色编码;" json:"nodeRoleCode"`            // 节点角色编码
	BomCode             string   `excel:"name:BOM编码;" json:"bomCode"`                  // BOM编码
	ConfigurationInfo   string   `excel:"name:配置概要;" json:"configurationInfo"`         // 配置概要
	Spec                string   `excel:"name:规格;" json:"spec"`                        // 规格
	CpuType             string   `excel:"name:CPU类型;" json:"cpuType"`                  // CPU类型
	Cpu                 int      `excel:"name:vCPU;" json:"cpu"`                       // CPU核数
	Gpu                 string   `excel:"name:GPU;" json:"gpu"`                        // GPU
	Memory              int      `excel:"name:内存;" json:"memory"`                      // 内存
	SystemDiskType      string   `excel:"name:系统盘类型;" json:"systemDiskType"`           // 系统盘类型
	SystemDisk          string   `excel:"name:系统盘;" json:"systemDisk"`                 // 系统盘
	StorageDiskType     string   `excel:"name:存储盘类型;" json:"storageDiskType"`          // 存储盘类型
	StorageDiskNum      int      `excel:"name:存储盘个数;" json:"storageDiskNum"`           // 存储盘个数
	StorageDiskCapacity int      `excel:"name:存储盘单盘容量（G）;" json:"storageDiskCapacity"` // 存储盘单盘容量（G）
	RamDisk             string   `excel:"name:缓存盘;" json:"ramDisk"`                    // 缓存盘
	NetworkCardNum      int      `excel:"name:网卡数量;" json:"networkCardNum"`            // 网卡数量
	Power               int      `excel:"name:功率（W）;" json:"power"`                    // 功率
	NodeRoleCodes       []string `json:"nodeRoleCodes"`                                // 节点角色编码数组
}

type NetworkDeviceRoleBaselineExcel struct {
	DeviceType       string   `excel:"name:设备类型;" json:"deviceType"`      // 设备类型
	FuncType         string   `excel:"name:类型;" json:"funcType"`          // 类型
	FuncCompoName    string   `excel:"name:功能组件;" json:"funcCompoName"`   // 功能组件
	FuncCompoCode    string   `excel:"name:功能组件命名;" json:"funcCompoCode"` // 功能组件命名
	Description      string   `excel:"name:描述;" json:"description"`       // 描述
	TwoNetworkIso    string   `excel:"name:两网分离;" json:"twoNetworkIso"`   // 两网分离
	ThreeNetworkIso  string   `excel:"name:三网分离;" json:"threeNetworkIso"` // 三网分离
	TriplePlay       string   `excel:"name:三网合一;" json:"triplePlay"`      // 三网合一
	MinimumNumUnit   int      `excel:"name:最小单元数;" json:"minimumNumUnit"` // 最小单元数
	UnitDeviceNum    int      `excel:"name:单元设备数量;" json:"unitDeviceNum"` // 单元设备数量
	DesignSpec       string   `excel:"name:设计规格;" json:"designSpec"`      // 设计规格
	TwoNetworkIsos   []string `json:"twoNetworkIsos"`                     // 两网分离数组
	ThreeNetworkIsos []string `json:"threeNetworkIsos"`                   // 三网分离数组
	TriplePlays      []string `json:"triplePlays"`                        // 三网合一数组
}

type NetworkDeviceBaselineExcel struct {
	DeviceModel            string   `excel:"name:设备型号;" json:"deviceModel"`             // 设备型号
	Manufacturer           string   `excel:"name:厂商;" json:"manufacturer"`              // 厂商
	DeviceType             string   `excel:"name:信创/商用;" json:"deviceType"`             // 信创/商用
	NetworkModel           string   `excel:"name:网络模型;" json:"networkModel"`            // 网络模型
	ConfOverview           string   `excel:"name:配置概述;" json:"confOverview"`            // 配置概述
	NetworkDeviceRoleCode  string   `excel:"name:功能组件命名;" json:"networkDeviceRoleCode"` // 功能组件命名
	Purpose                string   `excel:"name:网络设备角色（备注）;" json:"purpose"`           // 用途
	NetworkDeviceRoleCodes []string `json:"networkDeviceRoleCodes"`                     // 网络设备角色编码数组
}

type IPDemandBaselineExcel struct {
	Vlan                   string   `excel:"name:Vlan Id;" json:"vlan"`                // vlan id
	Explain                string   `excel:"name:说明;" json:"explain"`                  // 说明
	NetworkType            string   `excel:"name:网络类型;" json:"networkType"`            // 网络类型
	Description            string   `excel:"name:描述;" json:"description"`              // 描述
	IPSuggestion           string   `excel:"name:IP地址规划建议;" json:"IPSuggestion"`       // IP地址规划建议
	NetworkDeviceRoleCode  string   `excel:"name:关联设备组;" json:"networkDeviceRoleCode"` // 关联设备组
	AssignNum              string   `excel:"name:数量（C）;" json:"assignNum"`             // 分配数量
	Remark                 string   `excel:"name:备注;" json:"remark"`                   // 备注
	NetworkDeviceRoleCodes []string `json:"networkDeviceRoleCodes"`                    // 关联设备组数组
}

type CapConvertBaselineExcel struct {
	ProductName      string `excel:"name:云服务;" json:"productName"`         // 产品名称
	ProductCode      string `excel:"name:服务编码;" json:"productCode"`        // 产品编码
	SellSpecs        string `excel:"name:售卖规格;" json:"sellSpecs"`          // 售卖规格
	CapPlanningInput string `excel:"name:容量规划输入;" json:"capPlanningInput"` // 容量规划输入
	Unit             string `excel:"name:单位;" json:"unit"`                 // 单位
	Features         string `excel:"name:特性;" json:"features"`             // 特性
	Description      string `excel:"name:说明;" json:"description"`          // 说明
}

type CapActualResBaselineExcel struct {
	ProductCode   string `excel:"name:服务编码;" json:"productCode"`     // 产品编码
	SellSpecs     string `excel:"name:售卖规格;" json:"sellSpecs"`       // 售卖规格
	SellUnit      string `excel:"name:售卖单元;" json:"sellUnit"`        // 售卖单元
	ExpendRes     string `excel:"name:消耗资源;" json:"expendRes"`       // 消耗资源
	ExpendResCode string `excel:"name:消耗资源编码;" json:"expendResCode"` // 消耗资源编码
	Features      string `excel:"name:特性;" json:"features"`          // 特性
	OccRatio      string `excel:"name:占用比例;" json:"occRatio"`        // 占用比例
	Remarks       string `excel:"name:备注;" json:"remarks"`           // 备注
}

type CapServerCalcBaselineExcel struct {
	ExpendRes           string `excel:"name:消耗资源;" json:"expendRes"`               // 消耗资源
	ExpendResCode       string `excel:"name:消耗资源编码;" json:"expendResCode"`         // 消耗资源编码
	ExpendNodeRoleCode  string `excel:"name:消耗节点角色编码;" json:"expendNodeRoleCode"`  // 消耗节点角色编码
	OccNodeRes          string `excel:"name:占用节点资源;" json:"occNodeRes"`            // 占用节点资源
	OccNodeResCode      string `excel:"name:占用节点资源编码;" json:"occNodeResCode"`      // 占用节点资源编码
	NodeWastage         string `excel:"name:节点损耗;" json:"nodeWastage"`             // 节点损耗
	NodeWastageCalcType string `excel:"name:节点损耗计算类型;" json:"nodeWastageCalcType"` // 节点损耗计算类型，1：数量，2：百分比
	WaterLevel          string `excel:"name:水位;" json:"waterLevel"`                // 水位
}
