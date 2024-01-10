package az

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"time"
)

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

type Az struct {
	Id              int64                 `gorm:"column:id" json:"id"`                       //azId
	Code            string                `gorm:"column:code" json:"code"`                   //az编码
	RegionId        int64                 `gorm:"column:region_id" json:"regionId"`          //regionId
	CreateUserId    string                `gorm:"column:create_user_id" json:"createUserId"` //创建人id
	CreateTime      time.Time             `gorm:"column:create_time" json:"createTime"`      //创建时间
	UpdateUserId    string                `gorm:"column:update_user_id" json:"updateUserId"` //更新人id
	UpdateTime      time.Time             `gorm:"column:update_time" json:"updateTime"`      //更新时间
	DeleteState     int                   `gorm:"column:delete_state" json:"-"`              //作废状态：1，作废；0，正常
	MachineRoomList []*entity.MachineRoom `gorm:"-" json:"machineRoomList"`
}
