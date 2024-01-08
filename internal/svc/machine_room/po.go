package machine_room

type PageRequest struct {
	PlanId   int64 `form:"planId"`
	Current  int   `json:"current"`
	PageSize int   `json:"pageSize"`
}
