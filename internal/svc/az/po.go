package az

type Request struct {
	Id              int64
	UserId          string
	Code            string `form:"code"`
	RegionId        int64  `form:"regionId"`
	MachineRoomName string `json:"machineRoomName"`
	MachineRoomCode string `json:"machineRoomCode"`
	Province        string `json:"province"`
	City            string `json:"city"`
	Address         string `json:"address"`
	SortField       string `form:"sortField"`
	Sort            string `form:"sort"`
}
