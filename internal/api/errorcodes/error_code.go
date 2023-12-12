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

	InvalidUserError              = "PlanningM.User.InvalidUsernameOrPassword"
	CloudProductDependenciesError = "PlanningM.CloudProduct.RequiredError"
	CloudProductRequireError      = "PlanningM.CloudProduct.DependError"
)
