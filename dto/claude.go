package dto

import "encoding/json"

type ClaudeMetadata struct {
	UserId string `json:"user_id"`
}

type ClaudeMediaMessage struct {
	Type        string               `json:"type,omitempty"`
	Text        *string              `json:"text,omitempty"`
	Model       string               `json:"model,omitempty"`
	Source      *ClaudeMessageSource `json:"source,omitempty"`
	Usage       *ClaudeUsage         `json:"usage,omitempty"`
	StopReason  *string              `json:"stop_reason,omitempty"`
	PartialJson *string              `json:"partial_json,omitempty"`
	Role        string               `json:"role,omitempty"`
	Thinking    string               `json:"thinking,omitempty"`
	Signature   string               `json:"signature,omitempty"`
	Delta       string               `json:"delta,omitempty"`
	// tool_calls
	Id        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     any             `json:"input,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	ToolUseId string          `json:"tool_use_id,omitempty"`
}

func (c *ClaudeMediaMessage) SetText(s string) {
	c.Text = &s
}

func (c *ClaudeMediaMessage) GetText() string {
	if c.Text == nil {
		return ""
	}
	return *c.Text
}

func (c *ClaudeMediaMessage) IsStringContent() bool {
	var content string
	return json.Unmarshal(c.Content, &content) == nil
}

func (c *ClaudeMediaMessage) GetStringContent() string {
	var content string
	if err := json.Unmarshal(c.Content, &content); err == nil {
		return content
	}
	return ""
}

func (c *ClaudeMediaMessage) GetJsonRowString() string {
	jsonContent, _ := json.Marshal(c)
	return string(jsonContent)
}

func (c *ClaudeMediaMessage) SetContent(content any) {
	jsonContent, _ := json.Marshal(content)
	c.Content = jsonContent
}

func (c *ClaudeMediaMessage) ParseMediaContent() []ClaudeMediaMessage {
	var mediaContent []ClaudeMediaMessage
	if err := json.Unmarshal(c.Content, &mediaContent); err == nil {
		return mediaContent
	}
	return make([]ClaudeMediaMessage, 0)
}

type ClaudeMessageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      any    `json:"data"`
}

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

func (c *ClaudeMessage) IsStringContent() bool {
	_, ok := c.Content.(string)
	return ok
}

func (c *ClaudeMessage) GetStringContent() string {
	if c.IsStringContent() {
		return c.Content.(string)
	}
	return ""
}

func (c *ClaudeMessage) SetStringContent(content string) {
	c.Content = content
}

func (c *ClaudeMessage) ParseContent() ([]ClaudeMediaMessage, error) {
	// map content to []ClaudeMediaMessage
	// parse to json
	jsonContent, _ := json.Marshal(c.Content)
	var contentList []ClaudeMediaMessage
	err := json.Unmarshal(jsonContent, &contentList)
	if err != nil {
		return make([]ClaudeMediaMessage, 0), err
	}
	return contentList, nil
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type InputSchema struct {
	Type       string `json:"type"`
	Properties any    `json:"properties,omitempty"`
	Required   any    `json:"required,omitempty"`
}

type ClaudeRequest struct {
	Model             string          `json:"model"`
	Prompt            string          `json:"prompt,omitempty"`
	System            any             `json:"system,omitempty"`
	Messages          []ClaudeMessage `json:"messages,omitempty"`
	MaxTokens         uint            `json:"max_tokens,omitempty"`
	MaxTokensToSample uint            `json:"max_tokens_to_sample,omitempty"`
	StopSequences     []string        `json:"stop_sequences,omitempty"`
	Temperature       *float64        `json:"temperature,omitempty"`
	TopP              float64         `json:"top_p,omitempty"`
	TopK              int             `json:"top_k,omitempty"`
	//ClaudeMetadata    `json:"metadata,omitempty"`
	Stream     bool      `json:"stream,omitempty"`
	Tools      any       `json:"tools,omitempty"`
	ToolChoice any       `json:"tool_choice,omitempty"`
	Thinking   *Thinking `json:"thinking,omitempty"`
}

type Thinking struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens"`
}

func (c *ClaudeRequest) IsStringSystem() bool {
	_, ok := c.System.(string)
	return ok
}

func (c *ClaudeRequest) GetStringSystem() string {
	if c.IsStringSystem() {
		return c.System.(string)
	}
	return ""
}

func (c *ClaudeRequest) SetStringSystem(system string) {
	c.System = system
}

func (c *ClaudeRequest) ParseSystem() []ClaudeMediaMessage {
	// map content to []ClaudeMediaMessage
	// parse to json
	jsonContent, _ := json.Marshal(c.System)
	var contentList []ClaudeMediaMessage
	if err := json.Unmarshal(jsonContent, &contentList); err == nil {
		return contentList
	}
	return make([]ClaudeMediaMessage, 0)
}

type ClaudeError struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message,omitempty"`
}

type ClaudeErrorWithStatusCode struct {
	Error      ClaudeError `json:"error"`
	StatusCode int         `json:"status_code"`
	LocalError bool
}

type ClaudeResponse struct {
	Id           string               `json:"id,omitempty"`
	Type         string               `json:"type"`
	Role         string               `json:"role,omitempty"`
	Content      []ClaudeMediaMessage `json:"content,omitempty"`
	Completion   string               `json:"completion,omitempty"`
	StopReason   string               `json:"stop_reason,omitempty"`
	Model        string               `json:"model,omitempty"`
	Error        *ClaudeError         `json:"error,omitempty"`
	Usage        *ClaudeUsage         `json:"usage,omitempty"`
	Index        *int                 `json:"index,omitempty"`
	ContentBlock *ClaudeMediaMessage  `json:"content_block,omitempty"`
	Delta        *ClaudeMediaMessage  `json:"delta,omitempty"`
	Message      *ClaudeMediaMessage  `json:"message,omitempty"`
}

// set index
func (c *ClaudeResponse) SetIndex(i int) {
	c.Index = &i
}

// get index
func (c *ClaudeResponse) GetIndex() int {
	if c.Index == nil {
		return 0
	}
	return *c.Index
}

type ClaudeUsage struct {
	InputTokens              int `json:"input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	OutputTokens             int `json:"output_tokens"`
}
