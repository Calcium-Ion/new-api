package ollama

import "one-api/dto"

func requestOpenAI2Ollama(request dto.GeneralOpenAIRequest) *OllamaRequest {
	messages := make([]dto.Message, 0, len(request.Messages))
	for _, message := range request.Messages {
		messages = append(messages, dto.Message{
			Role:    message.Role,
			Content: message.Content,
		})
	}
	str, ok := request.Stop.(string)
	var Stop []string
	if ok {
		Stop = []string{str}
	} else {
		Stop, _ = request.Stop.([]string)
	}
	return &OllamaRequest{
		Model:    request.Model,
		Messages: messages,
		Stream:   request.Stream,
		Options: &OllamaOptions{
			Temperature: request.Temperature,
			Seed:        request.Seed,
			Topp:        request.TopP,
			Stop:        Stop,
		},
	}
}
