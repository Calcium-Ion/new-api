package model_setting

import (
	"net/http"
	"one-api/setting/config"
)

//var claudeHeadersSettings = map[string][]string{}
//
//var ClaudeThinkingAdapterEnabled = true
//var ClaudeThinkingAdapterMaxTokens = 8192
//var ClaudeThinkingAdapterBudgetTokensPercentage = 0.8

// ClaudeSettings 定义Claude模型的配置
type ClaudeSettings struct {
	HeadersSettings                       map[string][]string `json:"headers_settings"`
	ThinkingAdapterEnabled                bool                `json:"thinking_adapter_enabled"`
	ThinkingAdapterMaxTokens              int                 `json:"thinking_adapter_max_tokens"`
	ThinkingAdapterBudgetTokensPercentage float64             `json:"thinking_adapter_budget_tokens_percentage"`
}

// 默认配置
var defaultClaudeSettings = ClaudeSettings{
	HeadersSettings:                       map[string][]string{},
	ThinkingAdapterEnabled:                true,
	ThinkingAdapterMaxTokens:              8192,
	ThinkingAdapterBudgetTokensPercentage: 0.8,
}

// 全局实例
var claudeSettings = defaultClaudeSettings

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("claude", &claudeSettings)
}

// GetClaudeSettings 获取Claude配置
func GetClaudeSettings() *ClaudeSettings {
	return &claudeSettings
}

func (c *ClaudeSettings) WriteHeaders(headers *http.Header) {
	for key, values := range c.HeadersSettings {
		headers.Del(key)
		for _, value := range values {
			headers.Add(key, value)
		}
	}
}
