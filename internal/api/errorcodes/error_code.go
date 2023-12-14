package errorcodes

const (
	UnknownError                     = "PlanningM.UnknownError"
	InvalidData                      = "PlanningM.InvalidData"
	InvalidParam                     = "PlanningM.InvalidParam"
	InvalidUnauthorized              = "PlanningM.InvalidUnauthorized"
	NotFound                         = "PlanningM.NotFound"
	ReadOnly                         = "PlanningM.ReadOnly"
	SystemError                      = "PlanningM.SystemError"
	NodeRoleMustImportFirst          = "PlanningM.NodeRoleMustImportFirst"
	NetworkDeviceRoleMustImportFirst = "PlanningM.NetworkDeviceRoleMustImportFirst"

	InvalidUserError          = "PlanningM.User.InvalidUsernameOrPassword"
	CustomerNameExistsError   = "PlanningM.Customer.CustomerNameExistsError"
	CloudProductDependError   = "PlanningM.CloudProduct.DependError"
	CloudProductRequiredError = "PlanningM.CloudProduct.RequiredError"
)
