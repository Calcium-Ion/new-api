package controller

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"strings"
	"sync"
	"time"
)

// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/flfmc9do2

type BaiduTokenResponse struct {
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

type BaiduMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type BaiduChatRequest struct {
	Messages []BaiduMessage `json:"messages"`
	Stream   bool           `json:"stream"`
	UserId   string         `json:"user_id,omitempty"`
}

type BaiduError struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

type BaiduChatResponse struct {
	Id               string `json:"id"`
	Object           string `json:"object"`
	Created          int64  `json:"created"`
	Result           string `json:"result"`
	IsTruncated      bool   `json:"is_truncated"`
	NeedClearHistory bool   `json:"need_clear_history"`
	Usage            Usage  `json:"usage"`
	BaiduError
}

type BaiduChatStreamResponse struct {
	BaiduChatResponse
	SentenceId int  `json:"sentence_id"`
	IsEnd      bool `json:"is_end"`
}

type BaiduEmbeddingRequest struct {
	Input []string `json:"input"`
}

type BaiduEmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type BaiduEmbeddingResponse struct {
	Id      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Data    []BaiduEmbeddingData `json:"data"`
	Usage   Usage                `json:"usage"`
	BaiduError
}

type BaiduAccessToken struct {
	AccessToken      string    `json:"access_token"`
	Error            string    `json:"error,omitempty"`
	ErrorDescription string    `json:"error_description,omitempty"`
	ExpiresIn        int64     `json:"expires_in,omitempty"`
	ExpiresAt        time.Time `json:"-"`
}

var baiduTokenStore sync.Map

func requestOpenAI2Baidu(request GeneralOpenAIRequest) *BaiduChatRequest {
	messages := make([]BaiduMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		if message.Role == "system" {
			messages = append(messages, BaiduMessage{
				Role:    "user",
				Content: message.StringContent(),
			})
			messages = append(messages, BaiduMessage{
				Role:    "assistant",
				Content: "Okay",
			})
		} else {
			messages = append(messages, BaiduMessage{
				Role:    message.Role,
				Content: message.StringContent(),
			})
		}
	}
	return &BaiduChatRequest{
		Messages: messages,
		Stream:   request.Stream,
	}
}

func responseBaidu2OpenAI(response *BaiduChatResponse) *OpenAITextResponse {
	content, _ := json.Marshal(response.Result)
	choice := OpenAITextResponseChoice{
		Index: 0,
		Message: Message{
			Role:    "assistant",
			Content: content,
		},
		FinishReason: "stop",
	}
	fullTextResponse := OpenAITextResponse{
		Id:      response.Id,
		Object:  "chat.completion",
		Created: response.Created,
		Choices: []OpenAITextResponseChoice{choice},
		Usage:   response.Usage,
	}
	return &fullTextResponse
}

func streamResponseBaidu2OpenAI(baiduResponse *BaiduChatStreamResponse) *ChatCompletionsStreamResponse {
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = baiduResponse.Result
	if baiduResponse.IsEnd {
		choice.FinishReason = &stopFinishReason
	}
	response := ChatCompletionsStreamResponse{
		Id:      baiduResponse.Id,
		Object:  "chat.completion.chunk",
		Created: baiduResponse.Created,
		Model:   "ernie-bot",
		Choices: []ChatCompletionsStreamResponseChoice{choice},
	}
	return &response
}

func embeddingRequestOpenAI2Baidu(request GeneralOpenAIRequest) *BaiduEmbeddingRequest {
	return &BaiduEmbeddingRequest{
		Input: request.ParseInput(),
	}
}

func embeddingResponseBaidu2OpenAI(response *BaiduEmbeddingResponse) *OpenAIEmbeddingResponse {
	openAIEmbeddingResponse := OpenAIEmbeddingResponse{
		Object: "list",
		Data:   make([]OpenAIEmbeddingResponseItem, 0, len(response.Data)),
		Model:  "baidu-embedding",
		Usage:  response.Usage,
	}
	for _, item := range response.Data {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, OpenAIEmbeddingResponseItem{
			Object:    item.Object,
			Index:     item.Index,
			Embedding: item.Embedding,
		})
	}
	return &openAIEmbeddingResponse
}

func baiduStreamHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
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
			if len(data) < 6 { // ignore blank line or wrong format
				continue
			}
			data = data[6:]
			dataChan <- data
		}
		stopChan <- true
	}()
	setEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			var baiduResponse BaiduChatStreamResponse
			err := json.Unmarshal([]byte(data), &baiduResponse)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			if baiduResponse.Usage.TotalTokens != 0 {
				usage.TotalTokens = baiduResponse.Usage.TotalTokens
				usage.PromptTokens = baiduResponse.Usage.PromptTokens
				usage.CompletionTokens = baiduResponse.Usage.TotalTokens - baiduResponse.Usage.PromptTokens
			}
			response := streamResponseBaidu2OpenAI(&baiduResponse)
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

func baiduHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var baiduResponse BaiduChatResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &baiduResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if baiduResponse.ErrorMsg != "" {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
				Message: baiduResponse.ErrorMsg,
				Type:    "baidu_error",
				Param:   "",
				Code:    baiduResponse.ErrorCode,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseBaidu2OpenAI(&baiduResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func baiduEmbeddingHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var baiduResponse BaiduEmbeddingResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &baiduResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if baiduResponse.ErrorMsg != "" {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
				Message: baiduResponse.ErrorMsg,
				Type:    "baidu_error",
				Param:   "",
				Code:    baiduResponse.ErrorCode,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := embeddingResponseBaidu2OpenAI(&baiduResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func getBaiduAccessToken(apiKey string) (string, error) {
	if val, ok := baiduTokenStore.Load(apiKey); ok {
		var accessToken BaiduAccessToken
		if accessToken, ok = val.(BaiduAccessToken); ok {
			// soon this will expire
			if time.Now().Add(time.Hour).After(accessToken.ExpiresAt) {
				go func() {
					_, _ = getBaiduAccessTokenHelper(apiKey)
				}()
			}
			return accessToken.AccessToken, nil
		}
	}
	accessToken, err := getBaiduAccessTokenHelper(apiKey)
	if err != nil {
		return "", err
	}
	if accessToken == nil {
		return "", errors.New("getBaiduAccessToken return a nil token")
	}
	return (*accessToken).AccessToken, nil
}

func getBaiduAccessTokenHelper(apiKey string) (*BaiduAccessToken, error) {
	parts := strings.Split(apiKey, "|")
	if len(parts) != 2 {
		return nil, errors.New("invalid baidu apikey")
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=%s&client_secret=%s",
		parts[0], parts[1]), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	res, err := impatientHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var accessToken BaiduAccessToken
	err = json.NewDecoder(res.Body).Decode(&accessToken)
	if err != nil {
		return nil, err
	}
	if accessToken.Error != "" {
		return nil, errors.New(accessToken.Error + ": " + accessToken.ErrorDescription)
	}
	if accessToken.AccessToken == "" {
		return nil, errors.New("getBaiduAccessTokenHelper get empty access token")
	}
	accessToken.ExpiresAt = time.Now().Add(time.Duration(accessToken.ExpiresIn) * time.Second)
	baiduTokenStore.Store(apiKey, accessToken)
	return &accessToken, nil
}
