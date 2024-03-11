package ollama

import "one-api/dto"

type OllamaRequest struct {
	Model    string         `json:"model,omitempty"`
	Messages []dto.Message  `json:"messages,omitempty"`
	Stream   bool           `json:"stream,omitempty"`
	Options  *OllamaOptions `json:"options,omitempty"`
}

type OllamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	Seed        float64 `json:"seed,omitempty"`
	Topp        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
	Stop        any     `json:"stop,omitempty"`
}
