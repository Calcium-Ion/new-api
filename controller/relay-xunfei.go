package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/url"
	"one-api/common"
	"strings"
	"time"
)

// https://console.xfyun.cn/services/cbm
// https://www.xfyun.cn/doc/spark/Web.html

type XunfeiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type XunfeiChatRequest struct {
	Header struct {
		AppId string `json:"app_id"`
	} `json:"header"`
	Parameter struct {
		Chat struct {
			Domain      string  `json:"domain,omitempty"`
			Temperature float64 `json:"temperature,omitempty"`
			TopK        int     `json:"top_k,omitempty"`
			MaxTokens   uint    `json:"max_tokens,omitempty"`
			Auditing    bool    `json:"auditing,omitempty"`
		} `json:"chat"`
	} `json:"parameter"`
	Payload struct {
		Message struct {
			Text []XunfeiMessage `json:"text"`
		} `json:"message"`
	} `json:"payload"`
}

type XunfeiChatResponseTextItem struct {
	Content string `json:"content"`
	Role    string `json:"role"`
	Index   int    `json:"index"`
}

type XunfeiChatResponse struct {
	Header struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Sid     string `json:"sid"`
		Status  int    `json:"status"`
	} `json:"header"`
	Payload struct {
		Choices struct {
			Status int                          `json:"status"`
			Seq    int                          `json:"seq"`
			Text   []XunfeiChatResponseTextItem `json:"text"`
		} `json:"choices"`
		Usage struct {
			//Text struct {
			//	QuestionTokens   string `json:"question_tokens"`
			//	PromptTokens     string `json:"prompt_tokens"`
			//	CompletionTokens string `json:"completion_tokens"`
			//	TotalTokens      string `json:"total_tokens"`
			//} `json:"text"`
			Text Usage `json:"text"`
		} `json:"usage"`
	} `json:"payload"`
}

func requestOpenAI2Xunfei(request GeneralOpenAIRequest, xunfeiAppId string, domain string) *XunfeiChatRequest {
	messages := make([]XunfeiMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		if message.Role == "system" {
			messages = append(messages, XunfeiMessage{
				Role:    "user",
				Content: message.StringContent(),
			})
			messages = append(messages, XunfeiMessage{
				Role:    "assistant",
				Content: "Okay",
			})
		} else {
			messages = append(messages, XunfeiMessage{
				Role:    message.Role,
				Content: message.StringContent(),
			})
		}
	}
	xunfeiRequest := XunfeiChatRequest{}
	xunfeiRequest.Header.AppId = xunfeiAppId
	xunfeiRequest.Parameter.Chat.Domain = domain
	xunfeiRequest.Parameter.Chat.Temperature = request.Temperature
	xunfeiRequest.Parameter.Chat.TopK = request.N
	xunfeiRequest.Parameter.Chat.MaxTokens = request.MaxTokens
	xunfeiRequest.Payload.Message.Text = messages
	return &xunfeiRequest
}

func responseXunfei2OpenAI(response *XunfeiChatResponse) *OpenAITextResponse {
	if len(response.Payload.Choices.Text) == 0 {
		response.Payload.Choices.Text = []XunfeiChatResponseTextItem{
			{
				Content: "",
			},
		}
	}
	content, _ := json.Marshal(response.Payload.Choices.Text[0].Content)
	choice := OpenAITextResponseChoice{
		Index: 0,
		Message: Message{
			Role:    "assistant",
			Content: content,
		},
		FinishReason: stopFinishReason,
	}
	fullTextResponse := OpenAITextResponse{
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: []OpenAITextResponseChoice{choice},
		Usage:   response.Payload.Usage.Text,
	}
	return &fullTextResponse
}

func streamResponseXunfei2OpenAI(xunfeiResponse *XunfeiChatResponse) *ChatCompletionsStreamResponse {
	if len(xunfeiResponse.Payload.Choices.Text) == 0 {
		xunfeiResponse.Payload.Choices.Text = []XunfeiChatResponseTextItem{
			{
				Content: "",
			},
		}
	}
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = xunfeiResponse.Payload.Choices.Text[0].Content
	if xunfeiResponse.Payload.Choices.Status == 2 {
		choice.FinishReason = &stopFinishReason
	}
	response := ChatCompletionsStreamResponse{
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "SparkDesk",
		Choices: []ChatCompletionsStreamResponseChoice{choice},
	}
	return &response
}

func buildXunfeiAuthUrl(hostUrl string, apiKey, apiSecret string) string {
	HmacWithShaToBase64 := func(algorithm, data, key string) string {
		mac := hmac.New(sha256.New, []byte(key))
		mac.Write([]byte(data))
		encodeData := mac.Sum(nil)
		return base64.StdEncoding.EncodeToString(encodeData)
	}
	ul, err := url.Parse(hostUrl)
	if err != nil {
		fmt.Println(err)
	}
	date := time.Now().UTC().Format(time.RFC1123)
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}
	sign := strings.Join(signString, "\n")
	sha := HmacWithShaToBase64("hmac-sha256", sign, apiSecret)
	authUrl := fmt.Sprintf("hmac username=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
		"hmac-sha256", "host date request-line", sha)
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))
	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)
	callUrl := hostUrl + "?" + v.Encode()
	return callUrl
}

func xunfeiStreamHandler(c *gin.Context, textRequest GeneralOpenAIRequest, appId string, apiSecret string, apiKey string) (*OpenAIErrorWithStatusCode, *Usage) {
	domain, authUrl := getXunfeiAuthUrl(c, apiKey, apiSecret)
	dataChan, stopChan, err := xunfeiMakeRequest(textRequest, domain, authUrl, appId)
	if err != nil {
		return errorWrapper(err, "make xunfei request err", http.StatusInternalServerError), nil
	}
	setEventStreamHeaders(c)
	var usage Usage
	c.Stream(func(w io.Writer) bool {
		select {
		case xunfeiResponse := <-dataChan:
			usage.PromptTokens += xunfeiResponse.Payload.Usage.Text.PromptTokens
			usage.CompletionTokens += xunfeiResponse.Payload.Usage.Text.CompletionTokens
			usage.TotalTokens += xunfeiResponse.Payload.Usage.Text.TotalTokens
			response := streamResponseXunfei2OpenAI(&xunfeiResponse)
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
	return nil, &usage
}

func xunfeiHandler(c *gin.Context, textRequest GeneralOpenAIRequest, appId string, apiSecret string, apiKey string) (*OpenAIErrorWithStatusCode, *Usage) {
	domain, authUrl := getXunfeiAuthUrl(c, apiKey, apiSecret)
	dataChan, stopChan, err := xunfeiMakeRequest(textRequest, domain, authUrl, appId)
	if err != nil {
		return errorWrapper(err, "make xunfei request err", http.StatusInternalServerError), nil
	}
	var usage Usage
	var content string
	var xunfeiResponse XunfeiChatResponse
	stop := false
	for !stop {
		select {
		case xunfeiResponse = <-dataChan:
			if len(xunfeiResponse.Payload.Choices.Text) == 0 {
				continue
			}
			content += xunfeiResponse.Payload.Choices.Text[0].Content
			usage.PromptTokens += xunfeiResponse.Payload.Usage.Text.PromptTokens
			usage.CompletionTokens += xunfeiResponse.Payload.Usage.Text.CompletionTokens
			usage.TotalTokens += xunfeiResponse.Payload.Usage.Text.TotalTokens
		case stop = <-stopChan:
		}
	}

	xunfeiResponse.Payload.Choices.Text[0].Content = content

	response := responseXunfei2OpenAI(&xunfeiResponse)
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	_, _ = c.Writer.Write(jsonResponse)
	return nil, &usage
}

func xunfeiMakeRequest(textRequest GeneralOpenAIRequest, domain, authUrl, appId string) (chan XunfeiChatResponse, chan bool, error) {
	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	conn, resp, err := d.Dial(authUrl, nil)
	if err != nil || resp.StatusCode != 101 {
		return nil, nil, err
	}
	data := requestOpenAI2Xunfei(textRequest, appId, domain)
	err = conn.WriteJSON(data)
	if err != nil {
		return nil, nil, err
	}

	dataChan := make(chan XunfeiChatResponse)
	stopChan := make(chan bool)
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				common.SysError("error reading stream response: " + err.Error())
				break
			}
			var response XunfeiChatResponse
			err = json.Unmarshal(msg, &response)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				break
			}
			dataChan <- response
			if response.Payload.Choices.Status == 2 {
				err := conn.Close()
				if err != nil {
					common.SysError("error closing websocket connection: " + err.Error())
				}
				break
			}
		}
		stopChan <- true
	}()

	return dataChan, stopChan, nil
}

func getXunfeiAuthUrl(c *gin.Context, apiKey string, apiSecret string) (string, string) {
	query := c.Request.URL.Query()
	apiVersion := query.Get("api-version")
	if apiVersion == "" {
		apiVersion = c.GetString("api_version")
	}
	if apiVersion == "" {
		apiVersion = "v1.1"
		common.SysLog("api_version not found, use default: " + apiVersion)
	}
	domain := "general"
	if apiVersion != "v1.1" {
		domain += strings.Split(apiVersion, ".")[0]
	}
	authUrl := buildXunfeiAuthUrl(fmt.Sprintf("wss://spark-api.xf-yun.com/%s/chat", apiVersion), apiKey, apiSecret)
	return domain, authUrl
}
