package software_bom

import "code.cestc.cn/ccos/common/planning-manage/internal/entity"

type SoftwareData struct {
	ServiceYear                       int
	CloudProductBaselineList          []*entity.CloudProductBaseline
	ServerPlanningMap                 map[string]*entity.ServerPlanning
	ServerBaselineMap                 map[int64]*entity.ServerBaseline
	ServerCapPlanningMap              map[string]*entity.ServerCapPlanning
	SoftwareBomLicenseBaselineMap     map[string]*entity.SoftwareBomLicenseBaseline
	SoftwareBomLicenseBaselineListMap map[string][]*entity.SoftwareBomLicenseBaseline
}
