package model_setting

import (
	"encoding/json"
	"one-api/common"
)

var geminiSafetySettings = map[string]string{
	"default":                       "OFF",
	"HARM_CATEGORY_CIVIC_INTEGRITY": "BLOCK_NONE",
}

func GetGeminiSafetySetting(key string) string {
	if value, ok := geminiSafetySettings[key]; ok {
		return value
	}
	return geminiSafetySettings["default"]
}

func GeminiSafetySettingFromJsonString(jsonString string) {
	geminiSafetySettings = map[string]string{}
	err := json.Unmarshal([]byte(jsonString), &geminiSafetySettings)
	if err != nil {
		geminiSafetySettings = map[string]string{
			"default":                       "OFF",
			"HARM_CATEGORY_CIVIC_INTEGRITY": "BLOCK_NONE",
		}
	}
	// check must have default
	if _, ok := geminiSafetySettings["default"]; !ok {
		geminiSafetySettings["default"] = common.GeminiSafetySetting
	}
}

func GeminiSafetySettingsJsonString() string {
	// check must have default
	if _, ok := geminiSafetySettings["default"]; !ok {
		geminiSafetySettings["default"] = common.GeminiSafetySetting
	}
	jsonString, err := json.Marshal(geminiSafetySettings)
	if err != nil {
		return "{}"
	}
	return string(jsonString)
}

var geminiVersionSettings = map[string]string{
	"default":        "v1beta",
	"gemini-1.0-pro": "v1",
}

func GetGeminiVersionSetting(key string) string {
	if value, ok := geminiVersionSettings[key]; ok {
		return value
	}
	return geminiVersionSettings["default"]
}

func GeminiVersionSettingFromJsonString(jsonString string) {
	geminiVersionSettings = map[string]string{}
	err := json.Unmarshal([]byte(jsonString), &geminiVersionSettings)
	if err != nil {
		geminiVersionSettings = map[string]string{
			"default": "v1beta",
		}
	}
	// check must have default
	if _, ok := geminiVersionSettings["default"]; !ok {
		geminiVersionSettings["default"] = "v1beta"
	}
}

func GeminiVersionSettingsJsonString() string {
	// check must have default
	if _, ok := geminiVersionSettings["default"]; !ok {
		geminiVersionSettings["default"] = "v1beta"
	}
	jsonString, err := json.Marshal(geminiVersionSettings)
	if err != nil {
		return "{}"
	}
	return string(jsonString)
}
