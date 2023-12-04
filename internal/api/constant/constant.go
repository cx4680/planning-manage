package constant

const (
	XRequestID = "x-request-id"
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
	StateDeleted = "deleted"
	StateNormal  = "normal"
	StateEnable  = "enable"
	StateDisable = "disable"
	Manual       = "manual" // 手动
	Auto         = "auto"   // 自动
	Waiting      = "Waiting"
	Running      = "Running"
	Succeed      = "Succeed"     // 成功
	Failed       = "Failed"      // 失败
	Timeout      = "Timeout"     // 超时
	PartSucceed  = "PartSucceed" // 超时
)

const (
	True  = "true"
	False = "false"
)

const (
	QuestionMarkCorn = "?"
	SpaceCorn        = " "
	StarCorn         = "*"
	CommaCorn        = ","
	LeftSlash        = "/"
)

const (
	Hour  = "h"
	Day   = "d"
	Month = "m"
)

const (
	EnvGlobalBaseDomain = "GLOBAL_BASE_DOMAIN"
	BaseDomain          = "BASE_DOMAIN"
	EnvRegion           = "REGION"
	EnvCellID           = "CELLID"
)

const (
	CronjobPeriod = "period"      // 周期
	CronjobTiming = "timing"      // 定时
	Immediately   = "Immediately" // 实时
)

const (
	NameSpace = "ccos-ops-app"
)

const (
	UserSecretPrivateKeyPath = "/app/secret/userSecretPrivateKey"
)

const (
	LogResourceDisplayName = "resourceDisplayName"
	LogResourceType        = "resourceType"
	LogResourceID          = "resourceId"
	LogResourceTypeID      = "resourceTypeId"
	LogServiceCode         = "serviceCode"
)

const (
	SystemAtom    = "SystemAtom"
	CustomAtom    = "CustomAtom"
	SystemCompose = "SystemCompose"
	CustomCompose = "CustomCompose"
)

const (
	Shell  = "Shell"
	Python = "Python"
	Bat    = "Bat"
)

const (
	Readonly = "readonly"
	Editable = "editable"
)

const (
	Atom    = "atom"
	Compose = "compose"
)

const (
	ReportTopic      = "cell-agent-task-topic_repord"
	ExecTopic        = "cell-agent-task-topic_exec"
	KafkaBroken      = "KAFKA_BROKEN"
	JOB_REPORT_GROUP = "jobm_report_group"
)

const (
	LogLine = "0"
	LogFull = "1"
	LogNo   = "-1"
)

const (
	SEPARATION_OF_TWO_NETWORKS = "2" // 两网分离
	TRIPLE_NETWORK_SEPARATION  = "3" // 三网分离
	TRIPLE_PLAY                = "1" // 三网合一
)

const (
	MASW      = "MASW"      // 管理接入交换机
	VASW      = "VASW"      // 业务内网接入交换机
	StorSASW  = "StorSASW"  // 存储业务接入交换机
	StoreCASW = "StoreCASW" //存储集群接入交换机
	BMSASW    = "BMSASW"    //裸金属接入交换机
	ISW       = "ISW"       //业务外网出口交换机
	OASW      = "OASW"      //服务器带外接入交换机
)

const (
	YES = "1"
	NO  = "0"
)
