package entity

const ConfigItemTable = "config_item"

type ConfigItem struct {
	Id   int    `gorm:"column:id" json:"id"`     //配置Id
	PId  int    `gorm:"column:p_id" json:"PId"`  //上级配置Id
	Name string `gorm:"column:name" json:"name"` //配置名称
	Code string `gorm:"column:code" json:"code"` //配置编码
	Data string `gorm:"column:data" json:"data"` //配置值
	Sort int    `gorm:"column:sort" json:"sort"` //排序
}

func (entity *ConfigItem) TableName() string {
	return ConfigItemTable
}
