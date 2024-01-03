package region

type Request struct {
	Id              int64
	UserId          string
	Name            string `form:"name"`
	Code            string `form:"code"`
	Type            string `form:"type"`
	CloudPlatformId int64  `form:"cloudPlatformId"`
	SortField       string `form:"sortField"`
	Sort            string `form:"sort"`
}
