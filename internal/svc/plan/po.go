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
