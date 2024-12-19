package common

import (
	"encoding/json"
	"strings"
	"sync"
)

var (
	modelPriceMap      map[string]float64 = nil
	modelPriceMapMutex                    = sync.RWMutex{}
)
var (
	modelRatioMap      map[string]float64 = nil
	modelRatioMapMutex                    = sync.RWMutex{}
)

var CompletionRatio map[string]float64 = nil
var defaultCompletionRatio = map[string]float64{
	"gpt-4-gizmo-*":  2,
	"gpt-4o-gizmo-*": 3,
	"gpt-4-all":      2,
}

func GetModelPriceMap() map[string]float64 {
	modelPriceMapMutex.Lock()
	defer modelPriceMapMutex.Unlock()
	if modelPriceMap == nil {
		modelPriceMap = defaultModelPrice
	}
	return modelPriceMap
}

func ModelPrice2JSONString() string {
	GetModelPriceMap()
	jsonBytes, err := json.Marshal(modelPriceMap)
	if err != nil {
		SysError("error marshalling model price: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelPriceByJSONString(jsonStr string) error {
	modelPriceMapMutex.Lock()
	defer modelPriceMapMutex.Unlock()
	modelPriceMap = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &modelPriceMap)
}

// GetModelPrice 返回模型的价格，如果模型不存在则返回-1，false
func GetModelPrice(name string, printErr bool) (float64, bool) {
	GetModelPriceMap()
	if strings.HasPrefix(name, "gpt-4-gizmo") {
		name = "gpt-4-gizmo-*"
	}
	if strings.HasPrefix(name, "gpt-4o-gizmo") {
		name = "gpt-4o-gizmo-*"
	}
	price, ok := modelPriceMap[name]
	if !ok {
		if printErr {
			SysError("model price not found: " + name)
		}
		return -1, false
	}
	return price, true
}

func GetModelRatioMap() map[string]float64 {
	modelRatioMapMutex.Lock()
	defer modelRatioMapMutex.Unlock()
	if modelRatioMap == nil {
		modelRatioMap = defaultModelRatio
	}
	return modelRatioMap
}

func ModelRatio2JSONString() string {
	GetModelRatioMap()
	jsonBytes, err := json.Marshal(modelRatioMap)
	if err != nil {
		SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRatioByJSONString(jsonStr string) error {
	modelRatioMapMutex.Lock()
	defer modelRatioMapMutex.Unlock()
	modelRatioMap = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &modelRatioMap)
}

func GetModelRatio(name string) float64 {
	GetModelRatioMap()
	if strings.HasPrefix(name, "gpt-4-gizmo") {
		name = "gpt-4-gizmo-*"
	}
	ratio, ok := modelRatioMap[name]
	if !ok {
		SysError("model ratio not found: " + name)
		return 30
	}
	return ratio
}

func DefaultModelRatio2JSONString() string {
	jsonBytes, err := json.Marshal(defaultModelRatio)
	if err != nil {
		SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func GetDefaultModelRatioMap() map[string]float64 {
	return defaultModelRatio
}

func CompletionRatio2JSONString() string {
	if CompletionRatio == nil {
		CompletionRatio = defaultCompletionRatio
	}
	jsonBytes, err := json.Marshal(CompletionRatio)
	if err != nil {
		SysError("error marshalling completion ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateCompletionRatioByJSONString(jsonStr string) error {
	CompletionRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &CompletionRatio)
}

func GetCompletionRatio(name string) float64 {
	// Custom Completion Ratio
	if ratio, ok := CompletionRatio[name]; ok && ratio > 0 {
		return ratio
	}

	// OpenAI Models
	if strings.HasPrefix(name, "gpt") || strings.HasPrefix(name, "chatgpt") || strings.HasPrefix(name, "o1") {
		return getOpenAICompletionRatioData(name)
	}

	// Anthropic Models
	if strings.Contains(name, "claude-3") {
		// V3 v3.5 均为5倍
		return 5
	}
	if strings.Contains(name, "claude-2") || strings.Contains(name, "claude-instant-1") {
		// V1 V2 为3倍
		return 3
	}

	// Mistral Models
	if strings.HasPrefix(name, "mistral-") {
		return 3
	}

	// Google Models
	if strings.HasPrefix(name, "gemini-") {
		return 4
	}

	// Cohere Models
	if strings.HasPrefix(name, "command") {
		return getCohereCompletionRatioData(name)
	}

	// Deepseek Models
	if strings.HasPrefix(name, "deepseek") {
		return 2
	}

	// ERNIE Models
	if strings.HasPrefix(name, "ERNIE-") {
		return getERNIECompletionRatioData(name)
	}

	switch name {
	case "llama2-70b-4096":
		return 0.8 / 0.64
	case "llama3-8b-8192":
		return 2
	case "llama3-70b-8192":
		return 0.79 / 0.59
	}

	return 1
}

func GetAudioRatio(name string) float64 {
	if strings.HasPrefix(name, "gpt-4o-realtime") {
		return 20
	} else if strings.HasPrefix(name, "gpt-4o-audio") {
		return 40
	}
	return 20
}

func GetAudioCompletionRatio(name string) float64 {
	if strings.HasPrefix(name, "gpt-4o-realtime") {
		return 2
	}
	return 2
}

//func GetAudioPricePerMinute(name string) float64 {
//	if strings.HasPrefix(name, "gpt-4o-realtime") {
//		return 0.06
//	}
//	return 0.06
//}
//
//func GetAudioCompletionPricePerMinute(name string) float64 {
//	if strings.HasPrefix(name, "gpt-4o-realtime") {
//		return 0.24
//	}
//	return 0.24
//}

func GetCompletionRatioMap() map[string]float64 {
	if CompletionRatio == nil {
		CompletionRatio = defaultCompletionRatio
	}
	return CompletionRatio
}
