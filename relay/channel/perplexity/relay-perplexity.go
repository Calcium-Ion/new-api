package perplexity

import "one-api/dto"

func requestOpenAI2Perplexity(request dto.GeneralOpenAIRequest) *dto.GeneralOpenAIRequest {
	messages := make([]dto.Message, 0, len(request.Messages))
	for _, message := range request.Messages {
		messages = append(messages, dto.Message{
			Role:    message.Role,
			Content: message.Content,
		})
	}
	return &dto.GeneralOpenAIRequest{
		Model:       request.Model,
		Stream:      request.Stream,
		Messages:    messages,
		Temperature: request.Temperature,
		TopP:        request.TopP,
		MaxTokens:   request.MaxTokens,
	}
}
