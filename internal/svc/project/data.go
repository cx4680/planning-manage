package project

import (
	"errors"

	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
)

func PageProject(request *Request) ([]*Project, int64, error) {
	// 缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	screenSql, screenParams, orderSql := " delete_state = ? ", []interface{}{0}, " update_time "
	if request.Id != 0 {
		screenSql += " AND id = ? "
		screenParams = append(screenParams, request.Id)
	}
	if request.CustomerId != 0 {
		screenSql += " AND customer_id = ? "
		screenParams = append(screenParams, request.CustomerId)
	} else {
		// 查询账号下关联的所有客户，客户接口人和项目成员
		customerIdList, err := getCustomerIdList(db, request)
		if err != nil {
			return nil, 0, err
		}
		if len(customerIdList) == 0 {
			return nil, 0, nil
		}
		screenSql += " AND customer_id IN (?) "
		screenParams = append(screenParams, customerIdList)
	}
	if util.IsNotBlank(request.Name) {
		screenSql += " AND name LIKE CONCAT('%',?,'%') "
		screenParams = append(screenParams, request.Name)
	}
	if util.IsNotBlank(request.CustomerName) {
		var customerList []*entity.CustomerManage
		if err := db.Where("customer_name LIKE CONCAT('%',?,'%')", request.CustomerName).Find(&customerList).Error; err != nil {
			return nil, 0, err
		}
		var customerIdList []int64
		for _, v := range customerList {
			customerIdList = append(customerIdList, v.ID)
		}
		screenSql += " AND customer_id IN (?) "
		screenParams = append(screenParams, customerIdList)
	}
	if util.IsNotBlank(request.Type) {
		screenSql += " AND type = ? "
		screenParams = append(screenParams, request.Type)
	}
	if util.IsNotBlank(request.Stage) {
		screenSql += " AND stage = ? "
		screenParams = append(screenParams, request.Stage)
	}
	switch request.SortField {
	case "createTime":
		orderSql = " create_time "
	case "updateTime":
		orderSql = " update_time "
	}
	switch request.Sort {
	case "ascend":
		orderSql += " asc "
	case "descend":
		orderSql += " desc "
	default:
		orderSql += " desc "
	}
	var count int64
	if err := db.Model(&entity.ProjectManage{}).Where(screenSql, screenParams...).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	var list []*Project
	if err := db.Model(&entity.ProjectManage{}).Where(screenSql, screenParams...).Order(orderSql).Offset((request.Current - 1) * request.PageSize).Limit(request.PageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	list, err := buildResponse(list)
	if err != nil {
		return nil, 0, err
	}
	return list, count, nil
}

func CreateProject(request *Request) error {
	if err := checkBusiness(request, true); err != nil {
		return err
	}
	now := datetime.GetNow()
	projectEntity := &entity.ProjectManage{
		Name:            request.Name,
		CloudPlatformId: request.CloudPlatformId,
		RegionId:        request.RegionId,
		AzId:            request.AzId,
		CellId:          request.CellId,
		CustomerId:      request.CustomerId,
		Type:            request.Type,
		Stage:           constant.ProjectStagePlanning,
		DeleteState:     0,
		CreateUserId:    request.UserId,
		CreateTime:      now,
		UpdateUserId:    request.UserId,
		UpdateTime:      now,
	}
	CloudPlatformEntity := &entity.CloudPlatformManage{
		Id:   request.CloudPlatformId,
		Type: request.CloudPlatformType,
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&projectEntity).Error; err != nil {
			return err
		}
		if err := tx.Updates(&CloudPlatformEntity).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func UpdateProject(request *Request) error {
	if err := checkBusiness(request, false); err != nil {
		return err
	}
	now := datetime.GetNow()
	projectEntity := &entity.ProjectManage{
		Id:           request.Id,
		Name:         request.Name,
		Stage:        request.Stage,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
	}
	if err := data.DB.Updates(&projectEntity).Error; err != nil {
		return err
	}
	return nil
}

func DeleteProject(request *Request) error {
	// 校验该项目下是否有方案
	var planCount int64
	if err := data.DB.Model(&entity.PlanManage{}).Where("project_id = ? AND delete_state = ?", request.Id, 0).Count(&planCount).Error; err != nil {
		return err
	}
	if planCount != 0 {
		return errors.New("无法删除，该项目有方案未删除")
	}
	now := datetime.GetNow()
	projectEntity := &entity.ProjectManage{
		Id:           request.Id,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  1,
	}
	if err := data.DB.Updates(&projectEntity).Error; err != nil {
		return err
	}
	return nil
}

func checkBusiness(request *Request, isCreate bool) error {
	if isCreate {
		// 校验customerId
		var customerCount int64
		if err := data.DB.Model(&entity.CustomerManage{}).Where("id = ? AND delete_state = ?", request.CustomerId, 0).Count(&customerCount).Error; err != nil {
			return err
		}
		if customerCount == 0 {
			return errors.New("customerId参数错误")
		}
		// 校验projectType
		var projectTypeCount int64
		if err := data.DB.Model(&entity.ConfigItem{}).Where("p_id = ? AND code = ?", "3", request.Type).Count(&projectTypeCount).Error; err != nil {
			return err
		}
		if projectTypeCount == 0 {
			return errors.New("type参数错误")
		}
		// 校验cloudPlatform
		var cloudPlatform = &entity.CloudPlatformManage{}
		if err := data.DB.Where("id = ? AND delete_state = ?", request.CloudPlatformId, 0).Find(&cloudPlatform).Error; err != nil {
			return err
		}
		if cloudPlatform.Id == 0 {
			return errors.New("cloudPlatformId参数错误")
		}
		if util.IsBlank(cloudPlatform.Type) && util.IsBlank(request.CloudPlatformType) {
			return errors.New("云平台类型未设置")
		}
		// 校验regionId
		var regionCount int64
		if err := data.DB.Model(&entity.RegionManage{}).Where("id = ? AND delete_state = ?", request.RegionId, 0).Count(&regionCount).Error; err != nil {
			return err
		}
		if regionCount == 0 {
			return errors.New("region不存在")
		}
		// 校验azId
		var azCount int64
		if err := data.DB.Model(&entity.AzManage{}).Where("id = ? AND delete_state = ?", request.AzId, 0).Count(&azCount).Error; err != nil {
			return err
		}
		if azCount == 0 {
			return errors.New("az不存在")
		}
	} else {
		// 校验projectId
		var projectCount int64
		if err := data.DB.Model(&entity.ProjectManage{}).Where("id = ? AND delete_state = ?", request.Id, 0).Count(&projectCount).Error; err != nil {
			return err
		}
		if projectCount == 0 {
			return errors.New("project不存在")
		}
		// 校验projectStage
		if util.IsNotBlank(request.Stage) {
			var projectStageCount int64
			if err := data.DB.Model(&entity.ConfigItem{}).Where("p_id = ? AND code = ?", "4", request.Stage).Count(&projectStageCount).Error; err != nil {
				return err
			}
			if projectStageCount == 0 {
				return errors.New("stage参数错误")
			}
		}
	}
	return nil
}

func buildResponse(list []*Project) ([]*Project, error) {
	if len(list) == 0 {
		return list, nil
	}
	var projectIdList, customerIdList, cloudPlatformIdList, regionIdList, azIdList, cellIdList []int64
	for _, v := range list {
		projectIdList = append(projectIdList, v.Id)
		customerIdList = append(customerIdList, v.CustomerId)
		cloudPlatformIdList = append(cloudPlatformIdList, v.CloudPlatformId)
		regionIdList = append(regionIdList, v.RegionId)
		azIdList = append(azIdList, v.AzId)
		cellIdList = append(cellIdList, v.CellId)
	}
	// 查询方案数量
	var planList []*entity.PlanManage
	if err := data.DB.Model(&entity.PlanManage{}).Where("project_id IN (?) AND delete_state = ?", projectIdList, 0).Find(&planList).Error; err != nil {
		return nil, err
	}
	var planCountMap = make(map[int64]int)
	for _, v := range planList {
		planCountMap[v.ProjectId]++
	}
	// 查询客户名称
	var customerList []*entity.CustomerManage
	if err := data.DB.Where("id IN (?)", customerIdList).Find(&customerList).Error; err != nil {
		return nil, err
	}
	var customerMap = make(map[int64]*entity.CustomerManage)
	for _, v := range customerList {
		customerMap[v.ID] = v
	}
	// 查询云平台名称
	var cloudPlatformList []*entity.CloudPlatformManage
	if err := data.DB.Where("id IN (?)", cloudPlatformIdList).Find(&cloudPlatformList).Error; err != nil {
		return nil, err
	}
	var cloudPlatformMap = make(map[int64]*entity.CloudPlatformManage)
	for _, v := range cloudPlatformList {
		cloudPlatformMap[v.Id] = v
	}
	// 查询region名称
	var regionList []*entity.RegionManage
	if err := data.DB.Where("id IN (?)", regionIdList).Find(&regionList).Error; err != nil {
		return nil, err
	}
	var regionMap = make(map[int64]*entity.RegionManage)
	for _, v := range regionList {
		regionMap[v.Id] = v
	}
	// 查询az名称
	var azList []*entity.AzManage
	if err := data.DB.Where("id IN (?)", azIdList).Find(&azList).Error; err != nil {
		return nil, err
	}
	var azMap = make(map[int64]*entity.AzManage)
	for _, v := range azList {
		azMap[v.Id] = v
	}
	// 查询cell名称
	var cellList []*entity.CellManage
	if err := data.DB.Where("id IN (?)", cellIdList).Find(&cellList).Error; err != nil {
		return nil, err
	}
	var cellMap = make(map[int64]*entity.CellManage)
	for _, v := range cellList {
		cellMap[v.Id] = v
	}

	for i, v := range list {
		list[i].PlanCount = planCountMap[v.Id]
		if customerMap[v.CustomerId] != nil {
			list[i].CustomerName = customerMap[v.CustomerId].CustomerName
		}
		if cloudPlatformMap[v.CloudPlatformId] != nil {
			list[i].CloudPlatformName = cloudPlatformMap[v.CloudPlatformId].Name
			list[i].CloudPlatformType = cloudPlatformMap[v.CloudPlatformId].Type
		}
		if regionMap[v.RegionId] != nil {
			list[i].RegionName = regionMap[v.RegionId].Name
		}
		if azMap[v.AzId] != nil {
			list[i].AzCode = azMap[v.AzId].Code
		}
		if cellMap[v.CellId] != nil {
			list[i].CellName = cellMap[v.CellId].Name
		}
	}
	return list, nil
}

func getCustomerIdList(db *gorm.DB, request *Request) ([]int64, error) {
	var customerManageList []*entity.CustomerManage
	if err := db.Where("leader_id = ? AND delete_state = ?", request.UserId, 0).Find(&customerManageList).Error; err != nil {
		return nil, err
	}
	var permissionsManageList []*entity.PermissionsManage
	if err := db.Where("user_id = ? AND delete_state = ?", request.UserId, 0).Find(&permissionsManageList).Error; err != nil {
		return nil, err
	}
	var customerIdList []int64
	for _, v := range customerManageList {
		customerIdList = append(customerIdList, v.ID)
	}
	for _, v := range permissionsManageList {
		customerIdList = append(customerIdList, v.CustomerId)
	}
	return customerIdList, nil
}
