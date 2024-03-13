package cloud_platform

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
)

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

type CloudPlatform struct {
	entity.CloudPlatformManage
	LeaderId   string `gorm:"-" json:"leaderId"`
	LeaderName string `gorm:"-" json:"leaderName" `
}

type Region struct {
	entity.RegionManage
	AzList []*Az `gorm:"-" json:"azList"`
}

type Az struct {
	entity.AzManage
	MachineRoomList []*entity.MachineRoom `gorm:"-" json:"machineRoomList"`
	CellList        []*entity.CellManage  `gorm:"-" json:"cellList"`
}

type ResponseTree struct {
	RegionList []*ResponseTreeRegion `json:"regionList"`
}

type ResponseTreeRegion struct {
	Region *entity.RegionManage `json:"region"`
	AzList []*ResponseTreeAz    `json:"azList"`
}

type ResponseTreeAz struct {
	Az              *entity.AzManage      `json:"az"`
	MachineRoomList []*entity.MachineRoom `json:"machineRoomList"`
	CellList        []*ResponseTreeCell   `json:"cellList"`
}

type ResponseTreeCell struct {
	Cell         *entity.CellManage `json:"cell"`
	ProjectCount int                `json:"projectCount"`
}
