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
	Underline    = "_"

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

	NodeWastageCalcTypeNumCn         = "数量"
	NodeWastageCalcTypePercentCn     = "百分比"
	NodeWastageCalcTypeDataDiskNumCn = "数据盘数量"
	NodeWastageCalcTypeNum           = 1
	NodeWastageCalcTypePercent       = 2
	NodeWastageCalcTypeDataDiskNum   = 3

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

	NetworkInterface10GE = "10GE"
	NetworkInterface25GE = "25GE"

	ProductCodeCKE        = "CKE"
	ProductCodeECS        = "ECS"
	ProductCodeBMS        = "BMS"
	ProductCodeCCR        = "CCR"
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
	ProductCodeBGW        = "BGW"
	ProductCodeNLB        = "NLB"
	ProductCodeSLB        = "SLB"
	ProductCodeMONGODB    = "MONGODB"
	ProductCodeINFLUXDB   = "INFLUXDB"
	ProductCodeES         = "ES"
	ProductCodeCIK        = "CIK"
	ProductCodeRDSDM      = "RDSDM"
	ProductCodeHPC        = "HPC"

	CapPlanningInputVCpu                   = "vCPU"
	CapPlanningInputMemory                 = "内存"
	CapPlanningInputContainerCluster       = "容器集群数"
	CapPlanningInputStorageCapacity        = "存储容量"
	CapPlanningInputFirewall               = "防火墙数量"
	CapPlanningInputLogStorage             = "日志存储容量"
	CapPlanningInputAssetAccess            = "资产接入授权"
	CapPlanningInputVulnerabilityScanning  = "漏洞扫描服务数"
	CapPlanningInputDatabaseAudit          = "数据库审计"
	CapPlanningInputBusinessDataVolume     = "业务数据量"
	CapPlanningInputLinks                  = "链路数量"
	CapPlanningInputBroker                 = "broker节点数"
	CapPlanningInputDiskCapacity           = "磁盘容量"
	CapPlanningInputStandardEdition        = "标准版"
	CapPlanningInputProfessionalEdition    = "专业版"
	CapPlanningInputEnterpriseEdition      = "企业版"
	CapPlanningInputPlatinumEdition        = "铂金版"
	CapPlanningInputMicroservice           = "微服务实例"
	CapPlanningInputMonitoringNode         = "监控节点数量"
	CapPlanningInputInstances              = "实例数量"
	CapPlanningInputSingleInstanceCapacity = "单实例存储容量"
	CapPlanningInputBackupDataCapacity     = "备份数据容量"
	CapPlanningInputBasicType              = "基础型"
	CapPlanningInputStandardType           = "标准型"
	CapPlanningInputHighOrderType          = "高阶型"
	CapPlanningInputOverallocation         = "超分比"
	CapPlanningInputNetworkNLB             = "网络NLB数"
	CapPlanningInputVCpuTotal              = "vCPU总量"
	CapPlanningInputMemTotal               = "内存量"
	CapPlanningInputCopy                   = "副本数"
	CapPlanningInputSmall                  = "small"
	CapPlanningInputMiddle                 = "middle"
	CapPlanningInputLarge                  = "large"
	CapPlanningInputAgent                  = "代理数"
	CapPlanningInputCluster                = "集群数量"
	CapPlanningInputComputeVCpu            = "计算节点vCPU"
	CapPlanningInputComputeMemory          = "计算节点内存"
	CapPlanningInputComputeCount           = "单集群计算节点数量"

	CapPlanningInputOneHundred     = "100"
	CapPlanningInputOneHundredInt  = 100
	CapPlanningInputFiveHundred    = "500"
	CapPlanningInputFiveHundredInt = 500
	CapPlanningInputOneThousand    = "1000"
	CapPlanningInputOneThousandInt = 1000
	CapPlanningInputOpsAssets      = "运维资产数"

	ExpendResCodeECSVCpu           = "ECS_VCPU"
	ExpendResCodeECSMemory         = "ECS_MEM"
	ExpendResCodePAASComputeVCpu   = "PAAS-COMPUTE_VCPU"
	ExpendResCodePAASComputeMemory = "PAAS-COMPUTE_MEM"
	ExpendResCodePAASDataVCpu      = "PAAS-DATA_VCPU"
	ExpendResCodePAASDataMemory    = "PAAS-DATA_MEM"
	ExpendResCodePAASDataDisk      = "PAAS-DATA_DISK"
	ExpendResCodeEBSDisk           = "EBS_DISK"
	ExpendResCodeEFSDisk           = "EFS_DISK"
	ExpendResCodeOSSDisk           = "OSS_DISK"
	ExpendResCodeNFVVCpu           = "NFV_VCPU"
	ExpendResCodeNFVMemory         = "NFV_MEM"
	ExpendResCodeDBVCpu            = "DB_VCPU"
	ExpendResCodeDBMemory          = "DB_MEM"
	ExpendResCodeDBDisk            = "DB_DISK"
	ExpendResCodeBDVCpu            = "BD_VCPU"
	ExpendResCodeBDMemory          = "BD_MEM"
	ExpendResCodeBDDisk            = "BD_DISK"
	ExpendResCodeECSLDVCpu         = "ECS-LD_VCPU"
	ExpendResCodeECSLDMemory       = "ECS-LD_MEM"
	ExpendResCodeECSLDDisk         = "ECS-LD_DISK"
	ExpendResCodeCBRDisk           = "CBR_DISK"

	ExpendResCodeEndOfVCpu = "_VCPU"
	ExpendResCodeEndOfMem  = "_MEM"
	ExpendResCodeEndOfDisk = "_DISK"

	NodeRoleCodeCompute     = "COMPUTE"
	NodeRoleCodeBMS         = "BMS"
	NodeRoleCodeEBS         = "EBS"
	NodeRoleCodeEFS         = "EFS"
	NodeRoleCodeOSS         = "OSS"
	NodeRoleCodeNETWORK     = "NETWORK"
	NodeRoleCodeNFV         = "NFV"
	NodeRoleCodeBMSGW       = "BMSGW"
	NodeRoleCodeDATABASE    = "DATABASE"
	NodeRoleCodePAASCompute = "PAAS-COMPUTE"
	NodeRoleCodePAASData    = "PAAS-DATA"
	NodeRoleCodeCBR         = "CBR"
	NodeRoleCodeBIGDATA     = "BIG-DATA"
	NodeRoleCodeCOMPUTELD   = "COMPUTE-LD"

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

	CSOCSoftwareBomCalcMethod500GPackage           = "500G"
	DSPSoftwareBomCalcMethod1To5Instances          = "1-5实例"
	DSPSoftwareBomCalcMethod6To30Instances         = "6-30实例"
	DSPSoftwareBomCalcMethodOver30Instances        = "30实例以上"
	KAFKASoftwareBomCalcMethodBasePackage          = "基础包，200vCPU"
	KAFKASoftwareBomCalcMethodExpansionPackage     = "扩展包，100vCPU"
	CSPSoftwareBomCalcMethodBasePackage            = "基础包，500个微服务实例"
	CSPSoftwareBomCalcMethodExpansionPackage       = "扩展包，100个微服务实例"
	ROCKETMQSoftwareBomCalcMethodBasePackage       = "基础包，200vCPU"
	ROCKETMQSoftwareBomCalcMethodExpansionPackage  = "扩展包，100vCPU"
	APIMSoftwareBomCalcMethodBasePackage           = "基础包，200vCPU"
	APIMSoftwareBomCalcMethodExpansionPackage      = "扩容包，100vCPU"
	CONNECTSoftwareBomCalcMethodBasePackage        = "基础包，200个集成流"
	CONNECTSoftwareBomCalcMethodExpansionPackage   = "扩容包，100个集成流"
	CLCPSoftwareBomCalcMethodBasePackage           = "基础包，48C32个应用"
	CLCPSoftwareBomCalcMethodExpansionPackage      = "扩展包，16C8个应用"
	CLCPSoftwareBomCalcMethodBITool                = "BI工具"
	CLCPSoftwareBomCalcMethodVisualLargeScreenTool = "可视化大屏工具"
	COSSoftwareBomCalcMethodBasePackage            = "基础包，1000个监控节点"
	COSSoftwareBomCalcMethodExpansionPackage       = "扩容包，200个监控节点"
	CLSSoftwareBomCalcMethodBasePackage            = "基础包，10T"
	CLSSoftwareBomCalcMethodExpansionPackage       = "扩容包，5T"

	CopyPlanEndOfName = "的副本"

	SellSpecDPDK = "DPDK"

	ResourcePoolDefaultName = "资源池"

	CloseDpdk = 0
	OpenDpdk  = 1

	NFVResourcePoolNameKernel = "NFV-kernel资源池"
	NFVResourcePoolNameDpdk   = "NFV-DPDK资源池"
)
