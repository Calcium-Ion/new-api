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
	Delta        ChatCompletionsStreamResponseChoiceDelta `json:"delta,omitempty"`
	Logprobs     *any                                     `json:"logprobs"`
	FinishReason *string                                  `json:"finish_reason"`
	Index        int                                      `json:"index"`
}

type ChatCompletionsStreamResponseChoiceDelta struct {
	Content   *string    `json:"content,omitempty"`
	Role      string     `json:"role,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

func (c *ChatCompletionsStreamResponseChoiceDelta) IsEmpty() bool {
	return c.Content == nil && len(c.ToolCalls) == 0
}

func (c *ChatCompletionsStreamResponseChoiceDelta) SetContentString(s string) {
	c.Content = &s
}

func (c *ChatCompletionsStreamResponseChoiceDelta) GetContentString() string {
	if c.Content == nil {
		return ""
	}
	return *c.Content
}

type ToolCall struct {
	// Index is not nil only in chat completion chunk object
	Index    *int         `json:"index,omitempty"`
	ID       string       `json:"id"`
	Type     any          `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name string `json:"name,omitempty"`
	// call function with arguments in JSON format
	Arguments string `json:"arguments,omitempty"`
}

type ChatCompletionsStreamResponse struct {
	Id                string                                `json:"id"`
	Object            string                                `json:"object"`
	Created           int64                                 `json:"created"`
	Model             string                                `json:"model"`
	SystemFingerprint *string                               `json:"system_fingerprint"`
	Choices           []ChatCompletionsStreamResponseChoice `json:"choices"`
	Usage             *Usage                                `json:"usage"`
}

type ChatCompletionsStreamResponseSimple struct {
	Choices []ChatCompletionsStreamResponseChoice `json:"choices"`
	Usage   *Usage                                `json:"usage"`
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
