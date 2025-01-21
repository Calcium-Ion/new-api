package mokaai

import "one-api/dto"


type Request struct {
	Messages    []dto.Message `json:"messages,omitempty"`
	Lora        string        `json:"lora,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Prompt      string        `json:"prompt,omitempty"`
	Raw         bool          `json:"raw,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type Options struct {
	Seed             int      `json:"seed,omitempty"`
	Temperature      *float64 `json:"temperature,omitempty"`
	TopK             int      `json:"top_k,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	NumPredict       int      `json:"num_predict,omitempty"`
	NumCtx           int      `json:"num_ctx,omitempty"`
}

type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}