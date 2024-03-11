package ali

import (
	"bufio"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	"one-api/service"
	"strings"
)

// https://help.aliyun.com/document_detail/613695.html?spm=a2c4g.2399480.0.0.1adb778fAdzP9w#341800c0f8w0r

const EnableSearchModelSuffix = "-internet"

func requestOpenAI2Ali(request dto.GeneralOpenAIRequest) *AliChatRequest {
	messages := make([]AliMessage, 0, len(request.Messages))
	//prompt := ""
	for i := 0; i < len(request.Messages); i++ {
		message := request.Messages[i]
		messages = append(messages, AliMessage{
			Content: message.StringContent(),
			Role:    strings.ToLower(message.Role),
		})
	}
	enableSearch := false
	aliModel := request.Model
	if strings.HasSuffix(aliModel, EnableSearchModelSuffix) {
		enableSearch = true
		aliModel = strings.TrimSuffix(aliModel, EnableSearchModelSuffix)
	}
	return &AliChatRequest{
		Model: request.Model,
		Input: AliInput{
			//Prompt:  prompt,
			Messages: messages,
		},
		Parameters: AliParameters{
			IncrementalOutput: request.Stream,
			Seed:              uint64(request.Seed),
			EnableSearch:      enableSearch,
		},
	}
}

func embeddingRequestOpenAI2Ali(request dto.GeneralOpenAIRequest) *AliEmbeddingRequest {
	return &AliEmbeddingRequest{
		Model: "text-embedding-v1",
		Input: struct {
			Texts []string `json:"texts"`
		}{
			Texts: request.ParseInput(),
		},
	}
}

func aliEmbeddingHandler(c *gin.Context, resp *http.Response) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var aliResponse AliEmbeddingResponse
	err := json.NewDecoder(resp.Body).Decode(&aliResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	if aliResponse.Code != "" {
		return &dto.OpenAIErrorWithStatusCode{
			Error: dto.OpenAIError{
				Message: aliResponse.Message,
				Type:    aliResponse.Code,
				Param:   aliResponse.RequestId,
				Code:    aliResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	fullTextResponse := embeddingResponseAli2OpenAI(&aliResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func embeddingResponseAli2OpenAI(response *AliEmbeddingResponse) *dto.OpenAIEmbeddingResponse {
	openAIEmbeddingResponse := dto.OpenAIEmbeddingResponse{
		Object: "list",
		Data:   make([]dto.OpenAIEmbeddingResponseItem, 0, len(response.Output.Embeddings)),
		Model:  "text-embedding-v1",
		Usage:  dto.Usage{TotalTokens: response.Usage.TotalTokens},
	}

	for _, item := range response.Output.Embeddings {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, dto.OpenAIEmbeddingResponseItem{
			Object:    `embedding`,
			Index:     item.TextIndex,
			Embedding: item.Embedding,
		})
	}
	return &openAIEmbeddingResponse
}

func responseAli2OpenAI(response *AliChatResponse) *dto.OpenAITextResponse {
	content, _ := json.Marshal(response.Output.Text)
	choice := dto.OpenAITextResponseChoice{
		Index: 0,
		Message: dto.Message{
			Role:    "assistant",
			Content: content,
		},
		FinishReason: response.Output.FinishReason,
	}
	fullTextResponse := dto.OpenAITextResponse{
		Id:      response.RequestId,
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: []dto.OpenAITextResponseChoice{choice},
		Usage: dto.Usage{
			PromptTokens:     response.Usage.InputTokens,
			CompletionTokens: response.Usage.OutputTokens,
			TotalTokens:      response.Usage.InputTokens + response.Usage.OutputTokens,
		},
	}
	return &fullTextResponse
}

func streamResponseAli2OpenAI(aliResponse *AliChatResponse) *dto.ChatCompletionsStreamResponse {
	var choice dto.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = aliResponse.Output.Text
	if aliResponse.Output.FinishReason != "null" {
		finishReason := aliResponse.Output.FinishReason
		choice.FinishReason = &finishReason
	}
	response := dto.ChatCompletionsStreamResponse{
		Id:      aliResponse.RequestId,
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "ernie-bot",
		Choices: []dto.ChatCompletionsStreamResponseChoice{choice},
	}
	return &response
}

func aliStreamHandler(c *gin.Context, resp *http.Response) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var usage dto.Usage
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
			if len(data) < 5 { // ignore blank line or wrong format
				continue
			}
			if data[:5] != "data:" {
				continue
			}
			data = data[5:]
			dataChan <- data
		}
		stopChan <- true
	}()
	service.SetEventStreamHeaders(c)
	lastResponseText := ""
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			var aliResponse AliChatResponse
			err := json.Unmarshal([]byte(data), &aliResponse)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			if aliResponse.Usage.OutputTokens != 0 {
				usage.PromptTokens = aliResponse.Usage.InputTokens
				usage.CompletionTokens = aliResponse.Usage.OutputTokens
				usage.TotalTokens = aliResponse.Usage.InputTokens + aliResponse.Usage.OutputTokens
			}
			response := streamResponseAli2OpenAI(&aliResponse)
			response.Choices[0].Delta.Content = strings.TrimPrefix(response.Choices[0].Delta.Content, lastResponseText)
			lastResponseText = aliResponse.Output.Text
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonResponse)})
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
	return nil, &usage
}

func aliHandler(c *gin.Context, resp *http.Response) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var aliResponse AliChatResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &aliResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if aliResponse.Code != "" {
		return &dto.OpenAIErrorWithStatusCode{
			Error: dto.OpenAIError{
				Message: aliResponse.Message,
				Type:    aliResponse.Code,
				Param:   aliResponse.RequestId,
				Code:    aliResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseAli2OpenAI(&aliResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}
