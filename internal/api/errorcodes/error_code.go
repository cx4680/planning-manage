package errorcodes

const (
	UnknownError                     = "PlanningM.UnknownError"
	InvalidData                      = "PlanningM.InvalidData"
	InvalidParam                     = "PlanningM.InvalidParam"
	NotFound                         = "PlanningM.NotFound"
	ReadOnly                         = "PlanningM.ReadOnly"
	SystemError                      = "PlanningM.SystemError"
	NodeRoleMustImportFirst          = "PlanningM.NodeRoleMustImportFirst"
	NetworkDeviceRoleMustImportFirst = "PlanningM.NetworkDeviceRoleMustImportFirst"

	InvalidUserError                 = "PlanningM.User.InvalidUsernameOrPassword"
	CLOUD_PRODUCT_DEPENDENCIES_ERROR = "PlanningM.CloudProduct.DependError"
)
