package server

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"gorm.io/gorm"
)

type ResponseCapacity struct {
	Id              int64  `json:"id"`
	ProductId       int64  `json:"productId"`
	ProductName     string `json:"productName"`
	CapacitySpecs   string `json:"capacitySpecs"`
	SalesSpecs      string `json:"salesSpecs"`
	OverbookingRate string `json:"overbookingRate"`
	Number          string `json:"number"`
	Unit            string `json:"unit"`
}

type ResponseModel struct {
	Id                int64  `json:"id"`
	Model             string `json:"model"`
	HardwareVersion   string `json:"hardwareVersion"`
	ConfigurationInfo string `json:"configurationInfo"`
}

func ListServer(request *Request) ([]*entity.ServerPlanning, error) {
	//查询云产品规划表
	var cloudProductPlanningList []*entity.CloudProductPlanning
	if err := data.DB.Where("plan_id = ?", request.PlanId).Find(&cloudProductPlanningList).Error; err != nil {
		return nil, err
	}
	var productIdList []int64
	for _, v := range cloudProductPlanningList {
		productIdList = append(productIdList, v.ProductId)
	}
	//查询云产品和角色关联表
	var cloudProductNodeRoleRelList []*entity.CloudProductNodeRoleRel
	if err := data.DB.Where("product_id IN (?)", productIdList).Find(&cloudProductNodeRoleRelList).Error; err != nil {
		return nil, err
	}
	var nodeRoleIdList []int64
	for _, v := range cloudProductNodeRoleRelList {
		nodeRoleIdList = append(nodeRoleIdList, v.NodeRoleId)
	}
	//查询角色表
	var nodeRoleBaselineList []*entity.NodeRoleBaseline
	if err := data.DB.Where("id IN (?)", nodeRoleIdList).Find(&nodeRoleBaselineList).Error; err != nil {
		return nil, err
	}
	//查询服务器规划表
	var serverPlanningList []*entity.ServerPlanning
	if err := data.DB.Where("plan_id = ?", request.PlanId).Find(&serverPlanningList).Error; err != nil {
		return nil, err
	}
	var serverPlanningMap = make(map[int64]*entity.ServerPlanning)
	for _, v := range serverPlanningList {
		serverPlanningMap[v.NodeRoleId] = v
	}
	var list []*entity.ServerPlanning
	for _, v := range nodeRoleBaselineList {
		_, ok := serverPlanningMap[v.Id]
		if ok {
			serverPlanningMap[v.Id].NodeRoleName = v.NodeRoleName
			serverPlanningMap[v.Id].NodeRoleAnnotation = v.Annotation
			list = append(list, serverPlanningMap[v.Id])
		}
		serverPlanningList = append(serverPlanningList, &entity.ServerPlanning{
			PlanId:             request.PlanId,
			NodeRoleId:         v.Id,
			Number:             v.MinimumNum,
			NodeRoleName:       v.NodeRoleName,
			NodeRoleAnnotation: v.Annotation,
		})
	}
	return list, nil
}

func ListServerArch(request *Request) ([]string, error) {
	//查询云产品规划表
	var cloudProductPlanning = &entity.CloudProductPlanning{}
	if err := data.DB.Where("plan_id = ?", request.PlanId).First(&cloudProductPlanning).Error; err != nil {
		return nil, err
	}
	//查询云产品配置表
	var cloudProductBaseline = &entity.CloudProductBaseline{}
	if err := data.DB.Where("id = ?", cloudProductPlanning.ProductId).First(&cloudProductBaseline).Error; err != nil {
		return nil, err
	}
	//查询服务器配置表
	var serverBaselineList []*entity.ServerBaseline
	if err := data.DB.Where("version_id = ?", cloudProductBaseline.VersionId).Find(&serverBaselineList).Error; err != nil {
		return nil, err
	}
	var CpuTypeList []string
	for _, v := range serverBaselineList {
		CpuTypeList = append(CpuTypeList, v.CpuType)
	}
	return CpuTypeList, nil
}

func ListServerCapacity(request *Request) ([]*ResponseCapacity, error) {
	//云产品规划表
	var cloudProductPlanning = &entity.CloudProductPlanning{}
	if err := data.DB.Where("plan_id = ?", request.PlanId).First(&cloudProductPlanning).Error; err != nil {
		return nil, err
	}
	//查询云产品配置表
	var cloudProductBaseline = &entity.CloudProductBaseline{}
	if err := data.DB.Where("id = ?", cloudProductPlanning.ProductId).First(&cloudProductBaseline).Error; err != nil {
		return nil, err
	}
	//todo 缺少产品容量指标
	var capacityList []*ResponseCapacity

	return capacityList, nil
}

func ListServerModel(request *Request) ([]*ResponseModel, error) {
	//云产品规划表
	var cloudProductPlanning = &entity.CloudProductPlanning{}
	if err := data.DB.Where("plan_id = ?", request.PlanId).First(&cloudProductPlanning).Error; err != nil {
		return nil, err
	}
	//查询云产品配置表
	var cloudProductBaseline = &entity.CloudProductBaseline{}
	if err := data.DB.Where("id = ?", cloudProductPlanning.ProductId).First(&cloudProductBaseline).Error; err != nil {
		return nil, err
	}
	//todo 版本号、产品、角色，查询服务器基线可用机型，服务器基线表缺少角色
	var modelList []*ResponseModel

	return modelList, nil
}

func CreateServer(request *Request) error {
	if err := checkBusiness(request, true); err != nil {
		return err
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("planId = ?", request.PlanId).Delete(&entity.ServerPlanning{})

		return nil
	}); err != nil {
		return err
	}
	now := datetime.GetNow()
	cloudPlatformEntity := &entity.ServerPlanning{
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
	projectEntity := &entity.ServerPlanning{
		Id:           request.Id,
		UpdateUserId: request.UserId,
		UpdateTime:   datetime.GetNow(),
	}
	if err := data.DB.Updates(projectEntity).Error; err != nil {
		return err
	}
	return nil
}

func QueryServerPlanningListByPlanId(planId int64) ([]entity.ServerPlanning, error) {
	var serverPlanningList []entity.ServerPlanning
	if err := data.DB.Table(entity.ServerPlanningTable).Where("plan_id = ? and delete_state = 0", planId).Find(&serverPlanningList).Error; err != nil {
		return serverPlanningList, err
	}
	return serverPlanningList, nil
}

func checkBusiness(request *Request, isCreate bool) error {
	return nil
}
