package gemini

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/service"
	"strings"

	"github.com/gin-gonic/gin"
)

// Setting safety to the lowest possible values since Gemini is already powerless enough
func CovertGemini2OpenAI(textRequest dto.GeneralOpenAIRequest) (*GeminiChatRequest, error) {
	geminiRequest := GeminiChatRequest{
		Contents: make([]GeminiChatContent, 0, len(textRequest.Messages)),
		SafetySettings: []GeminiChatSafetySettings{
			{
				Category:  "HARM_CATEGORY_HARASSMENT",
				Threshold: common.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_HATE_SPEECH",
				Threshold: common.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				Threshold: common.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: common.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_CIVIC_INTEGRITY",
				Threshold: common.GeminiSafetySetting,
			},
		},
		GenerationConfig: GeminiChatGenerationConfig{
			Temperature:     textRequest.Temperature,
			TopP:            textRequest.TopP,
			MaxOutputTokens: textRequest.MaxTokens,
		},
	}
	if textRequest.Tools != nil {
		functions := make([]dto.FunctionCall, 0, len(textRequest.Tools))
		googleSearch := false
		for _, tool := range textRequest.Tools {
			if tool.Function.Name == "googleSearch" {
				googleSearch = true
				continue
			}
			if tool.Function.Parameters != nil {
				params, ok := tool.Function.Parameters.(map[string]interface{})
				if ok {
					if props, hasProps := params["properties"].(map[string]interface{}); hasProps {
						if len(props) == 0 {
							tool.Function.Parameters = nil
						}
					}
				}
			}
			functions = append(functions, tool.Function)
		}
		if len(functions) > 0 {
			geminiRequest.Tools = []GeminiChatTools{
				{
					FunctionDeclarations: functions,
				},
			}
		}
		if googleSearch {
			geminiRequest.Tools = append(geminiRequest.Tools, GeminiChatTools{
				GoogleSearch: make(map[string]string),
			})
		}
	} else if textRequest.Functions != nil {
		geminiRequest.Tools = []GeminiChatTools{
			{
				FunctionDeclarations: textRequest.Functions,
			},
		}
	}
	if textRequest.ResponseFormat != nil && (textRequest.ResponseFormat.Type == "json_schema" || textRequest.ResponseFormat.Type == "json_object") {
		geminiRequest.GenerationConfig.ResponseMimeType = "application/json"

		if textRequest.ResponseFormat.JsonSchema != nil && textRequest.ResponseFormat.JsonSchema.Schema != nil {
			cleanedSchema := removeAdditionalPropertiesWithDepth(textRequest.ResponseFormat.JsonSchema.Schema, 0)
			geminiRequest.GenerationConfig.ResponseSchema = cleanedSchema
		}
	}
	tool_call_ids := make(map[string]string)
	//shouldAddDummyModelMessage := false
	for _, message := range textRequest.Messages {

		if message.Role == "system" {
			geminiRequest.SystemInstructions = &GeminiChatContent{
				Parts: []GeminiPart{
					{
						Text: message.StringContent(),
					},
				},
			}
			continue
		} else if message.Role == "tool" {
			if len(geminiRequest.Contents) == 0 || geminiRequest.Contents[len(geminiRequest.Contents)-1].Role != "user" {
				geminiRequest.Contents = append(geminiRequest.Contents, GeminiChatContent{
					Role: "user",
				})
			}
			var parts = &geminiRequest.Contents[len(geminiRequest.Contents)-1].Parts
			name := ""
			if message.Name != nil {
				name = *message.Name
			} else if val, exists := tool_call_ids[message.ToolCallId]; exists {
				name = val
			}
			functionResp := &FunctionResponse{
				Name:     name,
				Response: common.StrToMap(message.StringContent()),
			}
			*parts = append(*parts, GeminiPart{
				FunctionResponse: functionResp,
			})
			continue
		}
		var parts []GeminiPart
		content := GeminiChatContent{
			Role: message.Role,
		}
		isToolCall := false
		if message.ToolCalls != nil {
			message.Role = "model"
			isToolCall = true
			for _, call := range message.ParseToolCalls() {
				toolCall := GeminiPart{
					FunctionCall: &FunctionCall{
						FunctionName: call.Function.Name,
						Arguments:    call.Function.Parameters,
					},
				}
				parts = append(parts, toolCall)
				tool_call_ids[call.ID] = call.Function.Name
			}
		}
		if !isToolCall {
			openaiContent := message.ParseContent()
			imageNum := 0
			for _, part := range openaiContent {
				if part.Type == dto.ContentTypeText {
					parts = append(parts, GeminiPart{
						Text: part.Text,
					})
				} else if part.Type == dto.ContentTypeImageURL {
					imageNum += 1

					if constant.GeminiVisionMaxImageNum != -1 && imageNum > constant.GeminiVisionMaxImageNum {
						return nil, fmt.Errorf("too many images in the message, max allowed is %d", constant.GeminiVisionMaxImageNum)
					}
					// 判断是否是url
					if strings.HasPrefix(part.ImageUrl.(dto.MessageImageUrl).Url, "http") {
						// 是url，获取图片的类型和base64编码的数据
						mimeType, data, _ := service.GetImageFromUrl(part.ImageUrl.(dto.MessageImageUrl).Url)
						parts = append(parts, GeminiPart{
							InlineData: &GeminiInlineData{
								MimeType: mimeType,
								Data:     data,
							},
						})
					} else {
						_, format, base64String, err := service.DecodeBase64ImageData(part.ImageUrl.(dto.MessageImageUrl).Url)
						if err != nil {
							return nil, fmt.Errorf("decode base64 image data failed: %s", err.Error())
						}
						parts = append(parts, GeminiPart{
							InlineData: &GeminiInlineData{
								MimeType: "image/" + format,
								Data:     base64String,
							},
						})
					}
				}
			}
		}

		content.Parts = parts

		// there's no assistant role in gemini and API shall vomit if Role is not user or model
		if content.Role == "assistant" {
			content.Role = "model"
		}
		geminiRequest.Contents = append(geminiRequest.Contents, content)
	}
	return &geminiRequest, nil
}

func removeAdditionalPropertiesWithDepth(schema interface{}, depth int) interface{} {
	if depth >= 5 {
		return schema
	}

	v, ok := schema.(map[string]interface{})
	if !ok || len(v) == 0 {
		return schema
	}

	// 如果type不为object和array，则直接返回
	if typeVal, exists := v["type"]; !exists || (typeVal != "object" && typeVal != "array") {
		return schema
	}
	delete(v, "title")
	switch v["type"] {
	case "object":
		delete(v, "additionalProperties")
		// 处理 properties
		if properties, ok := v["properties"].(map[string]interface{}); ok {
			for key, value := range properties {
				properties[key] = removeAdditionalPropertiesWithDepth(value, depth+1)
			}
		}
		for _, field := range []string{"allOf", "anyOf", "oneOf"} {
			if nested, ok := v[field].([]interface{}); ok {
				for i, item := range nested {
					nested[i] = removeAdditionalPropertiesWithDepth(item, depth+1)
				}
			}
		}
	case "array":
		if items, ok := v["items"].(map[string]interface{}); ok {
			v["items"] = removeAdditionalPropertiesWithDepth(items, depth+1)
		}
	}

	return v
}

func (g *GeminiChatResponse) GetResponseText() string {
	if g == nil {
		return ""
	}
	if len(g.Candidates) > 0 && len(g.Candidates[0].Content.Parts) > 0 {
		return g.Candidates[0].Content.Parts[0].Text
	}
	return ""
}

func getToolCall(item *GeminiPart) *dto.ToolCall {
	argsBytes, err := json.Marshal(item.FunctionCall.Arguments)
	if err != nil {
		//common.SysError("getToolCall failed: " + err.Error())
		return nil
	}
	return &dto.ToolCall{
		ID:   fmt.Sprintf("call_%s", common.GetUUID()),
		Type: "function",
		Function: dto.FunctionCall{
			Arguments: string(argsBytes),
			Name:      item.FunctionCall.FunctionName,
		},
	}
}

// func getToolCalls(candidate *GeminiChatCandidate, index int) []dto.ToolCall {
// 	var toolCalls []dto.ToolCall

// 	item := candidate.Content.Parts[index]
// 	if item.FunctionCall == nil {
// 		return toolCalls
// 	}
// 	argsBytes, err := json.Marshal(item.FunctionCall.Arguments)
// 	if err != nil {
// 		//common.SysError("getToolCalls failed: " + err.Error())
// 		return toolCalls
// 	}
// 	toolCall := dto.ToolCall{
// 		ID:   fmt.Sprintf("call_%s", common.GetUUID()),
// 		Type: "function",
// 		Function: dto.FunctionCall{
// 			Arguments: string(argsBytes),
// 			Name:      item.FunctionCall.FunctionName,
// 		},
// 	}
// 	toolCalls = append(toolCalls, toolCall)
// 	return toolCalls
// }

func responseGeminiChat2OpenAI(response *GeminiChatResponse) *dto.OpenAITextResponse {
	fullTextResponse := dto.OpenAITextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: make([]dto.OpenAITextResponseChoice, 0, len(response.Candidates)),
	}
	content, _ := json.Marshal("")
	for i, candidate := range response.Candidates {
		// jsonData, _ := json.MarshalIndent(candidate, "", "  ")
		// common.SysLog(fmt.Sprintf("candidate: %v", string(jsonData)))
		choice := dto.OpenAITextResponseChoice{
			Index: i,
			Message: dto.Message{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: constant.FinishReasonStop,
		}
		if len(candidate.Content.Parts) > 0 {
			var texts []string
			var tool_calls []dto.ToolCall
			for _, part := range candidate.Content.Parts {
				if part.FunctionCall != nil {
					choice.FinishReason = constant.FinishReasonToolCalls
					if call := getToolCall(&part); call != nil {
						tool_calls = append(tool_calls, *call)
					}
				} else {
					texts = append(texts, part.Text)
				}
			}
			choice.Message.SetStringContent(strings.Join(texts, "\n"))
			choice.Message.SetToolCalls(tool_calls)
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	}
	return &fullTextResponse
}

func streamResponseGeminiChat2OpenAI(geminiResponse *GeminiChatResponse) *dto.ChatCompletionsStreamResponse {
	var choice dto.ChatCompletionsStreamResponseChoice
	//choice.Delta.SetContentString(geminiResponse.GetResponseText())
	if len(geminiResponse.Candidates) > 0 && len(geminiResponse.Candidates[0].Content.Parts) > 0 {
		var texts []string
		var tool_calls []dto.ToolCall
		for _, part := range geminiResponse.Candidates[0].Content.Parts {
			if part.FunctionCall != nil {
				if call := getToolCall(&part); call != nil {
					tool_calls = append(tool_calls, *call)
				}
			} else {
				texts = append(texts, part.Text)
			}
		}
		if len(texts) > 0 {
			choice.Delta.SetContentString(strings.Join(texts, "\n"))
		}
		if len(tool_calls) > 0 {
			choice.Delta.ToolCalls = tool_calls
		}
	}
	var response dto.ChatCompletionsStreamResponse
	response.Object = "chat.completion.chunk"
	response.Model = "gemini"
	response.Choices = []dto.ChatCompletionsStreamResponseChoice{choice}
	return &response
}

func GeminiChatStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	responseText := ""
	id := fmt.Sprintf("chatcmpl-%s", common.GetUUID())
	createAt := common.GetTimestamp()
	var usage = &dto.Usage{}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	service.SetEventStreamHeaders(c)
	for scanner.Scan() {
		data := scanner.Text()
		info.SetFirstResponseTime()
		data = strings.TrimSpace(data)
		if !strings.HasPrefix(data, "data: ") {
			continue
		}
		data = strings.TrimPrefix(data, "data: ")
		data = strings.TrimSuffix(data, "\"")
		var geminiResponse GeminiChatResponse
		err := json.Unmarshal([]byte(data), &geminiResponse)
		if err != nil {
			common.LogError(c, "error unmarshalling stream response: "+err.Error())
			continue
		}

		response := streamResponseGeminiChat2OpenAI(&geminiResponse)
		if response == nil {
			continue
		}
		response.Id = id
		response.Created = createAt
		response.Model = info.UpstreamModelName
		responseText += response.Choices[0].Delta.GetContentString()
		if geminiResponse.UsageMetadata.TotalTokenCount != 0 {
			usage.PromptTokens = geminiResponse.UsageMetadata.PromptTokenCount
			usage.CompletionTokens = geminiResponse.UsageMetadata.CandidatesTokenCount
		}
		err = service.ObjectData(c, response)
		if err != nil {
			common.LogError(c, err.Error())
		}
	}

	response := service.GenerateStopResponse(id, createAt, info.UpstreamModelName, constant.FinishReasonStop)
	service.ObjectData(c, response)

	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens

	if info.ShouldIncludeUsage {
		response = service.GenerateFinalUsageResponse(id, createAt, info.UpstreamModelName, *usage)
		err := service.ObjectData(c, response)
		if err != nil {
			common.SysError("send final response failed: " + err.Error())
		}
	}
	service.Done(c)
	resp.Body.Close()
	return nil, usage
}

func GeminiChatHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var geminiResponse GeminiChatResponse
	err = json.Unmarshal(responseBody, &geminiResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if len(geminiResponse.Candidates) == 0 {
		return &dto.OpenAIErrorWithStatusCode{
			Error: dto.OpenAIError{
				Message: "No candidates returned",
				Type:    "server_error",
				Param:   "",
				Code:    500,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseGeminiChat2OpenAI(&geminiResponse)
	fullTextResponse.Model = info.UpstreamModelName
	usage := dto.Usage{
		PromptTokens:     geminiResponse.UsageMetadata.PromptTokenCount,
		CompletionTokens: geminiResponse.UsageMetadata.CandidatesTokenCount,
		TotalTokens:      geminiResponse.UsageMetadata.TotalTokenCount,
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
