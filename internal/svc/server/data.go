package server

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"gorm.io/gorm"
)

func ListServer(request *Request) ([]*entity.ServerPlanningManage, error) {
	var list []*entity.ServerPlanningManage
	return list, nil
}

func CreateServer(request *Request) error {
	if err := checkBusiness(request, true); err != nil {
		return err
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("planId = ?", request.PlanId).Delete(&entity.ServerPlanningManage{})

		return nil
	}); err != nil {
		return err
	}
	now := datetime.GetNow()
	cloudPlatformEntity := &entity.ServerPlanningManage{
		CreateUserId: request.UserId,
		CreateTime:   now,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  0,
	}
	if err := data.DB.Create(cloudPlatformEntity).Error; err != nil {
		return err
	}
	return nil
}

func UpdateServer(request *Request) error {
	if err := checkBusiness(request, true); err != nil {
		return err
	}
	projectEntity := &entity.ServerPlanningManage{
		Id:           request.Id,
		UpdateUserId: request.UserId,
		UpdateTime:   datetime.GetNow(),
	}
	if err := data.DB.Updates(projectEntity).Error; err != nil {
		return err
	}
	return nil
}

func QueryServerPlanningListByPlanId(planId int64) ([]entity.ServerPlanningManage, error) {
	var serverPlanningList []entity.ServerPlanningManage
	if err := data.DB.Table(entity.ServerPlanningTable).Where("plan_id = ? and delete_state = 0", planId).Find(&serverPlanningList).Error; err != nil {
		return serverPlanningList, err
	}
	return serverPlanningList, nil
}

func checkBusiness(request *Request, isCreate bool) error {
	return nil
}
