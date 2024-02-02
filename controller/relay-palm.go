package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
)

// https://developers.generativeai.google/api/rest/generativelanguage/models/generateMessage#request-body
// https://developers.generativeai.google/api/rest/generativelanguage/models/generateMessage#response-body

type PaLMChatMessage struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type PaLMFilter struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type PaLMPrompt struct {
	Messages []PaLMChatMessage `json:"messages"`
}

type PaLMChatRequest struct {
	Prompt         PaLMPrompt `json:"prompt"`
	Temperature    float64    `json:"temperature,omitempty"`
	CandidateCount int        `json:"candidateCount,omitempty"`
	TopP           float64    `json:"topP,omitempty"`
	TopK           uint       `json:"topK,omitempty"`
}

type PaLMError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type PaLMChatResponse struct {
	Candidates []PaLMChatMessage `json:"candidates"`
	Messages   []Message         `json:"messages"`
	Filters    []PaLMFilter      `json:"filters"`
	Error      PaLMError         `json:"error"`
}

func requestOpenAI2PaLM(textRequest GeneralOpenAIRequest) *PaLMChatRequest {
	palmRequest := PaLMChatRequest{
		Prompt: PaLMPrompt{
			Messages: make([]PaLMChatMessage, 0, len(textRequest.Messages)),
		},
		Temperature:    textRequest.Temperature,
		CandidateCount: textRequest.N,
		TopP:           textRequest.TopP,
		TopK:           textRequest.MaxTokens,
	}
	for _, message := range textRequest.Messages {
		palmMessage := PaLMChatMessage{
			Content: message.StringContent(),
		}
		if message.Role == "user" {
			palmMessage.Author = "0"
		} else {
			palmMessage.Author = "1"
		}
		palmRequest.Prompt.Messages = append(palmRequest.Prompt.Messages, palmMessage)
	}
	return &palmRequest
}

func responsePaLM2OpenAI(response *PaLMChatResponse) *OpenAITextResponse {
	fullTextResponse := OpenAITextResponse{
		Choices: make([]OpenAITextResponseChoice, 0, len(response.Candidates)),
	}
	for i, candidate := range response.Candidates {
		content, _ := json.Marshal(candidate.Content)
		choice := OpenAITextResponseChoice{
			Index: i,
			Message: Message{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: "stop",
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	}
	return &fullTextResponse
}

func streamResponsePaLM2OpenAI(palmResponse *PaLMChatResponse) *ChatCompletionsStreamResponse {
	var choice ChatCompletionsStreamResponseChoice
	if len(palmResponse.Candidates) > 0 {
		choice.Delta.Content = palmResponse.Candidates[0].Content
	}
	choice.FinishReason = &stopFinishReason
	var response ChatCompletionsStreamResponse
	response.Object = "chat.completion.chunk"
	response.Model = "palm2"
	response.Choices = []ChatCompletionsStreamResponseChoice{choice}
	return &response
}

func palmStreamHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, string) {
	responseText := ""
	responseId := fmt.Sprintf("chatcmpl-%s", common.GetUUID())
	createdTime := common.GetTimestamp()
	dataChan := make(chan string)
	stopChan := make(chan bool)
	go func() {
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			common.SysError("error reading stream response: " + err.Error())
			stopChan <- true
			return
		}
		err = resp.Body.Close()
		if err != nil {
			common.SysError("error closing stream response: " + err.Error())
			stopChan <- true
			return
		}
		var palmResponse PaLMChatResponse
		err = json.Unmarshal(responseBody, &palmResponse)
		if err != nil {
			common.SysError("error unmarshalling stream response: " + err.Error())
			stopChan <- true
			return
		}
		fullTextResponse := streamResponsePaLM2OpenAI(&palmResponse)
		fullTextResponse.Id = responseId
		fullTextResponse.Created = createdTime
		if len(palmResponse.Candidates) > 0 {
			responseText = palmResponse.Candidates[0].Content
		}
		jsonResponse, err := json.Marshal(fullTextResponse)
		if err != nil {
			common.SysError("error marshalling stream response: " + err.Error())
			stopChan <- true
			return
		}
		dataChan <- string(jsonResponse)
		stopChan <- true
	}()
	setEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			c.Render(-1, common.CustomEvent{Data: "data: " + data})
			return true
		case <-stopChan:
			c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}
	})
	err := resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), ""
	}
	return nil, responseText
}

func palmHandler(c *gin.Context, resp *http.Response, promptTokens int, model string) (*OpenAIErrorWithStatusCode, *Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var palmResponse PaLMChatResponse
	err = json.Unmarshal(responseBody, &palmResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if palmResponse.Error.Code != 0 || len(palmResponse.Candidates) == 0 {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
				Message: palmResponse.Error.Message,
				Type:    palmResponse.Error.Status,
				Param:   "",
				Code:    palmResponse.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responsePaLM2OpenAI(&palmResponse)
	completionTokens := countTokenText(palmResponse.Candidates[0].Content, model)
	usage := Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
	fullTextResponse.Usage = usage
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &usage
}
