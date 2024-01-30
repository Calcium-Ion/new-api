package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkoukk/tiktoken-go"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"net/http"
	"one-api/common"
	"strconv"
	"strings"
	"unicode/utf8"
)

var stopFinishReason = "stop"

// tokenEncoderMap won't grow after initialization
var tokenEncoderMap = map[string]*tiktoken.Tiktoken{}
var defaultTokenEncoder *tiktoken.Tiktoken

func InitTokenEncoders() {
	common.SysLog("initializing token encoders")
	gpt35TokenEncoder, err := tiktoken.EncodingForModel("gpt-3.5-turbo")
	if err != nil {
		common.FatalLog(fmt.Sprintf("failed to get gpt-3.5-turbo token encoder: %s", err.Error()))
	}
	defaultTokenEncoder = gpt35TokenEncoder
	gpt4TokenEncoder, err := tiktoken.EncodingForModel("gpt-4")
	if err != nil {
		common.FatalLog(fmt.Sprintf("failed to get gpt-4 token encoder: %s", err.Error()))
	}
	for model, _ := range common.ModelRatio {
		if strings.HasPrefix(model, "gpt-3.5") {
			tokenEncoderMap[model] = gpt35TokenEncoder
		} else if strings.HasPrefix(model, "gpt-4") {
			tokenEncoderMap[model] = gpt4TokenEncoder
		} else {
			tokenEncoderMap[model] = nil
		}
	}
	common.SysLog("token encoders initialized")
}

func getTokenEncoder(model string) *tiktoken.Tiktoken {
	tokenEncoder, ok := tokenEncoderMap[model]
	if ok && tokenEncoder != nil {
		return tokenEncoder
	}
	if ok {
		tokenEncoder, err := tiktoken.EncodingForModel(model)
		if err != nil {
			common.SysError(fmt.Sprintf("failed to get token encoder for model %s: %s, using encoder for gpt-3.5-turbo", model, err.Error()))
			tokenEncoder = defaultTokenEncoder
		}
		tokenEncoderMap[model] = tokenEncoder
		return tokenEncoder
	}
	return defaultTokenEncoder
}

func getTokenNum(tokenEncoder *tiktoken.Tiktoken, text string) int {
	return len(tokenEncoder.Encode(text, nil, nil))
}

func getImageToken(imageUrl *MessageImageUrl) (int, error) {
	if imageUrl.Detail == "low" {
		return 85, nil
	}
	var config image.Config
	var err error
	var format string
	if strings.HasPrefix(imageUrl.Url, "http") {
		common.SysLog(fmt.Sprintf("downloading image: %s", imageUrl.Url))
		config, format, err = common.DecodeUrlImageData(imageUrl.Url)
	} else {
		common.SysLog(fmt.Sprintf("decoding image"))
		config, format, err = common.DecodeBase64ImageData(imageUrl.Url)
	}
	if err != nil {
		return 0, err
	}

	if config.Width == 0 || config.Height == 0 {
		return 0, errors.New(fmt.Sprintf("fail to decode image config: %s", imageUrl.Url))
	}
	// TODO: 适配官方auto计费
	if config.Width < 512 && config.Height < 512 {
		if imageUrl.Detail == "auto" || imageUrl.Detail == "" {
			// 如果图片尺寸小于512，强制使用low
			imageUrl.Detail = "low"
			return 85, nil
		}
	}

	shortSide := config.Width
	otherSide := config.Height
	log.Printf("format: %s, width: %d, height: %d", format, config.Width, config.Height)
	// 缩放倍数
	scale := 1.0
	if config.Height < shortSide {
		shortSide = config.Height
		otherSide = config.Width
	}

	// 将最小变的尺寸缩小到768以下，如果大于768，则缩放到768
	if shortSide > 768 {
		scale = float64(shortSide) / 768
		shortSide = 768
	}
	// 将另一边按照相同的比例缩小，向上取整
	otherSide = int(math.Ceil(float64(otherSide) / scale))
	log.Printf("shortSide: %d, otherSide: %d, scale: %f", shortSide, otherSide, scale)
	// 计算图片的token数量(边的长度除以512，向上取整)
	tiles := (shortSide + 511) / 512 * ((otherSide + 511) / 512)
	log.Printf("tiles: %d", tiles)
	return tiles*170 + 85, nil
}

func countTokenMessages(messages []Message, model string) (int, error) {
	//recover when panic
	tokenEncoder := getTokenEncoder(model)
	// Reference:
	// https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
	// https://github.com/pkoukk/tiktoken-go/issues/6
	//
	// Every message follows <|start|>{role/name}\n{content}<|end|>\n
	var tokensPerMessage int
	var tokensPerName int
	if model == "gpt-3.5-turbo-0301" {
		tokensPerMessage = 4
		tokensPerName = -1 // If there's a name, the role is omitted
	} else {
		tokensPerMessage = 3
		tokensPerName = 1
	}
	tokenNum := 0
	for _, message := range messages {
		tokenNum += tokensPerMessage
		tokenNum += getTokenNum(tokenEncoder, message.Role)
		if len(message.Content) > 0 {
			var arrayContent []MediaMessage
			if err := json.Unmarshal(message.Content, &arrayContent); err != nil {
				var stringContent string
				if err := json.Unmarshal(message.Content, &stringContent); err != nil {
					return 0, err
				} else {
					tokenNum += getTokenNum(tokenEncoder, stringContent)
					if message.Name != nil {
						tokenNum += tokensPerName
						tokenNum += getTokenNum(tokenEncoder, *message.Name)
					}
				}
			} else {
				for _, m := range arrayContent {
					if m.Type == "image_url" {
						var imageTokenNum int
						if str, ok := m.ImageUrl.(string); ok {
							imageTokenNum, err = getImageToken(&MessageImageUrl{Url: str, Detail: "auto"})
						} else {
							imageUrlMap := m.ImageUrl.(map[string]interface{})
							detail, ok := imageUrlMap["detail"]
							if ok {
								imageUrlMap["detail"] = detail.(string)
							} else {
								imageUrlMap["detail"] = "auto"
							}
							imageUrl := MessageImageUrl{
								Url:    imageUrlMap["url"].(string),
								Detail: imageUrlMap["detail"].(string),
							}
							imageTokenNum, err = getImageToken(&imageUrl)
						}
						if err != nil {
							return 0, err
						}

						tokenNum += imageTokenNum
						log.Printf("image token num: %d", imageTokenNum)
					} else {
						tokenNum += getTokenNum(tokenEncoder, m.Text)
					}
				}
			}
		}
	}
	tokenNum += 3 // Every reply is primed with <|start|>assistant<|message|>
	return tokenNum, nil
}

func countTokenInput(input any, model string) int {
	switch v := input.(type) {
	case string:
		return countTokenText(v, model)
	case []string:
		text := ""
		for _, s := range v {
			text += s
		}
		return countTokenText(text, model)
	}
	return 0
}

func countAudioToken(text string, model string) int {
	if strings.HasPrefix(model, "tts") {
		return utf8.RuneCountInString(text)
	} else {
		return countTokenText(text, model)
	}
}

func countTokenText(text string, model string) int {
	tokenEncoder := getTokenEncoder(model)
	return getTokenNum(tokenEncoder, text)
}

func errorWrapper(err error, code string, statusCode int) *OpenAIErrorWithStatusCode {
	text := err.Error()
	// 定义一个正则表达式匹配URL
	if strings.Contains(text, "Post") {
		common.SysLog(fmt.Sprintf("error: %s", text))
		text = "请求上游地址失败"
	}
	//避免暴露内部错误

	openAIError := OpenAIError{
		Message: text,
		Type:    "new_api_error",
		Code:    code,
	}
	return &OpenAIErrorWithStatusCode{
		OpenAIError: openAIError,
		StatusCode:  statusCode,
	}
}

func shouldDisableChannel(err *OpenAIError, statusCode int) bool {
	if !common.AutomaticDisableChannelEnabled {
		return false
	}
	if err == nil {
		return false
	}
	if statusCode == http.StatusUnauthorized {
		return true
	}
	if err.Type == "insufficient_quota" || err.Code == "invalid_api_key" || err.Code == "account_deactivated" || err.Code == "billing_not_active" {
		return true
	}
	return false
}

func setEventStreamHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}

func relayErrorHandler(resp *http.Response) (openAIErrorWithStatusCode *OpenAIErrorWithStatusCode) {
	openAIErrorWithStatusCode = &OpenAIErrorWithStatusCode{
		StatusCode: resp.StatusCode,
		OpenAIError: OpenAIError{
			Message: fmt.Sprintf("bad response status code %d", resp.StatusCode),
			Type:    "upstream_error",
			Code:    "bad_response_status_code",
			Param:   strconv.Itoa(resp.StatusCode),
		},
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = resp.Body.Close()
	if err != nil {
		return
	}
	var textResponse TextResponse
	err = json.Unmarshal(responseBody, &textResponse)
	if err != nil {
		return
	}
	openAIErrorWithStatusCode.OpenAIError = textResponse.Error
	return
}

func getFullRequestURL(baseURL string, requestURL string, channelType int) string {
	fullRequestURL := fmt.Sprintf("%s%s", baseURL, requestURL)
	if channelType == common.ChannelTypeOpenAI {
		if strings.HasPrefix(baseURL, "https://gateway.ai.cloudflare.com") {
			fullRequestURL = fmt.Sprintf("%s%s", baseURL, strings.TrimPrefix(requestURL, "/v1"))
		}
	}
	return fullRequestURL
}

func GetAPIVersion(c *gin.Context) string {
	query := c.Request.URL.Query()
	apiVersion := query.Get("api-version")
	if apiVersion == "" {
		apiVersion = c.GetString("api_version")
	}
	return apiVersion
}
