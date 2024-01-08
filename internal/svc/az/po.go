package az

type Request struct {
	Id              int64
	UserId          string
	Code            string         `form:"code"`
	RegionId        int64          `form:"regionId"`
	MachineRoomList []*MachineRoom `form:"machineRoomList"`
	SortField       string         `form:"sortField"`
	Sort            string         `form:"sort"`
}

type MachineRoom struct {
	Name     string `json:"name"`
	Abbr     string `json:"abbr"`
	Province string `json:"province"`
	City     string `json:"city"`
	Address  string `json:"address"`
}
