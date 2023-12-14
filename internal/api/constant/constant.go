package constant

const (
	XRequestID    = "x-request-id"
	CurrentUserId = "currentUserId"
)

const (
	Size         = "size"
	SizeValue    = 10
	Current      = "current"
	CurrentValue = 1
)

const (
	Comma  = ","
	Colon  = ":"
	Hyphen = "-"
)

const (
	True  = "true"
	False = "false"
)

const (
	EnvGlobalBaseDomain = "GLOBAL_BASE_DOMAIN"
	BaseDomain          = "BASE_DOMAIN"
	EnvRegion           = "REGION"
	EnvCellID           = "CELLID"
)

const (
	NameSpace = "planning-manage"
)

const (
	UserSecretPrivateKeyPath = "/app/secret/userSecretPrivateKey"
)

const (
	SEPARATION_OF_TWO_NETWORKS = 2 // 两网分离
	TRIPLE_NETWORK_SEPARATION  = 3 // 三网分离
	TRIPLE_PLAY                = 1 // 三网合一
)

const (
	MASW      = "MASW"      // 管理接入交换机
	VASW      = "VASW"      // 业务内网接入交换机
	StorSASW  = "StorSASW"  // 存储业务接入交换机
	StoreCASW = "StoreCASW" // 存储集群接入交换机
	BMSASW    = "BMSASW"    // 裸金属接入交换机
	ISW       = "ISW"       // 业务外网出口交换机
	OASW      = "OASW"      // 服务器带外接入交换机
)

const (
	PLAN                = "plan"     // 待规划
	PLANNING            = "planning" // 规划中
	PLANNED             = "planned"  // 规划完成
	BUSINESS_START      = 0          // 业务规划开始阶段
	CLOUD_PRODUCT_CONF  = 1          // 云产品配置阶段
	SERVER_PLAN         = 2          // 服务器规划阶段
	NETWORK_DEVICE_PLAN = 3          // 网络设备规划阶段
	BUSINESS_END        = 4          // 业务规划结束
)

const (
	SplitLineBreak    = "\n"
	SplitLineAsterisk = "*"
	SplitLineColon    = ":"
)

const (
	WhetherRequiredNo         = 0
	WhetherRequiredYes        = 1
	WhetherRequiredNoChinese  = "否"
	WhetherRequiredYesChinese = "是"
)

const (
	ResNodeRoleType     = 0
	ControlNodeRoleType = 1
)

const (
	NetworkModelNo         = 0
	NetworkModelYes        = 1
	NeedQueryOtherTable    = 2
	NetworkModelNoChinese  = "否"
	NetworkModelYesChinese = "是"
)

const (
	NodeRoleType          = 0
	NetworkDeviceRoleType = 1
)

const (
	NetworkDeviceTypeXinchuangCn  = "信创"
	NetworkDeviceTypeCommercialCn = "商用"
	NetworkDeviceTypeXinchuang    = 0
	NetworkDeviceTypeCommercial   = 1
)

const (
	NodeRoleSupportDPDK      = 1
	NodeRoleNotSupportDPDK   = 0
	NodeRoleSupportDPDKCn    = "是"
	NodeRoleNotSupportDPDKCn = "否"
)

const (
	NodeWastageCalcTypeNumCn     = "数量"
	NodeWastageCalcTypePercentCn = "百分比"
	NodeWastageCalcTypeNum       = 1
	NodeWastageCalcTypePercent   = 2
)
