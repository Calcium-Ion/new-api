package dto

type ContextRequest struct {
	Model    string    `json:"model,omitempty"`
	Messages []Message `json:"messages,omitempty"`
	Ttl      int       `json:"ttl,omitempty"`
	Mode     string    `json:"mode,omitempty"`
}
