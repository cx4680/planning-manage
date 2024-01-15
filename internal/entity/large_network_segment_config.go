package entity

import "time"

const LargeNetworkSegmentConfigTable = "large_network_segment_config"

type LargeNetworkSegmentConfig struct {
	Id                             int64     `gorm:"column:id" json:"id"`                                                             // 主键Id
	PlanId                         int64     `gorm:"column:plan_id" json:"planId"`                                                    // 方案Id
	StorageNetworkSegmentRoute     string    `gorm:"column:storage_network_segment_route" json:"storageNetworkSegmentRoute"`          // 存储前端网规划网段明细路由
	BizIntranetNetworkSegmentRoute string    `gorm:"column:biz_intranet_network_segment_route" json:"bizIntranetNetworkSegmentRoute"` // 业务内网规划网段明细路由
	BizExternalLargeNetworkSegment string    `gorm:"column:biz_external_large_network_segment" json:"bizExternalLargeNetworkSegment"` // 业务外网大网网段
	BmcNetworkSegmentRoute         string    `gorm:"column:bmc_network_segment_route" json:"bmcNetworkSegmentRoute"`                  // bmc规划网段明细路由
	CreateUserId                   string    `gorm:"column:create_user_id" json:"createUserId"`                                       // 创建人id
	CreateTime                     time.Time `gorm:"column:create_time" json:"createTime"`                                            // 创建时间
	UpdateUserId                   string    `gorm:"column:update_user_id" json:"updateUserId"`                                       // 更新人id
	UpdateTime                     time.Time `gorm:"column:update_time" json:"updateTime"`                                            // 更新时间
}

func (entity *LargeNetworkSegmentConfig) TableName() string {
	return LargeNetworkSegmentConfigTable
}
