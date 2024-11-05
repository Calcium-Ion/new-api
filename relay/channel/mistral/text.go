package mistral

import (
	"encoding/json"
	"one-api/dto"
)

func requestOpenAI2Mistral(request dto.GeneralOpenAIRequest) *dto.GeneralOpenAIRequest {
	messages := make([]dto.Message, 0, len(request.Messages))
	for _, message := range request.Messages {
		if !message.IsStringContent() {
			mediaMessages := message.ParseContent()
			for j, mediaMessage := range mediaMessages {
				if mediaMessage.Type == dto.ContentTypeImageURL {
					imageUrl := mediaMessage.ImageUrl.(dto.MessageImageUrl)
					mediaMessage.ImageUrl = imageUrl.Url
					mediaMessages[j] = mediaMessage
				}
			}
			messageRaw, _ := json.Marshal(mediaMessages)
			message.Content = messageRaw
		}
		messages = append(messages, dto.Message{
			Role:       message.Role,
			Content:    message.Content,
			ToolCalls:  message.ToolCalls,
			ToolCallId: message.ToolCallId,
		})
	}
	return &dto.GeneralOpenAIRequest{
		Model:       request.Model,
		Stream:      request.Stream,
		Messages:    messages,
		Temperature: request.Temperature,
		TopP:        request.TopP,
		MaxTokens:   request.MaxTokens,
		Tools:       request.Tools,
		ToolChoice:  request.ToolChoice,
	}
}
