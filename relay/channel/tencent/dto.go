package tencent

import "one-api/dto"

type TencentMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TencentChatRequest struct {
	AppId    int64  `json:"app_id"`    // 腾讯云账号的 APPID
	SecretId string `json:"secret_id"` // 官网 SecretId
	// Timestamp当前 UNIX 时间戳，单位为秒，可记录发起 API 请求的时间。
	// 例如1529223702，如果与当前时间相差过大，会引起签名过期错误
	Timestamp int64 `json:"timestamp"`
	// Expired 签名的有效期，是一个符合 UNIX Epoch 时间戳规范的数值，
	// 单位为秒；Expired 必须大于 Timestamp 且 Expired-Timestamp 小于90天
	Expired int64  `json:"expired"`
	QueryID string `json:"query_id"` //请求 Id，用于问题排查
	// Temperature 较高的数值会使输出更加随机，而较低的数值会使其更加集中和确定
	// 默认 1.0，取值区间为[0.0,2.0]，非必要不建议使用,不合理的取值会影响效果
	// 建议该参数和 top_p 只设置1个，不要同时更改 top_p
	Temperature float64 `json:"temperature"`
	// TopP 影响输出文本的多样性，取值越大，生成文本的多样性越强
	// 默认1.0，取值区间为[0.0, 1.0]，非必要不建议使用, 不合理的取值会影响效果
	// 建议该参数和 temperature 只设置1个，不要同时更改
	TopP float64 `json:"top_p"`
	// Stream 0：同步，1：流式 （默认，协议：SSE)
	// 同步请求超时：60s，如果内容较长建议使用流式
	Stream int `json:"stream"`
	// Messages 会话内容, 长度最多为40, 按对话时间从旧到新在数组中排列
	// 输入 content 总数最大支持 3000 token。
	Messages []TencentMessage `json:"messages"`
	Model    string           `json:"model"` // 模型名称
}

type TencentError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type TencentUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type TencentResponseChoices struct {
	FinishReason string         `json:"finish_reason,omitempty"` // 流式结束标志位，为 stop 则表示尾包
	Messages     TencentMessage `json:"messages,omitempty"`      // 内容，同步模式返回内容，流模式为 null 输出 content 内容总数最多支持 1024token。
	Delta        TencentMessage `json:"delta,omitempty"`         // 内容，流模式返回内容，同步模式为 null 输出 content 内容总数最多支持 1024token。
}

type TencentChatResponse struct {
	Choices []TencentResponseChoices `json:"choices,omitempty"` // 结果
	Created string                   `json:"created,omitempty"` // unix 时间戳的字符串
	Id      string                   `json:"id,omitempty"`      // 会话 id
	Usage   dto.Usage                `json:"usage,omitempty"`   // token 数量
	Error   TencentError             `json:"error,omitempty"`   // 错误信息 注意：此字段可能返回 null，表示取不到有效值
	Note    string                   `json:"note,omitempty"`    // 注释
	ReqID   string                   `json:"req_id,omitempty"`  // 唯一请求 Id，每次请求都会返回。用于反馈接口入参
}
