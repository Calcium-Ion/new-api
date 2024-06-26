package tencent

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/service"
	"sort"
	"strconv"
	"strings"
)

// https://cloud.tencent.com/document/product/1729/97732

func requestOpenAI2Tencent(request dto.GeneralOpenAIRequest) *TencentChatRequest {
	messages := make([]TencentMessage, 0, len(request.Messages))
	for i := 0; i < len(request.Messages); i++ {
		message := request.Messages[i]
		if message.Role == "system" {
			messages = append(messages, TencentMessage{
				Role:    "user",
				Content: message.StringContent(),
			})
			messages = append(messages, TencentMessage{
				Role:    "assistant",
				Content: "Okay",
			})
			continue
		}
		messages = append(messages, TencentMessage{
			Content: message.StringContent(),
			Role:    message.Role,
		})
	}
	stream := 0
	if request.Stream {
		stream = 1
	}
	return &TencentChatRequest{
		Timestamp:   common.GetTimestamp(),
		Expired:     common.GetTimestamp() + 24*60*60,
		QueryID:     common.GetUUID(),
		Temperature: request.Temperature,
		TopP:        request.TopP,
		Stream:      stream,
		Messages:    messages,
		Model:       request.Model,
	}
}

func responseTencent2OpenAI(response *TencentChatResponse) *dto.OpenAITextResponse {
	fullTextResponse := dto.OpenAITextResponse{
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Usage:   response.Usage,
	}
	if len(response.Choices) > 0 {
		content, _ := json.Marshal(response.Choices[0].Messages.Content)
		choice := dto.OpenAITextResponseChoice{
			Index: 0,
			Message: dto.Message{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: response.Choices[0].FinishReason,
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	}
	return &fullTextResponse
}

func streamResponseTencent2OpenAI(TencentResponse *TencentChatResponse) *dto.ChatCompletionsStreamResponse {
	response := dto.ChatCompletionsStreamResponse{
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "tencent-hunyuan",
	}
	if len(TencentResponse.Choices) > 0 {
		var choice dto.ChatCompletionsStreamResponseChoice
		choice.Delta.SetContentString(TencentResponse.Choices[0].Delta.Content)
		if TencentResponse.Choices[0].FinishReason == "stop" {
			choice.FinishReason = &relaycommon.StopFinishReason
		}
		response.Choices = append(response.Choices, choice)
	}
	return &response
}

func tencentStreamHandler(c *gin.Context, resp *http.Response) (*dto.OpenAIErrorWithStatusCode, string) {
	var responseText string
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
	service.SetEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			var TencentResponse TencentChatResponse
			err := json.Unmarshal([]byte(data), &TencentResponse)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			response := streamResponseTencent2OpenAI(&TencentResponse)
			if len(response.Choices) != 0 {
				responseText += response.Choices[0].Delta.GetContentString()
			}
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
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), ""
	}
	return nil, responseText
}

func tencentHandler(c *gin.Context, resp *http.Response) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var TencentResponse TencentChatResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &TencentResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if TencentResponse.Error.Code != 0 {
		return &dto.OpenAIErrorWithStatusCode{
			Error: dto.OpenAIError{
				Message: TencentResponse.Error.Message,
				Code:    TencentResponse.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseTencent2OpenAI(&TencentResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func parseTencentConfig(config string) (appId int64, secretId string, secretKey string, err error) {
	parts := strings.Split(config, "|")
	if len(parts) != 3 {
		err = errors.New("invalid tencent config")
		return
	}
	appId, err = strconv.ParseInt(parts[0], 10, 64)
	secretId = parts[1]
	secretKey = parts[2]
	return
}

func getTencentSign(req TencentChatRequest, secretKey string) string {
	params := make([]string, 0)
	params = append(params, "app_id="+strconv.FormatInt(req.AppId, 10))
	params = append(params, "secret_id="+req.SecretId)
	params = append(params, "timestamp="+strconv.FormatInt(req.Timestamp, 10))
	params = append(params, "query_id="+req.QueryID)
	params = append(params, "temperature="+strconv.FormatFloat(req.Temperature, 'f', -1, 64))
	params = append(params, "top_p="+strconv.FormatFloat(req.TopP, 'f', -1, 64))
	params = append(params, "stream="+strconv.Itoa(req.Stream))
	params = append(params, "expired="+strconv.FormatInt(req.Expired, 10))

	var messageStr string
	for _, msg := range req.Messages {
		messageStr += fmt.Sprintf(`{"role":"%s","content":"%s"},`, msg.Role, msg.Content)
	}
	messageStr = strings.TrimSuffix(messageStr, ",")
	params = append(params, "messages=["+messageStr+"]")

	sort.Sort(sort.StringSlice(params))
	url := "hunyuan.cloud.tencent.com/hyllm/v1/chat/completions?" + strings.Join(params, "&")
	mac := hmac.New(sha1.New, []byte(secretKey))
	signURL := url
	mac.Write([]byte(signURL))
	sign := mac.Sum([]byte(nil))
	return base64.StdEncoding.EncodeToString(sign)
}
