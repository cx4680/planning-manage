package server

type Request struct {
	Id             int64
	UserId         string
	PlanId         int64            `form:"planId"`
	NetworkVersion string           `form:"networkVersion"`
	CpuType        string           `form:"cpuType"`
	serverList     []*RequestServer `form:"serverList"`
}

type RequestServer struct {
	NodeRoleId       int64 `form:"nodeRoleId"`
	MixedNodeRoleId  int64 `form:"mixedNodeRoleId"`
	ServerBaselineId int64 `form:"serverBaselineId"`
	Number           int   `form:"number"`
	OpenDpdk         int   `form:"openDpdk"`
}
