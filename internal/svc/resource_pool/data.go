package resource_pool

import (
	"errors"
	"fmt"

	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
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

func CreateResourcePool(request *Request) error {
	var nodeRoleBaseline *entity.NodeRoleBaseline
	if err := data.DB.Where("id = ?", request.NodeRoleId).Find(&nodeRoleBaseline).Error; err != nil {
		return err
	}
	if nodeRoleBaseline == nil {
		return errors.New("节点角色不存在")
	}
	if nodeRoleBaseline.SupportMultiResourcePool == constant.NodeRoleNotSupportMultiResourcePool {
		return errors.New("该节点角色不支持多资源池")
	}
	var count int64
	if err := data.DB.Table(entity.ResourcePoolTable).Where("plan_id = ? and node_role_id = ?", request.PlanId, request.NodeRoleId).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("该方案和节点角色下没有资源池")
	}
	resourcePoolName := fmt.Sprintf("%s-%s-%d", nodeRoleBaseline.NodeRoleName, constant.ResourcePoolDefaultName, count+1)
	resourcePool := &entity.ResourcePool{
		ResourcePoolName: resourcePoolName,
		PlanId:           request.PlanId,
		NodeRoleId:       request.NodeRoleId,
		OpenDpdk:         constant.CloseDpdk,
	}
	if err := data.DB.Create(&resourcePool).Error; err != nil {
		return err
	}
	return nil
}

func DeleteResourcePool(request *Request) error {
	var resourcePool = &entity.ResourcePool{}
	if err := data.DB.Where("id = ?", request.Id).Find(&resourcePool).Error; err != nil {
		return err
	}
	if resourcePool.Id == 0 {
		return errors.New("资源池不存在")
	}
	var resourcePools []*entity.ResourcePool
	if err := data.DB.Where("plan_id = ? and node_role_id = ?", resourcePool.PlanId, resourcePool.NodeRoleId).Find(&resourcePools).Error; err != nil {
		return err
	}
	if len(resourcePools) < 2 {
		return errors.New("该节点角色至少有一个资源池")
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", request.Id).Delete(&entity.ResourcePool{}).Error; err != nil {
			return err
		}
		if err := tx.Where("resource_pool_id = ?", request.Id).Delete(&entity.ServerPlanning{}).Error; err != nil {
			return err
		}
		if err := tx.Where("resource_pool_id = ?", request.Id).Delete(&entity.ServerCapPlanning{}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Errorf("[deleteResourcePool] error, %v", err)
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
