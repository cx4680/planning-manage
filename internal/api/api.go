package api

type ExecInParam struct {
	Name        string `json:"name" validate:"required,checkParamName"`
	ParamType   string `json:"paramType" validate:"required,oneof=Text Date File"`
	ParamValue  string `json:"paramValue"`
	Required    string `json:"required"`
	Description string `json:"description" validate:"lte=255"`
}

type ExecParam struct {
	InParam []ExecInParam `json:"inParam" validate:"unique=Name,dive"`
	Targets []ExecTarget  `json:"targets" validate:"required,dive"`
}

type InPutParam struct {
	Name        string `json:"name" validate:"required,checkParamName"`
	ParamType   string `json:"paramType" validate:"required,oneof=Text Date File"`
	ParamValue  string `json:"paramValue" validate:"lte=255"`
	Required    string `json:"required" validate:"required"`
	Description string `json:"description" validate:"lte=255"`
}

type OutPutParam struct {
	Name        string `json:"name" validate:"required,checkParamName"`
	ParamType   string `json:"paramType" validate:"required,oneof=Text Date File"`
	ParamValue  string `json:"paramValue"`
	Description string `json:"description" validate:"lte=255"`
}

type ExecTarget struct {
	HostName string `form:"hostName" json:"hostName" validate:"required"` // 主机名称
	IP       string `form:"ip" json:"ip" validate:"required"`             // ip地址
	Label    string `form:"label" json:"label"`
	Region   string `form:"region" json:"region"`                 // 区域
	Az       string `form:"az" json:"az" validate:"required"`     // 可用区
	Cell     string `form:"cell" json:"cell" validate:"required"` // 集群
	OS       string `form:"os" json:"os"`                         // 操作系统
}
