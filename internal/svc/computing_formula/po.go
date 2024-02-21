package computing_formula

import "code.cestc.cn/ccos/common/planning-manage/internal/entity"

type SoftwareData struct {
	CloudProductPlanningList          []*entity.CloudProductPlanning
	CloudProductNodeRoleRelList       []*entity.CloudProductNodeRoleRel
	CloudProductBaselineMap           map[int64]*entity.CloudProductBaseline
	ServerPlanningMap                 map[int64]*entity.ServerPlanning
	ServerBaselineMap                 map[int64]*entity.ServerBaseline
	ServerCapPlanningMap              map[string]*entity.ServerCapPlanning
	SoftwareBomLicenseBaselineListMap map[string][]*entity.SoftwareBomLicenseBaseline
}
