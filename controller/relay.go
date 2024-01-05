package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"one-api/common"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Message struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
	Name    *string         `json:"name,omitempty"`
}

type MediaMessage struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	ImageUrl any    `json:"image_url,omitempty"`
}

type MessageImageUrl struct {
	Url    string `json:"url"`
	Detail string `json:"detail"`
}

const (
	ContentTypeText     = "text"
	ContentTypeImageURL = "image_url"
)

func (m Message) ParseContent() []MediaMessage {
	var contentList []MediaMessage
	var stringContent string
	if err := json.Unmarshal(m.Content, &stringContent); err == nil {
		contentList = append(contentList, MediaMessage{
			Type: ContentTypeText,
			Text: stringContent,
		})
		return contentList
	}
	var arrayContent []json.RawMessage
	if err := json.Unmarshal(m.Content, &arrayContent); err == nil {
		for _, contentItem := range arrayContent {
			var contentMap map[string]any
			if err := json.Unmarshal(contentItem, &contentMap); err != nil {
				continue
			}
			switch contentMap["type"] {
			case ContentTypeText:
				if subStr, ok := contentMap["text"].(string); ok {
					contentList = append(contentList, MediaMessage{
						Type: ContentTypeText,
						Text: subStr,
					})
				}
			case ContentTypeImageURL:
				if subObj, ok := contentMap["image_url"].(map[string]any); ok {
					detail, ok := subObj["detail"]
					if ok {
						subObj["detail"] = detail.(string)
					} else {
						subObj["detail"] = "auto"
					}
					contentList = append(contentList, MediaMessage{
						Type: ContentTypeImageURL,
						ImageUrl: MessageImageUrl{
							Url:    subObj["url"].(string),
							Detail: subObj["detail"].(string),
						},
					})
				}
			}
		}
		return contentList
	}

	return nil
}

const (
	RelayModeUnknown = iota
	RelayModeChatCompletions
	RelayModeCompletions
	RelayModeEmbeddings
	RelayModeModerations
	RelayModeImagesGenerations
	RelayModeEdits
	RelayModeMidjourneyImagine
	RelayModeMidjourneyDescribe
	RelayModeMidjourneyBlend
	RelayModeMidjourneyChange
	RelayModeMidjourneySimpleChange
	RelayModeMidjourneyNotify
	RelayModeMidjourneyTaskFetch
	RelayModeMidjourneyTaskFetchByCondition
	RelayModeAudio
)

// https://platform.openai.com/docs/api-reference/chat

type GeneralOpenAIRequest struct {
	Model       string    `json:"model,omitempty"`
	Messages    []Message `json:"messages,omitempty"`
	Prompt      any       `json:"prompt,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	MaxTokens   uint      `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	N           int       `json:"n,omitempty"`
	Input       any       `json:"input,omitempty"`
	Instruction string    `json:"instruction,omitempty"`
	Size        string    `json:"size,omitempty"`
	Functions   any       `json:"functions,omitempty"`
}

func (r GeneralOpenAIRequest) ParseInput() []string {
	if r.Input == nil {
		return nil
	}
	var input []string
	switch r.Input.(type) {
	case string:
		input = []string{r.Input.(string)}
	case []any:
		input = make([]string, 0, len(r.Input.([]any)))
		for _, item := range r.Input.([]any) {
			if str, ok := item.(string); ok {
				input = append(input, str)
			}
		}
	}
	return input
}

type AudioRequest struct {
	Model string `json:"model"`
	Voice string `json:"voice"`
	Input string `json:"input"`
}

type ChatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens uint      `json:"max_tokens"`
}

type TextRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	Prompt    string    `json:"prompt"`
	MaxTokens uint      `json:"max_tokens"`
	//Stream   bool      `json:"stream"`
}

type ImageRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n"`
	Size           string `json:"size"`
	Quality        string `json:"quality,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	Style          string `json:"style,omitempty"`
}

type AudioResponse struct {
	Text string `json:"text,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    any    `json:"code"`
}

type OpenAIErrorWithStatusCode struct {
	OpenAIError
	StatusCode int `json:"status_code"`
}

type TextResponse struct {
	Choices []OpenAITextResponseChoice `json:"choices"`
	Usage   `json:"usage"`
	Error   OpenAIError `json:"error"`
}

type OpenAITextResponseChoice struct {
	Index        int `json:"index"`
	Message      `json:"message"`
	FinishReason string `json:"finish_reason"`
}

type OpenAITextResponse struct {
	Id      string                     `json:"id"`
	Object  string                     `json:"object"`
	Created int64                      `json:"created"`
	Choices []OpenAITextResponseChoice `json:"choices"`
	Usage   `json:"usage"`
}

type OpenAIEmbeddingResponseItem struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

type OpenAIEmbeddingResponse struct {
	Object string                        `json:"object"`
	Data   []OpenAIEmbeddingResponseItem `json:"data"`
	Model  string                        `json:"model"`
	Usage  `json:"usage"`
}

type ImageResponse struct {
	Created int `json:"created"`
	Data    []struct {
		Url     string `json:"url"`
		B64Json string `json:"b64_json"`
	}
}

type ChatCompletionsStreamResponseChoice struct {
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
	FinishReason *string `json:"finish_reason"`
}

type ChatCompletionsStreamResponse struct {
	Id      string                                `json:"id"`
	Object  string                                `json:"object"`
	Created int64                                 `json:"created"`
	Model   string                                `json:"model"`
	Choices []ChatCompletionsStreamResponseChoice `json:"choices"`
}

type ChatCompletionsStreamResponseSimple struct {
	Choices []ChatCompletionsStreamResponseChoice `json:"choices"`
}

type CompletionsStreamResponse struct {
	Choices []struct {
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type MidjourneyRequest struct {
	Prompt      string   `json:"prompt"`
	NotifyHook  string   `json:"notifyHook"`
	Action      string   `json:"action"`
	Index       int      `json:"index"`
	State       string   `json:"state"`
	TaskId      string   `json:"taskId"`
	Base64Array []string `json:"base64Array"`
	Content     string   `json:"content"`
}

type MidjourneyResponse struct {
	Code        int         `json:"code"`
	Description string      `json:"description"`
	Properties  interface{} `json:"properties"`
	Result      string      `json:"result"`
}

func Relay(c *gin.Context) {
	relayMode := RelayModeUnknown
	if strings.HasPrefix(c.Request.URL.Path, "/v1/chat/completions") {
		relayMode = RelayModeChatCompletions
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/completions") {
		relayMode = RelayModeCompletions
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/embeddings") {
		relayMode = RelayModeEmbeddings
	} else if strings.HasSuffix(c.Request.URL.Path, "embeddings") {
		relayMode = RelayModeEmbeddings
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/moderations") {
		relayMode = RelayModeModerations
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations") {
		relayMode = RelayModeImagesGenerations
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/edits") {
		relayMode = RelayModeEdits
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/audio") {
		relayMode = RelayModeAudio
	}
	var err *OpenAIErrorWithStatusCode
	switch relayMode {
	case RelayModeImagesGenerations:
		err = relayImageHelper(c, relayMode)
	case RelayModeAudio:
		err = relayAudioHelper(c, relayMode)
	default:
		err = relayTextHelper(c, relayMode)
	}
	if err != nil {
		requestId := c.GetString(common.RequestIdKey)
		retryTimesStr := c.Query("retry")
		retryTimes, _ := strconv.Atoi(retryTimesStr)
		if retryTimesStr == "" {
			retryTimes = common.RetryTimes
		}
		if retryTimes > 0 {
			c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?retry=%d", c.Request.URL.Path, retryTimes-1))
		} else {
			if err.StatusCode == http.StatusTooManyRequests {
				//err.OpenAIError.Message = "当前分组上游负载已饱和，请稍后再试"
			}
			err.OpenAIError.Message = common.MessageWithRequestId(err.OpenAIError.Message, requestId)
			c.JSON(err.StatusCode, gin.H{
				"error": err.OpenAIError,
			})
		}
		channelId := c.GetInt("channel_id")
		autoBan := c.GetBool("auto_ban")
		common.LogError(c.Request.Context(), fmt.Sprintf("relay error (channel #%d): %s", channelId, err.Message))
		// https://platform.openai.com/docs/guides/error-codes/api-errors
		if shouldDisableChannel(&err.OpenAIError, err.StatusCode) && autoBan {
			channelId := c.GetInt("channel_id")
			channelName := c.GetString("channel_name")
			disableChannel(channelId, channelName, err.Message)
		}
	}
}

func RelayMidjourney(c *gin.Context) {
	relayMode := RelayModeUnknown
	if strings.HasPrefix(c.Request.URL.Path, "/mj/submit/imagine") {
		relayMode = RelayModeMidjourneyImagine
	} else if strings.HasPrefix(c.Request.URL.Path, "/mj/submit/blend") {
		relayMode = RelayModeMidjourneyBlend
	} else if strings.HasPrefix(c.Request.URL.Path, "/mj/submit/describe") {
		relayMode = RelayModeMidjourneyDescribe
	} else if strings.HasPrefix(c.Request.URL.Path, "/mj/notify") {
		relayMode = RelayModeMidjourneyNotify
	} else if strings.HasPrefix(c.Request.URL.Path, "/mj/submit/change") {
		relayMode = RelayModeMidjourneyChange
	} else if strings.HasPrefix(c.Request.URL.Path, "/mj/submit/simple-change") {
		relayMode = RelayModeMidjourneyChange
	} else if strings.HasSuffix(c.Request.URL.Path, "/fetch") {
		relayMode = RelayModeMidjourneyTaskFetch
	} else if strings.HasSuffix(c.Request.URL.Path, "/list-by-condition") {
		relayMode = RelayModeMidjourneyTaskFetchByCondition
	}

	var err *MidjourneyResponse
	switch relayMode {
	case RelayModeMidjourneyNotify:
		err = relayMidjourneyNotify(c)
	case RelayModeMidjourneyTaskFetch, RelayModeMidjourneyTaskFetchByCondition:
		err = relayMidjourneyTask(c, relayMode)
	default:
		err = relayMidjourneySubmit(c, relayMode)
	}
	//err = relayMidjourneySubmit(c, relayMode)
	log.Println(err)
	if err != nil {
		retryTimesStr := c.Query("retry")
		retryTimes, _ := strconv.Atoi(retryTimesStr)
		if retryTimesStr == "" {
			retryTimes = common.RetryTimes
		}
		if retryTimes > 0 {
			c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?retry=%d", c.Request.URL.Path, retryTimes-1))
		} else {
			if err.Code == 30 {
				err.Result = "当前分组负载已饱和，请稍后再试，或升级账户以提升服务质量。"
			}
			c.JSON(400, gin.H{
				"error": err.Description + " " + err.Result,
			})
		}
		channelId := c.GetInt("channel_id")
		common.SysError(fmt.Sprintf("relay error (channel #%d): %s", channelId, err.Result))
		//if shouldDisableChannel(&err.OpenAIError) {
		//	channelId := c.GetInt("channel_id")
		//	channelName := c.GetString("channel_name")
		//	disableChannel(channelId, channelName, err.Result)
		//};''''''''''''''''''''''''''''''''
	}
}

func RelayNotImplemented(c *gin.Context) {
	err := OpenAIError{
		Message: "API not implemented",
		Type:    "new_api_error",
		Param:   "",
		Code:    "api_not_implemented",
	}
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": err,
	})
}

func RelayNotFound(c *gin.Context) {
	err := OpenAIError{
		Message: fmt.Sprintf("Invalid URL (%s %s)", c.Request.Method, c.Request.URL.Path),
		Type:    "invalid_request_error",
		Param:   "",
		Code:    "",
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": err,
	})
}
