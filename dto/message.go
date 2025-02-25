package dto

type CreateMessageRequest struct {
	ConversationID string
	Role           string
	Content        string
	ContentType    string
}
