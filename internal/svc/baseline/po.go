package baseline

const (
	CloudProductBaselineSheetName  = "云产品售卖清单"
	ServerBaselineSheetName        = "服务器基线"
	NetworkDeviceBaselineSheetName = "网络设备基线"
	NodeRoleBaselineSheetName      = "节点角色基线"
)

const (
	CloudProductBaselineType  = "cloudProductListBaseline"
	ServerBaselineType        = "serverBaseline"
	NetworkDeviceBaselineType = "networkDeviceBaseline"
	NodeRoleBaselineType      = "nodeRoleBaseline"
)

type ImportBaselineRequest struct {
	CloudPlatformType string `json:"cloudPlatformType" validate:"required"` // 云平台类型，1：运营云，0：交付云
	BaselineVersion   string `json:"baselineVersion" validate:"required"`   // 基线版本
	BaselineType      string `json:"baselineType" validate:"required"`      // 基线类型
	ReleaseTime       string `json:"releaseTime"`                           // 发布时间
}

type CloudProductBaselineExcel struct {
	ProductType         string   `excel:"name:云服务类型;" json:"productType"`           // 产品类型
	ProductName         string   `excel:"name:云服务;" json:"productName"`             // 产品名称
	ProductCode         string   `excel:"name:服务编码;" json:"productCode"`            // 产品编码
	SellSpecs           string   `excel:"name:售卖规格;" json:"sellSpecs"`              // 售卖规格
	AuthorizedUnit      string   `excel:"name:授权单元;" json:"authorizedUnit"`         // 授权单元
	WhetherRequired     string   `excel:"name:是否必选;" json:"whetherRequired"`        // 是否必选，0：否，1：是
	Instructions        string   `excel:"name:说明;" json:"instructions"`             // 说明
	DependProductCode   string   `excel:"name:依赖服务编码;" json:"dependProductCode"`    // 依赖产品Code
	ControlResNodeRole  string   `excel:"name:管控资源节点角色;" json:"controlResNodeRole"` // 管控资源节点角色
	ResNodeRole         string   `excel:"name:资源节点角色;" json:"resNodeRole"`          // 资源节点角色
	DependProductCodes  []string `json:"dependProductCodes"`                        // 依赖产品Code数组
	ControlResNodeRoles []string `json:"controlResNodeRoles"`                       // 管控资源节点角色数组
	ResNodeRoles        []string `json:"resNodeRoles"`                              // 资源节点角色数组
}

type NodeRoleBaselineExcel struct {
	NodeRoleCode string   `excel:"name:角色code;" json:"nodeRoleCode"`   // 角色code
	NodeRoleName string   `excel:"name:角色名称;" json:"nodeRoleName"`     // 角色名称
	MinimumCount int      `excel:"name:单独部署最小数量;" json:"minimumCount"` // 单独部署最小数量
	DeployMethod string   `excel:"name:部署方式;" json:"deployMethod"`     // 部署方式
	Classify     string   `gorm:"name:分类;" json:"classify"`            // 分类
	MixedDeploy  string   `excel:"name:节点混部;" json:"mixedDeploy"`      // 节点混部
	Annotation   string   `excel:"name:节点说明;" json:"annotation"`       // 节点说明
	BusinessType string   `excel:"name:业务类型;" json:"businessType"`     // 业务类型
	MixedDeploys []string `json:"mixedDeploys"`                        // 节点混部数组
}

type ServerBaselineExcel struct {
	Arch                string   `excel:"name:硬件架构" json:"Arch"`                      // 硬件架构
	NetworkInterface    string   `excel:"name:网络接口" json:"networkInterface"`          // 网络接口
	NodeRole            string   `excel:"name:节点角色" json:"nodeRole"`                  // 节点角色
	ServerModel         string   `excel:"name:机型" json:"serverModel"`                 // 机型
	ConfigurationInfo   string   `excel:"name:配置概要" json:"configurationInfo"`         // 配置概要
	Spec                string   `excel:"name:规格" json:"spec"`                        // 规格
	CpuType             string   `excel:"name:硬件架构" json:"cpuType"`                   // CPU类型
	Cpu                 int      `excel:"name:vCPU" json:"cpu"`                       // CPU核数
	Gpu                 string   `excel:"name:GPU" json:"gpu"`                        // GPU
	Memory              int      `excel:"name:内存" json:"memory"`                      // 内存
	SystemDiskType      string   `excel:"name:系统盘类型" json:"systemDiskType"`           // 系统盘类型
	SystemDisk          string   `excel:"name:系统盘" json:"systemDisk"`                 // 系统盘
	StorageDiskType     string   `excel:"name:存储盘类型" json:"storageDiskType"`          // 存储盘类型
	StorageDiskNum      int      `excel:"name:存储盘个数" json:"storageDiskNum"`           // 存储盘个数
	StorageDiskCapacity int      `excel:"name:存储盘单盘容量（G）" json:"storageDiskCapacity"` // 存储盘单盘容量（G）
	RamDisk             string   `excel:"name:缓存盘" json:"ramDisk"`                    // 缓存盘
	NetworkCardNum      int      `excel:"name:网卡数量" json:"networkCardNum"`            // 网卡数量
	Power               int      `excel:"name:功率（W）" json:"power"`                    // 功率
	NodeRoles           []string `json:"nodeRoles"`                                   // 节点角色数组
}
