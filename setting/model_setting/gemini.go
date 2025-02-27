package model_setting

import (
	"one-api/setting/config"
)

// GeminiSettings 定义Gemini模型的配置
type GeminiSettings struct {
	SafetySettings  map[string]string `json:"safety_settings"`
	VersionSettings map[string]string `json:"version_settings"`
}

// 默认配置
var defaultGeminiSettings = GeminiSettings{
	SafetySettings: map[string]string{
		"default":                       "OFF",
		"HARM_CATEGORY_CIVIC_INTEGRITY": "BLOCK_NONE",
	},
	VersionSettings: map[string]string{
		"default":        "v1beta",
		"gemini-1.0-pro": "v1",
	},
}

// 全局实例
var geminiSettings = defaultGeminiSettings

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("gemini", &geminiSettings)
}

// GetGeminiSettings 获取Gemini配置
func GetGeminiSettings() *GeminiSettings {
	return &geminiSettings
}

// GetGeminiSafetySetting 获取安全设置
func GetGeminiSafetySetting(key string) string {
	if value, ok := geminiSettings.SafetySettings[key]; ok {
		return value
	}
	return geminiSettings.SafetySettings["default"]
}

// GetGeminiVersionSetting 获取版本设置
func GetGeminiVersionSetting(key string) string {
	if value, ok := geminiSettings.VersionSettings[key]; ok {
		return value
	}
	return geminiSettings.VersionSettings["default"]
}
