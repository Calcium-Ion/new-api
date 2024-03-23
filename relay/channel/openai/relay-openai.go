package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	relayconstant "one-api/relay/constant"
	"one-api/service"
	"strings"
	"sync"
	"time"
)

func OpenaiStreamHandler(c *gin.Context, resp *http.Response, relayMode int) (*dto.OpenAIErrorWithStatusCode, string) {
	checkSensitive := constant.ShouldCheckCompletionSensitive()
	var responseTextBuilder strings.Builder
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
	dataChan := make(chan string, 5)
	stopChan := make(chan bool, 2)
	defer close(stopChan)
	defer close(dataChan)
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		var streamItems []string // store stream items
		for scanner.Scan() {
			data := scanner.Text()
			if len(data) < 6 { // ignore blank line or wrong format
				continue
			}
			if data[:6] != "data: " && data[:6] != "[DONE]" {
				continue
			}
			sensitive := false
			if checkSensitive {
				// check sensitive
				sensitive, _, data = service.SensitiveWordReplace(data, false)
			}
			dataChan <- data
			data = data[6:]
			if !strings.HasPrefix(data, "[DONE]") {
				streamItems = append(streamItems, data)
			}
			if sensitive && constant.StopOnSensitiveEnabled {
				dataChan <- "data: [DONE]"
				break
			}
		}
		streamResp := "[" + strings.Join(streamItems, ",") + "]"
		switch relayMode {
		case relayconstant.RelayModeChatCompletions:
			var streamResponses []dto.ChatCompletionsStreamResponseSimple
			err := json.Unmarshal(common.StringToByteSlice(streamResp), &streamResponses)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				for _, item := range streamItems {
					var streamResponse dto.ChatCompletionsStreamResponseSimple
					err := json.Unmarshal(common.StringToByteSlice(item), &streamResponse)
					if err == nil {
						for _, choice := range streamResponse.Choices {
							responseTextBuilder.WriteString(choice.Delta.Content)
						}
					}
				}
			} else {
				for _, streamResponse := range streamResponses {
					for _, choice := range streamResponse.Choices {
						responseTextBuilder.WriteString(choice.Delta.Content)
					}
				}
			}
		case relayconstant.RelayModeCompletions:
			var streamResponses []dto.CompletionsStreamResponse
			err := json.Unmarshal(common.StringToByteSlice(streamResp), &streamResponses)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				for _, item := range streamItems {
					var streamResponse dto.CompletionsStreamResponse
					err := json.Unmarshal(common.StringToByteSlice(item), &streamResponse)
					if err == nil {
						for _, choice := range streamResponse.Choices {
							responseTextBuilder.WriteString(choice.Text)
						}
					}
				}
			} else {
				for _, streamResponse := range streamResponses {
					for _, choice := range streamResponse.Choices {
						responseTextBuilder.WriteString(choice.Text)
					}
				}
			}
		}
		if len(dataChan) > 0 {
			// wait data out
			time.Sleep(2 * time.Second)
		}
		common.SafeSend(stopChan, true)
	}()
	service.SetEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			if strings.HasPrefix(data, "data: [DONE]") {
				data = data[:12]
			}
			// some implementations may add \r at the end of data
			data = strings.TrimSuffix(data, "\r")
			c.Render(-1, common.CustomEvent{Data: data})
			return true
		case <-stopChan:
			return false
		}
	})
	err := resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), ""
	}
	wg.Wait()
	return nil, responseTextBuilder.String()
}

func OpenaiHandler(c *gin.Context, resp *http.Response, promptTokens int, model string, relayMode int) (*dto.OpenAIErrorWithStatusCode, *dto.Usage, *dto.SensitiveResponse) {
	var responseWithError dto.TextResponseWithError
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil, nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil, nil
	}
	err = json.Unmarshal(responseBody, &responseWithError)
	if err != nil {
		log.Printf("unmarshal_response_body_failed: body: %s, err: %v", string(responseBody), err)
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil, nil
	}
	if responseWithError.Error.Type != "" {
		return &dto.OpenAIErrorWithStatusCode{
			Error:      responseWithError.Error,
			StatusCode: resp.StatusCode,
		}, nil, nil
	}

	checkSensitive := constant.ShouldCheckCompletionSensitive()
	sensitiveWords := make([]string, 0)
	triggerSensitive := false

	usage := &responseWithError.Usage

	//textResponse := &dto.TextResponse{
	//	Choices: responseWithError.Choices,
	//	Usage:   responseWithError.Usage,
	//}
	var doResponseBody []byte

	switch relayMode {
	case relayconstant.RelayModeEmbeddings:
		embeddingResponse := &dto.OpenAIEmbeddingResponse{
			Object: responseWithError.Object,
			Data:   responseWithError.Data,
			Model:  responseWithError.Model,
			Usage:  *usage,
		}
		doResponseBody, err = json.Marshal(embeddingResponse)
	default:
		if responseWithError.Usage.TotalTokens == 0 || checkSensitive {
			completionTokens := 0
			for i, choice := range responseWithError.Choices {
				stringContent := string(choice.Message.Content)
				ctkm, _, _ := service.CountTokenText(stringContent, model, false)
				completionTokens += ctkm
				if checkSensitive {
					sensitive, words, stringContent := service.SensitiveWordReplace(stringContent, false)
					if sensitive {
						triggerSensitive = true
						msg := choice.Message
						msg.Content = common.StringToByteSlice(stringContent)
						responseWithError.Choices[i].Message = msg
						sensitiveWords = append(sensitiveWords, words...)
					}
				}
			}
			responseWithError.Usage = dto.Usage{
				PromptTokens:     promptTokens,
				CompletionTokens: completionTokens,
				TotalTokens:      promptTokens + completionTokens,
			}
		}
		textResponse := &dto.TextResponse{
			Id:      responseWithError.Id,
			Created: responseWithError.Created,
			Object:  responseWithError.Object,
			Choices: responseWithError.Choices,
			Model:   responseWithError.Model,
			Usage:   *usage,
		}
		doResponseBody, err = json.Marshal(textResponse)
	}

	if checkSensitive && triggerSensitive && constant.StopOnSensitiveEnabled {
		sensitiveWords = common.RemoveDuplicate(sensitiveWords)
		return service.OpenAIErrorWrapper(errors.New(fmt.Sprintf("sensitive words detected on response: %s",
				strings.Join(sensitiveWords, ", "))), "sensitive_words_detected", http.StatusBadRequest),
			usage, &dto.SensitiveResponse{
				SensitiveWords: sensitiveWords,
			}
	} else {
		// Reset response body
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
			return service.OpenAIErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil, nil
		}
		err = resp.Body.Close()
		if err != nil {
			return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil, nil
		}
	}
	return nil, usage, nil
}
