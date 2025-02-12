package vertex

import "one-api/relay/channel/claude"

type VertexAIClaudeRequest struct {
	AnthropicVersion string                 `json:"anthropic_version"`
	Messages         []claude.ClaudeMessage `json:"messages"`
	System           string                 `json:"system,omitempty"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	StopSequences    []string               `json:"stop_sequences,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	Temperature      *float64               `json:"temperature,omitempty"`
	TopP             float64                `json:"top_p,omitempty"`
	TopK             int                    `json:"top_k,omitempty"`
	Tools            []claude.Tool          `json:"tools,omitempty"`
	ToolChoice       any                    `json:"tool_choice,omitempty"`
}
