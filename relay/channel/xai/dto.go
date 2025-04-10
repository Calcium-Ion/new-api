package xai

import "one-api/dto"

// ChatCompletionResponse represents the response from XAI chat completion API
type ChatCompletionResponse struct {
	Id                string `json:"id"`
	Object            string `json:"object"`
	Created           int64  `json:"created"`
	Model             string `json:"model"`
	Choices           []dto.ChatCompletionsStreamResponseChoice
	Usage             *dto.Usage `json:"usage"`
	SystemFingerprint string     `json:"system_fingerprint"`
}
