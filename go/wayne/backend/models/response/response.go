package response

type ResponseBase struct {
	Code int `json:"code"`
}

// OpenAPI 通用 成功 返回接口
// swagger:response responseSuccess
type Success struct {
	// in: body
	// Required: true
	Body struct {
		ResponseBase
	}
}

// OpenAPI 通用 失败 返回接口
// swagger:response responseState
type Failure struct {
	// in: body
	// Required: true
	Body struct {
		ResponseBase
		Errors []string `json:"errors"`
	}
}
