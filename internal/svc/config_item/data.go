package config_item

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
)

func ListConfigItem(code string) ([]*entity.ConfigItem, error) {
	//查询父节点配置数据
	var configItem = &entity.ConfigItem{}
	if err := data.DB.Where("code = ?", code).Find(&configItem).Error; err != nil {
		return nil, err
	}
	//查询子节点配置数据
	var list []*entity.ConfigItem
	if err := data.DB.Where("p_id = ?", configItem.Id).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
