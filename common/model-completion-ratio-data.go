package common

import "strings"

var OpenAICompletionRatioDataDiff = map[string]float64{
	"gpt-4o-2024-05-13":  3, // 第一版4o 为 3倍
	"chatgpt-4o-latest":  3,
	"gpt-3.5-turbo-0125": 3, // 最后一版3.5 turbo 为 3倍
	"gpt-3.5-turbo-1106": 2, // 1106 为 2倍
}

func GetOpenAICompletionRatioData(name string) float64 {
	// 特殊倍率
	if v, ok := OpenAICompletionRatioDataDiff[name]; ok {
		return v
	}

	if strings.HasPrefix(name, "gpt-4-gizmo") {
		name = "gpt-4-gizmo-*"
	}
	if strings.HasPrefix(name, "gpt-4o-gizmo") {
		name = "gpt-4o-gizmo-*"
	}

	// 逆向大手子
	if strings.HasSuffix(name, "-all") || strings.HasSuffix(name, "-gizmo-*") {
		return 1
	}

	// 4o 大家族基本上为4倍 , 05-13和chatgpt-4o-latest为3倍
	if strings.HasPrefix(name, "gpt-4o") {
		return 4
	}

	// o1
	if strings.HasPrefix(name, "o1-") {
		return 4
	}

	// gpt-4 大家族基本上为2倍 带preview的为3倍
	if strings.HasPrefix(name, "gpt-4") {
		if strings.HasSuffix(name, "-preview") {
			return 3
		}
		return 2
	}

	// gpt-4-turbo 为3倍
	if strings.HasPrefix(name, "gpt-4-turbo") {
		return 3
	}

	// gpt-3.5 老版本均为1.33 (3/4)
	if strings.HasPrefix(name, "gpt-3.5") {
		return 3.0 / 4.0
	}

	// 默认倍率
	return 1

}
