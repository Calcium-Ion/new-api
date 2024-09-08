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
	"one-api/constant"
	"one-api/dto"
	"strings"
	"unicode/utf8"
)

// tokenEncoderMap won't grow after initialization
var tokenEncoderMap = map[string]*tiktoken.Tiktoken{}
var defaultTokenEncoder *tiktoken.Tiktoken
var cl200kTokenEncoder *tiktoken.Tiktoken

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
	cl200kTokenEncoder, err = tiktoken.EncodingForModel("gpt-4o")
	if err != nil {
		common.FatalLog(fmt.Sprintf("failed to get gpt-4o token encoder: %s", err.Error()))
	}
	for model, _ := range common.GetDefaultModelRatioMap() {
		if strings.HasPrefix(model, "gpt-3.5") {
			tokenEncoderMap[model] = gpt35TokenEncoder
		} else if strings.HasPrefix(model, "gpt-4") {
			if strings.HasPrefix(model, "gpt-4o") {
				tokenEncoderMap[model] = cl200kTokenEncoder
			} else {
				tokenEncoderMap[model] = gpt4TokenEncoder
			}
		} else {
			tokenEncoderMap[model] = nil
		}
	}
	common.SysLog("token encoders initialized")
}

func getModelDefaultTokenEncoder(model string) *tiktoken.Tiktoken {
	if strings.HasPrefix(model, "gpt-4o") || strings.HasPrefix(model, "chatgpt-4o") {
		return cl200kTokenEncoder
	}
	return defaultTokenEncoder
}

func getTokenEncoder(model string) *tiktoken.Tiktoken {
	tokenEncoder, ok := tokenEncoderMap[model]
	if ok && tokenEncoder != nil {
		return tokenEncoder
	}
	// 如果ok（即model在tokenEncoderMap中），但是tokenEncoder为nil，说明可能是自定义模型
	if ok {
		tokenEncoder, err := tiktoken.EncodingForModel(model)
		if err != nil {
			common.SysError(fmt.Sprintf("failed to get token encoder for model %s: %s, using encoder for gpt-3.5-turbo", model, err.Error()))
			tokenEncoder = getModelDefaultTokenEncoder(model)
		}
		tokenEncoderMap[model] = tokenEncoder
		return tokenEncoder
	}
	// 如果model不在tokenEncoderMap中，直接返回默认的tokenEncoder
	return getModelDefaultTokenEncoder(model)
}

func getTokenNum(tokenEncoder *tiktoken.Tiktoken, text string) int {
	return len(tokenEncoder.Encode(text, nil, nil))
}

func getImageToken(imageUrl *dto.MessageImageUrl, model string, stream bool) (int, error) {
	baseTokens := 85
	if model == "glm-4v" {
		return 1047, nil
	}
	if imageUrl.Detail == "low" {
		return baseTokens, nil
	}
	// TODO: 非流模式下不计算图片token数量
	if !constant.GetMediaTokenNotStream && !stream {
		return 1000, nil
	}
	// 是否统计图片token
	if !constant.GetMediaToken {
		return 1000, nil
	}
	// 同步One API的图片计费逻辑
	if imageUrl.Detail == "auto" || imageUrl.Detail == "" {
		imageUrl.Detail = "high"
	}

	tileTokens := 170
	if strings.HasPrefix(model, "gpt-4o-mini") {
		tileTokens = 5667
		baseTokens = 2833
	}
	var config image.Config
	var err error
	var format string
	if strings.HasPrefix(imageUrl.Url, "http") {
		config, format, err = DecodeUrlImageData(imageUrl.Url)
	} else {
		common.SysLog(fmt.Sprintf("decoding image"))
		config, format, _, err = DecodeBase64ImageData(imageUrl.Url)
	}
	if err != nil {
		return 0, err
	}

	if config.Width == 0 || config.Height == 0 {
		return 0, errors.New(fmt.Sprintf("fail to decode image config: %s", imageUrl.Url))
	}
	//// TODO: 适配官方auto计费
	//if config.Width < 512 && config.Height < 512 {
	//	if imageUrl.Detail == "auto" || imageUrl.Detail == "" {
	//		// 如果图片尺寸小于512，强制使用low
	//		imageUrl.Detail = "low"
	//		return 85, nil
	//	}
	//}

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
	return tiles*tileTokens + baseTokens, nil
}

func CountTokenChatRequest(request dto.GeneralOpenAIRequest, model string) (int, error) {
	tkm := 0
	msgTokens, err := CountTokenMessages(request.Messages, model, request.Stream)
	if err != nil {
		return 0, err
	}
	tkm += msgTokens
	if request.Tools != nil {
		toolsData, _ := json.Marshal(request.Tools)
		var openaiTools []dto.OpenAITools
		err := json.Unmarshal(toolsData, &openaiTools)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("count_tools_token_fail: %s", err.Error()))
		}
		countStr := ""
		for _, tool := range openaiTools {
			countStr = tool.Function.Name
			if tool.Function.Description != "" {
				countStr += tool.Function.Description
			}
			if tool.Function.Parameters != nil {
				countStr += fmt.Sprintf("%v", tool.Function.Parameters)
			}
		}
		toolTokens, err := CountTokenInput(countStr, model)
		if err != nil {
			return 0, err
		}
		tkm += 8
		tkm += toolTokens
	}

	return tkm, nil
}

func CountTokenMessages(messages []dto.Message, model string, stream bool) (int, error) {
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
			if message.IsStringContent() {
				stringContent := message.StringContent()
				tokenNum += getTokenNum(tokenEncoder, stringContent)
				if message.Name != nil {
					tokenNum += tokensPerName
					tokenNum += getTokenNum(tokenEncoder, *message.Name)
				}
			} else {
				arrayContent := message.ParseContent()
				for _, m := range arrayContent {
					if m.Type == "image_url" {
						imageUrl := m.ImageUrl.(dto.MessageImageUrl)
						imageTokenNum, err := getImageToken(&imageUrl, model, stream)
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

func CountTokenInput(input any, model string) (int, error) {
	switch v := input.(type) {
	case string:
		return CountTokenText(v, model)
	case []string:
		text := ""
		for _, s := range v {
			text += s
		}
		return CountTokenText(text, model)
	}
	return CountTokenInput(fmt.Sprintf("%v", input), model)
}

func CountTokenStreamChoices(messages []dto.ChatCompletionsStreamResponseChoice, model string) int {
	tokens := 0
	for _, message := range messages {
		tkm, _ := CountTokenInput(message.Delta.GetContentString(), model)
		tokens += tkm
		if message.Delta.ToolCalls != nil {
			for _, tool := range message.Delta.ToolCalls {
				tkm, _ := CountTokenInput(tool.Function.Name, model)
				tokens += tkm
				tkm, _ = CountTokenInput(tool.Function.Arguments, model)
				tokens += tkm
			}
		}
	}
	return tokens
}

func CountAudioToken(text string, model string) (int, error) {
	if strings.HasPrefix(model, "tts") {
		return utf8.RuneCountInString(text), nil
	} else {
		return CountTokenText(text, model)
	}
}

// CountTokenText 统计文本的token数量，仅当文本包含敏感词，返回错误，同时返回token数量
func CountTokenText(text string, model string) (int, error) {
	var err error
	tokenEncoder := getTokenEncoder(model)
	return getTokenNum(tokenEncoder, text), err
}
