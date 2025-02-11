package dto

type EmbeddingOptions struct {
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

type EmbeddingResponseItem struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

type EmbeddingResponse struct {
	Object string                        `json:"object"`
	Data   []EmbeddingResponseItem `json:"data"`
	Model  string                        `json:"model"`
	Usage  `json:"usage"`
}