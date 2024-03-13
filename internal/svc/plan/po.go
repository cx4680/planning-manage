package plan

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
)

type Request struct {
	UserId    string
	Id        int64  `form:"id"`
	Name      string `form:"name"`
	ProjectId int64  `form:"projectId"`
	Type      string `form:"type"`
	Stage     string `form:"stage"`
	SortField string `form:"sortField"`
	Sort      string `form:"sort"`
	Current   int    `json:"current"`
	PageSize  int    `json:"pageSize"`
}

type Plan struct {
	entity.PlanManage
	Alternative int `gorm:"-" json:"alternative"` // 是否有备选方案
}

type SendBomsRequest struct {
	ProductConfigLibId string                 `json:"productConfigLibId"`
	Steps              []*SendBomsRequestStep `json:"steps"`
}

type SendBomsRequestStep struct {
	StepName string                    `json:"stepName"`
	Features []*SendBomsRequestFeature `json:"features"`
}

type SendBomsRequestFeature struct {
	FeatureName string                `json:"featureName"`
	FeatureCode string                `json:"featureCode"`
	Boms        []*SendBomsRequestBom `json:"boms"`
}

type SendBomsRequestBom struct {
	Code  string `json:"code"`
	Count int    `json:"count"`
}

type SendBomsResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Desc    string      `json:"desc"`
	Message string      `json:"message"`
	Success bool        `json:"success"`
}

type DownloadResponse struct {
	SheetName string      `json:"sheetName"`
	Data      interface{} `json:"data"`
}

type CloudProductPlanningExportResponse struct {
	ProductType       string `excel:"name:产品类型;" gorm:"column:product_type" json:"productType"`
	ProductName       string `excel:"name:产品名称;" gorm:"column:product_name" json:"productName"`
	SellSpec          string `excel:"name:售卖规格;" gorm:"column:sell_spec" json:"sellSpec"`
	ValueAddedService string `excel:"name:增值服务;" gorm:"column:value_added_service" json:"valueAddedService"`
	Instructions      string `excel:"name:说明;" gorm:"column:instructions" json:"instructions"`
}

type ResponseDownloadServer struct {
	NodeRole   string `json:"nodeRole" excel:"name:角色;"`
	ServerType string `json:"serverType" excel:"name:设备类型;"`
	BomCode    string `json:"bomCode" excel:"name:机型;"`
	Spec       string `json:"spec" excel:"name:规格;"`
	Number     string `json:"number" excel:"name:数量;"`
}

type NetworkDeviceRoleIdNum struct {
	NetworkDeviceRoleId int64 `gorm:"column:network_device_role_id" json:"networkDeviceRoleId"`
	Num                 int   `gorm:"column:num" json:"num"`
}

type NetworkDeviceListExportResponse struct {
	NetworkDeviceRoleName string `gorm:"column:network_device_role_name" json:"networkDeviceRoleName" excel:"name:设备类型;index:0"`
	NetworkDeviceRole     string `gorm:"column:network_device_role" json:"networkDeviceRole" excel:"name:设备角色;index:1"`
	Brand                 string `gorm:"column:brand" json:"brand" excel:"name:厂商;index:2"`
	DeviceModel           string `gorm:"column:device_model" json:"deviceModel" excel:"name:机型;index:3"`
	ConfOverview          string `gorm:"column:conf_overview" json:"confOverview" excel:"name:规格参数;index:4"`
	Num                   string `gorm:"column:num" json:"num" excel:"name:数量;index:5"`
}

type BomListDownload struct {
	Category       string `json:"category"  excel:"name:分类;"`
	CloudProduct   string `json:"cloudProduct" excel:"name:云产品;"`
	SellType       string `json:"sellType" excel:"name:类型;"`
	BomId          string `json:"bomId" excel:"name:BOMID;"`
	AuthorizedUnit string `json:"authorizedUnit" excel:"name:授权单元;"`
	Number         string `json:"number" excel:"name:数量;"`
}
