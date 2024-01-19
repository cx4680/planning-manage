package cloud_product

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"
	"strings"
	"time"
)

func getVersionListByProjectId(projectId int64) ([]entity.SoftwareVersion, error) {
	var project entity.ProjectManage
	if err := data.DB.Where("id = ? AND delete_state = 0", projectId).Find(&project).Error; err != nil {
		return nil, err
	}
	var cloudPlatform = &entity.CloudPlatformManage{}
	if err := data.DB.Where("id = ? AND delete_state = 0", project.CloudPlatformId).Find(&cloudPlatform).Error; err != nil {
		return nil, err
	}
	var versionList []entity.SoftwareVersion
	if err := data.DB.Where("cloud_platform_type = ?", cloudPlatform.Type).Find(&versionList).Error; err != nil {
		log.Errorf("[getVersionListByProjectId] get version by platformType error,%v", err)
		return nil, err
	}
	return versionList, nil
}

func getCloudProductBaseListByVersionId(versionId int64) ([]CloudProductBaselineResponse, error) {
	var baselineList []entity.CloudProductBaseline
	if err := data.DB.Where("version_id = ? ", versionId).Find(&baselineList).Error; err != nil {
		return nil, err
	}
	// 查询依赖的云产品
	cloudProductDependList, err := getDependProductIds()
	if err != nil {
		return nil, err
	}
	var responseList []CloudProductBaselineResponse
	for _, baseline := range baselineList {
		sellSpecs := make([]string, 1)
		if strings.Index(baseline.SellSpecs, ",") > 0 {
			sellSpecs = strings.Split(baseline.SellSpecs, ",")
		} else {
			sellSpecs[0] = baseline.SellSpecs
		}
		var dependProductId int64
		var dependProductName string
		for _, depend := range cloudProductDependList {
			if depend.ID == baseline.Id {
				dependProductId = depend.DependId
				dependProductName = depend.DependProductName
			}
		}
		responseData := CloudProductBaselineResponse{
			ID:                baseline.Id,
			VersionId:         baseline.VersionId,
			ProductType:       baseline.ProductType,
			ProductName:       baseline.ProductName,
			ProductCode:       baseline.ProductCode,
			SellSpecs:         sellSpecs,
			AuthorizedUnit:    baseline.AuthorizedUnit,
			WhetherRequired:   baseline.WhetherRequired,
			Instructions:      baseline.Instructions,
			DependProductId:   dependProductId,
			DependProductName: dependProductName,
		}
		responseList = append(responseList, responseData)
	}
	return responseList, nil
}

func saveCloudProductPlanning(request CloudProductPlanningRequest, currentUserId string) error {
	var cloudProductPlanningList []entity.CloudProductPlanning
	for _, cloudProduct := range request.ProductList {
		cloudProductPlanning := entity.CloudProductPlanning{
			PlanId:      request.PlanId,
			ProductId:   cloudProduct.ProductId,
			VersionId:   request.VersionId,
			SellSpec:    cloudProduct.SellSpec,
			ServiceYear: request.ServiceYear,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}
		cloudProductPlanningList = append(cloudProductPlanningList, cloudProductPlanning)
	}

	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		//删除原有的清单
		if err := tx.Delete(&entity.CloudProductPlanning{}, "plan_id=?", request.PlanId).Error; err != nil {
			log.Errorf("[saveCloudProductPlanning] delete cloudProductPlanning by planId error,%v", err)
			return err
		}
		// 保存云服务规划清单
		if err := tx.Table(entity.CloudProductPlanningTable).CreateInBatches(&cloudProductPlanningList, len(cloudProductPlanningList)).Error; err != nil {
			log.Errorf("[saveCloudProductPlanning] batch create cloudProductPlanning error, %v", err)
			return err
		}
		// 更新方案业务规划阶段
		if err := tx.Table(entity.PlanManageTable).Where("id = ?", request.PlanId).
			Update("business_plan_stage", constant.BusinessPlanningServer).
			Update("stage", constant.PlanStagePlanning).
			Update("update_user_id", currentUserId).
			Update("update_time", time.Now()).Error; err != nil {
			log.Errorf("[saveCloudProductPlanning] update plan business stage error, %v", err)
			return err
		}
		return nil
	}); err != nil {
		log.Errorf("[saveCloudProductPlanning] error, %v", err)
		return err
	}
	return nil
}

func listCloudProductPlanningByPlanId(planId int64) ([]entity.CloudProductPlanning, error) {
	var cloudProductPlanningList []entity.CloudProductPlanning
	if err := data.DB.Table(entity.CloudProductPlanningTable).Where("plan_id=?", planId).Scan(&cloudProductPlanningList).Error; err != nil {
		log.Errorf("[listCloudProductPlanningByPlanId] error, %v", err)
		return nil, err
	}
	return cloudProductPlanningList, nil
}

func exportCloudProductPlanningByPlanId(planId int64) (string, []CloudProductPlanningExportResponse, error) {
	var planManage entity.PlanManage
	if err := data.DB.Table(entity.PlanManageTable).Where("id=?", planId).Scan(&planManage).Error; err != nil {
		log.Errorf("[exportCloudProductPlanningByPlanId] get planManage by id err, %v", err)
		return "", nil, err
	}

	var projectManage entity.ProjectManage
	if err := data.DB.Table(entity.ProjectManageTable).Where("id=?", planManage.ProjectId).Scan(&projectManage).Error; err != nil {
		log.Errorf("[exportCloudProductPlanningByPlanId] get projectManage by id err, %v", err)
		return "", nil, err
	}

	var cloudProductPlanningList []entity.CloudProductPlanning
	if err := data.DB.Table(entity.CloudProductPlanningTable).Where("plan_id=?", planId).Scan(&cloudProductPlanningList).Error; err != nil {
		log.Errorf("[exportCloudProductPlanningByPlanId] error, %v", err)
		return "", nil, err
	}

	var response []CloudProductPlanningExportResponse
	if err := data.DB.Table("cloud_product_planning cpp").Select("cpb.product_type,cpb.product_name, cpb.instructions, cpp.sell_spec").
		Joins("LEFT JOIN cloud_product_baseline cpb ON cpb.id = cpp.product_id").
		Where("cpp.plan_id=?", planId).
		Find(&response).Error; err != nil {
		log.Errorf("[exportCloudProductPlanningByPlanId] query db error")
		return "", nil, err
	}
	return projectManage.Name + "-" + planManage.Name + "-" + "云产品清单", response, nil
}

func getDependProductIds() ([]CloudProductBaselineDependResponse, error) {
	var cloudProductDependList []CloudProductBaselineDependResponse
	if err := data.DB.Table("cloud_product_depend_rel cpdr").
		Select("cpdr.product_id id, cpb.id dependId, cpb.product_name dependProductName, cpb.product_code dependProductCode").
		Joins("LEFT JOIN cloud_product_baseline cpb ON cpb.id = cpdr.depend_product_id").
		Find(&cloudProductDependList).Error; err != nil {

		log.Errorf("[getDependProductIds] query db err, %v", err)
		return nil, err
	}
	return cloudProductDependList, nil
}
