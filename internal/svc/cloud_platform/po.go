package cloud_platform

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"time"
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

type CloudPlatformManage struct {
	Id           int64           `gorm:"column:id" json:"id"`                       //云平台id
	Name         string          `gorm:"column:name" json:"name"`                   //云平台名称
	Type         string          `gorm:"column:type" json:"type"`                   //云平台类型（运营云、交付云）
	CustomerId   int64           `gorm:"column:customer_id" json:"customerId"`      //客户id
	CreateUserId string          `gorm:"column:create_user_id" json:"createUserId"` //创建人id
	CreateTime   time.Time       `gorm:"column:create_time" json:"createTime"`      //创建时间
	UpdateUserId string          `gorm:"column:update_user_id" json:"updateUserId"` //更新人id
	UpdateTime   time.Time       `gorm:"column:update_time" json:"updateTime"`      //更新时间
	DeleteState  int             `gorm:"column:delete_state" json:"-"`              //作废状态：1，作废；0，正常
	RegionList   []*RegionManage `gorm:"-" json:"regionList"`
	LeaderId     string          `gorm:"-" json:"leaderId"`
	LeaderName   string          `gorm:"-" json:"leaderName" `
}

type AzManage struct {
	Id              int64                 `gorm:"column:id" json:"id"`                       //azId
	Code            string                `gorm:"column:code" json:"code"`                   //az编码
	RegionId        int64                 `gorm:"column:region_id" json:"regionId"`          //regionId
	CreateUserId    string                `gorm:"column:create_user_id" json:"createUserId"` //创建人id
	CreateTime      time.Time             `gorm:"column:create_time" json:"createTime"`      //创建时间
	UpdateUserId    string                `gorm:"column:update_user_id" json:"updateUserId"` //更新人id
	UpdateTime      time.Time             `gorm:"column:update_time" json:"updateTime"`      //更新时间
	DeleteState     int                   `gorm:"column:delete_state" json:"-"`              //作废状态：1，作废；0，正常
	MachineRoomList []*entity.MachineRoom `gorm:"-" json:"machineRoomList"`
	CellList        []*entity.CellManage  `gorm:"-" json:"cellList"`
}

type RegionManage struct {
	Id              int64       `gorm:"column:id" json:"id"`                             //regionId
	Code            string      `gorm:"column:code" json:"code"`                         //region编码
	Name            string      `gorm:"column:name" json:"name"`                         //region名称
	Type            string      `gorm:"column:type" json:"type"`                         //region类型
	CloudPlatformId int64       `gorm:"column:cloud_platform_id" json:"cloudPlatformId"` //云平台id
	CreateUserId    string      `gorm:"column:create_user_id" json:"createUserId"`       //创建人id
	CreateTime      time.Time   `gorm:"column:create_time" json:"createTime"`            //创建时间
	UpdateUserId    string      `gorm:"column:update_user_id" json:"updateUserId"`       //更新人id
	UpdateTime      time.Time   `gorm:"column:update_time" json:"updateTime"`            //更新时间
	DeleteState     int         `gorm:"column:delete_state" json:"-"`                    //作废状态：1，作废；0，正常
	AzList          []*AzManage `gorm:"-" json:"azList"`
}
