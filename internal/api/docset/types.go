package docset

type DataResponse struct {
	RequestId string      `json:"requestId" example:"12345678" swaggertype:"string"` // 请求id，由api网关生成，透传给后端服务。
	Code      string      `json:"code" example:"Success" swaggertype:"string"`       // 由各服务定义的错误码，需带上服务前缀{PRODUCT}，如“ECS.EcsLocked”。
	Data      interface{} `json:"data"`
}

type NoDataResponse struct {
	RequestId string `json:"requestId" example:"12345678" swaggertype:"string"`
	Code      string `json:"code" example:"Success" swaggertype:"string"`
}

type UnknownErrorResponse struct {
	RequestId string `json:"requestId" example:"12345678" swaggertype:"string"`
	Code      string `json:"code" example:"JobM.UnknownError" swaggertype:"string"`
}

type InvalidDataResponse struct {
	RequestId string `json:"requestId" example:"12345678" swaggertype:"string"`
	Code      string `json:"code" example:"JobM.InvalidData" swaggertype:"string"`
}

type AddAtomOpsErrorResponse struct {
	RequestId string `json:"requestId" example:"12345678" swaggertype:"string"`
	Code      string `json:"code" example:"JobM.AtomOps.AddFailure" swaggertype:"string"`
}

type AtomOpsCheckNameErrorResponse struct {
	RequestId string `json:"requestId" example:"12345678" swaggertype:"string"`
	Code      string `json:"code" example:"JobM.AtomOps.CheckNameFailure" swaggertype:"string"`
}

type AtomOpsCheckIdErrorResponse struct {
	RequestId string `json:"requestId" example:"12345678" swaggertype:"string"`
	Code      string `json:"code" example:"JobM.AtomOps.CheckIdFailure" swaggertype:"string"`
}

type AtomOpsCheckParamNameErrorResponse struct {
	RequestId string `json:"requestId" example:"12345678" swaggertype:"string"`
	Code      string `json:"code" example:"JobM.AtomOps.CheckParamNameFailure" swaggertype:"string"`
}

type UpdateAtomOpsErrorResponse struct {
	RequestId string `json:"requestId" example:"12345678" swaggertype:"string"`
	Code      string `json:"code" example:"JobM.AtomOps.UpdateFailure" swaggertype:"string"`
}

type CopyAtomOpsErrorResponse struct {
	RequestId string `json:"requestId" example:"12345678" swaggertype:"string"`
	Code      string `json:"code" example:"JobM.AtomOps.CopyFailure" swaggertype:"string"`
}

type DeleteAtomOpsErrorResponse struct {
	RequestId string `json:"requestId" example:"12345678" swaggertype:"string"`
	Code      string `json:"code" example:"JobM.AtomOps.DeleteFailure" swaggertype:"string"`
}

type ExecAtomOpsErrorResponse struct {
	RequestId string `json:"requestId" example:"12345678" swaggertype:"string"`
	Code      string `json:"code" example:"JobM.AtomOps.ExecFailure" swaggertype:"string"`
}

type GetAtomOpsDetailErrorResponse struct {
	RequestId string `json:"requestId" example:"12345678" swaggertype:"string"`
	Code      string `json:"code" example:"JobM.AtomOps.GetDetailFailure" swaggertype:"string"`
}
