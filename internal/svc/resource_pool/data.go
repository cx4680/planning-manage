package resource_pool

import (
	"errors"
	"fmt"

	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
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
	var defaultServerBaselineId int64
	if err := data.DB.Table(entity.ServerNodeRoleRelTable).Select("server_id").Where("node_role_id = ?", nodeRoleBaseline.Id).Order("server_id asc").Limit(1).Find(&defaultServerBaselineId).Error; err != nil {
		return err
	}
	var serverBaseline *entity.ServerBaseline
	if err := data.DB.Where("id = ?", defaultServerBaselineId).Find(&serverBaseline).Error; err != nil {
		return err
	}
	resourcePoolName := fmt.Sprintf("%s-%s-%d", nodeRoleBaseline.NodeRoleName, constant.ResourcePoolDefaultName, count+1)
	resourcePool := &entity.ResourcePool{
		ResourcePoolName:    resourcePoolName,
		PlanId:              request.PlanId,
		NodeRoleId:          request.NodeRoleId,
		OpenDpdk:            constant.CloseDpdk,
		DefaultResourcePool: constant.No,
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&resourcePool).Error; err != nil {
			return err
		}
		now := datetime.GetNow()
		serverPlanning := &entity.ServerPlanning{
			PlanId:           request.PlanId,
			NodeRoleId:       request.NodeRoleId,
			MixedNodeRoleId:  request.NodeRoleId,
			ServerBaselineId: serverBaseline.Id,
			Number:           nodeRoleBaseline.MinimumNum,
			OpenDpdk:         constant.CloseDpdk,
			NetworkInterface: serverBaseline.NetworkInterface,
			CpuType:          serverBaseline.CpuType,
			ResourcePoolId:   resourcePool.Id,
			CreateUserId:     request.UserId,
			UpdateUserId:     request.UserId,
			CreateTime:       now,
			UpdateTime:       now,
		}
		if err := tx.Create(&serverPlanning).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Errorf("[createResourcePool] error, %v", err)
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
	if resourcePool.DefaultResourcePool == constant.Yes {
		return errors.New("默认资源池不可以删除")
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
