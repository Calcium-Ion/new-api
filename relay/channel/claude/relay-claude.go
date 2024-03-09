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
		return "length"
	default:
		return reason
	}
}

func requestOpenAI2ClaudeComplete(textRequest dto.GeneralOpenAIRequest) *ClaudeRequest {
	claudeRequest := ClaudeRequest{
		Model:             textRequest.Model,
		Prompt:            "",
		MaxTokensToSample: textRequest.MaxTokens,
		StopSequences:     nil,
		Temperature:       textRequest.Temperature,
		TopP:              textRequest.TopP,
		Stream:            textRequest.Stream,
	}
	if claudeRequest.MaxTokensToSample == 0 {
		claudeRequest.MaxTokensToSample = 1000000
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

func requestOpenAI2ClaudeMessage(textRequest dto.GeneralOpenAIRequest) (*ClaudeRequest, error) {
	claudeRequest := ClaudeRequest{
		Model:         textRequest.Model,
		MaxTokens:     textRequest.MaxTokens,
		StopSequences: nil,
		Temperature:   textRequest.Temperature,
		TopP:          textRequest.TopP,
		Stream:        textRequest.Stream,
	}
	if claudeRequest.MaxTokens == 0 {
		claudeRequest.MaxTokens = 4096
	}
	claudeMessages := make([]ClaudeMessage, 0)
	for _, message := range textRequest.Messages {
		if message.Role == "system" {
			claudeRequest.System = message.StringContent()
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
							mimeType, data, _ := common.GetImageFromUrl(imageUrl.Url)
							claudeMediaMessage.Source.MediaType = mimeType
							claudeMediaMessage.Source.Data = data
						} else {
							_, format, base64String, err := common.DecodeBase64ImageData(imageUrl.Url)
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

func streamResponseClaude2OpenAI(reqMode int, claudeResponse *ClaudeResponse) (*dto.ChatCompletionsStreamResponse, *ClaudeUsage) {
	var response dto.ChatCompletionsStreamResponse
	var claudeUsage *ClaudeUsage
	response.Object = "chat.completion.chunk"
	response.Model = claudeResponse.Model
	response.Choices = make([]dto.ChatCompletionsStreamResponseChoice, 0)
	var choice dto.ChatCompletionsStreamResponseChoice
	if reqMode == RequestModeCompletion {
		choice.Delta.Content = claudeResponse.Completion
		finishReason := stopReasonClaude2OpenAI(claudeResponse.StopReason)
		if finishReason != "null" {
			choice.FinishReason = &finishReason
		}
	} else {
		if claudeResponse.Type == "message_start" {
			response.Id = claudeResponse.Message.Id
			response.Model = claudeResponse.Message.Model
			claudeUsage = &claudeResponse.Message.Usage
		} else if claudeResponse.Type == "content_block_delta" {
			choice.Index = claudeResponse.Index
			choice.Delta.Content = claudeResponse.Delta.Text
		} else if claudeResponse.Type == "message_delta" {
			finishReason := stopReasonClaude2OpenAI(*claudeResponse.Delta.StopReason)
			if finishReason != "null" {
				choice.FinishReason = &finishReason
			}
			claudeUsage = &claudeResponse.Usage
		}
	}
	response.Choices = append(response.Choices, choice)
	return &response, claudeUsage
}

func responseClaude2OpenAI(reqMode int, claudeResponse *ClaudeResponse) *dto.OpenAITextResponse {
	choices := make([]dto.OpenAITextResponseChoice, 0)
	fullTextResponse := dto.OpenAITextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
	}
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
		for i, message := range claudeResponse.Content {
			content, _ := json.Marshal(message.Text)
			choice := dto.OpenAITextResponseChoice{
				Index: i,
				Message: dto.Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: stopReasonClaude2OpenAI(claudeResponse.StopReason),
			}
			choices = append(choices, choice)
		}
	}

	fullTextResponse.Choices = choices
	return &fullTextResponse
}

func claudeStreamHandler(requestMode int, modelName string, promptTokens int, c *gin.Context, resp *http.Response) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	responseId := fmt.Sprintf("chatcmpl-%s", common.GetUUID())
	var usage dto.Usage
	responseText := ""
	createdTime := common.GetTimestamp()
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "\n"); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	dataChan := make(chan string)
	stopChan := make(chan bool)
	go func() {
		for scanner.Scan() {
			data := scanner.Text()
			if !strings.HasPrefix(data, "data: ") {
				continue
			}
			data = strings.TrimPrefix(data, "data: ")
			dataChan <- data
		}
		stopChan <- true
	}()
	service.SetEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			// some implementations may add \r at the end of data
			data = strings.TrimSuffix(data, "\r")
			var claudeResponse ClaudeResponse
			err := json.Unmarshal([]byte(data), &claudeResponse)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}

			response, claudeUsage := streamResponseClaude2OpenAI(requestMode, &claudeResponse)
			if requestMode == RequestModeCompletion {
				responseText += claudeResponse.Completion
				responseId = response.Id
			} else {
				if claudeResponse.Type == "message_start" {
					// message_start, 获取usage
					responseId = claudeResponse.Message.Id
					modelName = claudeResponse.Message.Model
					usage.PromptTokens = claudeUsage.InputTokens
				} else if claudeResponse.Type == "content_block_delta" {
					responseText += claudeResponse.Delta.Text
				} else if claudeResponse.Type == "message_delta" {
					usage.CompletionTokens = claudeUsage.OutputTokens
					usage.TotalTokens = claudeUsage.InputTokens + claudeUsage.OutputTokens
				} else {
					return true
				}
			}
			//response.Id = responseId
			response.Id = responseId
			response.Created = createdTime
			response.Model = modelName

			jsonStr, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonStr)})
			return true
		case <-stopChan:
			c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}
	})
	err := resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	if requestMode == RequestModeCompletion {
		usage = *service.ResponseText2Usage(responseText, modelName, promptTokens)
	} else {
		if usage.CompletionTokens == 0 {
			usage = *service.ResponseText2Usage(responseText, modelName, usage.PromptTokens)
		}
	}
	return nil, &usage
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
	fullTextResponse := responseClaude2OpenAI(requestMode, &claudeResponse)
	completionTokens := service.CountTokenText(claudeResponse.Completion, model)
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
