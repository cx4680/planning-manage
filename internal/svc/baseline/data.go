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

func CreateSoftwareVersion(softwareVersion *entity.SoftwareVersion) error {
	if err := data.DB.Table(entity.SoftwareVersionTable).Create(softwareVersion).Scan(softwareVersion).Error; err != nil {
		log.Errorf("insert software version error: %v", err)
		return err
	}
	return nil
}

func UpdateSoftwareVersion(softwareVersion entity.SoftwareVersion) error {
	if err := data.DB.Table(entity.SoftwareVersionTable).Updates(&softwareVersion).Error; err != nil {
		log.Errorf("update software error: %v", err)
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
	if len(nodeRoleBaselines) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.NodeRoleBaselineTable).Create(&nodeRoleBaselines).Scan(&nodeRoleBaselines).Error; err != nil {
		log.Errorf("batch insert nodeRoleBaseline error: %v", err)
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
	if len(nodeRoleMixedDeploys) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.NodeRoleMixedDeployTable).Create(&nodeRoleMixedDeploys).Scan(&nodeRoleMixedDeploys).Error; err != nil {
		log.Errorf("batch insert nodeRoleMixedDeploy error: %v", err)
		return err
	}
	return nil
}

func BatchCreateCloudProductBaseline(cloudProductBaselines []entity.CloudProductBaseline) error {
	if len(cloudProductBaselines) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.CloudProductBaselineTable).Create(&cloudProductBaselines).Scan(&cloudProductBaselines).Error; err != nil {
		log.Errorf("batch insert cloudProductBaseline error: %v", err)
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
		log.Errorf("batch insert cloudProductDependRel error: %v", err)
		return err
	}
	return nil
}

func BatchCreateCloudProductNodeRoleRel(cloudProductNodeRoleRels []entity.CloudProductNodeRoleRel) error {
	if err := data.DB.Table(entity.CloudProductNodeRoleTable).Create(&cloudProductNodeRoleRels).Scan(&cloudProductNodeRoleRels).Error; err != nil {
		log.Errorf("batch insert cloudProductNodeRoleRel error: %v", err)
		return err
	}
	return nil
}

func QueryServerBaselineByVersionId(versionId int64) ([]entity.ServerBaseline, error) {
	var serverBaselines []entity.ServerBaseline
	if err := data.DB.Table(entity.ServerBaselineTable).Where("version_id = ? ", versionId).Find(&serverBaselines).Error; err != nil {
		return serverBaselines, err
	}
	return serverBaselines, nil
}

func BatchCreateServerBaseline(serverBaselines []entity.ServerBaseline) error {
	if err := data.DB.Table(entity.ServerBaselineTable).Create(&serverBaselines).Scan(&serverBaselines).Error; err != nil {
		log.Errorf("batch insert serverBaseline error: %v", err)
		return err
	}
	return nil
}

func BatchCreateServerNodeRoleRel(serverNodeRoleRels []entity.ServerNodeRoleRel) error {
	if err := data.DB.Table(entity.ServerNodeRoleRelTable).Create(&serverNodeRoleRels).Scan(&serverNodeRoleRels).Error; err != nil {
		log.Errorf("batch insert serverNodeRoleRel error: %v", err)
		return err
	}
	return nil
}

func QueryNetworkDeviceRoleBaselineByVersionId(versionId int64) ([]entity.NetworkDeviceRoleBaseline, error) {
	var networkDeviceRoleBaselines []entity.NetworkDeviceRoleBaseline
	if err := data.DB.Table(entity.NetworkDeviceRoleBaselineTable).Where("version_id = ? ", versionId).Find(&networkDeviceRoleBaselines).Error; err != nil {
		return networkDeviceRoleBaselines, err
	}
	return networkDeviceRoleBaselines, nil
}

func BatchCreateNetworkDeviceRoleBaseline(networkDeviceRoleBaselines []entity.NetworkDeviceRoleBaseline) error {
	if err := data.DB.Table(entity.NetworkDeviceRoleBaselineTable).Create(&networkDeviceRoleBaselines).Scan(&networkDeviceRoleBaselines).Error; err != nil {
		log.Errorf("batch insert networkDeviceRoleBaseline error: %v", err)
		return err
	}
	return nil
}

func BatchCreateNetworkModelRoleRel(networkModelRoleRels []entity.NetworkModelRoleRel) error {
	if len(networkModelRoleRels) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.NetworkModelRoleRelTable).Create(&networkModelRoleRels).Scan(&networkModelRoleRels).Error; err != nil {
		log.Errorf("batch insert networkModelRoleRel error: %v", err)
		return err
	}
	return nil
}

func QueryNetworkDeviceBaselineByVersionId(versionId int64) ([]entity.NetworkDeviceBaseline, error) {
	var networkDeviceBaselines []entity.NetworkDeviceBaseline
	if err := data.DB.Table(entity.NetworkDeviceBaselineTable).Where("version_id = ? ", versionId).Find(&networkDeviceBaselines).Error; err != nil {
		return networkDeviceBaselines, err
	}
	return networkDeviceBaselines, nil
}

func BatchCreateNetworkDeviceBaseline(networkDeviceBaselines []entity.NetworkDeviceBaseline) error {
	if err := data.DB.Table(entity.NetworkDeviceBaselineTable).Create(&networkDeviceBaselines).Scan(&networkDeviceBaselines).Error; err != nil {
		log.Errorf("batch insert networkDeviceBaseline error: %v", err)
		return err
	}
	return nil
}

func BatchCreateNetworkDeviceRoleRel(networkDeviceRoleRels []entity.NetworkDeviceRoleRel) error {
	if err := data.DB.Table(entity.NetworkDeviceRoleRelTable).Create(&networkDeviceRoleRels).Scan(&networkDeviceRoleRels).Error; err != nil {
		log.Errorf("batch insert networkDeviceRoleRel error: %v", err)
		return err
	}
	return nil
}

func QueryIPDemandBaselineByVersionId(versionId int64) ([]entity.IPDemandBaseline, error) {
	var ipDemandBaselines []entity.IPDemandBaseline
	if err := data.DB.Table(entity.IPDemandBaselineTable).Where("version_id = ? ", versionId).Find(&ipDemandBaselines).Error; err != nil {
		return ipDemandBaselines, err
	}
	return ipDemandBaselines, nil
}

func BatchCreateIPDemandBaseline(ipDemandBaselines []entity.IPDemandBaseline) error {
	if err := data.DB.Table(entity.IPDemandBaselineTable).Create(&ipDemandBaselines).Scan(&ipDemandBaselines).Error; err != nil {
		log.Errorf("batch insert ipDemandBaseline error: %v", err)
		return err
	}
	return nil
}

func BatchCreateIPDemandDeviceRoleRel(ipDemandDeviceRoleRels []entity.IPDemandDeviceRoleRel) error {
	if err := data.DB.Table(entity.IPDemandDeviceRoleRelTable).Create(&ipDemandDeviceRoleRels).Scan(&ipDemandDeviceRoleRels).Error; err != nil {
		log.Errorf("batch insert ipDemandDeviceRoleRel error: %v", err)
		return err
	}
	return nil
}

func QueryCapConvertBaselineByVersionId(versionId int64) ([]entity.CapConvertBaseline, error) {
	var capConvertBaselines []entity.CapConvertBaseline
	if err := data.DB.Table(entity.CapConvertBaselineTable).Where("version_id = ? ", versionId).Find(&capConvertBaselines).Error; err != nil {
		return capConvertBaselines, err
	}
	return capConvertBaselines, nil
}

func BatchCreateCapConvertBaseline(capConvertBaselines []entity.CapConvertBaseline) error {
	if err := data.DB.Table(entity.CapConvertBaselineTable).Create(&capConvertBaselines).Scan(&capConvertBaselines).Error; err != nil {
		log.Errorf("batch insert capConvertBaseline error: %v", err)
		return err
	}
	return nil
}

func QueryCapActualResBaselineByVersionId(versionId int64) ([]entity.CapActualResBaseline, error) {
	var capActualResBaselines []entity.CapActualResBaseline
	if err := data.DB.Table(entity.CapActualResBaselineTable).Where("version_id = ? ", versionId).Find(&capActualResBaselines).Error; err != nil {
		return capActualResBaselines, err
	}
	return capActualResBaselines, nil
}

func BatchCreateCapActualResBaseline(capActualResBaselines []entity.CapActualResBaseline) error {
	if err := data.DB.Table(entity.CapActualResBaselineTable).Create(&capActualResBaselines).Scan(&capActualResBaselines).Error; err != nil {
		log.Errorf("batch insert capActualResBaseline error: %v", err)
		return err
	}
	return nil
}

func QueryCapServerCalcBaselineByVersionId(versionId int64) ([]entity.CapServerCalcBaseline, error) {
	var capServerCalcBaselines []entity.CapServerCalcBaseline
	if err := data.DB.Table(entity.CapServerCalcBaselineTable).Where("version_id = ? ", versionId).Find(&capServerCalcBaselines).Error; err != nil {
		return capServerCalcBaselines, err
	}
	return capServerCalcBaselines, nil
}

func BatchCreateCapServerCalcBaseline(capServerCalcBaselines []entity.CapServerCalcBaseline) error {
	if err := data.DB.Table(entity.CapServerCalcBaselineTable).Create(&capServerCalcBaselines).Scan(&capServerCalcBaselines).Error; err != nil {
		log.Errorf("batch insert capServerCalcBaseline error: %v", err)
		return err
	}
	return nil
}

func DeleteNodeRoleBaseline(nodeRoleBaselines []entity.NodeRoleBaseline) error {
	if len(nodeRoleBaselines) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.NodeRoleBaselineTable).Delete(&nodeRoleBaselines).Error; err != nil {
		log.Errorf("delete nodeRoleBaseline error: %v", err)
		return err
	}
	return nil
}

func DeleteNodeRoleMixedDeploy() error {
	if err := data.DB.Table(entity.NodeRoleMixedDeployTable).Delete(&entity.NodeRoleMixedDeploy{}).Error; err != nil {
		log.Errorf("delete nodeRoleMixedDeploy error: %v", err)
		return err
	}
	return nil
}

func QueryNodeRoleMixedDeployByNodeRoleIds(nodeRoleIds []int64) ([]entity.NodeRoleMixedDeploy, error) {
	var nodeRoleMixedDeploys []entity.NodeRoleMixedDeploy
	if err := data.DB.Table(entity.NodeRoleMixedDeployTable).Where("node_role_id in (?) ", nodeRoleIds).Find(&nodeRoleMixedDeploys).Error; err != nil {
		return nodeRoleMixedDeploys, err
	}
	return nodeRoleMixedDeploys, nil
}

func UpdateNodeRoleBaseline(nodeRoleBaselines []entity.NodeRoleBaseline) error {
	if len(nodeRoleBaselines) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.NodeRoleBaselineTable).Updates(&nodeRoleBaselines).Error; err != nil {
		log.Errorf("update nodeRoleBaseline error: %v", err)
		return err
	}
	return nil
}

func UpdateCloudProductBaseline(cloudProductBaselines []entity.CloudProductBaseline) error {
	if len(cloudProductBaselines) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.CloudProductBaselineTable).Updates(&cloudProductBaselines).Error; err != nil {
		log.Errorf("update cloudProductBaseline error: %v", err)
		return err
	}
	return nil
}

func DeleteCloudProductDependRel() error {
	if err := data.DB.Table(entity.CloudProductDependRelTable).Delete(&entity.CloudProductDependRel{}).Error; err != nil {
		log.Errorf("delete cloudProductDependRel error: %v", err)
		return err
	}
	return nil
}

func DeleteCloudProductNodeRoleRel() error {
	if err := data.DB.Table(entity.CloudProductNodeRoleTable).Delete(&entity.CloudProductNodeRoleRel{}).Error; err != nil {
		log.Errorf("delete cloudProductNodeRoleRel error: %v", err)
		return err
	}
	return nil
}

func DeleteCloudProductBaseline(cloudProductBaselines []entity.CloudProductBaseline) error {
	if len(cloudProductBaselines) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.CloudProductBaselineTable).Delete(&cloudProductBaselines).Error; err != nil {
		log.Errorf("delete cloudProductBaseline error: %v", err)
		return err
	}
	return nil
}

func UpdateServerBaseline(serverBaselines []entity.ServerBaseline) error {
	if len(serverBaselines) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.ServerBaselineTable).Updates(&serverBaselines).Error; err != nil {
		log.Errorf("update serverBaseline error: %v", err)
		return err
	}
	return nil
}

func DeleteServerBaseline(serverBaselines []entity.ServerBaseline) error {
	if len(serverBaselines) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.ServerBaselineTable).Delete(&serverBaselines).Error; err != nil {
		log.Errorf("delete serverBaseline error: %v", err)
		return err
	}
	return nil
}

func DeleteServerNodeRoleRel() error {
	if err := data.DB.Table(entity.ServerNodeRoleRelTable).Delete(&entity.ServerNodeRoleRel{}).Error; err != nil {
		log.Errorf("delete serverNodeRoleRel error: %v", err)
		return err
	}
	return nil
}

func UpdateNetworkDeviceRoleBaseline(networkDeviceRoleBaselines []entity.NetworkDeviceRoleBaseline) error {
	if len(networkDeviceRoleBaselines) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.NetworkDeviceRoleBaselineTable).Updates(&networkDeviceRoleBaselines).Error; err != nil {
		log.Errorf("update networkDeviceRoleBaseline error: %v", err)
		return err
	}
	return nil
}

func DeleteNetworkModelRoleRel() error {
	if err := data.DB.Table(entity.NetworkModelRoleRelTable).Delete(&entity.NetworkModelRoleRel{}).Error; err != nil {
		log.Errorf("delete networkModelRoleRel error: %v", err)
		return err
	}
	return nil
}

func DeleteNetworkDeviceRoleBaseline(networkDeviceRoleBaselines []entity.NetworkDeviceRoleBaseline) error {
	if len(networkDeviceRoleBaselines) == 0 {
		return nil
	}
	if err := data.DB.Table(entity.NetworkDeviceRoleBaselineTable).Delete(&networkDeviceRoleBaselines).Error; err != nil {
		log.Errorf("delete networkDeviceRoleBaseline error: %v", err)
		return err
	}
	return nil
}
