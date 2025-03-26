package dto

type OpenAIResponsesOutput struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Status  string `json:"status"`
	Role    string `json:"role"`
	Content []struct {
		Type        string `json:"type"`
		Text        string `json:"text"`
		Annotations []any  `json:"annotations"`
	} `json:"content"`
}

type OpenAIResponsesResponse struct {
	ID                 string                  `json:"id"`
	Object             string                  `json:"object"`
	CreatedAt          int64                   `json:"created_at"`
	Status             string                  `json:"status"`
	Error              *OpenAIError            `json:"error,omitempty"`
	IncompleteDetails  *any                    `json:"incomplete_details,omitempty"`
	Instructions       *any                    `json:"instructions,omitempty"`
	MaxOutputTokens    *int                    `json:"max_output_tokens,omitempty"`
	Model              string                  `json:"model"`
	Output             []OpenAIResponsesOutput `json:"output"`
	ParallelToolCalls  bool                    `json:"parallel_tool_calls"`
	PreviousResponseID *string                 `json:"previous_response_id,omitempty"`
	Reasoning          struct {
		Effort          *string `json:"effort,omitempty"`
		GenerateSummary *bool   `json:"generate_summary,omitempty"`
	} `json:"reasoning"`
	Store       bool    `json:"store"`
	Temperature float64 `json:"temperature"`
	Text        struct {
		Format struct {
			Type string `json:"type"`
		} `json:"format"`
	} `json:"text"`
	ToolChoice string  `json:"tool_choice"`
	Tools      []any   `json:"tools"`
	TopP       float64 `json:"top_p"`
	Truncation string  `json:"truncation"`
	Usage      struct {
		InputTokens        int `json:"input_tokens"`
		InputTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"input_tokens_details"`
		OutputTokens        int `json:"output_tokens"`
		OutputTokensDetails struct {
			ReasoningTokens int `json:"reasoning_tokens"`
		} `json:"output_tokens_details"`
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	User     *string        `json:"user,omitempty"`
	Metadata map[string]any `json:"metadata"`
}
