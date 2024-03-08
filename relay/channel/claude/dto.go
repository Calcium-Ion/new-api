package claude

type ClaudeMetadata struct {
	UserId string `json:"user_id"`
}

type ClaudeMediaMessage struct {
	Type       string               `json:"type"`
	Text       string               `json:"text,omitempty"`
	Source     *ClaudeMessageSource `json:"source,omitempty"`
	Usage      *ClaudeUsage         `json:"usage,omitempty"`
	StopReason *string              `json:"stop_reason,omitempty"`
}

type ClaudeMessageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type ClaudeRequest struct {
	Model             string          `json:"model"`
	Prompt            string          `json:"prompt,omitempty"`
	System            string          `json:"system,omitempty"`
	Messages          []ClaudeMessage `json:"messages,omitempty"`
	MaxTokensToSample uint            `json:"max_tokens_to_sample,omitempty"`
	MaxTokens         uint            `json:"max_tokens,omitempty"`
	StopSequences     []string        `json:"stop_sequences,omitempty"`
	Temperature       float64         `json:"temperature,omitempty"`
	TopP              float64         `json:"top_p,omitempty"`
	TopK              int             `json:"top_k,omitempty"`
	//ClaudeMetadata    `json:"metadata,omitempty"`
	Stream bool `json:"stream,omitempty"`
}

type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ClaudeResponse struct {
	Id         string               `json:"id"`
	Type       string               `json:"type"`
	Content    []ClaudeMediaMessage `json:"content"`
	Completion string               `json:"completion"`
	StopReason string               `json:"stop_reason"`
	Model      string               `json:"model"`
	Error      ClaudeError          `json:"error"`
	Usage      ClaudeUsage          `json:"usage"`
	Index      int                  `json:"index"`   // stream only
	Delta      *ClaudeMediaMessage  `json:"delta"`   // stream only
	Message    *ClaudeResponse      `json:"message"` // stream only: message_start
}

//type ClaudeResponseChoice struct {
//	Index   int                `json:"index"`
//	Type    string             `json:"type"`
//}

type ClaudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
