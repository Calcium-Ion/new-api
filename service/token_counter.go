package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"log"
	"math"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/setting/operation_setting"
	"strings"
	"unicode/utf8"

	"github.com/pkoukk/tiktoken-go"
)

// tokenEncoderMap won't grow after initialization
var tokenEncoderMap = map[string]*tiktoken.Tiktoken{}
var defaultTokenEncoder *tiktoken.Tiktoken
var o200kTokenEncoder *tiktoken.Tiktoken

func InitTokenEncoders() {
	common.SysLog("initializing token encoders")
	cl100TokenEncoder, err := tiktoken.GetEncoding(tiktoken.MODEL_CL100K_BASE)
	if err != nil {
		common.FatalLog(fmt.Sprintf("failed to get gpt-3.5-turbo token encoder: %s", err.Error()))
	}
	defaultTokenEncoder = cl100TokenEncoder
	o200kTokenEncoder, err = tiktoken.GetEncoding(tiktoken.MODEL_O200K_BASE)
	if err != nil {
		common.FatalLog(fmt.Sprintf("failed to get gpt-4o token encoder: %s", err.Error()))
	}
	for model, _ := range operation_setting.GetDefaultModelRatioMap() {
		if strings.HasPrefix(model, "gpt-3.5") {
			tokenEncoderMap[model] = cl100TokenEncoder
		} else if strings.HasPrefix(model, "gpt-4") {
			if strings.HasPrefix(model, "gpt-4o") {
				tokenEncoderMap[model] = o200kTokenEncoder
			} else {
				tokenEncoderMap[model] = defaultTokenEncoder
			}
		} else if strings.HasPrefix(model, "o1") {
			tokenEncoderMap[model] = o200kTokenEncoder
		} else {
			tokenEncoderMap[model] = defaultTokenEncoder
		}
	}
	common.SysLog("token encoders initialized")
}

func getModelDefaultTokenEncoder(model string) *tiktoken.Tiktoken {
	if strings.HasPrefix(model, "gpt-4o") || strings.HasPrefix(model, "chatgpt-4o") || strings.HasPrefix(model, "o1") {
		return o200kTokenEncoder
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
	if text == "" {
		return 0
	}
	return len(tokenEncoder.Encode(text, nil, nil))
}

func getImageToken(info *relaycommon.RelayInfo, imageUrl *dto.MessageImageUrl, model string, stream bool) (int, error) {
	if imageUrl == nil {
		return 0, fmt.Errorf("image_url_is_nil")
	}
	baseTokens := 85
	if model == "glm-4v" {
		return 1047, nil
	}
	if imageUrl.Detail == "low" {
		return baseTokens, nil
	}
	if !constant.GetMediaTokenNotStream && !stream {
		return 3 * baseTokens, nil
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
	// 是否统计图片token
	if !constant.GetMediaToken {
		return 3 * baseTokens, nil
	}
	if info.ChannelType == common.ChannelTypeGemini || info.ChannelType == common.ChannelTypeVertexAi || info.ChannelType == common.ChannelTypeAnthropic {
		return 3 * baseTokens, nil
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
	imageUrl.MimeType = format

	if config.Width == 0 || config.Height == 0 {
		return 0, errors.New(fmt.Sprintf("fail to decode image config: %s", imageUrl.Url))
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
	return tiles*tileTokens + baseTokens, nil
}

func CountTokenChatRequest(info *relaycommon.RelayInfo, request dto.GeneralOpenAIRequest) (int, error) {
	tkm := 0
	msgTokens, err := CountTokenMessages(info, request.Messages, request.Model, request.Stream)
	if err != nil {
		return 0, err
	}
	tkm += msgTokens
	if request.Tools != nil {
		openaiTools := request.Tools
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
		toolTokens, err := CountTokenInput(countStr, request.Model)
		if err != nil {
			return 0, err
		}
		tkm += 8
		tkm += toolTokens
	}

	return tkm, nil
}

func CountTokenClaudeRequest(request dto.ClaudeRequest, model string) (int, error) {
	tkm := 0

	// Count tokens in messages
	msgTokens, err := CountTokenClaudeMessages(request.Messages, model, request.Stream)
	if err != nil {
		return 0, err
	}
	tkm += msgTokens

	// Count tokens in system message
	if request.System != "" {
		systemTokens, err := CountTokenInput(request.System, model)
		if err != nil {
			return 0, err
		}
		tkm += systemTokens
	}

	if request.Tools != nil {
		// check is array
		if tools, ok := request.Tools.([]any); ok {
			if len(tools) > 0 {
				parsedTools, err1 := common.Any2Type[[]dto.Tool](request.Tools)
				if err1 != nil {
					return 0, fmt.Errorf("tools: Input should be a valid list: %v", err)
				}
				toolTokens, err2 := CountTokenClaudeTools(parsedTools, model)
				if err2 != nil {
					return 0, fmt.Errorf("tools: %v", err)
				}
				tkm += toolTokens
			}
		} else {
			return 0, errors.New("tools: Input should be a valid list")
		}
	}

	return tkm, nil
}

func CountTokenClaudeMessages(messages []dto.ClaudeMessage, model string, stream bool) (int, error) {
	tokenEncoder := getTokenEncoder(model)
	tokenNum := 0

	for _, message := range messages {
		// Count tokens for role
		tokenNum += getTokenNum(tokenEncoder, message.Role)
		if message.IsStringContent() {
			tokenNum += getTokenNum(tokenEncoder, message.GetStringContent())
		} else {
			content, err := message.ParseContent()
			if err != nil {
				return 0, err
			}
			for _, mediaMessage := range content {
				switch mediaMessage.Type {
				case "text":
					tokenNum += getTokenNum(tokenEncoder, mediaMessage.GetText())
				case "image":
					//imageTokenNum, err := getClaudeImageToken(mediaMsg.Source, model, stream)
					//if err != nil {
					//	return 0, err
					//}
					tokenNum += 1000
				case "tool_use":
					tokenNum += getTokenNum(tokenEncoder, mediaMessage.Name)
					inputJSON, _ := json.Marshal(mediaMessage.Input)
					tokenNum += getTokenNum(tokenEncoder, string(inputJSON))
				case "tool_result":
					contentJSON, _ := json.Marshal(mediaMessage.Content)
					tokenNum += getTokenNum(tokenEncoder, string(contentJSON))
				}
			}
		}
	}

	// Add a constant for message formatting (this may need adjustment based on Claude's exact formatting)
	tokenNum += len(messages) * 2 // Assuming 2 tokens per message for formatting

	return tokenNum, nil
}

func CountTokenClaudeTools(tools []dto.Tool, model string) (int, error) {
	tokenEncoder := getTokenEncoder(model)
	tokenNum := 0

	for _, tool := range tools {
		tokenNum += getTokenNum(tokenEncoder, tool.Name)
		tokenNum += getTokenNum(tokenEncoder, tool.Description)

		schemaJSON, err := json.Marshal(tool.InputSchema)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("marshal_tool_schema_fail: %s", err.Error()))
		}
		tokenNum += getTokenNum(tokenEncoder, string(schemaJSON))
	}

	// Add a constant for tool formatting (this may need adjustment based on Claude's exact formatting)
	tokenNum += len(tools) * 3 // Assuming 3 tokens per tool for formatting

	return tokenNum, nil
}

func CountTokenRealtime(info *relaycommon.RelayInfo, request dto.RealtimeEvent, model string) (int, int, error) {
	audioToken := 0
	textToken := 0
	switch request.Type {
	case dto.RealtimeEventTypeSessionUpdate:
		if request.Session != nil {
			msgTokens, err := CountTextToken(request.Session.Instructions, model)
			if err != nil {
				return 0, 0, err
			}
			textToken += msgTokens
		}
	case dto.RealtimeEventResponseAudioDelta:
		// count audio token
		atk, err := CountAudioTokenOutput(request.Delta, info.OutputAudioFormat)
		if err != nil {
			return 0, 0, fmt.Errorf("error counting audio token: %v", err)
		}
		audioToken += atk
	case dto.RealtimeEventResponseAudioTranscriptionDelta, dto.RealtimeEventResponseFunctionCallArgumentsDelta:
		// count text token
		tkm, err := CountTextToken(request.Delta, model)
		if err != nil {
			return 0, 0, fmt.Errorf("error counting text token: %v", err)
		}
		textToken += tkm
	case dto.RealtimeEventInputAudioBufferAppend:
		// count audio token
		atk, err := CountAudioTokenInput(request.Audio, info.InputAudioFormat)
		if err != nil {
			return 0, 0, fmt.Errorf("error counting audio token: %v", err)
		}
		audioToken += atk
	case dto.RealtimeEventConversationItemCreated:
		if request.Item != nil {
			switch request.Item.Type {
			case "message":
				for _, content := range request.Item.Content {
					if content.Type == "input_text" {
						tokens, err := CountTextToken(content.Text, model)
						if err != nil {
							return 0, 0, err
						}
						textToken += tokens
					}
				}
			}
		}
	case dto.RealtimeEventTypeResponseDone:
		// count tools token
		if !info.IsFirstRequest {
			if info.RealtimeTools != nil && len(info.RealtimeTools) > 0 {
				for _, tool := range info.RealtimeTools {
					toolTokens, err := CountTokenInput(tool, model)
					if err != nil {
						return 0, 0, err
					}
					textToken += 8
					textToken += toolTokens
				}
			}
		}
	}
	return textToken, audioToken, nil
}

func CountTokenMessages(info *relaycommon.RelayInfo, messages []dto.Message, model string, stream bool) (int, error) {
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
			if message.Name != nil {
				tokenNum += tokensPerName
				tokenNum += getTokenNum(tokenEncoder, *message.Name)
			}
			arrayContent := message.ParseContent()
			for _, m := range arrayContent {
				if m.Type == dto.ContentTypeImageURL {
					imageUrl := m.GetImageMedia()
					imageTokenNum, err := getImageToken(info, imageUrl, model, stream)
					if err != nil {
						return 0, err
					}
					tokenNum += imageTokenNum
					log.Printf("image token num: %d", imageTokenNum)
				} else if m.Type == dto.ContentTypeInputAudio {
					// TODO: 音频token数量计算
					tokenNum += 100
				} else if m.Type == dto.ContentTypeFile {
					tokenNum += 5000
				} else {
					tokenNum += getTokenNum(tokenEncoder, m.Text)
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
		return CountTextToken(v, model)
	case []string:
		text := ""
		for _, s := range v {
			text += s
		}
		return CountTextToken(text, model)
	case []interface{}:
		text := ""
		for _, item := range v {
			text += fmt.Sprintf("%v", item)
		}
		return CountTextToken(text, model)
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

func CountTTSToken(text string, model string) (int, error) {
	if strings.HasPrefix(model, "tts") {
		return utf8.RuneCountInString(text), nil
	} else {
		return CountTextToken(text, model)
	}
}

func CountAudioTokenInput(audioBase64 string, audioFormat string) (int, error) {
	if audioBase64 == "" {
		return 0, nil
	}
	duration, err := parseAudio(audioBase64, audioFormat)
	if err != nil {
		return 0, err
	}
	return int(duration / 60 * 100 / 0.06), nil
}

func CountAudioTokenOutput(audioBase64 string, audioFormat string) (int, error) {
	if audioBase64 == "" {
		return 0, nil
	}
	duration, err := parseAudio(audioBase64, audioFormat)
	if err != nil {
		return 0, err
	}
	return int(duration / 60 * 200 / 0.24), nil
}

//func CountAudioToken(sec float64, audioType string) {
//	if audioType == "input" {
//
//	}
//}

// CountTextToken 统计文本的token数量，仅当文本包含敏感词，返回错误，同时返回token数量
func CountTextToken(text string, model string) (int, error) {
	var err error
	tokenEncoder := getTokenEncoder(model)
	return getTokenNum(tokenEncoder, text), err
}
