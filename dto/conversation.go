package dto

type CreateConversationRequest struct {
	Title string `json:"title"`
	Model string `json:"model"`
}
