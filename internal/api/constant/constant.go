package constant

const (
	XRequestID    = "x-request-id"
	CurrentUserId = "currentUserId"

	Size         = "size"
	SizeValue    = 10
	Current      = "current"
	CurrentValue = 1

	Comma        = ","
	Colon        = ":"
	Hyphen       = "-"
	ForwardSlash = "/"

	EnvGlobalBaseDomain = "GLOBAL_BASE_DOMAIN"
	BaseDomain          = "BASE_DOMAIN"
	EnvRegion           = "REGION"
	EnvCellID           = "CELLID"

	UserCenterUrl       = "USER_CENTER_URL"
	ProductCode         = "PRODUCT_CODE"
	UserCenterSecretKey = "USER_CENTER_SECRET_KEY"
	FrontUrl            = "FRONT_URL"

	NameSpace = "planning-manage"

	UserSecretPrivateKeyPath = "/app/secret/userSecretPrivateKey"

	SeparationOfTwoNetworks   = 2      // 两网分离
	TripleNetworkSeparation   = 3      // 三网分离
	TriplePlay                = 1      // 三网合一
	SeparationOfTwoNetworksCn = "两网分离" // 两网分离
	TripleNetworkSeparationCn = "三网分离" // 三网分离
	TriplePlayCn              = "三网合一" // 三网合一

	ProjectStagePlanning  = "planning"  // 规划阶段
	ProjectStageDelivery  = "delivery"  // 交付阶段
	ProjectStageDelivered = "delivered" // 已交付

	PlanStagePlan       = "plan"       // 待规划
	PlanStagePlanning   = "planning"   // 规划中
	PlanStagePlanned    = "planned"    // 规划完成
	PlanStageDelivering = "delivering" // 交付中
	PlanStageDelivered  = "delivered"  // 交付完成

	BusinessPlanningStart         = 0 // 业务规划开始
	BusinessPlanningCloudProduct  = 1 // 业务规划-云产品配置阶段
	BusinessPlanningServer        = 2 // 业务规划-服务器规划阶段
	BusinessPlanningNetworkDevice = 3 // 业务规划-网络设备规划阶段
	BusinessPlanningEnd           = 4 // 业务规划结束

	DeliverPlanningStart               = 0 // 交付规划开始
	DeliverPlanningMachineRoom         = 1 // 交付规划-机房规划
	DeliverPlanningNetworkDevice       = 2 // 交付规划-网络设备上架
	DeliverPlanningServer              = 3 // 交付规划-服务器上架
	DeliverPlanningIp                  = 4 // 交付规划-IP规划
	DeliverPlanningGlobalConfiguration = 5 // 交付规划-全局配置
	DeliverPlanningEnd                 = 6 // 交付规划结束

	General   = "general"   // 普通方案
	Alternate = "alternate" // 备选方案
	Delivery  = "delivery"  // 交付方案

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

	AccessNetworkDeviceRoleKeyword = "接入交换机"

	NetworkDevicePxeVlanId      = "89"
	NetworkDeviceManageVlanId   = "91"
	NetworkDeviceBizVlanId      = "92"
	NetworkDeviceStorageVlanId  = "909"
	ServerManageNetworkIpVlanId = "91"
	ServerBizIntranetIpVlanId   = "92"
	ServerStorageIpVlanId       = "909"

	NodeRoleDeployMethodInCluster = "inCluster"

	KernelArchArm = "aarch64"
	KernelArchX86 = "x86_64"

	CpuArchX86 = "x86"
	CpuArchARM = "ARM"
	CpuArchXC  = "xc"

	CpuTypeIntel = "intel"
	CpuTypeHygon = "hygon"

	ProductCodeCKE        = "CKE"
	ProductCodeECS        = "ECS"
	ProductCodeBMS        = "BMS"
	ProductCodeCBR        = "CBR"
	ProductCodeEBS        = "EBS"
	ProductCodeEFS        = "EFS"
	ProductCodeOSS        = "OSS"
	ProductCodeVPC        = "VPC"
	ProductCodeCNFW       = "CNFW"
	ProductCodeCWAF       = "CWAF"
	ProductCodeCSOC       = "CSOC"
	ProductCodeDSP        = "DSP"
	ProductCodeCNBH       = "CNBH"
	ProductCodeCWP        = "CWP"
	ProductCodeDES        = "DES"
	ProductCodeCEASQLTX   = "CEASQLTX"
	ProductCodeMYSQL      = "MYSQL"
	ProductCodeCEASQLDW   = "CEASQLDW"
	ProductCodeCEASQLCK   = "CEASQLCK"
	ProductCodeDTS        = "DTS"
	ProductCodeREDIS      = "REDIS"
	ProductCodePOSTGRESQL = "POSTGRESQL"
	ProductCodeKAFKA      = "KAFKA"
	ProductCodeCSP        = "CSP"
	ProductCodeROCKETMQ   = "ROCKETMQ"
	ProductCodeRABBITMQ   = "RABBITMQ"
	ProductCodeAPIM       = "APIM"
	ProductCodeCONNECT    = "CONNECT"
	ProductCodeCLCP       = "CLCP"
	ProductCodeCOS        = "COS"
	ProductCodeCLS        = "CLS"

	CapPlanningInputVCpu                      = "vCPU"
	CapPlanningInputMemory                    = "内存"
	CapPlanningInputContainerCluster          = "容器集群数"
	CapPlanningInputStorageCapacity           = "存储容量"
	CapPlanningInputFirewall                  = "防火墙数量"
	CapPlanningInputLogStorage                = "日志存储空间"
	CapPlanningInputAssetAccess               = "资产接入授权"
	CapPlanningInputVulnerabilityScanning     = "漏洞扫描服务数"
	CapPlanningInputDatabaseAudit             = "数据库审计"
	CapPlanningInputBusinessDataVolume        = "业务数据量"
	CapPlanningInputLinks                     = "链路数量"
	CapPlanningInputBroker                    = "broker节点数"
	CapPlanningInputDiskCapacity              = "磁盘容量"
	CapPlanningInputStandardEdition           = "标准版"
	CapPlanningInputProfessionalEdition       = "专业版"
	CapPlanningInputEnterpriseEdition         = "企业版"
	CapPlanningInputPlatinumEdition           = "铂金版"
	CapPlanningInputMicroservice              = "微服务实例"
	CapPlanningInputComputingResourceCapacity = "计算资源容量"
	CapPlanningInputApplications              = "计算资源容量"

	CapPlanningInputOneHundred     = "100"
	CapPlanningInputOneHundredInt  = 100
	CapPlanningInputFiveHundred    = "500"
	CapPlanningInputFiveHundredInt = 500
	CapPlanningInputOneThousand    = "1000"
	CapPlanningInputOneThousandInt = 1000

	ExpendResCodeECSVCpu   = "ECS_VCPU"
	ExpendResCodeECSMemory = "ECS_MEM"

	NodeRoleCodeCompute  = "COMPUTE"
	NodeRoleCodeNETWORK  = "NETWORK"
	NodeRoleCodeNFV      = "NFV"
	NodeRoleCodeBMSGW    = "BMSGW"
	NodeRoleCodeDATABASE = "DATABASE"

	SellSpecsStandardEdition = "标准版"
	SellSpecsUltimateEdition = "旗舰版"

	FeaturesNameOverallocation = "超分比"
	FeaturesNameThreeCopies    = "三副本"
	FeaturesNameEC             = "EC纠删码"

	SoftwareBomLicense     = "License"
	SoftwareBomMaintenance = "升级维保"

	SoftwareBomAuthorizedUnitAssetAccess = "资产接入"
	SoftwareBomAuthorizedUnitLogStorage  = "日志存储空间"

	SoftwareBomValueAddedServiceVulnerabilityScanning = "漏洞扫描"

	SoftwareBomAuthorizedUnit500G = "500G"
)
