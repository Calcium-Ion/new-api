package cloudflare

import "one-api/dto"

type CfRequest struct {
	Messages    []dto.Message `json:"messages,omitempty"`
	Lora        string        `json:"lora,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Prompt      string        `json:"prompt,omitempty"`
	Raw         bool          `json:"raw,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
}

type CfAudioResponse struct {
	Result CfSTTResult `json:"result"`
}

type CfSTTResult struct {
	Text string `json:"text"`
}
