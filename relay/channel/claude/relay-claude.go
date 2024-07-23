package claude

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/service"
	"strings"
)

func stopReasonClaude2OpenAI(reason string) string {
	switch reason {
	case "stop_sequence":
		return "stop"
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "max_tokens"
	default:
		return reason
	}
}

func RequestOpenAI2ClaudeComplete(textRequest dto.GeneralOpenAIRequest) *ClaudeRequest {

	claudeRequest := ClaudeRequest{
		Model:         textRequest.Model,
		Prompt:        "",
		StopSequences: nil,
		Temperature:   textRequest.Temperature,
		TopP:          textRequest.TopP,
		TopK:          textRequest.TopK,
		Stream:        textRequest.Stream,
	}
	if claudeRequest.MaxTokensToSample == 0 {
		claudeRequest.MaxTokensToSample = 4096
	}
	prompt := ""
	for _, message := range textRequest.Messages {
		if message.Role == "user" {
			prompt += fmt.Sprintf("\n\nHuman: %s", message.Content)
		} else if message.Role == "assistant" {
			prompt += fmt.Sprintf("\n\nAssistant: %s", message.Content)
		} else if message.Role == "system" {
			if prompt == "" {
				prompt = message.StringContent()
			}
		}
	}
	prompt += "\n\nAssistant:"
	claudeRequest.Prompt = prompt
	return &claudeRequest
}

func RequestOpenAI2ClaudeMessage(textRequest dto.GeneralOpenAIRequest) (*ClaudeRequest, error) {
	claudeTools := make([]Tool, 0, len(textRequest.Tools))

	for _, tool := range textRequest.Tools {
		if params, ok := tool.Function.Parameters.(map[string]any); ok {
			claudeTools = append(claudeTools, Tool{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				InputSchema: InputSchema{
					Type:       params["type"].(string),
					Properties: params["properties"],
					Required:   params["required"],
				},
			})
		}
	}

	claudeRequest := ClaudeRequest{
		Model:         textRequest.Model,
		MaxTokens:     textRequest.MaxTokens,
		StopSequences: nil,
		Temperature:   textRequest.Temperature,
		TopP:          textRequest.TopP,
		TopK:          textRequest.TopK,
		Stream:        textRequest.Stream,
		Tools:         claudeTools,
	}
	if claudeRequest.MaxTokens == 0 {
		claudeRequest.MaxTokens = 4096
	}
	if textRequest.Stop != nil {
		// stop maybe string/array string, convert to array string
		switch textRequest.Stop.(type) {
		case string:
			claudeRequest.StopSequences = []string{textRequest.Stop.(string)}
		case []interface{}:
			stopSequences := make([]string, 0)
			for _, stop := range textRequest.Stop.([]interface{}) {
				stopSequences = append(stopSequences, stop.(string))
			}
			claudeRequest.StopSequences = stopSequences
		}
	}
	formatMessages := make([]dto.Message, 0)
	var lastMessage *dto.Message
	for i, message := range textRequest.Messages {
		//if message.Role == "system" {
		//	if i != 0 {
		//		message.Role = "user"
		//	}
		//}
		if message.Role == "" {
			textRequest.Messages[i].Role = "user"
		}
		fmtMessage := dto.Message{
			Role:    message.Role,
			Content: message.Content,
		}
		if lastMessage != nil && lastMessage.Role == message.Role {
			if lastMessage.IsStringContent() && message.IsStringContent() {
				content, _ := json.Marshal(strings.Trim(fmt.Sprintf("%s %s", lastMessage.StringContent(), message.StringContent()), "\""))
				fmtMessage.Content = content
				// delete last message
				formatMessages = formatMessages[:len(formatMessages)-1]
			}
		}
		if fmtMessage.Content == nil {
			content, _ := json.Marshal("...")
			fmtMessage.Content = content
		}
		formatMessages = append(formatMessages, fmtMessage)
		lastMessage = &textRequest.Messages[i]
	}

	claudeMessages := make([]ClaudeMessage, 0)
	for _, message := range formatMessages {
		if message.Role == "system" {
			if message.IsStringContent() {
				claudeRequest.System = message.StringContent()
			} else {
				contents := message.ParseContent()
				content := ""
				for _, ctx := range contents {
					if ctx.Type == "text" {
						content += ctx.Text
					}
				}
				claudeRequest.System = content
			}
		} else {
			claudeMessage := ClaudeMessage{
				Role: message.Role,
			}
			if message.IsStringContent() {
				claudeMessage.Content = message.StringContent()
			} else {
				claudeMediaMessages := make([]ClaudeMediaMessage, 0)
				for _, mediaMessage := range message.ParseContent() {
					claudeMediaMessage := ClaudeMediaMessage{
						Type: mediaMessage.Type,
					}
					if mediaMessage.Type == "text" {
						claudeMediaMessage.Text = mediaMessage.Text
					} else {
						imageUrl := mediaMessage.ImageUrl.(dto.MessageImageUrl)
						claudeMediaMessage.Type = "image"
						claudeMediaMessage.Source = &ClaudeMessageSource{
							Type: "base64",
						}
						// 判断是否是url
						if strings.HasPrefix(imageUrl.Url, "http") {
							// 是url，获取图片的类型和base64编码的数据
							mimeType, data, _ := service.GetImageFromUrl(imageUrl.Url)
							claudeMediaMessage.Source.MediaType = mimeType
							claudeMediaMessage.Source.Data = data
						} else {
							_, format, base64String, err := service.DecodeBase64ImageData(imageUrl.Url)
							if err != nil {
								return nil, err
							}
							claudeMediaMessage.Source.MediaType = "image/" + format
							claudeMediaMessage.Source.Data = base64String
						}
					}
					claudeMediaMessages = append(claudeMediaMessages, claudeMediaMessage)
				}
				claudeMessage.Content = claudeMediaMessages
			}
			claudeMessages = append(claudeMessages, claudeMessage)
		}
	}
	claudeRequest.Prompt = ""
	claudeRequest.Messages = claudeMessages
	return &claudeRequest, nil
}

func StreamResponseClaude2OpenAI(reqMode int, claudeResponse *ClaudeResponse) (*dto.ChatCompletionsStreamResponse, *ClaudeUsage) {
	var response dto.ChatCompletionsStreamResponse
	var claudeUsage *ClaudeUsage
	response.Object = "chat.completion.chunk"
	response.Model = claudeResponse.Model
	response.Choices = make([]dto.ChatCompletionsStreamResponseChoice, 0)
	tools := make([]dto.ToolCall, 0)
	var choice dto.ChatCompletionsStreamResponseChoice
	if reqMode == RequestModeCompletion {
		choice.Delta.SetContentString(claudeResponse.Completion)
		finishReason := stopReasonClaude2OpenAI(claudeResponse.StopReason)
		if finishReason != "null" {
			choice.FinishReason = &finishReason
		}
	} else {
		if claudeResponse.Type == "message_start" {
			response.Id = claudeResponse.Message.Id
			response.Model = claudeResponse.Message.Model
			claudeUsage = &claudeResponse.Message.Usage
			choice.Delta.SetContentString("")
			choice.Delta.Role = "assistant"
		} else if claudeResponse.Type == "content_block_start" {
			if claudeResponse.ContentBlock != nil {
				//choice.Delta.SetContentString(claudeResponse.ContentBlock.Text)
				if claudeResponse.ContentBlock.Type == "tool_use" {
					tools = append(tools, dto.ToolCall{
						ID:   claudeResponse.ContentBlock.Id,
						Type: "function",
						Function: dto.FunctionCall{
							Name:      claudeResponse.ContentBlock.Name,
							Arguments: "",
						},
					})
				}
			} else {
				return nil, nil
			}
		} else if claudeResponse.Type == "content_block_delta" {
			if claudeResponse.Delta != nil {
				choice.Index = claudeResponse.Index
				choice.Delta.SetContentString(claudeResponse.Delta.Text)
				if claudeResponse.Delta.Type == "input_json_delta" {
					tools = append(tools, dto.ToolCall{
						Function: dto.FunctionCall{
							Arguments: claudeResponse.Delta.PartialJson,
						},
					})
				}
			}
		} else if claudeResponse.Type == "message_delta" {
			finishReason := stopReasonClaude2OpenAI(*claudeResponse.Delta.StopReason)
			if finishReason != "null" {
				choice.FinishReason = &finishReason
			}
			claudeUsage = &claudeResponse.Usage
		} else if claudeResponse.Type == "message_stop" {
			return nil, nil
		} else {
			return nil, nil
		}
	}
	if claudeUsage == nil {
		claudeUsage = &ClaudeUsage{}
	}
	if len(tools) > 0 {
		choice.Delta.Content = nil // compatible with other OpenAI derivative applications, like LobeOpenAICompatibleFactory ...
		choice.Delta.ToolCalls = tools
	}
	response.Choices = append(response.Choices, choice)

	return &response, claudeUsage
}

func ResponseClaude2OpenAI(reqMode int, claudeResponse *ClaudeResponse) *dto.OpenAITextResponse {
	choices := make([]dto.OpenAITextResponseChoice, 0)
	fullTextResponse := dto.OpenAITextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
	}
	var responseText string
	if len(claudeResponse.Content) > 0 {
		responseText = claudeResponse.Content[0].Text
	}
	tools := make([]dto.ToolCall, 0)
	if reqMode == RequestModeCompletion {
		content, _ := json.Marshal(strings.TrimPrefix(claudeResponse.Completion, " "))
		choice := dto.OpenAITextResponseChoice{
			Index: 0,
			Message: dto.Message{
				Role:    "assistant",
				Content: content,
				Name:    nil,
			},
			FinishReason: stopReasonClaude2OpenAI(claudeResponse.StopReason),
		}
		choices = append(choices, choice)
	} else {
		fullTextResponse.Id = claudeResponse.Id
		for _, message := range claudeResponse.Content {
			if message.Type == "tool_use" {
				args, _ := json.Marshal(message.Input)
				tools = append(tools, dto.ToolCall{
					ID:   message.Id,
					Type: "function", // compatible with other OpenAI derivative applications
					Function: dto.FunctionCall{
						Name:      message.Name,
						Arguments: string(args),
					},
				})
			}
		}
	}
	choice := dto.OpenAITextResponseChoice{
		Index: 0,
		Message: dto.Message{
			Role: "assistant",
		},
		FinishReason: stopReasonClaude2OpenAI(claudeResponse.StopReason),
	}
	choice.SetStringContent(responseText)
	if len(tools) > 0 {
		choice.Message.ToolCalls = tools
	}
	choices = append(choices, choice)
	fullTextResponse.Choices = choices
	return &fullTextResponse
}

func claudeStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo, requestMode int) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	responseId := fmt.Sprintf("chatcmpl-%s", common.GetUUID())
	var usage *dto.Usage
	usage = &dto.Usage{}
	responseText := ""
	createdTime := common.GetTimestamp()
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)
	service.SetEventStreamHeaders(c)

	for scanner.Scan() {
		data := scanner.Text()
		info.SetFirstResponseTime()
		if len(data) < 6 || !strings.HasPrefix(data, "data:") {
			continue
		}
		data = strings.TrimPrefix(data, "data:")
		data = strings.TrimSpace(data)
		var claudeResponse ClaudeResponse
		err := json.Unmarshal([]byte(data), &claudeResponse)
		if err != nil {
			common.SysError("error unmarshalling stream response: " + err.Error())
			continue
		}

		response, claudeUsage := StreamResponseClaude2OpenAI(requestMode, &claudeResponse)
		if response == nil {
			continue
		}
		if requestMode == RequestModeCompletion {
			responseText += claudeResponse.Completion
			responseId = response.Id
		} else {
			if claudeResponse.Type == "message_start" {
				// message_start, 获取usage
				responseId = claudeResponse.Message.Id
				info.UpstreamModelName = claudeResponse.Message.Model
				usage.PromptTokens = claudeUsage.InputTokens
			} else if claudeResponse.Type == "content_block_delta" {
				responseText += claudeResponse.Delta.Text
			} else if claudeResponse.Type == "message_delta" {
				usage.CompletionTokens = claudeUsage.OutputTokens
				usage.TotalTokens = claudeUsage.InputTokens + claudeUsage.OutputTokens
			} else if claudeResponse.Type == "content_block_start" {

			} else {
				continue
			}
		}
		//response.Id = responseId
		response.Id = responseId
		response.Created = createdTime
		response.Model = info.UpstreamModelName

		err = service.ObjectData(c, response)
		if err != nil {
			common.LogError(c, "send_stream_response_failed: "+err.Error())
		}
	}

	if requestMode == RequestModeCompletion {
		usage, _ = service.ResponseText2Usage(responseText, info.UpstreamModelName, info.PromptTokens)
	} else {
		if usage.PromptTokens == 0 {
			usage.PromptTokens = info.PromptTokens
		}
		if usage.CompletionTokens == 0 {
			usage, _ = service.ResponseText2Usage(responseText, info.UpstreamModelName, usage.PromptTokens)
		}
	}
	if info.ShouldIncludeUsage {
		response := service.GenerateFinalUsageResponse(responseId, createdTime, info.UpstreamModelName, *usage)
		err := service.ObjectData(c, response)
		if err != nil {
			common.SysError("send final response failed: " + err.Error())
		}
	}
	service.Done(c)
	resp.Body.Close()
	return nil, usage
}

func claudeHandler(requestMode int, c *gin.Context, resp *http.Response, promptTokens int, model string) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var claudeResponse ClaudeResponse
	err = json.Unmarshal(responseBody, &claudeResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if claudeResponse.Error.Type != "" {
		return &dto.OpenAIErrorWithStatusCode{
			Error: dto.OpenAIError{
				Message: claudeResponse.Error.Message,
				Type:    claudeResponse.Error.Type,
				Param:   "",
				Code:    claudeResponse.Error.Type,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := ResponseClaude2OpenAI(requestMode, &claudeResponse)
	completionTokens, err := service.CountTokenText(claudeResponse.Completion, model)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "count_token_text_failed", http.StatusInternalServerError), nil
	}
	usage := dto.Usage{}
	if requestMode == RequestModeCompletion {
		usage.PromptTokens = promptTokens
		usage.CompletionTokens = completionTokens
		usage.TotalTokens = promptTokens + completionTokens
	} else {
		usage.PromptTokens = claudeResponse.Usage.InputTokens
		usage.CompletionTokens = claudeResponse.Usage.OutputTokens
		usage.TotalTokens = claudeResponse.Usage.InputTokens + claudeResponse.Usage.OutputTokens
	}
	fullTextResponse.Usage = usage
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &usage
}
