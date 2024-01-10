package az

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
)

type Request struct {
	Id              int64
	UserId          string
	Code            string                `form:"code"`
	RegionId        int64                 `form:"regionId"`
	MachineRoomList []*RequestMachineRoom `form:"machineRoomList"`
	SortField       string                `form:"sortField"`
	Sort            string                `form:"sort"`
}

type RequestMachineRoom struct {
	Name     string `form:"name"`
	Abbr     string `form:"abbr"`
	Province string `form:"province"`
	City     string `form:"city"`
	Address  string `form:"address"`
}

type Response struct {
	Az              *entity.AzManage      `json:"az"`
	MachineRoomList []*entity.MachineRoom `json:"machineRoomList"`
}
