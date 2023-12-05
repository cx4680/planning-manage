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

func BatchCreateCloudProductBaseline(cloudProductBaselines []entity.CloudProductBaseline) error {
	if err := data.DB.Table(entity.CloudProductBaselineTable).Create(&cloudProductBaselines).Scan(&cloudProductBaselines).Error; err != nil {
		log.Errorf("batch insert cloudProductBaseline error: ", err)
		return err
	}
	return nil
}

func QueryCloudProductBaselineByVersionId(versionId int64) ([]entity.CloudProductBaseline, error) {
	var cloudProductBaselines []entity.CloudProductBaseline
	if err := data.DB.Table(entity.CloudProductBaselineTable).Where("version_id = ? ", versionId).Find(&cloudProductBaselines).Error; err != nil {
		return cloudProductBaselines, err
	}
	return cloudProductBaselines, nil
}

func BatchCreateCloudProductDependRel(cloudProductDependRels []entity.CloudProductDependRel) error {
	if err := data.DB.Table(entity.CloudProductDependRelTable).Create(&cloudProductDependRels).Scan(&cloudProductDependRels).Error; err != nil {
		log.Errorf("batch insert cloudProductDependRel error: ", err)
		return err
	}
	return nil
}

func BatchCreateCloudProductNodeRoleRel(cloudProductNodeRoleRels []entity.CloudProductNodeRoleRel) error {
	if err := data.DB.Table(entity.CloudProductNodeRoleTable).Create(&cloudProductNodeRoleRels).Scan(&cloudProductNodeRoleRels).Error; err != nil {
		log.Errorf("batch insert cloudProductNodeRoleRel error: ", err)
		return err
	}
	return nil
}
