package cloud_product

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"
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

func getCloudProductBaseListByVersionId(versionId int64) ([]entity.CloudProductBaseline, error) {
	var baselineList []entity.CloudProductBaseline
	if err := data.DB.Where("version_id = ? ", versionId).Find(&baselineList).Error; err != nil {
		return nil, err
	}
	return baselineList, nil
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
		if err := tx.Model(entity.PlanManage{}).Where("id = ?", request.PlanId).
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
