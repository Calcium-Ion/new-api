package ollama

import "one-api/dto"

type OllamaRequest struct {
	Model            string                `json:"model,omitempty"`
	Messages         []dto.Message         `json:"messages,omitempty"`
	Stream           bool                  `json:"stream,omitempty"`
	Temperature      *float64              `json:"temperature,omitempty"`
	Seed             float64               `json:"seed,omitempty"`
	Topp             float64               `json:"top_p,omitempty"`
	TopK             int                   `json:"top_k,omitempty"`
	Stop             any                   `json:"stop,omitempty"`
	MaxTokens        uint                  `json:"max_tokens,omitempty"`
	Tools            []dto.ToolCallRequest `json:"tools,omitempty"`
	ResponseFormat   any                   `json:"response_format,omitempty"`
	FrequencyPenalty float64               `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64               `json:"presence_penalty,omitempty"`
	Suffix           any                   `json:"suffix,omitempty"`
	StreamOptions    *dto.StreamOptions    `json:"stream_options,omitempty"`
	Prompt           any                   `json:"prompt,omitempty"`
}

type Options struct {
	Seed             int      `json:"seed,omitempty"`
	Temperature      *float64 `json:"temperature,omitempty"`
	TopK             int      `json:"top_k,omitempty"`
	TopP             float64  `json:"top_p,omitempty"`
	FrequencyPenalty float64  `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64  `json:"presence_penalty,omitempty"`
	NumPredict       int      `json:"num_predict,omitempty"`
	NumCtx           int      `json:"num_ctx,omitempty"`
}

type OllamaEmbeddingRequest struct {
	Model   string   `json:"model,omitempty"`
	Input   []string `json:"input"`
	Options *Options `json:"options,omitempty"`
}

type OllamaEmbeddingResponse struct {
	Error     string      `json:"error,omitempty"`
	Model     string      `json:"model"`
	Embedding [][]float64 `json:"embeddings,omitempty"`
}
