package dto

type SimpleResponse struct {
	Usage `json:"usage"`
	Error *OpenAIError `json:"error"`
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
	Model   string                     `json:"model"`
	Object  string                     `json:"object"`
	Created int64                      `json:"created"`
	Choices []OpenAITextResponseChoice `json:"choices"`
	Error   *OpenAIError               `json:"error,omitempty"`
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
	Content          *string            `json:"content,omitempty"`
	ReasoningContent *string            `json:"reasoning_content,omitempty"`
	Reasoning        *string            `json:"reasoning,omitempty"`
	Role             string             `json:"role,omitempty"`
	ToolCalls        []ToolCallResponse `json:"tool_calls,omitempty"`
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

func (c *ChatCompletionsStreamResponseChoiceDelta) GetReasoningContent() string {
	if c.ReasoningContent == nil && c.Reasoning == nil {
		return ""
	}
	if c.ReasoningContent != nil {
		return *c.ReasoningContent
	}
	return *c.Reasoning
}

func (c *ChatCompletionsStreamResponseChoiceDelta) SetReasoningContent(s string) {
	c.ReasoningContent = &s
	c.Reasoning = &s
}

type ToolCallResponse struct {
	// Index is not nil only in chat completion chunk object
	Index    *int             `json:"index,omitempty"`
	ID       string           `json:"id,omitempty"`
	Type     any              `json:"type"`
	Function FunctionResponse `json:"function"`
}

func (c *ToolCallResponse) SetIndex(i int) {
	c.Index = &i
}

type FunctionResponse struct {
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
	// call function with arguments in JSON format
	Parameters any    `json:"parameters,omitempty"` // request
	Arguments  string `json:"arguments"`            // response
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

func (c *ChatCompletionsStreamResponse) IsToolCall() bool {
	if len(c.Choices) == 0 {
		return false
	}
	return len(c.Choices[0].Delta.ToolCalls) > 0
}

func (c *ChatCompletionsStreamResponse) GetFirstToolCall() *ToolCallResponse {
	if c.IsToolCall() {
		return &c.Choices[0].Delta.ToolCalls[0]
	}
	return nil
}

func (c *ChatCompletionsStreamResponse) Copy() *ChatCompletionsStreamResponse {
	choices := make([]ChatCompletionsStreamResponseChoice, len(c.Choices))
	copy(choices, c.Choices)
	return &ChatCompletionsStreamResponse{
		Id:                c.Id,
		Object:            c.Object,
		Created:           c.Created,
		Model:             c.Model,
		SystemFingerprint: c.SystemFingerprint,
		Choices:           choices,
		Usage:             c.Usage,
	}
}

func (c *ChatCompletionsStreamResponse) GetSystemFingerprint() string {
	if c.SystemFingerprint == nil {
		return ""
	}
	return *c.SystemFingerprint
}

func (c *ChatCompletionsStreamResponse) SetSystemFingerprint(s string) {
	c.SystemFingerprint = &s
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
	PromptTokens           int                `json:"prompt_tokens"`
	CompletionTokens       int                `json:"completion_tokens"`
	TotalTokens            int                `json:"total_tokens"`
	PromptCacheHitTokens   int                `json:"prompt_cache_hit_tokens,omitempty"`
	PromptTokensDetails    InputTokenDetails  `json:"prompt_tokens_details"`
	CompletionTokenDetails OutputTokenDetails `json:"completion_tokens_details"`
}
