package project

import "time"

type Request struct {
	UserId            string
	Id                int64  `form:"id"`
	Name              string `form:"name"`
	CustomerName      string `form:"customerName"`
	Type              string `form:"type"`
	Stage             string `form:"stage"`
	CloudPlatformId   int64  `form:"cloudPlatformId"`
	CloudPlatformType string `form:"cloudPlatformType"`
	RegionId          int64  `form:"regionId"`
	AzId              int64  `form:"azId"`
	CellId            int64  `form:"cellId"`
	CustomerId        int64  `form:"customerId"`
	SortField         string `form:"sortField"`
	Sort              string `form:"sort"`
	Current           int    `json:"current"`
	PageSize          int    `json:"pageSize"`
}

type Project struct {
	Id                int64     `gorm:"column:id" json:"id"`                             //项目id
	Name              string    `gorm:"column:name" json:"name"`                         //项目名称
	CloudPlatformId   int64     `gorm:"column:cloud_platform_id" json:"cloudPlatformId"` //云平台id
	RegionId          int64     `gorm:"column:region_id" json:"regionId"`                //regionId
	AzId              int64     `gorm:"column:az_id" json:"azId"`                        //azId
	CellId            int64     `gorm:"column:cell_id" json:"cellId"`                    //cell Id
	CustomerId        int64     `gorm:"column:customer_id" json:"customerId"`            //客户id
	Type              string    `gorm:"column:type" json:"type"`                         //项目类型
	Stage             string    `gorm:"column:stage" json:"stage"`                       //项目进度
	CreateUserId      string    `gorm:"column:create_user_id" json:"createUserId"`       //创建人id
	CreateTime        time.Time `gorm:"column:create_time" json:"createTime"`            //创建时间
	UpdateUserId      string    `gorm:"column:update_user_id" json:"updateUserId"`       //更新人id
	UpdateTime        time.Time `gorm:"column:update_time" json:"updateTime"`            //更新时间
	DeleteState       int       `gorm:"column:delete_state" json:"-"`                    //作废状态：1，作废；0，正常
	CustomerName      string    `gorm:"-" json:"customerName"`                           //客户名称
	CloudPlatformName string    `gorm:"-" json:"cloudPlatformName"`                      //云平台名称
	CloudPlatformType string    `gorm:"-" json:"cloudPlatformType"`                      //云平台类型
	RegionName        string    `gorm:"-" json:"regionName"`                             //region名称
	AzCode            string    `gorm:"-" json:"azCode"`                                 //az编码
	CellName          string    `gorm:"-" json:"cellName"`                               //cell名称
	PlanCount         int       `gorm:"-" json:"planCount"`                              //cell名称
}
