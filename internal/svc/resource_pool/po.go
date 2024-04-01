package resource_pool

type Request struct {
	Id               int64
	ResourcePoolName string `form:"resourcePoolName"`
	PlanId           int64  `form:"planId"`
	NodeRoleId       int64  `form:"nodeRoleId"`
}
