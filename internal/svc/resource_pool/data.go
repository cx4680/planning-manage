package resource_pool

import (
	"errors"

	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
)

func UpdateResourcePool(request *Request) error {
	if err := checkBusiness(request); err != nil {
		return err
	}
	resourcePool := &entity.ResourcePool{
		Id:               request.Id,
		ResourcePoolName: request.ResourcePoolName,
	}
	if err := data.DB.Updates(&resourcePool).Error; err != nil {
		return err
	}
	return nil
}

func checkBusiness(request *Request) error {
	var resourcePool = &entity.ResourcePool{}
	if err := data.DB.Where("id = ?", request.Id).Find(&resourcePool).Error; err != nil {
		return err
	}
	if resourcePool.Id == 0 {
		return errors.New("资源池不存在")
	}
	return nil
}
