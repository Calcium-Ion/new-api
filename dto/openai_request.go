package dto

import "encoding/json"

type ResponseFormat struct {
	Type string `json:"type,omitempty"`
}

type GeneralOpenAIRequest struct {
	Model               string         `json:"model,omitempty"`
	Messages            []Message      `json:"messages,omitempty"`
	Prompt              any            `json:"prompt,omitempty"`
	Stream              bool           `json:"stream,omitempty"`
	StreamOptions       *StreamOptions `json:"stream_options,omitempty"`
	MaxTokens           uint           `json:"max_tokens,omitempty"`
	MaxCompletionTokens uint           `json:"max_completion_tokens,omitempty"`
	Temperature         float64        `json:"temperature,omitempty"`
	TopP                float64        `json:"top_p,omitempty"`
	TopK                int            `json:"top_k,omitempty"`
	Stop                any            `json:"stop,omitempty"`
	N                   int            `json:"n,omitempty"`
	Input               any            `json:"input,omitempty"`
	Instruction         string         `json:"instruction,omitempty"`
	Size                string         `json:"size,omitempty"`
	Functions           any            `json:"functions,omitempty"`
	FrequencyPenalty    float64        `json:"frequency_penalty,omitempty"`
	PresencePenalty     float64        `json:"presence_penalty,omitempty"`
	ResponseFormat      any            `json:"response_format,omitempty"`
	EncodingFormat      any            `json:"encoding_format,omitempty"`
	Seed                float64        `json:"seed,omitempty"`
	Tools               []ToolCall     `json:"tools,omitempty"`
	ToolChoice          any            `json:"tool_choice,omitempty"`
	User                string         `json:"user,omitempty"`
	LogProbs            bool           `json:"logprobs,omitempty"`
	TopLogProbs         int            `json:"top_logprobs,omitempty"`
	Dimensions          int            `json:"dimensions,omitempty"`
	Modalities          any            `json:"modalities,omitempty"`
	Audio               any            `json:"audio,omitempty"`
}

type OpenAITools struct {
	Type     string         `json:"type"`
	Function OpenAIFunction `json:"function"`
}

type OpenAIFunction struct {
	Description string `json:"description,omitempty"`
	Name        string `json:"name"`
	Parameters  any    `json:"parameters,omitempty"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

func (r GeneralOpenAIRequest) GetMaxTokens() int {
	return int(r.MaxTokens)
}

func (r GeneralOpenAIRequest) ParseInput() []string {
	if r.Input == nil {
		return nil
	}
	var input []string
	switch r.Input.(type) {
	case string:
		input = []string{r.Input.(string)}
	case []any:
		input = make([]string, 0, len(r.Input.([]any)))
		for _, item := range r.Input.([]any) {
			if str, ok := item.(string); ok {
				input = append(input, str)
			}
		}
	}
	return input
}

type Message struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content"`
	Name       *string         `json:"name,omitempty"`
	ToolCalls  any             `json:"tool_calls,omitempty"`
	ToolCallId string          `json:"tool_call_id,omitempty"`
}

type MediaMessage struct {
	Type       string `json:"type"`
	Text       string `json:"text"`
	ImageUrl   any    `json:"image_url,omitempty"`
	InputAudio any    `json:"input_audio,omitempty"`
}

type MessageImageUrl struct {
	Url    string `json:"url"`
	Detail string `json:"detail"`
}

type MessageInputAudio struct {
	Data   string `json:"data"` //base64
	Format string `json:"format"`
}

const (
	ContentTypeText       = "text"
	ContentTypeImageURL   = "image_url"
	ContentTypeInputAudio = "input_audio"
)

func (m Message) StringContent() string {
	var stringContent string
	if err := json.Unmarshal(m.Content, &stringContent); err == nil {
		return stringContent
	}
	return string(m.Content)
}

func (m *Message) SetStringContent(content string) {
	jsonContent, _ := json.Marshal(content)
	m.Content = jsonContent
}

func (m Message) IsStringContent() bool {
	var stringContent string
	if err := json.Unmarshal(m.Content, &stringContent); err == nil {
		return true
	}
	return false
}

func (m Message) ParseContent() []MediaMessage {
	var contentList []MediaMessage
	var stringContent string
	if err := json.Unmarshal(m.Content, &stringContent); err == nil {
		contentList = append(contentList, MediaMessage{
			Type: ContentTypeText,
			Text: stringContent,
		})
		return contentList
	}
	var arrayContent []json.RawMessage
	if err := json.Unmarshal(m.Content, &arrayContent); err == nil {
		for _, contentItem := range arrayContent {
			var contentMap map[string]any
			if err := json.Unmarshal(contentItem, &contentMap); err != nil {
				continue
			}
			switch contentMap["type"] {
			case ContentTypeText:
				if subStr, ok := contentMap["text"].(string); ok {
					contentList = append(contentList, MediaMessage{
						Type: ContentTypeText,
						Text: subStr,
					})
				}
			case ContentTypeImageURL:
				if subObj, ok := contentMap["image_url"].(map[string]any); ok {
					detail, ok := subObj["detail"]
					if ok {
						subObj["detail"] = detail.(string)
					} else {
						subObj["detail"] = "high"
					}
					contentList = append(contentList, MediaMessage{
						Type: ContentTypeImageURL,
						ImageUrl: MessageImageUrl{
							Url:    subObj["url"].(string),
							Detail: subObj["detail"].(string),
						},
					})
				} else if url, ok := contentMap["image_url"].(string); ok {
					contentList = append(contentList, MediaMessage{
						Type: ContentTypeImageURL,
						ImageUrl: MessageImageUrl{
							Url:    url,
							Detail: "high",
						},
					})
				}
			case ContentTypeInputAudio:
				if subObj, ok := contentMap["input_audio"].(map[string]any); ok {
					contentList = append(contentList, MediaMessage{
						Type: ContentTypeInputAudio,
						InputAudio: MessageInputAudio{
							Data:   subObj["data"].(string),
							Format: subObj["format"].(string),
						},
					})
				}
			}
		}
		return contentList
	}
	return nil
}
