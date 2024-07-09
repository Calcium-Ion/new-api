package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	relayconstant "one-api/relay/constant"
	"one-api/service"
	"strings"
	"sync"
	"time"
)

func OpenaiStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage, string, int) {
	//checkSensitive := constant.ShouldCheckCompletionSensitive()
	var responseTextBuilder strings.Builder
	var usage dto.Usage
	toolCount := 0
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
			if !common.SafeSendStringTimeout(dataChan, data, constant.StreamingTimeout) {
				// send data timeout, stop the stream
				common.LogError(c, "send data timeout, stop the stream")
				break
			}
			data = data[6:]
			if !strings.HasPrefix(data, "[DONE]") {
				streamItems = append(streamItems, data)
			}
		}
		// 计算token
		streamResp := "[" + strings.Join(streamItems, ",") + "]"
		switch info.RelayMode {
		case relayconstant.RelayModeChatCompletions:
			var streamResponses []dto.ChatCompletionsStreamResponseSimple
			err := json.Unmarshal(common.StringToByteSlice(streamResp), &streamResponses)
			if err != nil {
				// 一次性解析失败，逐个解析
				common.SysError("error unmarshalling stream response: " + err.Error())
				for _, item := range streamItems {
					var streamResponse dto.ChatCompletionsStreamResponseSimple
					err := json.Unmarshal(common.StringToByteSlice(item), &streamResponse)
					if err == nil {
						if streamResponse.Usage != nil {
							if streamResponse.Usage.TotalTokens != 0 {
								usage = *streamResponse.Usage
							}
						}
						for _, choice := range streamResponse.Choices {
							responseTextBuilder.WriteString(choice.Delta.GetContentString())
							if choice.Delta.ToolCalls != nil {
								if len(choice.Delta.ToolCalls) > toolCount {
									toolCount = len(choice.Delta.ToolCalls)
								}
								for _, tool := range choice.Delta.ToolCalls {
									responseTextBuilder.WriteString(tool.Function.Name)
									responseTextBuilder.WriteString(tool.Function.Arguments)
								}
							}
						}
					}
				}
			} else {
				for _, streamResponse := range streamResponses {
					if streamResponse.Usage != nil {
						if streamResponse.Usage.TotalTokens != 0 {
							usage = *streamResponse.Usage
						}
					}
					for _, choice := range streamResponse.Choices {
						responseTextBuilder.WriteString(choice.Delta.GetContentString())
						if choice.Delta.ToolCalls != nil {
							if len(choice.Delta.ToolCalls) > toolCount {
								toolCount = len(choice.Delta.ToolCalls)
							}
							for _, tool := range choice.Delta.ToolCalls {
								responseTextBuilder.WriteString(tool.Function.Name)
								responseTextBuilder.WriteString(tool.Function.Arguments)
							}
						}
					}
				}
			}
		case relayconstant.RelayModeCompletions:
			var streamResponses []dto.CompletionsStreamResponse
			err := json.Unmarshal(common.StringToByteSlice(streamResp), &streamResponses)
			if err != nil {
				// 一次性解析失败，逐个解析
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
		common.SafeSendBool(stopChan, true)
	}()
	service.SetEventStreamHeaders(c)
	isFirst := true
	ticker := time.NewTicker(time.Duration(constant.StreamingTimeout) * time.Second)
	defer ticker.Stop()
	c.Stream(func(w io.Writer) bool {
		select {
		case <-ticker.C:
			common.LogError(c, "reading data from upstream timeout")
			return false
		case data := <-dataChan:
			if isFirst {
				isFirst = false
				info.FirstResponseTime = time.Now()
			}
			ticker.Reset(time.Duration(constant.StreamingTimeout) * time.Second)
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
		return service.OpenAIErrorWrapperLocal(err, "close_response_body_failed", http.StatusInternalServerError), nil, "", toolCount
	}
	wg.Wait()
	return nil, &usage, responseTextBuilder.String(), toolCount
}

func OpenaiHandler(c *gin.Context, resp *http.Response, promptTokens int, model string) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var simpleResponse dto.SimpleResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &simpleResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if simpleResponse.Error.Type != "" {
		return &dto.OpenAIErrorWithStatusCode{
			Error:      simpleResponse.Error,
			StatusCode: resp.StatusCode,
		}, nil
	}
	// Reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the httpClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	if simpleResponse.Usage.TotalTokens == 0 || (simpleResponse.Usage.PromptTokens == 0 && simpleResponse.Usage.CompletionTokens == 0) {
		completionTokens := 0
		for _, choice := range simpleResponse.Choices {
			ctkm, _ := service.CountTokenText(string(choice.Message.Content), model)
			completionTokens += ctkm
		}
		simpleResponse.Usage = dto.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		}
	}
	return nil, &simpleResponse.Usage
}
