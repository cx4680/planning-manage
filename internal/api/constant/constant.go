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

	ProjectStagePlanning  = "planning"  //规划阶段
	ProjectStageDelivery  = "delivery"  //交付阶段
	ProjectStageDelivered = "delivered" //已交付

	PlanStagePlan       = "plan"       // 待规划
	PlanStagePlanning   = "planning"   // 规划中
	PlanStagePlanned    = "planned"    // 规划完成
	PlanStageDelivering = "delivering" //交付中
	PlanStageDelivered  = "delivered"  //交付完成

	BusinessPlanningStart         = 0 // 业务规划开始阶段
	BusinessPlanningCloudProduct  = 1 // 业务规划-云产品配置阶段
	BusinessPlanningServer        = 2 // 业务规划-服务器规划阶段
	BusinessPlanningNetworkDevice = 3 // 业务规划-网络设备规划阶段
	BusinessPlanningEnd           = 4 // 业务规划结束

	DeliverPlanningStart               = 0 // 交付规划开始阶段
	DeliverPlanningMachineRoom         = 1 // 交付规划-机房规划
	DeliverPlanningNetworkDevice       = 2 // 交付规划-网络设备上架
	DeliverPlanningServer              = 3 // 交付规划-服务器上架
	DeliverPlanningIp                  = 4 // 交付规划-IP规划
	DeliverPlanningGlobalConfiguration = 5 // 交付规划-全局配置
	DeliverPlanningEnd                 = 6 // 交付规划-全局配置

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

	YesCn = "是"
	Yes   = 1
	NoCn  = "否"
	No    = 0

	Ipv6YesCn = "是"
	Ipv6Yes   = "1"
	Ipv6NoCn  = "否"
	Ipv6No    = "0"

	NetworkModeStandardCn = "标准模式"
	NetworkModeStandard   = 0
	NetworkMode2NetworkCn = "纯二层组网模式"
	NetworkMode2Network   = 1

	RegionTypeCode = "regionType"
	CellTypeCode   = "cellType"

	IpDemandNetworkTypeIpv4Cn = "ipv4"
	IpDemandNetworkTypeIpv4   = 0
	IpDemandNetworkTypeIpv6Cn = "ipv6"
	IpDemandNetworkTypeIpv6   = 1
)
