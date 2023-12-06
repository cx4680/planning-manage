package cloud_product

import (
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
	var cloudProductDependRelList []entity.CloudProductDependRel
	if err := data.DB.Find(&cloudProductDependRelList).Error; err != nil {
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
		for _, rel := range cloudProductDependRelList {
			if rel.ProductId == baseline.Id {
				dependProductId = rel.DependProductId
			}
		}
		responseData := CloudProductBaselineResponse{
			Id:              baseline.Id,
			VersionId:       baseline.VersionId,
			ProductType:     baseline.ProductType,
			ProductName:     baseline.ProductName,
			ProductCode:     baseline.ProductCode,
			SellSpecs:       sellSpecs,
			AuthorizedUnit:  baseline.AuthorizedUnit,
			WhetherRequired: baseline.WhetherRequired,
			Instructions:    baseline.Instructions,
			DependProductId: dependProductId,
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
			SellSpec:    cloudProduct.SellSpec,
			ServiceYear: request.ServiceYear,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}
		cloudProductPlanningList = append(cloudProductPlanningList, cloudProductPlanning)
	}
	if err := data.DB.Transaction(func(tx *gorm.DB) error {
		// 保存云服务规划清单
		if err := tx.Table(entity.CloudProductPlanningTable).CreateInBatches(&cloudProductPlanningList, len(cloudProductPlanningList)).Error; err != nil {
			log.Errorf("[saveCloudProductPlanning] batch create cloudProductPlanning error, %v", err)
			return err
		}
		// 更新方案业务规划阶段
		if err := tx.Table(entity.PlanManageTable).Where("id = ?", request.PlanId).
			Update("business_plan_stage", 1).
			Update("update_user_id", currentUserId).
			Update("update_time", time.Now()).Error; err != nil {
			log.Errorf("[saveCloudProductPlanning] update plan business stage error, %v", err)
			return err
		}
		// TODO 删除服务器规划清单列表

		// TODO 删除容量规划列表

		return nil
	}); err != nil {
		log.Errorf("[saveCloudProductPlanning] error, %v", err)
		return err
	}
	return nil
}

func listCloudProductPlanningByPlanId(planId int64) ([]entity.CloudProductPlanning, error) {
	var cloudProductPlanningList []entity.CloudProductPlanning
	if err := data.DB.Where("plan_id=?", planId).Scan(&cloudProductPlanningList).Error; err != nil {
		log.Errorf("[listCloudProductPlanningByPlanId] error, %v", err)
		return nil, err
	}
	return cloudProductPlanningList, nil
}

func exportCloudProductPlanningByPlanId(planId int64) (string, []CloudProductPlanningExportResponse, error) {
	var planManage entity.PlanManage
	if err := data.DB.Where("id=?", planId).Scan(planManage).Error; err != nil {
		log.Errorf("[exportCloudProductPlanningByPlanId] get planManage by id err, %v", err)
		return "", nil, err
	}

	var projectManage entity.ProjectManage
	if err := data.DB.Where("id=?", planManage.ProjectId).Scan(projectManage).Error; err != nil {
		log.Errorf("[exportCloudProductPlanningByPlanId] get projectManage by id err, %v", err)
		return "", nil, err
	}

	var cloudProductPlanningList []entity.CloudProductPlanning
	if err := data.DB.Where("plan_id=?", planId).Scan(&cloudProductPlanningList).Error; err != nil {
		log.Errorf("[exportCloudProductPlanningByPlanId] error, %v", err)
		return "", nil, err
	}

	var response []CloudProductPlanningExportResponse
	if err := data.DB.Table("cloud_product_planning").Select("cloud_product_baseline.product_type,cloud_product_baseline.product_name, cloud_product_baseline.instructions, cloud_product_planning.sell_specs").
		Joins("LEFT JOIN cloud_product_baseline ON cloud_product_baseline.id = cloud_product_planning.product_id").
		Where("cloud_product_planning.plan_id=?", planId).
		Find(&response).Error; err != nil {
		log.Errorf("[exportCloudProductPlanningByPlanId] query db error")
		return "", nil, err
	}
	return projectManage.Name + "-" + planManage.Name + "-" + "云产品清单", response, nil
}
