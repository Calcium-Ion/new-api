package mokaai

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/dto"
	"one-api/service"
)

func embeddingRequestOpenAI2Moka(request dto.GeneralOpenAIRequest) *dto.EmbeddingRequest {
	var input []string // Change input to []string

	switch v := request.Input.(type) {
	case string:
		input = []string{v} // Convert string to []string
	case []string:
		input = v // Already a []string, no conversion needed
	case []interface{}:
		for _, part := range v {
			if str, ok := part.(string); ok {
				input = append(input, str) // Append each string to the slice
			}
		}
	}
	return &dto.EmbeddingRequest{
		Input: input,
		Model:  request.Model,
	}
}

func embeddingResponseMoka2OpenAI(response *dto.EmbeddingResponse) *dto.OpenAIEmbeddingResponse {
	openAIEmbeddingResponse := dto.OpenAIEmbeddingResponse{
		Object: "list",
		Data:   make([]dto.OpenAIEmbeddingResponseItem, 0, len(response.Data)),
		Model:  "baidu-embedding",
		Usage:  response.Usage,
	}
	for _, item := range response.Data {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, dto.OpenAIEmbeddingResponseItem{
			Object:    item.Object,
			Index:     item.Index,
			Embedding: item.Embedding,
		})
	}
	return &openAIEmbeddingResponse
}

func mokaEmbeddingHandler(c *gin.Context, resp *http.Response) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var baiduResponse dto.EmbeddingResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &baiduResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	// if baiduResponse.ErrorMsg != "" {
	// 	return &dto.OpenAIErrorWithStatusCode{
	// 		Error: dto.OpenAIError{
	// 			Type:    "baidu_error",
	// 			Param:   "",
	// 		},
	// 		StatusCode: resp.StatusCode,
	// 	}, nil
	// }
	fullTextResponse := embeddingResponseMoka2OpenAI(&baiduResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

