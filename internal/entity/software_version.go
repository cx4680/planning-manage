package entity

import (
	"time"
)

const SoftwareVersionTable = "software_version"

type SoftwareVersion struct {
	Id              int64     `gorm:"column:id" json:"id"`                            // 主键id
	SoftwareVersion string    `gorm:"column:software_version" json:"softwareVersion"` // 软件版本
	ReleaseTime     time.Time `gorm:"column:release_time" json:"releaseTime"`         // 发布时间
	CreateTime      time.Time `gorm:"column:create_time" json:"createTime"`           // 创建时间
}

func (entity *SoftwareVersion) TableName() string {
	return SoftwareVersionTable
}
