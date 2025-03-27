package operation_setting

import (
	"encoding/json"
	"one-api/common"
	"strings"
	"sync"
)

// from songquanpeng/one-api
const (
	USD2RMB = 7.3 // 暂定 1 USD = 7.3 RMB
	USD     = 500 // $0.002 = 1 -> $1 = 500
	RMB     = USD / USD2RMB
)

// modelRatio
// https://platform.openai.com/docs/models/model-endpoint-compatibility
// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Blfmc9dlf
// https://openai.com/pricing
// TODO: when a new api is enabled, check the pricing here
// 1 === $0.002 / 1K tokens
// 1 === ￥0.014 / 1k tokens

var defaultModelRatio = map[string]float64{
	//"midjourney":                50,
	"gpt-4-gizmo-*":  15,
	"gpt-4o-gizmo-*": 2.5,
	"gpt-4-all":      15,
	"gpt-4o-all":     15,
	"gpt-4":          15,
	//"gpt-4-0314":                   15, //deprecated
	"gpt-4-0613": 15,
	"gpt-4-32k":  30,
	//"gpt-4-32k-0314":               30, //deprecated
	"gpt-4-32k-0613":                          30,
	"gpt-4-1106-preview":                      5,    // $10 / 1M tokens
	"gpt-4-0125-preview":                      5,    // $10 / 1M tokens
	"gpt-4-turbo-preview":                     5,    // $10 / 1M tokens
	"gpt-4-vision-preview":                    5,    // $10 / 1M tokens
	"gpt-4-1106-vision-preview":               5,    // $10 / 1M tokens
	"chatgpt-4o-latest":                       2.5,  // $5 / 1M tokens
	"gpt-4o":                                  1.25, // $2.5 / 1M tokens
	"gpt-4o-audio-preview":                    1.25, // $2.5 / 1M tokens
	"gpt-4o-audio-preview-2024-10-01":         1.25, // $2.5 / 1M tokens
	"gpt-4o-2024-05-13":                       2.5,  // $5 / 1M tokens
	"gpt-4o-2024-08-06":                       1.25, // $2.5 / 1M tokens
	"gpt-4o-2024-11-20":                       1.25, // $2.5 / 1M tokens
	"gpt-4o-realtime-preview":                 2.5,
	"gpt-4o-realtime-preview-2024-10-01":      2.5,
	"gpt-4o-realtime-preview-2024-12-17":      2.5,
	"gpt-4o-mini-realtime-preview":            0.3,
	"gpt-4o-mini-realtime-preview-2024-12-17": 0.3,
	"o1":                         7.5,
	"o1-2024-12-17":              7.5,
	"o1-preview":                 7.5,
	"o1-preview-2024-09-12":      7.5,
	"o1-mini":                    0.55,
	"o1-mini-2024-09-12":         0.55,
	"o3-mini":                    0.55,
	"o3-mini-2025-01-31":         0.55,
	"o3-mini-high":               0.55,
	"o3-mini-2025-01-31-high":    0.55,
	"o3-mini-low":                0.55,
	"o3-mini-2025-01-31-low":     0.55,
	"o3-mini-medium":             0.55,
	"o3-mini-2025-01-31-medium":  0.55,
	"gpt-4o-mini":                0.075,
	"gpt-4o-mini-2024-07-18":     0.075,
	"gpt-4-turbo":                5, // $0.01 / 1K tokens
	"gpt-4-turbo-2024-04-09":     5, // $0.01 / 1K tokens
	"gpt-4.5-preview":            37.5,
	"gpt-4.5-preview-2025-02-27": 37.5,
	//"gpt-3.5-turbo-0301":           0.75, //deprecated
	"gpt-3.5-turbo":          0.25,
	"gpt-3.5-turbo-0613":     0.75,
	"gpt-3.5-turbo-16k":      1.5, // $0.003 / 1K tokens
	"gpt-3.5-turbo-16k-0613": 1.5,
	"gpt-3.5-turbo-instruct": 0.75, // $0.0015 / 1K tokens
	"gpt-3.5-turbo-1106":     0.5,  // $0.001 / 1K tokens
	"gpt-3.5-turbo-0125":     0.25,
	"babbage-002":            0.2, // $0.0004 / 1K tokens
	"davinci-002":            1,   // $0.002 / 1K tokens
	"text-ada-001":           0.2,
	"text-babbage-001":       0.25,
	"text-curie-001":         1,
	//"text-davinci-002":               10,
	//"text-davinci-003":               10,
	"text-davinci-edit-001":               10,
	"code-davinci-edit-001":               10,
	"whisper-1":                           15,  // $0.006 / minute -> $0.006 / 150 words -> $0.006 / 200 tokens -> $0.03 / 1k tokens
	"tts-1":                               7.5, // 1k characters -> $0.015
	"tts-1-1106":                          7.5, // 1k characters -> $0.015
	"tts-1-hd":                            15,  // 1k characters -> $0.03
	"tts-1-hd-1106":                       15,  // 1k characters -> $0.03
	"davinci":                             10,
	"curie":                               10,
	"babbage":                             10,
	"ada":                                 10,
	"text-embedding-3-small":              0.01,
	"text-embedding-3-large":              0.065,
	"text-embedding-ada-002":              0.05,
	"text-search-ada-doc-001":             10,
	"text-moderation-stable":              0.1,
	"text-moderation-latest":              0.1,
	"claude-instant-1":                    0.4,   // $0.8 / 1M tokens
	"claude-2.0":                          4,     // $8 / 1M tokens
	"claude-2.1":                          4,     // $8 / 1M tokens
	"claude-3-haiku-20240307":             0.125, // $0.25 / 1M tokens
	"claude-3-5-haiku-20241022":           0.5,   // $1 / 1M tokens
	"claude-3-sonnet-20240229":            1.5,   // $3 / 1M tokens
	"claude-3-5-sonnet-20240620":          1.5,
	"claude-3-5-sonnet-20241022":          1.5,
	"claude-3-7-sonnet-20250219":          1.5,
	"claude-3-7-sonnet-20250219-thinking": 1.5,
	"claude-3-opus-20240229":              7.5, // $15 / 1M tokens
	"ERNIE-4.0-8K":                        0.120 * RMB,
	"ERNIE-3.5-8K":                        0.012 * RMB,
	"ERNIE-3.5-8K-0205":                   0.024 * RMB,
	"ERNIE-3.5-8K-1222":                   0.012 * RMB,
	"ERNIE-Bot-8K":                        0.024 * RMB,
	"ERNIE-3.5-4K-0205":                   0.012 * RMB,
	"ERNIE-Speed-8K":                      0.004 * RMB,
	"ERNIE-Speed-128K":                    0.004 * RMB,
	"ERNIE-Lite-8K-0922":                  0.008 * RMB,
	"ERNIE-Lite-8K-0308":                  0.003 * RMB,
	"ERNIE-Tiny-8K":                       0.001 * RMB,
	"BLOOMZ-7B":                           0.004 * RMB,
	"Embedding-V1":                        0.002 * RMB,
	"bge-large-zh":                        0.002 * RMB,
	"bge-large-en":                        0.002 * RMB,
	"tao-8k":                              0.002 * RMB,
	"PaLM-2":                              1,
	"gemini-pro":                          1, // $0.00025 / 1k characters -> $0.001 / 1k tokens
	"gemini-pro-vision":                   1, // $0.00025 / 1k characters -> $0.001 / 1k tokens
	"gemini-1.0-pro-vision-001":           1,
	"gemini-1.0-pro-001":                  1,
	"gemini-1.5-pro-latest":               1.75, // $3.5 / 1M tokens
	"gemini-1.5-pro-exp-0827":             1.75, // $3.5 / 1M tokens
	"gemini-1.5-flash-latest":             1,
	"gemini-1.5-flash-exp-0827":           1,
	"gemini-1.0-pro-latest":               1,
	"gemini-1.0-pro-vision-latest":        1,
	"gemini-ultra":                        1,
	"chatglm_turbo":                       0.3572,     // ￥0.005 / 1k tokens
	"chatglm_pro":                         0.7143,     // ￥0.01 / 1k tokens
	"chatglm_std":                         0.3572,     // ￥0.005 / 1k tokens
	"chatglm_lite":                        0.1429,     // ￥0.002 / 1k tokens
	"glm-4":                               7.143,      // ￥0.1 / 1k tokens
	"glm-4v":                              0.05 * RMB, // ￥0.05 / 1k tokens
	"glm-4-alltools":                      0.1 * RMB,  // ￥0.1 / 1k tokens
	"glm-3-turbo":                         0.3572,
	"glm-4-plus":                          0.05 * RMB,
	"glm-4-0520":                          0.1 * RMB,
	"glm-4-air":                           0.001 * RMB,
	"glm-4-airx":                          0.01 * RMB,
	"glm-4-long":                          0.001 * RMB,
	"glm-4-flash":                         0,
	"glm-4v-plus":                         0.01 * RMB,
	"qwen-turbo":                          0.8572, // ￥0.012 / 1k tokens
	"qwen-plus":                           10,     // ￥0.14 / 1k tokens
	"text-embedding-v1":                   0.05,   // ￥0.0007 / 1k tokens
	"SparkDesk-v1.1":                      1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v2.1":                      1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.1":                      1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.5":                      1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v4.0":                      1.2858,
	"360GPT_S2_V9":                        0.8572, // ¥0.012 / 1k tokens
	"360gpt-turbo":                        0.0858, // ¥0.0012 / 1k tokens
	"360gpt-turbo-responsibility-8k":      0.8572, // ¥0.012 / 1k tokens
	"360gpt-pro":                          0.8572, // ¥0.012 / 1k tokens
	"360gpt2-pro":                         0.8572, // ¥0.012 / 1k tokens
	"embedding-bert-512-v1":               0.0715, // ¥0.001 / 1k tokens
	"embedding_s1_v1":                     0.0715, // ¥0.001 / 1k tokens
	"semantic_similarity_s1_v1":           0.0715, // ¥0.001 / 1k tokens
	"hunyuan":                             7.143,  // ¥0.1 / 1k tokens  // https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
	// https://platform.lingyiwanwu.com/docs#-计费单元
	// 已经按照 7.2 来换算美元价格
	"yi-34b-chat-0205":       0.18,
	"yi-34b-chat-200k":       0.864,
	"yi-vl-plus":             0.432,
	"yi-large":               20.0 / 1000 * RMB,
	"yi-medium":              2.5 / 1000 * RMB,
	"yi-vision":              6.0 / 1000 * RMB,
	"yi-medium-200k":         12.0 / 1000 * RMB,
	"yi-spark":               1.0 / 1000 * RMB,
	"yi-large-rag":           25.0 / 1000 * RMB,
	"yi-large-turbo":         12.0 / 1000 * RMB,
	"yi-large-preview":       20.0 / 1000 * RMB,
	"yi-large-rag-preview":   25.0 / 1000 * RMB,
	"command":                0.5,
	"command-nightly":        0.5,
	"command-light":          0.5,
	"command-light-nightly":  0.5,
	"command-r":              0.25,
	"command-r-plus":         1.5,
	"command-r-08-2024":      0.075,
	"command-r-plus-08-2024": 1.25,
	"deepseek-chat":          0.27 / 2,
	"deepseek-coder":         0.27 / 2,
	"deepseek-reasoner":      0.55 / 2, // 0.55 / 1k tokens
	// Perplexity online 模型对搜索额外收费，有需要应自行调整，此处不计入搜索费用
	"llama-3-sonar-small-32k-chat":   0.2 / 1000 * USD,
	"llama-3-sonar-small-32k-online": 0.2 / 1000 * USD,
	"llama-3-sonar-large-32k-chat":   1 / 1000 * USD,
	"llama-3-sonar-large-32k-online": 1 / 1000 * USD,
}

var defaultModelPrice = map[string]float64{
	"suno_music":        0.1,
	"suno_lyrics":       0.01,
	"dall-e-3":          0.04,
	"gpt-4-gizmo-*":     0.1,
	"mj_imagine":        0.1,
	"mj_variation":      0.1,
	"mj_reroll":         0.1,
	"mj_blend":          0.1,
	"mj_modal":          0.1,
	"mj_zoom":           0.1,
	"mj_shorten":        0.1,
	"mj_high_variation": 0.1,
	"mj_low_variation":  0.1,
	"mj_pan":            0.1,
	"mj_inpaint":        0,
	"mj_custom_zoom":    0,
	"mj_describe":       0.05,
	"mj_upscale":        0.05,
	"swap_face":         0.05,
	"mj_upload":         0.05,
}

var (
	modelPriceMap      map[string]float64 = nil
	modelPriceMapMutex                    = sync.RWMutex{}
)
var (
	modelRatioMap      map[string]float64 = nil
	modelRatioMapMutex                    = sync.RWMutex{}
)

var (
	CompletionRatio      map[string]float64 = nil
	CompletionRatioMutex                    = sync.RWMutex{}
)

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
		common.SysError("error marshalling model price: " + err.Error())
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
			common.SysError("model price not found: " + name)
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
		common.SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRatioByJSONString(jsonStr string) error {
	modelRatioMapMutex.Lock()
	defer modelRatioMapMutex.Unlock()
	modelRatioMap = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &modelRatioMap)
}

func GetModelRatio(name string) (float64, bool) {
	GetModelRatioMap()
	if strings.HasPrefix(name, "gpt-4-gizmo") {
		name = "gpt-4-gizmo-*"
	}
	ratio, ok := modelRatioMap[name]
	if !ok {
		return 37.5, SelfUseModeEnabled
	}
	return ratio, true
}

func DefaultModelRatio2JSONString() string {
	jsonBytes, err := json.Marshal(defaultModelRatio)
	if err != nil {
		common.SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func GetDefaultModelRatioMap() map[string]float64 {
	return defaultModelRatio
}

func GetCompletionRatioMap() map[string]float64 {
	CompletionRatioMutex.Lock()
	defer CompletionRatioMutex.Unlock()
	if CompletionRatio == nil {
		CompletionRatio = defaultCompletionRatio
	}
	return CompletionRatio
}

func CompletionRatio2JSONString() string {
	GetCompletionRatioMap()
	jsonBytes, err := json.Marshal(CompletionRatio)
	if err != nil {
		common.SysError("error marshalling completion ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateCompletionRatioByJSONString(jsonStr string) error {
	CompletionRatioMutex.Lock()
	defer CompletionRatioMutex.Unlock()
	CompletionRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &CompletionRatio)
}

func GetCompletionRatio(name string) float64 {
	GetCompletionRatioMap()

	if strings.Contains(name, "/") {
		if ratio, ok := CompletionRatio[name]; ok {
			return ratio
		}
	}
	hardCodedRatio, contain := getHardcodedCompletionModelRatio(name)
	if contain {
		return hardCodedRatio
	}
	if ratio, ok := CompletionRatio[name]; ok {
		return ratio
	}
	return hardCodedRatio
}

func getHardcodedCompletionModelRatio(name string) (float64, bool) {
	lowercaseName := strings.ToLower(name)
	if strings.HasPrefix(name, "gpt-4-gizmo") {
		name = "gpt-4-gizmo-*"
	}
	if strings.HasPrefix(name, "gpt-4o-gizmo") {
		name = "gpt-4o-gizmo-*"
	}
	if strings.HasPrefix(name, "gpt-4") && !strings.HasSuffix(name, "-all") && !strings.HasSuffix(name, "-gizmo-*") {
		if strings.HasPrefix(name, "gpt-4o") {
			if name == "gpt-4o-2024-05-13" {
				return 3, true
			}
			return 4, true
		}
		// gpt-4.5-preview匹配
		if strings.HasPrefix(name, "gpt-4.5-preview") {
			return 2, true
		}
		if strings.HasPrefix(name, "gpt-4-turbo") || strings.HasSuffix(name, "gpt-4-1106") || strings.HasSuffix(name, "gpt-4-1105") {
			return 3, true
		}
		// 没有特殊标记的 gpt-4 模型默认倍率为 2
		return 2, false
	}
	if strings.HasPrefix(name, "o1") || strings.HasPrefix(name, "o3") {
		return 4, true
	}
	if name == "chatgpt-4o-latest" {
		return 3, true
	}
	if strings.Contains(name, "claude-instant-1") {
		return 3, true
	} else if strings.Contains(name, "claude-2") {
		return 3, true
	} else if strings.Contains(name, "claude-3") {
		return 5, true
	}
	if strings.HasPrefix(name, "gpt-3.5") {
		if name == "gpt-3.5-turbo" || strings.HasSuffix(name, "0125") {
			// https://openai.com/blog/new-embedding-models-and-api-updates
			// Updated GPT-3.5 Turbo model and lower pricing
			return 3, true
		}
		if strings.HasSuffix(name, "1106") {
			return 2, true
		}
		return 4.0 / 3.0, true
	}
	if strings.HasPrefix(name, "mistral-") {
		return 3, true
	}
	if strings.HasPrefix(name, "gemini-") {
		return 4, true
	}
	if strings.HasPrefix(name, "command") {
		switch name {
		case "command-r":
			return 3, true
		case "command-r-plus":
			return 5, true
		case "command-r-08-2024":
			return 4, true
		case "command-r-plus-08-2024":
			return 4, true
		default:
			return 4, true
		}
	}
	// hint 只给官方上4倍率，由于开源模型供应商自行定价，不对其进行补全倍率进行强制对齐
	if lowercaseName == "deepseek-chat" || lowercaseName == "deepseek-reasoner" {
		return 4, true
	}
	if strings.HasPrefix(name, "ERNIE-Speed-") {
		return 2, true
	} else if strings.HasPrefix(name, "ERNIE-Lite-") {
		return 2, true
	} else if strings.HasPrefix(name, "ERNIE-Character") {
		return 2, true
	} else if strings.HasPrefix(name, "ERNIE-Functions") {
		return 2, true
	}
	switch name {
	case "llama2-70b-4096":
		return 0.8 / 0.64, true
	case "llama3-8b-8192":
		return 2, true
	case "llama3-70b-8192":
		return 0.79 / 0.59, true
	}
	return 1, false
}

func GetAudioRatio(name string) float64 {
	if strings.Contains(name, "-realtime") {
		if strings.HasSuffix(name, "gpt-4o-realtime-preview-2024-12-17") {
			return 8
		} else if strings.Contains(name, "mini") {
			return 10 / 0.6
		} else {
			return 20
		}
	}
	if strings.Contains(name, "-audio") {
		if strings.HasSuffix(name, "gpt-4o-audio-preview-2024-12-17") {
			return 16
		} else if strings.Contains(name, "mini") {
			return 10 / 0.15
		} else {
			return 40
		}
	}
	return 20
}

func GetAudioCompletionRatio(name string) float64 {
	if strings.HasPrefix(name, "gpt-4o-realtime") {
		return 2
	} else if strings.HasPrefix(name, "gpt-4o-mini-realtime") {
		return 2
	}
	return 2
}
