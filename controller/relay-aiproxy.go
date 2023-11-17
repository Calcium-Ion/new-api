package controller

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"strconv"
	"strings"
)

// https://docs.aiproxy.io/dev/library#使用已经定制好的知识库进行对话问答

type AIProxyLibraryRequest struct {
	Model     string `json:"model"`
	Query     string `json:"query"`
	LibraryId string `json:"libraryId"`
	Stream    bool   `json:"stream"`
}

type AIProxyLibraryError struct {
	ErrCode int    `json:"errCode"`
	Message string `json:"message"`
}

type AIProxyLibraryDocument struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type AIProxyLibraryResponse struct {
	Success   bool                     `json:"success"`
	Answer    string                   `json:"answer"`
	Documents []AIProxyLibraryDocument `json:"documents"`
	AIProxyLibraryError
}

type AIProxyLibraryStreamResponse struct {
	Content   string                   `json:"content"`
	Finish    bool                     `json:"finish"`
	Model     string                   `json:"model"`
	Documents []AIProxyLibraryDocument `json:"documents"`
}

func requestOpenAI2AIProxyLibrary(request GeneralOpenAIRequest) *AIProxyLibraryRequest {
	query := ""
	if len(request.Messages) != 0 {
		query = string(request.Messages[len(request.Messages)-1].Content)
	}
	return &AIProxyLibraryRequest{
		Model:  request.Model,
		Stream: request.Stream,
		Query:  query,
	}
}

func aiProxyDocuments2Markdown(documents []AIProxyLibraryDocument) string {
	if len(documents) == 0 {
		return ""
	}
	content := "\n\n参考文档：\n"
	for i, document := range documents {
		content += fmt.Sprintf("%d. [%s](%s)\n", i+1, document.Title, document.URL)
	}
	return content
}

func responseAIProxyLibrary2OpenAI(response *AIProxyLibraryResponse) *OpenAITextResponse {
	content, _ := json.Marshal(response.Answer + aiProxyDocuments2Markdown(response.Documents))
	choice := OpenAITextResponseChoice{
		Index: 0,
		Message: Message{
			Role:    "assistant",
			Content: content,
		},
		FinishReason: "stop",
	}
	fullTextResponse := OpenAITextResponse{
		Id:      common.GetUUID(),
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: []OpenAITextResponseChoice{choice},
	}
	return &fullTextResponse
}

func documentsAIProxyLibrary(documents []AIProxyLibraryDocument) *ChatCompletionsStreamResponse {
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = aiProxyDocuments2Markdown(documents)
	choice.FinishReason = &stopFinishReason
	return &ChatCompletionsStreamResponse{
		Id:      common.GetUUID(),
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "",
		Choices: []ChatCompletionsStreamResponseChoice{choice},
	}
}

func streamResponseAIProxyLibrary2OpenAI(response *AIProxyLibraryStreamResponse) *ChatCompletionsStreamResponse {
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = response.Content
	return &ChatCompletionsStreamResponse{
		Id:      common.GetUUID(),
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   response.Model,
		Choices: []ChatCompletionsStreamResponseChoice{choice},
	}
}

func aiProxyLibraryStreamHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
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
	var documents []AIProxyLibraryDocument
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			var AIProxyLibraryResponse AIProxyLibraryStreamResponse
			err := json.Unmarshal([]byte(data), &AIProxyLibraryResponse)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			if len(AIProxyLibraryResponse.Documents) != 0 {
				documents = AIProxyLibraryResponse.Documents
			}
			response := streamResponseAIProxyLibrary2OpenAI(&AIProxyLibraryResponse)
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonResponse)})
			return true
		case <-stopChan:
			response := documentsAIProxyLibrary(documents)
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonResponse)})
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

func aiProxyLibraryHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var AIProxyLibraryResponse AIProxyLibraryResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &AIProxyLibraryResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if AIProxyLibraryResponse.ErrCode != 0 {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
				Message: AIProxyLibraryResponse.Message,
				Type:    strconv.Itoa(AIProxyLibraryResponse.ErrCode),
				Code:    AIProxyLibraryResponse.ErrCode,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseAIProxyLibrary2OpenAI(&AIProxyLibraryResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}
