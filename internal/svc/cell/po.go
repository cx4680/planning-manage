package cell

type Request struct {
	Id        int64
	UserId    string
	Name      string `form:"name"`
	AzId      int64  `form:"azId"`
	SortField string `form:"sortField"`
	Sort      string `form:"sort"`
}
