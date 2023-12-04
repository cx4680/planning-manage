package baseline

const (
	CloudProductBaselineSheetName  = "云产品售卖清单"
	ServerBaselineSheetName        = "服务器基线"
	NetworkDeviceBaselineSheetName = "网络设备基线"
	NodeRoleBaselineSheetName      = "node_role_config"
)

const (
	CloudProductBaselineType  = "cloudProductListBaseline"
	ServerBaselineType        = "serverBaseline"
	NetworkDeviceBaselineType = "networkDeviceBaseline"
	NodeRoleBaselineType      = "nodeRoleBaseline"
)

type ImportBaselineRequest struct {
	CloudPlatformType string `json:"cloudPlatformType" validate:"required"` // 云平台类型
	BaselineVersion   string `json:"baselineVersion" validate:"required"`   // 基线版本
	BaselineType      string `json:"baselineType" validate:"required"`      // 基线类型
	ReleaseTime       string `json:"releaseTime"`                           // 发布时间
}

type CloudProductBaselineExcel struct {
	ProductType         string   `excel:"name:产品分类;" json:"productType"`            // 产品类型
	ProductName         string   `excel:"name:产品名称;" json:"productName"`            // 产品名称
	ProductCode         string   `excel:"name:产品Code;" json:"productCode"`          // 产品编码
	SellSpecs           string   `excel:"name:售卖规格;" json:"sellSpecs"`              // 售卖规格
	AuthorizedUnit      string   `excel:"name:授权单元;" json:"authorizedUnit"`         // 授权单元
	WhetherRequired     string   `excel:"name:是否必选;" json:"whetherRequired"`        // 是否必选，0：否，1：是
	Instructions        string   `excel:"name:说明;" json:"instructions"`             // 说明
	DependProductCode   string   `excel:"name:依赖产品Code;" json:"dependProductCode"`  // 依赖产品Code
	ControlResNodeRole  string   `excel:"name:管控资源节点角色;" json:"controlResNodeRole"` // 管控资源节点角色
	ResNodeRole         string   `excel:"name:资源节点角色;" json:"resNodeRole"`          // 资源节点角色
	ControlResNodeRoles []string `json:"controlResNodeRoles"`                       // 管控资源节点角色数组
	ResNodeRoles        []string `json:"resNodeRoles"`                              // 资源节点角色数组
}

type NodeRoleBaselineExcel struct {
	NodeRole     string `excel:"name:NodeRole;" json:"nodeRole"`     // NodeRole
	RoleName     string `excel:"name:角色;" json:"roleName"`           // 角色
	MinimumCount string `excel:"name:单独部署最小数量;" json:"minimumCount"` // 单独部署最小数量
	MixedDeploy  string `excel:"name:节点混部;" json:"mixedDeploy"`      // 节点混部
	Annotation   string `excel:"name:节点说明;" json:"annotation"`       // 节点说明
	BusinessType string `excel:"name:业务类型;" json:"businessType"`     // 业务类型
}
