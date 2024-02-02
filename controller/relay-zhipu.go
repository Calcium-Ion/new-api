package controller

import (
	"bufio"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"io"
	"net/http"
	"one-api/common"
	"strings"
	"sync"
	"time"
)

// https://open.bigmodel.cn/doc/api#chatglm_std
// chatglm_std, chatglm_lite
// https://open.bigmodel.cn/api/paas/v3/model-api/chatglm_std/invoke
// https://open.bigmodel.cn/api/paas/v3/model-api/chatglm_std/sse-invoke

type ZhipuMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ZhipuRequest struct {
	Prompt      []ZhipuMessage `json:"prompt"`
	Temperature float64        `json:"temperature,omitempty"`
	TopP        float64        `json:"top_p,omitempty"`
	RequestId   string         `json:"request_id,omitempty"`
	Incremental bool           `json:"incremental,omitempty"`
}

type ZhipuResponseData struct {
	TaskId     string         `json:"task_id"`
	RequestId  string         `json:"request_id"`
	TaskStatus string         `json:"task_status"`
	Choices    []ZhipuMessage `json:"choices"`
	Usage      `json:"usage"`
}

type ZhipuResponse struct {
	Code    int               `json:"code"`
	Msg     string            `json:"msg"`
	Success bool              `json:"success"`
	Data    ZhipuResponseData `json:"data"`
}

type ZhipuStreamMetaResponse struct {
	RequestId  string `json:"request_id"`
	TaskId     string `json:"task_id"`
	TaskStatus string `json:"task_status"`
	Usage      `json:"usage"`
}

type zhipuTokenData struct {
	Token      string
	ExpiryTime time.Time
}

var zhipuTokens sync.Map
var expSeconds int64 = 24 * 3600

func getZhipuToken(apikey string) string {
	data, ok := zhipuTokens.Load(apikey)
	if ok {
		tokenData := data.(zhipuTokenData)
		if time.Now().Before(tokenData.ExpiryTime) {
			return tokenData.Token
		}
	}

	split := strings.Split(apikey, ".")
	if len(split) != 2 {
		common.SysError("invalid zhipu key: " + apikey)
		return ""
	}

	id := split[0]
	secret := split[1]

	expMillis := time.Now().Add(time.Duration(expSeconds)*time.Second).UnixNano() / 1e6
	expiryTime := time.Now().Add(time.Duration(expSeconds) * time.Second)

	timestamp := time.Now().UnixNano() / 1e6

	payload := jwt.MapClaims{
		"api_key":   id,
		"exp":       expMillis,
		"timestamp": timestamp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	token.Header["alg"] = "HS256"
	token.Header["sign_type"] = "SIGN"

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return ""
	}

	zhipuTokens.Store(apikey, zhipuTokenData{
		Token:      tokenString,
		ExpiryTime: expiryTime,
	})

	return tokenString
}

func requestOpenAI2Zhipu(request GeneralOpenAIRequest) *ZhipuRequest {
	messages := make([]ZhipuMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		if message.Role == "system" {
			messages = append(messages, ZhipuMessage{
				Role:    "system",
				Content: message.StringContent(),
			})
			messages = append(messages, ZhipuMessage{
				Role:    "user",
				Content: "Okay",
			})
		} else {
			messages = append(messages, ZhipuMessage{
				Role:    message.Role,
				Content: message.StringContent(),
			})
		}
	}
	return &ZhipuRequest{
		Prompt:      messages,
		Temperature: request.Temperature,
		TopP:        request.TopP,
		Incremental: false,
	}
}

func responseZhipu2OpenAI(response *ZhipuResponse) *OpenAITextResponse {
	fullTextResponse := OpenAITextResponse{
		Id:      response.Data.TaskId,
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: make([]OpenAITextResponseChoice, 0, len(response.Data.Choices)),
		Usage:   response.Data.Usage,
	}
	for i, choice := range response.Data.Choices {
		content, _ := json.Marshal(strings.Trim(choice.Content, "\""))
		openaiChoice := OpenAITextResponseChoice{
			Index: i,
			Message: Message{
				Role:    choice.Role,
				Content: content,
			},
			FinishReason: "",
		}
		if i == len(response.Data.Choices)-1 {
			openaiChoice.FinishReason = "stop"
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, openaiChoice)
	}
	return &fullTextResponse
}

func streamResponseZhipu2OpenAI(zhipuResponse string) *ChatCompletionsStreamResponse {
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = zhipuResponse
	response := ChatCompletionsStreamResponse{
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "chatglm",
		Choices: []ChatCompletionsStreamResponseChoice{choice},
	}
	return &response
}

func streamMetaResponseZhipu2OpenAI(zhipuResponse *ZhipuStreamMetaResponse) (*ChatCompletionsStreamResponse, *Usage) {
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = ""
	choice.FinishReason = &stopFinishReason
	response := ChatCompletionsStreamResponse{
		Id:      zhipuResponse.RequestId,
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "chatglm",
		Choices: []ChatCompletionsStreamResponseChoice{choice},
	}
	return &response, &zhipuResponse.Usage
}

func zhipuStreamHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var usage *Usage
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "\n\n"); i >= 0 && strings.Index(string(data), ":") >= 0 {
			return i + 2, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	dataChan := make(chan string)
	metaChan := make(chan string)
	stopChan := make(chan bool)
	go func() {
		for scanner.Scan() {
			data := scanner.Text()
			lines := strings.Split(data, "\n")
			for i, line := range lines {
				if len(line) < 5 {
					continue
				}
				if line[:5] == "data:" {
					dataChan <- line[5:]
					if i != len(lines)-1 {
						dataChan <- "\n"
					}
				} else if line[:5] == "meta:" {
					metaChan <- line[5:]
				}
			}
		}
		stopChan <- true
	}()
	setEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			response := streamResponseZhipu2OpenAI(data)
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonResponse)})
			return true
		case data := <-metaChan:
			var zhipuResponse ZhipuStreamMetaResponse
			err := json.Unmarshal([]byte(data), &zhipuResponse)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			response, zhipuUsage := streamMetaResponseZhipu2OpenAI(&zhipuResponse)
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			usage = zhipuUsage
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
	return nil, usage
}

func zhipuHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var zhipuResponse ZhipuResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &zhipuResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if !zhipuResponse.Success {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
				Message: zhipuResponse.Msg,
				Type:    "zhipu_error",
				Param:   "",
				Code:    zhipuResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseZhipu2OpenAI(&zhipuResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}
