package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pkoukk/tiktoken-go"
	"image"
	"log"
	"math"
	"one-api/common"
	"one-api/dto"
	"strings"
	"unicode/utf8"
)

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
	for model, _ := range common.DefaultModelRatio {
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

func getImageToken(imageUrl *dto.MessageImageUrl) (int, error) {
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
		config, format, _, err = common.DecodeBase64ImageData(imageUrl.Url)
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

func CountTokenMessages(messages []dto.Message, model string, checkSensitive bool) (int, error, bool) {
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
			var arrayContent []dto.MediaMessage
			if err := json.Unmarshal(message.Content, &arrayContent); err != nil {
				var stringContent string
				if err := json.Unmarshal(message.Content, &stringContent); err != nil {
					return 0, err, false
				} else {
					if checkSensitive {
						contains, words := SensitiveWordContains(stringContent)
						if contains {
							err := fmt.Errorf("message contains sensitive words: [%s]", strings.Join(words, ", "))
							return 0, err, true
						}
					}
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
						if model == "glm-4v" {
							imageTokenNum = 1047
						} else {
							if str, ok := m.ImageUrl.(string); ok {
								imageTokenNum, err = getImageToken(&dto.MessageImageUrl{Url: str, Detail: "auto"})
							} else {
								imageUrlMap := m.ImageUrl.(map[string]interface{})
								detail, ok := imageUrlMap["detail"]
								if ok {
									imageUrlMap["detail"] = detail.(string)
								} else {
									imageUrlMap["detail"] = "auto"
								}
								imageUrl := dto.MessageImageUrl{
									Url:    imageUrlMap["url"].(string),
									Detail: imageUrlMap["detail"].(string),
								}
								imageTokenNum, err = getImageToken(&imageUrl)
							}
							if err != nil {
								return 0, err, false
							}
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
	return tokenNum, nil, false
}

func CountTokenInput(input any, model string, check bool) (int, error, bool) {
	switch v := input.(type) {
	case string:
		return CountTokenText(v, model, check)
	case []string:
		text := ""
		for _, s := range v {
			text += s
		}
		return CountTokenText(text, model, check)
	}
	return CountTokenInput(fmt.Sprintf("%v", input), model, check)
}

func CountTokenStreamChoices(messages []dto.ChatCompletionsStreamResponseChoice, model string) int {
	tokens := 0
	for _, message := range messages {
		tkm, _, _ := CountTokenInput(message.Delta.Content, model, false)
		tokens += tkm
		if message.Delta.ToolCalls != nil {
			for _, tool := range message.Delta.ToolCalls {
				tkm, _, _ := CountTokenInput(tool.Function.Name, model, false)
				tokens += tkm
				tkm, _, _ = CountTokenInput(tool.Function.Arguments, model, false)
				tokens += tkm
			}
		}
	}
	return tokens
}

func CountAudioToken(text string, model string, check bool) (int, error, bool) {
	if strings.HasPrefix(model, "tts") {
		contains, words := SensitiveWordContains(text)
		if contains {
			return utf8.RuneCountInString(text), fmt.Errorf("input contains sensitive words: [%s]", strings.Join(words, ",")), true
		}
		return utf8.RuneCountInString(text), nil, false
	} else {
		return CountTokenText(text, model, check)
	}
}

// CountTokenText 统计文本的token数量，仅当文本包含敏感词，返回错误，同时返回token数量
func CountTokenText(text string, model string, check bool) (int, error, bool) {
	var err error
	var trigger bool
	if check {
		contains, words := SensitiveWordContains(text)
		if contains {
			err = fmt.Errorf("input contains sensitive words: [%s]", strings.Join(words, ","))
			trigger = true
		}
	}
	tokenEncoder := getTokenEncoder(model)
	return getTokenNum(tokenEncoder, text), err, trigger
}
