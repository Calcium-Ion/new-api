package aws

import (
	"one-api/relay/channel/claude"
)

type AwsClaudeRequest struct {
	// AnthropicVersion should be "bedrock-2023-05-31"
	AnthropicVersion string                 `json:"anthropic_version"`
	System           string                 `json:"system"`
	Messages         []claude.ClaudeMessage `json:"messages"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	Temperature      float64                `json:"temperature,omitempty"`
	TopP             float64                `json:"top_p,omitempty"`
	TopK             int                    `json:"top_k,omitempty"`
	StopSequences    []string               `json:"stop_sequences,omitempty"`
	Tools            []claude.Tool          `json:"tools,omitempty"`
	ToolChoice       any                    `json:"tool_choice,omitempty"`
}
