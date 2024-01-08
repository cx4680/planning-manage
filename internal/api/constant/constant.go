package constant

const (
	XRequestID    = "x-request-id"
	CurrentUserId = "currentUserId"

	Size         = "size"
	SizeValue    = 10
	Current      = "current"
	CurrentValue = 1

	Comma  = ","
	Colon  = ":"
	Hyphen = "-"

	EnvGlobalBaseDomain = "GLOBAL_BASE_DOMAIN"
	BaseDomain          = "BASE_DOMAIN"
	EnvRegion           = "REGION"
	EnvCellID           = "CELLID"

	NameSpace = "planning-manage"

	UserSecretPrivateKeyPath = "/app/secret/userSecretPrivateKey"

	SeparationOfTwoNetworks = 2 // 两网分离
	TripleNetworkSeparation = 3 // 三网分离
	TriplePlay              = 1 // 三网合一

	Plan              = "plan"     // 待规划
	Planning          = "planning" // 规划中
	Planned           = "planned"  // 规划完成
	BusinessStart     = 0          // 业务规划开始阶段
	CloudProductConf  = 1          // 云产品配置阶段
	ServerPlan        = 2          // 服务器规划阶段
	NetworkDevicePlan = 3          // 网络设备规划阶段
	BusinessEnd       = 4          // 业务规划结束

	SplitLineBreak    = "\n"
	SplitLineAsterisk = "*"
	SplitLineColon    = ":"

	WhetherRequiredNo         = 0
	WhetherRequiredYes        = 1
	WhetherRequiredNoChinese  = "否"
	WhetherRequiredYesChinese = "是"

	ResNodeRoleType     = 0
	ControlNodeRoleType = 1

	NetworkModelNo         = 0
	NetworkModelYes        = 1
	NeedQueryOtherTable    = 2
	NetworkModelNoChinese  = "否"
	NetworkModelYesChinese = "是"

	NodeRoleType          = 0
	NetworkDeviceRoleType = 1

	NetworkDeviceTypeXinchuangCn  = "信创"
	NetworkDeviceTypeCommercialCn = "商用"
	NetworkDeviceTypeXinchuang    = 0
	NetworkDeviceTypeCommercial   = 1

	NodeRoleSupportDPDK      = 1
	NodeRoleNotSupportDPDK   = 0
	NodeRoleSupportDPDKCn    = "是"
	NodeRoleNotSupportDPDKCn = "否"

	NodeWastageCalcTypeNumCn     = "数量"
	NodeWastageCalcTypePercentCn = "百分比"
	NodeWastageCalcTypeNum       = 1
	NodeWastageCalcTypePercent   = 2

	CellTypeControl = "control"

	CabinetTypeNetworkCn  = "网络机柜"
	CabinetTypeBusinessCn = "业务机柜"
	CabinetTypeStorageCn  = "存储机柜"
	CabinetTypeNetwork    = 1
	CabinetTypeBusiness   = 2
	CabinetTypeStorage    = 3
)
