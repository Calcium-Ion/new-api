package service

import (
	"one-api/dto"
)

//func GetPromptTokens(textRequest dto.GeneralOpenAIRequest, relayMode int) (int, error) {
//	switch relayMode {
//	case constant.RelayModeChatCompletions:
//		return CountTokenMessages(textRequest.Messages, textRequest.Model)
//	case constant.RelayModeCompletions:
//		return CountTokenInput(textRequest.Prompt, textRequest.Model), nil
//	case constant.RelayModeModerations:
//		return CountTokenInput(textRequest.Input, textRequest.Model), nil
//	}
//	return 0, errors.New("unknown relay mode")
//}

func ResponseText2Usage(responseText string, modeName string, promptTokens int) (*dto.Usage, error) {
	usage := &dto.Usage{}
	usage.PromptTokens = promptTokens
	ctkm, err := CountTextToken(responseText, modeName)
	usage.CompletionTokens = ctkm
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	return usage, err
}

func ValidUsage(usage *dto.Usage) bool {
	return usage != nil && (usage.PromptTokens != 0 || usage.CompletionTokens != 0)
}
