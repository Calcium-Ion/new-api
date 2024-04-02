package dto

type TextResponseWithError struct {
	Id      string                        `json:"id"`
	Object  string                        `json:"object"`
	Created int64                         `json:"created"`
	Choices []OpenAITextResponseChoice    `json:"choices"`
	Data    []OpenAIEmbeddingResponseItem `json:"data"`
	Model   string                        `json:"model"`
	Usage   `json:"usage"`
	Error   OpenAIError `json:"error"`
}

type SimpleResponse struct {
	Usage   `json:"usage"`
	Error   OpenAIError                `json:"error"`
	Choices []OpenAITextResponseChoice `json:"choices"`
}

type TextResponse struct {
	Id      string                     `json:"id"`
	Object  string                     `json:"object"`
	Created int64                      `json:"created"`
	Model   string                     `json:"model"`
	Choices []OpenAITextResponseChoice `json:"choices"`
	Usage   `json:"usage"`
}

type OpenAITextResponseChoice struct {
	Index        int `json:"index"`
	Message      `json:"message"`
	FinishReason string `json:"finish_reason"`
}

type OpenAITextResponse struct {
	Id      string                     `json:"id"`
	Object  string                     `json:"object"`
	Created int64                      `json:"created"`
	Choices []OpenAITextResponseChoice `json:"choices"`
	Usage   `json:"usage"`
}

type OpenAIEmbeddingResponseItem struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

type OpenAIEmbeddingResponse struct {
	Object string                        `json:"object"`
	Data   []OpenAIEmbeddingResponseItem `json:"data"`
	Model  string                        `json:"model"`
	Usage  `json:"usage"`
}

type ChatCompletionsStreamResponseChoice struct {
	Delta struct {
		Content   string `json:"content"`
		Role      string `json:"role,omitempty"`
		ToolCalls any    `json:"tool_calls,omitempty"`
	} `json:"delta"`
	FinishReason *string `json:"finish_reason,omitempty"`
	Index        int     `json:"index,omitempty"`
}

type ChatCompletionsStreamResponse struct {
	Id      string                                `json:"id"`
	Object  string                                `json:"object"`
	Created int64                                 `json:"created"`
	Model   string                                `json:"model"`
	Choices []ChatCompletionsStreamResponseChoice `json:"choices"`
}

type ChatCompletionsStreamResponseSimple struct {
	Choices []ChatCompletionsStreamResponseChoice `json:"choices"`
}

type CompletionsStreamResponse struct {
	Choices []struct {
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
