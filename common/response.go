package common

// Response 统一API响应结构
type Response struct {
	Success bool        `json:"success" example:"true"` // 是否成功
	Message string      `json:"message" example:"操作成功"` // 响应消息
	Data    interface{} `json:"data,omitempty"`         // 响应数据
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"` // 是否成功
	Message string `json:"message" example:"操作失败"`  // 错误消息
}

// ListResponse 列表响应结构
type ListResponse struct {
	Success bool        `json:"success" example:"true"` // 是否成功
	Message string      `json:"message" example:""`     // 响应消息
	Data    interface{} `json:"data"`                   // 响应数据
	Count   int64       `json:"count" example:"100"`    // 总数
	Page    int         `json:"page" example:"1"`       // 当前页码
	Size    int         `json:"size" example:"10"`      // 每页数量
}
