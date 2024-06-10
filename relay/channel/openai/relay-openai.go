package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	relayconstant "one-api/relay/constant"
	"one-api/service"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var modelmapper = map[string]string{
	"gpt-4-turbo":         "gpt-4-turbo-2024-04-09",
	"gpt-4":               "gpt-4-0613",
	"gpt-3.5-turbo":       "gpt-3.5-turbo-0125",
	"gpt-3.5-turbo-16k":   "gpt-3.5-turbo-16k-0613",
	"gpt-4-32k":           "gpt-4-32k-0613",
	"gpt-4-turbo-preview": "gpt-4-0125-preview",
}

func OpenaiStreamHandler(c *gin.Context, resp *http.Response, relayMode int, model string) (*dto.OpenAIErrorWithStatusCode, string, int) {
	//checkSensitive := constant.ShouldCheckCompletionSensitive()
	modelName := model
	if v, ok := modelmapper[model]; ok {
		fmt.Println("modelName is in modelmapper change to ", v)
		modelName = v
	}
	var responseTextBuilder strings.Builder
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

			data = data[6:]
			if strings.HasPrefix(data, "[DONE]") {
				common.SafeSendString(dataChan, "data: [DONE]")
				continue
			}
			var jsonData map[string]interface{}
			err := json.Unmarshal([]byte(data), &jsonData)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				continue
			}
			if _, ok := jsonData["model"]; ok {
				jsonData["model"] = modelName
			}
			if choices, ok := jsonData["choices"].([]interface{}); ok {
				for _, choice := range choices {
					if choiceMap, ok := choice.(map[string]interface{}); ok {
						delete(choiceMap, "content_filter_results")
					}
				}
			}
			delete(jsonData, "prompt_filter_results")
			modifiedData, err := json.Marshal(jsonData)
			if err != nil {
				common.SysError("error marshalling modified response: " + err.Error())
				continue
			}
			streamItems = append(streamItems, string(modifiedData))
			common.SafeSendString(dataChan, "data: "+string(modifiedData))
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
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), "", toolCount
	}
	wg.Wait()
	return nil, responseTextBuilder.String(), toolCount
}

// func OpenaiStreamHandler(c *gin.Context, resp *http.Response, relayMode int, model string) (*dto.OpenAIErrorWithStatusCode, string, int) {
// 	modelName := model
// 	if v, ok := modelmapper[model]; ok {
// 		fmt.Println("modelName is in modelmapper change to ", v)
// 		modelName = v
// 	}
// 	var responseTextBuilder strings.Builder
// 	toolCount := 0
// 	scanner := bufio.NewScanner(resp.Body)
// 	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
// 		if atEOF && len(data) == 0 {
// 			return 0, nil, nil
// 		}
// 		if i := strings.Index(string(data), "\n"); i >= 0 {
// 			return i + 1, data[0:i], nil
// 		}
// 		if atEOF {
// 			return len(data), data, nil
// 		}
// 		return 0, nil, nil
// 	})
// 	dataChan := make(chan string, 5)
// 	stopChan := make(chan bool, 2)
// 	defer close(stopChan)
// 	defer close(dataChan)
// 	var wg sync.WaitGroup
// 	go func() {
// 		wg.Add(1)
// 		defer wg.Done()
// 		for scanner.Scan() {
// 			data := scanner.Text()
// 			if len(data) < 6 || (data[:6] != "data: " && data[:6] != "[DONE]") {
// 				continue
// 			}
// 			data = data[6:]
// 			if strings.HasPrefix(data, "[DONE]") {
// 				common.SafeSendString(dataChan, "data: [DONE]")
// 				continue
// 			}
// 			var jsonData map[string]interface{}
// 			err := json.Unmarshal([]byte(data), &jsonData)
// 			if err != nil {
// 				common.SysError("error unmarshalling stream response: " + err.Error())
// 				continue
// 			}
// 			if _, ok := jsonData["model"]; ok {
// 				jsonData["model"] = modelName
// 			}
// 			if choices, ok := jsonData["choices"].([]interface{}); ok {
// 				for _, choice := range choices {
// 					if choiceMap, ok := choice.(map[string]interface{}); ok {
// 						delete(choiceMap, "content_filter_results")
// 					}
// 				}
// 			}
// 			delete(jsonData, "prompt_filter_results")
// 			modifiedData, err := json.Marshal(jsonData)
// 			if err != nil {
// 				common.SysError("error marshalling modified response: " + err.Error())
// 				continue
// 			}
// 			common.SafeSendString(dataChan, "data: "+string(modifiedData))
// 		}
// 		if len(dataChan) > 0 {
// 			time.Sleep(2 * time.Second)
// 		}
// 		common.SafeSendBool(stopChan, true)
// 	}()
// 	service.SetEventStreamHeaders(c)
// 	c.Stream(func(w io.Writer) bool {
// 		select {
// 		case data := <-dataChan:
// 			data = strings.TrimSuffix(data, "\r")
// 			c.Render(-1, common.CustomEvent{Data: data})
// 			return true
// 		case <-stopChan:
// 			return false
// 		}
// 	})
// 	err := resp.Body.Close()
// 	if err != nil {
// 		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), "", toolCount
// 	}
// 	wg.Wait()
// 	return nil, responseTextBuilder.String(), toolCount
// }

func OpenaiHandler(c *gin.Context, resp *http.Response, promptTokens int, model string, originModel string) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	modelName := originModel
	if v, ok := modelmapper[originModel]; ok {
		fmt.Println("modelName is in modelmapper change to ", v)
		modelName = v
	}
	var simpleResponse dto.SimpleResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var jsonResponse map[string]interface{}
	err = json.Unmarshal(responseBody, &jsonResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	if errInfo, ok := jsonResponse["error"]; ok {
		if errType, exists := errInfo.(map[string]interface{})["type"]; exists && errType != "" {
			return service.OpenAIErrorWrapper(err, "openai_api_error", http.StatusBadGateway), nil
		} else {
			return service.OpenAIErrorWrapper(err, "unknown_error", http.StatusInternalServerError), nil
		}
	}

	// 删除 choices 中的 content_filter_results 字段
	if choices, ok := jsonResponse["choices"].([]interface{}); ok {
		for _, choice := range choices {
			if choiceMap, ok := choice.(map[string]interface{}); ok {
				delete(choiceMap, "content_filter_results")
			}
		}
	}

	// 删除外层的 prompt_filter_results 字段
	delete(jsonResponse, "prompt_filter_results")
	// 修改下model名称
	jsonResponse["model"] = modelName
	fmt.Println("modelName is ", modelName)
	fmt.Println("jsonResponse", jsonResponse)

	// 将修改后的 JSON 对象重新转换为字符串
	modifiedResponseBody, err := json.Marshal(jsonResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_modified_response_body_failed", http.StatusInternalServerError), nil
	}

	// 重置响应体
	resp.Body = io.NopCloser(bytes.NewBuffer(modifiedResponseBody))

	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	// 重新计算 Content-Length 并设置响应头
	c.Writer.Header().Set("Content-Length", fmt.Sprint(len(modifiedResponseBody)))
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}

	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &simpleResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	if simpleResponse.Usage.TotalTokens == 0 {
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
