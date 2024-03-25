package software_bom

import "code.cestc.cn/ccos/common/planning-manage/internal/entity"

type SoftwareData struct {
	ServiceYear                              int
	CloudProductBaselineList                 []*entity.CloudProductBaseline
	ServerPlanningsMap                       map[string][]*entity.ServerPlanning
	ServerBaselineMap                        map[int64]*entity.ServerBaseline
	ServerCapPlanningMap                     map[string]*entity.ServerCapPlanning
	BomIdSoftwareBomLicenseBaselineMap       map[string]*entity.SoftwareBomLicenseBaseline
	ServiceCodeSoftwareBomLicenseBaselineMap map[string][]*entity.SoftwareBomLicenseBaseline
}
