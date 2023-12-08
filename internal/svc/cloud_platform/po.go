package cloud_platform

type Request struct {
	Id              int64
	UserId          string
	Name            string `form:"name"`
	Type            string `form:"type"`
	CustomerId      int64  `form:"customerId"`
	CloudPlatformId int64  `form:"cloudPlatformId"`
	SortField       string `form:"sortField"`
	Sort            string `form:"sort"`
}
