package baseline

import (
	"github.com/opentrx/seata-golang/v2/pkg/util/log"

	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
)

func QueryCloudPlatformType() ([]string, error) {
	var cloudPlatformTypes []string
	var configItem entity.ConfigItem
	if err := data.DB.Table(entity.ConfigItemTable).Where("code = ?", "cloudPlatformType").Find(&configItem).Error; err != nil {
		return cloudPlatformTypes, err
	}
	if err := data.DB.Table(entity.ConfigItemTable).Where("p_id = ?", configItem.Id).Pluck("code", &cloudPlatformTypes).Error; err != nil {
		return cloudPlatformTypes, err
	}
	return cloudPlatformTypes, nil
}

func QuerySoftwareVersionByVersion(version string, cloudPlatformType string) (entity.SoftwareVersion, error) {
	var softwareVersion entity.SoftwareVersion
	if err := data.DB.Table(entity.SoftwareVersionTable).Where("software_version = ? AND cloud_platform_type = ?", version, cloudPlatformType).First(&softwareVersion).Error; err != nil {
		return softwareVersion, err
	}
	return softwareVersion, nil
}

func CreateSoftwareVersion(softwareVersion entity.SoftwareVersion) error {
	if err := data.DB.Table(entity.SoftwareVersionTable).Create(&softwareVersion).Scan(&softwareVersion).Error; err != nil {
		log.Errorf("insert software version error: ", err)
		return err
	}
	return nil
}

func UpdateSoftwareVersion(softwareVersion entity.SoftwareVersion) error {
	if err := data.DB.Table(entity.SoftwareVersionTable).Updates(&softwareVersion).Error; err != nil {
		log.Errorf("update software error: ", err)
		return err
	}
	return nil
}

func QueryNodeRoleBaselineByVersionId(versionId int64) ([]entity.NodeRoleBaseline, error) {
	var nodeRoleBaselines []entity.NodeRoleBaseline
	if err := data.DB.Table(entity.NodeRoleBaselineTable).Where("version_id = ? ", versionId).Find(&nodeRoleBaselines).Error; err != nil {
		return nodeRoleBaselines, err
	}
	return nodeRoleBaselines, nil
}

func BatchCreateNodeRoleBaseline(nodeRoleBaselines []entity.NodeRoleBaseline) error {
	if err := data.DB.Table(entity.SoftwareVersionTable).Create(&nodeRoleBaselines).Scan(&nodeRoleBaselines).Error; err != nil {
		log.Errorf("batch insert nodeRoleBaseline error: ", err)
		return err
	}
	return nil
}

func QueryServiceBaselineById(id int64) (*entity.ServerBaseline, error) {
	var serverBaseline entity.ServerBaseline
	if err := data.DB.Table(entity.ServerBaselineTable).Where("id=?", id).Scan(&serverBaseline).Error; err != nil {
		log.Errorf("[queryServiceBaselineById] query service baseline error, %v", err)
		return nil, err
	}
	return &serverBaseline, nil
}

func BatchCreateNodeRoleMixedDeploy(nodeRoleMixedDeploys []entity.NodeRoleMixedDeploy) error {
	if err := data.DB.Table(entity.NodeRoleMixedDeployTable).Create(&nodeRoleMixedDeploys).Scan(&nodeRoleMixedDeploys).Error; err != nil {
		log.Errorf("batch insert nodeRoleMixedDeploy error: ", err)
		return err
	}
	return nil
}

func GetVersionIdByPlanId(planId int64) (int64, error) {
	var versionId int64
	if err := data.DB.Raw("SELECT version_id FROM cloud_product_baseline WHERE id = ( SELECT product_id FROM cloud_product_planning WHERE plan_id = ? LIMIT 1 )", planId).Scan(&versionId).Error; err != nil {
		log.Errorf("[GetVersionIdByPlanId] error, %v", err)
		return 0, err
	}
	return versionId, nil
}
