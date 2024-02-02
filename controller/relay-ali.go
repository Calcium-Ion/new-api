package controller

import (
	"bufio"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"strings"
)

// https://help.aliyun.com/document_detail/613695.html?spm=a2c4g.2399480.0.0.1adb778fAdzP9w#341800c0f8w0r

type AliMessage struct {
	User string `json:"user"`
	Bot  string `json:"bot"`
}

type AliInput struct {
	Prompt  string       `json:"prompt"`
	History []AliMessage `json:"history"`
}

type AliParameters struct {
	TopP         float64 `json:"top_p,omitempty"`
	TopK         int     `json:"top_k,omitempty"`
	Seed         uint64  `json:"seed,omitempty"`
	EnableSearch bool    `json:"enable_search,omitempty"`
}

type AliChatRequest struct {
	Model      string        `json:"model"`
	Input      AliInput      `json:"input"`
	Parameters AliParameters `json:"parameters,omitempty"`
}

type AliEmbeddingRequest struct {
	Model string `json:"model"`
	Input struct {
		Texts []string `json:"texts"`
	} `json:"input"`
	Parameters *struct {
		TextType string `json:"text_type,omitempty"`
	} `json:"parameters,omitempty"`
}

type AliEmbedding struct {
	Embedding []float64 `json:"embedding"`
	TextIndex int       `json:"text_index"`
}

type AliEmbeddingResponse struct {
	Output struct {
		Embeddings []AliEmbedding `json:"embeddings"`
	} `json:"output"`
	Usage AliUsage `json:"usage"`
	AliError
}

type AliError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

type AliUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type AliOutput struct {
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason"`
}

type AliChatResponse struct {
	Output AliOutput `json:"output"`
	Usage  AliUsage  `json:"usage"`
	AliError
}

func requestOpenAI2Ali(request GeneralOpenAIRequest) *AliChatRequest {
	messages := make([]AliMessage, 0, len(request.Messages))
	prompt := ""
	for i := 0; i < len(request.Messages); i++ {
		message := request.Messages[i]
		if message.Role == "system" {
			messages = append(messages, AliMessage{
				User: message.StringContent(),
				Bot:  "Okay",
			})
			continue
		} else {
			if i == len(request.Messages)-1 {
				prompt = message.StringContent()
				break
			}
			messages = append(messages, AliMessage{
				User: message.StringContent(),
				Bot:  string(request.Messages[i+1].Content),
			})
			i++
		}
	}
	return &AliChatRequest{
		Model: request.Model,
		Input: AliInput{
			Prompt:  prompt,
			History: messages,
		},
		//Parameters: AliParameters{  // ChatGPT's parameters are not compatible with Ali's
		//	TopP: request.TopP,
		//	TopK: 50,
		//	//Seed:         0,
		//	//EnableSearch: false,
		//},
	}
}

func embeddingRequestOpenAI2Ali(request GeneralOpenAIRequest) *AliEmbeddingRequest {
	return &AliEmbeddingRequest{
		Model: "text-embedding-v1",
		Input: struct {
			Texts []string `json:"texts"`
		}{
			Texts: request.ParseInput(),
		},
	}
}

func aliEmbeddingHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var aliResponse AliEmbeddingResponse
	err := json.NewDecoder(resp.Body).Decode(&aliResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	if aliResponse.Code != "" {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
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
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func embeddingResponseAli2OpenAI(response *AliEmbeddingResponse) *OpenAIEmbeddingResponse {
	openAIEmbeddingResponse := OpenAIEmbeddingResponse{
		Object: "list",
		Data:   make([]OpenAIEmbeddingResponseItem, 0, len(response.Output.Embeddings)),
		Model:  "text-embedding-v1",
		Usage:  Usage{TotalTokens: response.Usage.TotalTokens},
	}

	for _, item := range response.Output.Embeddings {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, OpenAIEmbeddingResponseItem{
			Object:    `embedding`,
			Index:     item.TextIndex,
			Embedding: item.Embedding,
		})
	}
	return &openAIEmbeddingResponse
}

func responseAli2OpenAI(response *AliChatResponse) *OpenAITextResponse {
	content, _ := json.Marshal(response.Output.Text)
	choice := OpenAITextResponseChoice{
		Index: 0,
		Message: Message{
			Role:    "assistant",
			Content: content,
		},
		FinishReason: response.Output.FinishReason,
	}
	fullTextResponse := OpenAITextResponse{
		Id:      response.RequestId,
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: []OpenAITextResponseChoice{choice},
		Usage: Usage{
			PromptTokens:     response.Usage.InputTokens,
			CompletionTokens: response.Usage.OutputTokens,
			TotalTokens:      response.Usage.InputTokens + response.Usage.OutputTokens,
		},
	}
	return &fullTextResponse
}

func streamResponseAli2OpenAI(aliResponse *AliChatResponse) *ChatCompletionsStreamResponse {
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = aliResponse.Output.Text
	if aliResponse.Output.FinishReason != "null" {
		finishReason := aliResponse.Output.FinishReason
		choice.FinishReason = &finishReason
	}
	response := ChatCompletionsStreamResponse{
		Id:      aliResponse.RequestId,
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "ernie-bot",
		Choices: []ChatCompletionsStreamResponseChoice{choice},
	}
	return &response
}

func aliStreamHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var usage Usage
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
	setEventStreamHeaders(c)
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
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, &usage
}

func aliHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var aliResponse AliChatResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &aliResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if aliResponse.Code != "" {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
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
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}
