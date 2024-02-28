package claude

type ClaudeMetadata struct {
	UserId string `json:"user_id"`
}

type ClaudeRequest struct {
	Model             string   `json:"model"`
	Prompt            string   `json:"prompt"`
	MaxTokensToSample uint     `json:"max_tokens_to_sample"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
	Temperature       float64  `json:"temperature,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	//ClaudeMetadata    `json:"metadata,omitempty"`
	Stream bool `json:"stream,omitempty"`
}

type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ClaudeResponse struct {
	Completion string      `json:"completion"`
	StopReason string      `json:"stop_reason"`
	Model      string      `json:"model"`
	Error      ClaudeError `json:"error"`
}
