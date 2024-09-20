package constant

import (
	"fmt"
	"one-api/common"
	"os"
	"strings"
)

var StreamingTimeout = common.GetEnvOrDefault("STREAMING_TIMEOUT", 30)
var DifyDebug = common.GetEnvOrDefaultBool("DIFY_DEBUG", true)

// ForceStreamOption 覆盖请求参数，强制返回usage信息
var ForceStreamOption = common.GetEnvOrDefaultBool("FORCE_STREAM_OPTION", true)

var GetMediaToken = common.GetEnvOrDefaultBool("GET_MEDIA_TOKEN", true)

var GetMediaTokenNotStream = common.GetEnvOrDefaultBool("GET_MEDIA_TOKEN_NOT_STREAM", true)

var UpdateTask = common.GetEnvOrDefaultBool("UPDATE_TASK", true)

var GeminiModelMap = map[string]string{
	"gemini-1.5-pro-latest":     "v1beta",
	"gemini-1.5-pro-001":        "v1beta",
	"gemini-1.5-pro":            "v1beta",
	"gemini-1.5-pro-exp-0801":   "v1beta",
	"gemini-1.5-pro-exp-0827":   "v1beta",
	"gemini-1.5-flash-latest":   "v1beta",
	"gemini-1.5-flash-exp-0827": "v1beta",
	"gemini-1.5-flash-001":      "v1beta",
	"gemini-1.5-flash":          "v1beta",
	"gemini-ultra":              "v1beta",
}

// GeminiChatSafetySettingsArray 用于设置Gemini的安全性设置 测试
// https://ai.google.dev/api/generate-content#v1beta.HarmCategory
var GeminiChatSafetySettingsMap = map[string]string{}

var GeminiExpChatSafetySettingsMap = map[string]string{}

// 是否生成初始令牌，默认关闭。
var GenerateDefaultToken = common.GetEnvOrDefaultBool("GENERATE_DEFAULT_TOKEN", false)

func InitEnv() {
	InitGeminiEnv()
}

func InitGeminiEnv() {
	InitGeminiModelMap()
	InitGeminiChatSafetySettings()
	InitGeminiExpChatSafetySettings()
}

func InitGeminiModelMap() {
	modelVersionMapStr := strings.TrimSpace(os.Getenv("GEMINI_MODEL_MAP"))
	if modelVersionMapStr == "" {
		return
	}
	for _, pair := range strings.Split(modelVersionMapStr, ",") {
		parts := strings.Split(pair, ":")
		if len(parts) == 2 {
			GeminiModelMap[parts[0]] = parts[1]
		} else {
			common.SysError(fmt.Sprintf("invalid model version map: %s", pair))
		}
	}
}

func InitGeminiChatSafetySettings() {
	geminiSafetySettingArrayStr := strings.TrimSpace(os.Getenv("GEMINI_SAFETY_SETTINGS"))

	constants := []string{
		"HARM_CATEGORY_HARASSMENT",
		"HARM_CATEGORY_HATE_SPEECH",
		"HARM_CATEGORY_SEXUALLY_EXPLICIT",
		"HARM_CATEGORY_DANGEROUS_CONTENT",
	}

	if geminiSafetySettingArrayStr != "" {
		for _, pair := range strings.Split(geminiSafetySettingArrayStr, ",") {
			parts := strings.Split(pair, ":")
			category := parts[0]

			if found, index := Contains(constants, category); found {
				// remove from constants
				constants = append(constants[:index], constants[index+1:]...)
			}

			if len(parts) == 2 {
				GeminiChatSafetySettingsMap[category] = parts[1]

			} else {
				GeminiChatSafetySettingsMap[category] = common.GeminiSafetySetting
			}
		}
	}
	for _, category := range constants {
		GeminiChatSafetySettingsMap[category] = common.GeminiSafetySetting
	}
}

func InitGeminiExpChatSafetySettings() {

	geminiSafetySettingArrayStr := strings.TrimSpace(os.Getenv("GEMINI_EXP_SAFETY_SETTINGS"))

	constants := []string{
		"HARM_CATEGORY_CIVIC_INTEGRITY",
	}

	if geminiSafetySettingArrayStr != "" {
		for _, pair := range strings.Split(geminiSafetySettingArrayStr, ",") {
			parts := strings.Split(pair, ":")
			category := parts[0]

			if found, index := Contains(constants, category); found {
				// remove from constants
				constants = append(constants[:index], constants[index+1:]...)
			}

			if len(parts) == 2 {
				GeminiExpChatSafetySettingsMap[category] = parts[1]

			} else {
				GeminiExpChatSafetySettingsMap[category] = common.GeminiSafetySetting
			}
		}
	}
	for _, category := range constants {
		GeminiExpChatSafetySettingsMap[category] = common.GeminiSafetySetting
	}
}

func Contains(slice []string, item string) (bool, int) {
	for index, a := range slice {
		if a == item {
			return true, index
		}
	}
	return false, -1
}
