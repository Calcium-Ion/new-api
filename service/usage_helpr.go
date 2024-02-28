package service

import (
	"errors"
	"one-api/dto"
	"one-api/relay/constant"
)

func GetPromptTokens(textRequest dto.GeneralOpenAIRequest, relayMode int) (int, error) {
	switch relayMode {
	case constant.RelayModeChatCompletions:
		return CountTokenMessages(textRequest.Messages, textRequest.Model)
	case constant.RelayModeCompletions:
		return CountTokenInput(textRequest.Prompt, textRequest.Model), nil
	case constant.RelayModeModerations:
		return CountTokenInput(textRequest.Input, textRequest.Model), nil
	}
	return 0, errors.New("unknown relay mode")
}

func ResponseText2Usage(responseText string, modeName string, promptTokens int) *dto.Usage {
	usage := &dto.Usage{}
	usage.PromptTokens = promptTokens
	usage.CompletionTokens = CountTokenText(responseText, modeName)
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	return usage
}
