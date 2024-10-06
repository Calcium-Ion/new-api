package dify

import (
	"bufio"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/service"
	"strings"
)

func requestOpenAI2Dify(request dto.GeneralOpenAIRequest) *DifyChatRequest {
	content := ""
	for _, message := range request.Messages {
		if message.Role == "system" {
			content += "SYSTEM: \n" + message.StringContent() + "\n"
		} else if message.Role == "assistant" {
			content += "ASSISTANT: \n" + message.StringContent() + "\n"
		} else {
			content += "USER: \n" + message.StringContent() + "\n"
		}
	}
	mode := "blocking"
	if request.Stream {
		mode = "streaming"
	}
	user := request.User
	if user == "" {
		user = "api-user"
	}
	return &DifyChatRequest{
		Inputs:           make(map[string]interface{}),
		Query:            content,
		ResponseMode:     mode,
		User:             user,
		AutoGenerateName: false,
	}
}

func streamResponseDify2OpenAI(difyResponse DifyChunkChatCompletionResponse) *dto.ChatCompletionsStreamResponse {
	response := dto.ChatCompletionsStreamResponse{
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "dify",
	}
	var choice dto.ChatCompletionsStreamResponseChoice
	if constant.DifyDebug && difyResponse.Event == "workflow_started" {
		choice.Delta.SetContentString("Workflow: " + difyResponse.Data.WorkflowId + "\n")
	} else if constant.DifyDebug && difyResponse.Event == "node_started" {
		choice.Delta.SetContentString("Node: " + difyResponse.Data.NodeId + "\n")
	} else if difyResponse.Event == "message" || difyResponse.Event == "agent_message" {
		choice.Delta.SetContentString(difyResponse.Answer)
	}
	response.Choices = append(response.Choices, choice)
	return &response
}

func difyStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var responseText string
	usage := &dto.Usage{}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	service.SetEventStreamHeaders(c)

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < 5 || !strings.HasPrefix(data, "data:") {
			continue
		}
		data = strings.TrimPrefix(data, "data:")
		var difyResponse DifyChunkChatCompletionResponse
		err := json.Unmarshal([]byte(data), &difyResponse)
		if err != nil {
			common.SysError("error unmarshalling stream response: " + err.Error())
			continue
		}
		var openaiResponse dto.ChatCompletionsStreamResponse
		if difyResponse.Event == "message_end" {
			usage = &difyResponse.MetaData.Usage
			break
		} else if difyResponse.Event == "error" {
			break
		} else {
			openaiResponse = *streamResponseDify2OpenAI(difyResponse)
			if len(openaiResponse.Choices) != 0 {
				responseText += openaiResponse.Choices[0].Delta.GetContentString()
			}
		}
		err = service.ObjectData(c, openaiResponse)
		if err != nil {
			common.SysError(err.Error())
		}
	}
	if err := scanner.Err(); err != nil {
		common.SysError("error reading stream: " + err.Error())
	}
	service.Done(c)
	err := resp.Body.Close()
	if err != nil {
		//return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
		common.SysError("close_response_body_failed: " + err.Error())
	}
	if usage.TotalTokens == 0 {
		usage.PromptTokens = info.PromptTokens
		usage.CompletionTokens, _ = service.CountTextToken("gpt-3.5-turbo", responseText)
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	return nil, usage
}

func difyHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var difyResponse DifyChatCompletionResponse
	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &difyResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	fullTextResponse := dto.OpenAITextResponse{
		Id:      difyResponse.ConversationId,
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Usage:   difyResponse.MetaData.Usage,
	}
	content, _ := json.Marshal(difyResponse.Answer)
	choice := dto.OpenAITextResponseChoice{
		Index: 0,
		Message: dto.Message{
			Role:    "assistant",
			Content: content,
		},
		FinishReason: "stop",
	}
	fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &difyResponse.MetaData.Usage
}
