package cloudflare

import (
	"bufio"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/service"
	"strings"
	"time"
)

func convertCf2CompletionsRequest(textRequest dto.GeneralOpenAIRequest) *CfRequest {
	p, _ := textRequest.Prompt.(string)
	return &CfRequest{
		Prompt:      p,
		MaxTokens:   textRequest.GetMaxTokens(),
		Stream:      textRequest.Stream,
		Temperature: textRequest.Temperature,
	}
}

func cfStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	service.SetEventStreamHeaders(c)
	id := service.GetResponseID(c)
	var responseText string
	isFirst := true

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < len("data: ") {
			continue
		}
		data = strings.TrimPrefix(data, "data: ")
		data = strings.TrimSuffix(data, "\r")

		if data == "[DONE]" {
			break
		}

		var response dto.ChatCompletionsStreamResponse
		err := json.Unmarshal([]byte(data), &response)
		if err != nil {
			common.LogError(c, "error_unmarshalling_stream_response: "+err.Error())
			continue
		}
		for _, choice := range response.Choices {
			choice.Delta.Role = "assistant"
			responseText += choice.Delta.GetContentString()
		}
		response.Id = id
		response.Model = info.UpstreamModelName
		err = service.ObjectData(c, response)
		if isFirst {
			isFirst = false
			info.FirstResponseTime = time.Now()
		}
		if err != nil {
			common.LogError(c, "error_rendering_stream_response: "+err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		common.LogError(c, "error_scanning_stream_response: "+err.Error())
	}
	usage, _ := service.ResponseText2Usage(responseText, info.UpstreamModelName, info.PromptTokens)
	if info.ShouldIncludeUsage {
		response := service.GenerateFinalUsageResponse(id, info.StartTime.Unix(), info.UpstreamModelName, *usage)
		err := service.ObjectData(c, response)
		if err != nil {
			common.LogError(c, "error_rendering_final_usage_response: "+err.Error())
		}
	}
	service.Done(c)

	err := resp.Body.Close()
	if err != nil {
		common.LogError(c, "close_response_body_failed: "+err.Error())
	}

	return nil, usage
}

func cfHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var response dto.TextResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	response.Model = info.UpstreamModelName
	var responseText string
	for _, choice := range response.Choices {
		responseText += choice.Message.StringContent()
	}
	usage, _ := service.ResponseText2Usage(responseText, info.UpstreamModelName, info.PromptTokens)
	response.Usage = *usage
	response.Id = service.GetResponseID(c)
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResponse)
	return nil, usage
}

func cfSTTHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var cfResp CfAudioResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &cfResp)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	audioResp := &dto.AudioResponse{
		Text: cfResp.Result.Text,
	}

	jsonResponse, err := json.Marshal(audioResp)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResponse)

	usage := &dto.Usage{}
	usage.PromptTokens = info.PromptTokens
	usage.CompletionTokens, _ = service.CountTextToken(cfResp.Result.Text, info.UpstreamModelName)
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens

	return nil, usage
}
