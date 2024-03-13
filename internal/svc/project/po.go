package project

import "code.cestc.cn/ccos/common/planning-manage/internal/entity"

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
	entity.ProjectManage
	CustomerName      string `gorm:"-" json:"customerName"`      //客户名称
	CloudPlatformName string `gorm:"-" json:"cloudPlatformName"` //云平台名称
	CloudPlatformType string `gorm:"-" json:"cloudPlatformType"` //云平台类型
	RegionName        string `gorm:"-" json:"regionName"`        //region名称
	AzCode            string `gorm:"-" json:"azCode"`            //az编码
	CellName          string `gorm:"-" json:"cellName"`          //cell名称
	PlanCount         int    `gorm:"-" json:"planCount"`         //cell名称
}
