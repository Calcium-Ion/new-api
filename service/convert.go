package service

import (
	"encoding/json"
	"fmt"
	"one-api/common"
	"one-api/dto"
	relaycommon "one-api/relay/common"
)

func ClaudeToOpenAIRequest(claudeRequest dto.ClaudeRequest) (*dto.GeneralOpenAIRequest, error) {
	openAIRequest := dto.GeneralOpenAIRequest{
		Model:       claudeRequest.Model,
		MaxTokens:   claudeRequest.MaxTokens,
		Temperature: claudeRequest.Temperature,
		TopP:        claudeRequest.TopP,
		Stream:      claudeRequest.Stream,
	}

	// Convert stop sequences
	if len(claudeRequest.StopSequences) == 1 {
		openAIRequest.Stop = claudeRequest.StopSequences[0]
	} else if len(claudeRequest.StopSequences) > 1 {
		openAIRequest.Stop = claudeRequest.StopSequences
	}

	// Convert tools
	tools, _ := common.Any2Type[[]dto.Tool](claudeRequest.Tools)
	openAITools := make([]dto.ToolCallRequest, 0)
	for _, claudeTool := range tools {
		openAITool := dto.ToolCallRequest{
			Type: "function",
			Function: dto.FunctionRequest{
				Name:        claudeTool.Name,
				Description: claudeTool.Description,
				Parameters:  claudeTool.InputSchema,
			},
		}
		openAITools = append(openAITools, openAITool)
	}
	openAIRequest.Tools = openAITools

	// Convert messages
	openAIMessages := make([]dto.Message, 0)

	// Add system message if present
	if claudeRequest.System != nil {
		if claudeRequest.IsStringSystem() {
			openAIMessage := dto.Message{
				Role: "system",
			}
			openAIMessage.SetStringContent(claudeRequest.GetStringSystem())
			openAIMessages = append(openAIMessages, openAIMessage)
		} else {
			systems := claudeRequest.ParseSystem()
			if len(systems) > 0 {
				systemStr := ""
				openAIMessage := dto.Message{
					Role: "system",
				}
				for _, system := range systems {
					systemStr += system.Type
				}
				openAIMessage.SetStringContent(systemStr)
				openAIMessages = append(openAIMessages, openAIMessage)
			}
		}
	}
	for _, claudeMessage := range claudeRequest.Messages {
		openAIMessage := dto.Message{
			Role: claudeMessage.Role,
		}

		//log.Printf("claudeMessage.Content: %v", claudeMessage.Content)
		if claudeMessage.IsStringContent() {
			openAIMessage.SetStringContent(claudeMessage.GetStringContent())
		} else {
			content, err := claudeMessage.ParseContent()
			if err != nil {
				return nil, err
			}
			contents := content
			var toolCalls []dto.ToolCallRequest
			mediaMessages := make([]dto.MediaContent, 0, len(contents))

			for _, mediaMsg := range contents {
				switch mediaMsg.Type {
				case "text":
					message := dto.MediaContent{
						Type: "text",
						Text: mediaMsg.GetText(),
					}
					mediaMessages = append(mediaMessages, message)
				case "image":
					// Handle image conversion (base64 to URL or keep as is)
					imageData := fmt.Sprintf("data:%s;base64,%s", mediaMsg.Source.MediaType, mediaMsg.Source.Data)
					//textContent += fmt.Sprintf("[Image: %s]", imageData)
					mediaMessage := dto.MediaContent{
						Type:     "image_url",
						ImageUrl: &dto.MessageImageUrl{Url: imageData},
					}
					mediaMessages = append(mediaMessages, mediaMessage)
				case "tool_use":
					toolCall := dto.ToolCallRequest{
						ID:   mediaMsg.Id,
						Type: "function",
						Function: dto.FunctionRequest{
							Name:      mediaMsg.Name,
							Arguments: toJSONString(mediaMsg.Input),
						},
					}
					toolCalls = append(toolCalls, toolCall)
				case "tool_result":
					// Add tool result as a separate message
					oaiToolMessage := dto.Message{
						Role:       "tool",
						Name:       &mediaMsg.Name,
						ToolCallId: mediaMsg.ToolUseId,
					}
					//oaiToolMessage.SetStringContent(*mediaMsg.GetMediaContent().Text)
					if mediaMsg.IsStringContent() {
						oaiToolMessage.SetStringContent(mediaMsg.GetStringContent())
					} else {
						mediaContents := mediaMsg.ParseMediaContent()
						if len(mediaContents) > 0 && mediaContents[0].Text != nil {
							oaiToolMessage.SetStringContent(*mediaContents[0].Text)
						}
					}
					openAIMessages = append(openAIMessages, oaiToolMessage)
				}
			}

			if len(mediaMessages) > 0 {
				openAIMessage.SetMediaContent(mediaMessages)
			}

			if len(toolCalls) > 0 {
				openAIMessage.SetToolCalls(toolCalls)
			}
		}
		if len(openAIMessage.ParseContent()) > 0 {
			openAIMessages = append(openAIMessages, openAIMessage)
		}
	}

	openAIRequest.Messages = openAIMessages

	return &openAIRequest, nil
}

func OpenAIErrorToClaudeError(openAIError *dto.OpenAIErrorWithStatusCode) *dto.ClaudeErrorWithStatusCode {
	claudeError := dto.ClaudeError{
		Type:    "new_api_error",
		Message: openAIError.Error.Message,
	}
	return &dto.ClaudeErrorWithStatusCode{
		Error:      claudeError,
		StatusCode: openAIError.StatusCode,
	}
}

func ClaudeErrorToOpenAIError(claudeError *dto.ClaudeErrorWithStatusCode) *dto.OpenAIErrorWithStatusCode {
	openAIError := dto.OpenAIError{
		Message: claudeError.Error.Message,
		Type:    "new_api_error",
	}
	return &dto.OpenAIErrorWithStatusCode{
		Error:      openAIError,
		StatusCode: claudeError.StatusCode,
	}
}

func generateStopBlock(index int) *dto.ClaudeResponse {
	return &dto.ClaudeResponse{
		Type:  "content_block_stop",
		Index: common.GetPointer[int](index),
	}
}

func StreamResponseOpenAI2Claude(openAIResponse *dto.ChatCompletionsStreamResponse, info *relaycommon.RelayInfo) []*dto.ClaudeResponse {
	var claudeResponses []*dto.ClaudeResponse
	if info.SendResponseCount == 1 {
		msg := &dto.ClaudeMediaMessage{
			Id:    openAIResponse.Id,
			Model: openAIResponse.Model,
			Type:  "message",
			Role:  "assistant",
			Usage: &dto.ClaudeUsage{
				InputTokens:  info.PromptTokens,
				OutputTokens: 0,
			},
		}
		msg.SetContent(make([]any, 0))
		claudeResponses = append(claudeResponses, &dto.ClaudeResponse{
			Type:    "message_start",
			Message: msg,
		})
		claudeResponses = append(claudeResponses)
		//claudeResponses = append(claudeResponses, &dto.ClaudeResponse{
		//	Type: "ping",
		//})
		if openAIResponse.IsToolCall() {
			resp := &dto.ClaudeResponse{
				Type: "content_block_start",
				ContentBlock: &dto.ClaudeMediaMessage{
					Id:   openAIResponse.GetFirstToolCall().ID,
					Type: "tool_use",
					Name: openAIResponse.GetFirstToolCall().Function.Name,
				},
			}
			resp.SetIndex(0)
			claudeResponses = append(claudeResponses, resp)
		} else {
			resp := &dto.ClaudeResponse{
				Type: "content_block_start",
				ContentBlock: &dto.ClaudeMediaMessage{
					Type: "text",
					Text: common.GetPointer[string](""),
				},
			}
			resp.SetIndex(0)
			claudeResponses = append(claudeResponses, resp)
		}
		return claudeResponses
	}

	if len(openAIResponse.Choices) == 0 {
		// no choices
		// TODO: handle this case
		return claudeResponses
	} else {
		chosenChoice := openAIResponse.Choices[0]
		if chosenChoice.FinishReason != nil && *chosenChoice.FinishReason != "" {
			// should be done
			claudeResponses = append(claudeResponses, generateStopBlock(info.ClaudeConvertInfo.Index))
			if openAIResponse.Usage != nil {
				claudeResponses = append(claudeResponses, &dto.ClaudeResponse{
					Type: "message_delta",
					Usage: &dto.ClaudeUsage{
						InputTokens:  openAIResponse.Usage.PromptTokens,
						OutputTokens: openAIResponse.Usage.CompletionTokens,
					},
					Delta: &dto.ClaudeMediaMessage{
						StopReason: common.GetPointer[string](stopReasonOpenAI2Claude(*chosenChoice.FinishReason)),
					},
				})
			}
			claudeResponses = append(claudeResponses, &dto.ClaudeResponse{
				Type: "message_stop",
			})
		} else {
			var claudeResponse dto.ClaudeResponse
			claudeResponse.SetIndex(0)
			claudeResponse.Type = "content_block_delta"
			if len(chosenChoice.Delta.ToolCalls) > 0 {
				if info.ClaudeConvertInfo.LastMessagesType == relaycommon.LastMessageTypeText {
					claudeResponses = append(claudeResponses, generateStopBlock(info.ClaudeConvertInfo.Index))
					info.ClaudeConvertInfo.Index++
					claudeResponses = append(claudeResponses, &dto.ClaudeResponse{
						Index: &info.ClaudeConvertInfo.Index,
						Type:  "content_block_start",
						ContentBlock: &dto.ClaudeMediaMessage{
							Id:    openAIResponse.GetFirstToolCall().ID,
							Type:  "tool_use",
							Name:  openAIResponse.GetFirstToolCall().Function.Name,
							Input: map[string]interface{}{},
						},
					})
				}
				info.ClaudeConvertInfo.LastMessagesType = relaycommon.LastMessageTypeTools
				// tools delta
				claudeResponse.Delta = &dto.ClaudeMediaMessage{
					Type:        "input_json_delta",
					PartialJson: &chosenChoice.Delta.ToolCalls[0].Function.Arguments,
				}
			} else {
				info.ClaudeConvertInfo.LastMessagesType = relaycommon.LastMessageTypeText
				// text delta
				claudeResponse.Delta = &dto.ClaudeMediaMessage{
					Type: "text_delta",
					Text: common.GetPointer[string](chosenChoice.Delta.GetContentString()),
				}
			}
			claudeResponse.Index = &info.ClaudeConvertInfo.Index
			claudeResponses = append(claudeResponses, &claudeResponse)
		}
	}

	return claudeResponses
}

func ResponseOpenAI2Claude(openAIResponse *dto.OpenAITextResponse, info *relaycommon.RelayInfo) *dto.ClaudeResponse {
	var stopReason string
	contents := make([]dto.ClaudeMediaMessage, 0)
	claudeResponse := &dto.ClaudeResponse{
		Id:    openAIResponse.Id,
		Type:  "message",
		Role:  "assistant",
		Model: openAIResponse.Model,
	}
	for _, choice := range openAIResponse.Choices {
		stopReason = stopReasonOpenAI2Claude(choice.FinishReason)
		claudeContent := dto.ClaudeMediaMessage{}
		if choice.FinishReason == "tool_calls" {
			claudeContent.Type = "tool_use"
			claudeContent.Id = choice.Message.ToolCallId
			claudeContent.Name = choice.Message.ParseToolCalls()[0].Function.Name
			var mapParams map[string]interface{}
			if err := json.Unmarshal([]byte(choice.Message.ParseToolCalls()[0].Function.Arguments), &mapParams); err == nil {
				claudeContent.Input = mapParams
			} else {
				claudeContent.Input = choice.Message.ParseToolCalls()[0].Function.Arguments
			}
		} else {
			claudeContent.Type = "text"
			claudeContent.SetText(choice.Message.StringContent())
		}
		contents = append(contents, claudeContent)
	}
	claudeResponse.Content = contents
	claudeResponse.StopReason = stopReason
	claudeResponse.Usage = &dto.ClaudeUsage{
		InputTokens:  openAIResponse.PromptTokens,
		OutputTokens: openAIResponse.CompletionTokens,
	}

	return claudeResponse
}

func stopReasonOpenAI2Claude(reason string) string {
	switch reason {
	case "stop":
		return "end_turn"
	case "stop_sequence":
		return "stop_sequence"
	case "max_tokens":
		return "max_tokens"
	case "tool_calls":
		return "tool_use"
	default:
		return reason
	}
}

func toJSONString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
