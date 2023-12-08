package plan

type Request struct {
	Id        int64
	UserId    string
	Name      string `form:"name"`
	ProjectId int64  `form:"projectId"`
	Type      string `form:"type"`
	Stage     string `form:"stage"`
	SortField string `form:"sortField"`
	Sort      string `form:"sort"`
	Current   int    `json:"current"`
	PageSize  int    `json:"pageSize"`
}
