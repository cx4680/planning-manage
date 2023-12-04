package baseline

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
)

// func InsertSoftwareVersion(softwareVersion entity.SoftwareVersion) (int64, error) {
//
// }

// func InsertCloudProductBaseline(cloudProductBaselineExcelList []CloudProductBaselineExcel, importBaselineRequest ImportBaselineRequest) (int64, error) {
// }

func QueryNodeRoleList() ([]*entity.NodeRoleBaseline, error) {
	var nodeRoleBaselineList []*entity.NodeRoleBaseline
	if err := data.DB.Model(&entity.NodeRoleBaseline{}).Find(&nodeRoleBaselineList).Error; err != nil {
		return nil, err
	}
	return nodeRoleBaselineList, nil
}

func QueryServiceBaselineById(id int64) (*entity.ServerBaseline, error) {
	var serverBaseline entity.ServerBaseline
	if err := data.DB.Table(entity.ServerBaselineTable).Where("id=?", id).Scan(&serverBaseline).Error; err != nil {
		log.Errorf("[queryServiceBaselineById] query service baseline error, %v", err)
		return nil, err
	}
	return &serverBaseline, nil
}
