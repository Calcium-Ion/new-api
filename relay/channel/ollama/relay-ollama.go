package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/dto"
	"one-api/service"
	"strings"
)

func requestOpenAI2Ollama(request dto.GeneralOpenAIRequest) *OllamaRequest {
	messages := make([]dto.Message, 0, len(request.Messages))
	for _, message := range request.Messages {
		messages = append(messages, dto.Message{
			Role:    message.Role,
			Content: message.Content,
		})
	}
	str, ok := request.Stop.(string)
	var Stop []string
	if ok {
		Stop = []string{str}
	} else {
		Stop, _ = request.Stop.([]string)
	}
	return &OllamaRequest{
		Model:       request.Model,
		Messages:    messages,
		Stream:      request.Stream,
		Temperature: request.Temperature,
		Seed:        request.Seed,
		Topp:        request.TopP,
		TopK:        request.TopK,
		Stop:        Stop,
	}
}

func requestOpenAI2Embeddings(request dto.GeneralOpenAIRequest) *OllamaEmbeddingRequest {
	return &OllamaEmbeddingRequest{
		Model:  request.Model,
		Prompt: strings.Join(request.ParseInput(), " "),
	}
}

func ollamaEmbeddingHandler(c *gin.Context, resp *http.Response, promptTokens int, model string, relayMode int) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var ollamaEmbeddingResponse OllamaEmbeddingResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &ollamaEmbeddingResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	data := make([]dto.OpenAIEmbeddingResponseItem, 0, 1)
	data = append(data, dto.OpenAIEmbeddingResponseItem{
		Embedding: ollamaEmbeddingResponse.Embedding,
		Object:    "embedding",
	})
	usage := &dto.Usage{
		TotalTokens:      promptTokens,
		CompletionTokens: 0,
		PromptTokens:     promptTokens,
	}
	embeddingResponse := &dto.OpenAIEmbeddingResponse{
		Object: "list",
		Data:   data,
		Model:  model,
		Usage:  *usage,
	}
	doResponseBody, err := json.Marshal(embeddingResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(doResponseBody))
	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the httpClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	// Copy headers
	for k, v := range resp.Header {
		// 删除任何现有的相同头部，以防止重复添加头部
		c.Writer.Header().Del(k)
		for _, vv := range v {
			c.Writer.Header().Add(k, vv)
		}
	}
	// reset content length
	c.Writer.Header().Del("Content-Length")
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(doResponseBody)))
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, usage
}
