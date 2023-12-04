package baseline

import (
	"code.cestc.cn/zhangzhi/planning-manage/internal/data"
	"code.cestc.cn/zhangzhi/planning-manage/internal/entity"
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
