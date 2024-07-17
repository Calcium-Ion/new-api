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
	"time"
)

func OpenaiStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	hasStreamUsage := false
	responseId := ""
	var createAt int64 = 0
	var systemFingerprint string
	model := info.UpstreamModelName

	var responseTextBuilder strings.Builder
	var usage = &dto.Usage{}
	var streamItems []string // store stream items

	toolCount := 0
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	service.SetEventStreamHeaders(c)

	ticker := time.NewTicker(time.Duration(constant.StreamingTimeout) * time.Second)
	defer ticker.Stop()

	stopChan := make(chan bool)
	defer close(stopChan)

	go func() {
		for scanner.Scan() {
			info.SetFirstResponseTime()
			ticker.Reset(time.Duration(constant.StreamingTimeout) * time.Second)
			data := scanner.Text()
			if len(data) < 6 { // ignore blank line or wrong format
				continue
			}
			if data[:6] != "data: " && data[:6] != "[DONE]" {
				continue
			}
			data = data[6:]
			if !strings.HasPrefix(data, "[DONE]") {
				err := service.StringData(c, data)
				if err != nil {
					common.LogError(c, "streaming error: "+err.Error())
				}
				streamItems = append(streamItems, data)
			}
		}
		common.SafeSendBool(stopChan, true)
	}()

	select {
	case <-ticker.C:
		// 超时处理逻辑
		common.LogError(c, "streaming timeout")
	case <-stopChan:
		// 正常结束
	}

	// 计算token
	streamResp := "[" + strings.Join(streamItems, ",") + "]"
	switch info.RelayMode {
	case relayconstant.RelayModeChatCompletions:
		var streamResponses []dto.ChatCompletionsStreamResponse
		err := json.Unmarshal(common.StringToByteSlice(streamResp), &streamResponses)
		if err != nil {
			// 一次性解析失败，逐个解析
			common.SysError("error unmarshalling stream response: " + err.Error())
			for _, item := range streamItems {
				var streamResponse dto.ChatCompletionsStreamResponse
				err := json.Unmarshal(common.StringToByteSlice(item), &streamResponse)
				if err == nil {
					responseId = streamResponse.Id
					createAt = streamResponse.Created
					systemFingerprint = streamResponse.GetSystemFingerprint()
					model = streamResponse.Model
					if service.ValidUsage(streamResponse.Usage) {
						usage = streamResponse.Usage
						hasStreamUsage = true
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
				responseId = streamResponse.Id
				createAt = streamResponse.Created
				systemFingerprint = streamResponse.GetSystemFingerprint()
				model = streamResponse.Model
				if service.ValidUsage(streamResponse.Usage) {
					usage = streamResponse.Usage
					hasStreamUsage = true
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

	if !hasStreamUsage {
		usage, _ = service.ResponseText2Usage(responseTextBuilder.String(), info.UpstreamModelName, info.PromptTokens)
		usage.CompletionTokens += toolCount * 7
	}

	if info.ShouldIncludeUsage && !hasStreamUsage {
		response := service.GenerateFinalUsageResponse(responseId, createAt, model, *usage)
		response.SetSystemFingerprint(systemFingerprint)
		service.ObjectData(c, response)
	}

	service.Done(c)

	err := resp.Body.Close()
	if err != nil {
		common.LogError(c, "close_response_body_failed: "+err.Error())
	}
	return nil, usage
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
