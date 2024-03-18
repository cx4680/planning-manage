package resource_pool

type Request struct {
	Id               int64
	ResourcePoolName string `form:"resourcePoolName"`
}
