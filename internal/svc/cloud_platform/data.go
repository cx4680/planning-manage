package cloud_platform

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"errors"
	"gorm.io/gorm"
)

func ListCloudPlatform(request *Request) ([]*CloudPlatform, error) {
	screenSql, screenParams, orderSql := " delete_state = ? ", []interface{}{0}, " create_time "
	if request.CustomerId != 0 {
		screenSql += " AND customer_id = ? "
		screenParams = append(screenParams, request.CustomerId)
	}
	switch request.SortField {
	case "createTime":
		orderSql = " create_time "
	case "updateTime":
		orderSql = " update_time "
	}
	switch request.Sort {
	case "asc":
		orderSql += " asc "
	case "desc":
		orderSql += " desc "
	default:
		orderSql += " asc "
	}
	var list []*CloudPlatform
	if err := data.DB.Model(&entity.CloudPlatformManage{}).Where(screenSql, screenParams...).Order(orderSql).Find(&list).Error; err != nil {
		return nil, err
	}
	//查询负责人名称
	var customerIdList []int64
	for _, v := range list {
		customerIdList = append(customerIdList, v.CustomerId)
	}
	var customerList []*entity.CustomerManage
	if err := data.DB.Where("id IN (?)", customerIdList).Find(&customerList).Error; err != nil {
		return nil, err
	}
	var customerMap = make(map[int64]*entity.CustomerManage)
	for _, v := range customerList {
		customerMap[v.ID] = v
	}
	for i, v := range list {
		if customerMap[v.CustomerId] != nil {
			list[i].LeaderId = customerMap[v.CustomerId].LeaderId
			list[i].LeaderName = customerMap[v.CustomerId].LeaderName
		}
	}
	return list, nil
}

func CreateCloudPlatform(request *Request) error {
	if err := checkBusiness(request, true); err != nil {
		return err
	}
	now := datetime.GetNow()
	cloudPlatformEntity := &entity.CloudPlatformManage{
		Name:         request.Name,
		Type:         request.Type,
		CustomerId:   request.CustomerId,
		CreateUserId: request.UserId,
		CreateTime:   now,
		UpdateUserId: request.UserId,
		UpdateTime:   now,
		DeleteState:  0,
	}
	if err := data.DB.Create(&cloudPlatformEntity).Error; err != nil {
		return err
	}
	return nil
}

func UpdateCloudPlatform(request *Request) error {
	if err := checkBusiness(request, true); err != nil {
		return err
	}
	cloudPlatformEntity := &entity.CloudPlatformManage{
		Id:           request.Id,
		Name:         request.Name,
		Type:         request.Type,
		CustomerId:   request.CustomerId,
		UpdateUserId: request.UserId,
		UpdateTime:   datetime.GetNow(),
	}
	if err := data.DB.Updates(&cloudPlatformEntity).Error; err != nil {
		return err
	}
	return nil
}

func TreeCloudPlatform(request *Request) (*ResponseTree, error) {
	var responseTree = &ResponseTree{}
	//缓存预编译 会话模式
	db := data.DB.Session(&gorm.Session{PrepareStmt: true})
	//查询region
	var RegionList []*entity.RegionManage
	if err := db.Where(" delete_state = ? AND cloud_platform_id = ?", 0, request.CloudPlatformId).Find(&RegionList).Error; err != nil {
		return nil, err
	}
	var regionIdList []int64
	for _, v := range RegionList {
		responseTree.RegionList = append(responseTree.RegionList, &ResponseTreeRegion{Region: v})
		regionIdList = append(regionIdList, v.Id)
	}
	//查询az
	var azList []*entity.AzManage
	if err := db.Where(" delete_state = ? AND region_id IN (?)", 0, regionIdList).Find(&azList).Error; err != nil {
		return nil, err
	}
	var regionAzMap = make(map[int64][]*ResponseTreeAz)
	var azIdList []int64
	for _, v := range azList {
		regionAzMap[v.RegionId] = append(regionAzMap[v.RegionId], &ResponseTreeAz{Az: v})
		azIdList = append(azIdList, v.Id)
	}
	for i, v := range responseTree.RegionList {
		responseTree.RegionList[i].AzList = regionAzMap[v.Region.Id]
	}
	//查询机房信息
	var machineRoomList []*entity.MachineRoom
	if err := db.Model(&entity.MachineRoom{}).Where("az_id IN (?)", azIdList).Find(&machineRoomList).Error; err != nil {
		return nil, err
	}
	var machineRoomMap = make(map[int64][]*entity.MachineRoom)
	for _, v := range machineRoomList {
		machineRoomMap[v.AzId] = append(machineRoomMap[v.AzId], v)
	}
	//查询cell
	var cellList []*entity.CellManage
	if err := db.Model(&entity.CellManage{}).Where("delete_state = ? AND az_id IN (?)", 0, azIdList).Find(&cellList).Error; err != nil {
		return nil, err
	}
	var azCellMap = make(map[int64][]*ResponseTreeCell)
	var cellIdList []int64
	for _, v := range cellList {
		cellIdList = append(cellIdList, v.Id)
	}
	//查询方案数量
	var projectList []*entity.ProjectManage
	if err := data.DB.Model(&entity.ProjectManage{}).Where("cell_id IN (?)", cellIdList).Find(&projectList).Error; err != nil {
		return nil, err
	}
	var projectCountMap = make(map[int64]int)
	for _, v := range projectList {
		projectCountMap[v.CellId]++
	}
	for _, v := range cellList {
		azCellMap[v.AzId] = append(azCellMap[v.AzId], &ResponseTreeCell{Cell: v, ProjectCount: projectCountMap[v.Id]})
		cellIdList = append(cellIdList, v.Id)
	}
	for i, v := range responseTree.RegionList {
		for i2, v2 := range v.AzList {
			responseTree.RegionList[i].AzList[i2].MachineRoomList = machineRoomMap[v2.Az.Id]
			responseTree.RegionList[i].AzList[i2].CellList = azCellMap[v2.Az.Id]
		}
	}
	return responseTree, nil
}

func checkBusiness(request *Request, isCreate bool) error {
	//校验cloudPlatformType
	if util.IsNotBlank(request.Type) {
		var cloudPlatformTypeCount int64
		if err := data.DB.Model(&entity.ConfigItem{}).Where("p_id = ? AND code = ?", "1", request.Type).Count(&cloudPlatformTypeCount).Error; err != nil {
			return err
		}
		if cloudPlatformTypeCount == 0 {
			return errors.New("type参数错误")
		}
	}
	if !isCreate {
		//校验cloudPlatformId
		var cloudPlatformCount int64
		if err := data.DB.Model(&entity.CloudPlatformManage{}).Where("id = ? AND delete_state = ?", request.Id, 0).Count(&cloudPlatformCount).Error; err != nil {
			return err
		}
		if cloudPlatformCount == 0 {
			return errors.New("云平台不存在")
		}
	}
	return nil
}
