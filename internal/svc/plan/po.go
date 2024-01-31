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
	Alternative int `gorm:"-" json:"alternative"` //是否有备选方案
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
	Code    int    `json:"code"`
	Data    string `json:"data"`
	Desc    string `json:"desc"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}
