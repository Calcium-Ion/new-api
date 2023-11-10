package common

import (
	"encoding/json"
	"strings"
	"time"
)

// ModelRatio
// https://platform.openai.com/docs/models/model-endpoint-compatibility
// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Blfmc9dlf
// https://openai.com/pricing
// TODO: when a new api is enabled, check the pricing here
// 1 === $0.002 / 1K tokens
// 1 === ￥0.014 / 1k tokens
var ModelRatio = map[string]float64{
	"gpt-4":                     15,
	"gpt-4-0314":                15,
	"gpt-4-0613":                15,
	"gpt-4-32k":                 30,
	"gpt-4-32k-0314":            30,
	"gpt-4-32k-0613":            30,
	"gpt-4-1106-preview":        5,    // $0.01 / 1K tokens
	"gpt-4-vision-preview":      5,    // $0.01 / 1K tokens
	"gpt-3.5-turbo":             0.75, // $0.0015 / 1K tokens
	"gpt-3.5-turbo-0301":        0.75,
	"gpt-3.5-turbo-0613":        0.75,
	"gpt-3.5-turbo-16k":         1.5, // $0.003 / 1K tokens
	"gpt-3.5-turbo-16k-0613":    1.5,
	"gpt-3.5-turbo-instruct":    0.75, // $0.0015 / 1K tokens
	"gpt-3.5-turbo-1106":        0.5,  // $0.001 / 1K tokens
	"text-ada-001":              0.2,
	"text-babbage-001":          0.25,
	"text-curie-001":            1,
	"text-davinci-002":          10,
	"text-davinci-003":          10,
	"text-davinci-edit-001":     10,
	"code-davinci-edit-001":     10,
	"whisper-1":                 15, // $0.006 / minute -> $0.006 / 150 words -> $0.006 / 200 tokens -> $0.03 / 1k tokens
	"davinci":                   10,
	"curie":                     10,
	"babbage":                   10,
	"ada":                       10,
	"text-embedding-ada-002":    0.05,
	"text-search-ada-doc-001":   10,
	"text-moderation-stable":    0.1,
	"text-moderation-latest":    0.1,
	"dall-e":                    8,
	"claude-instant-1":          0.815,  // $1.63 / 1M tokens
	"claude-2":                  5.51,   // $11.02 / 1M tokens
	"ERNIE-Bot":                 0.8572, // ￥0.012 / 1k tokens
	"ERNIE-Bot-turbo":           0.5715, // ￥0.008 / 1k tokens
	"ERNIE-Bot-4":               8.572,  // ￥0.12 / 1k tokens
	"Embedding-V1":              0.1429, // ￥0.002 / 1k tokens
	"PaLM-2":                    1,
	"chatglm_turbo":             0.3572, // ￥0.005 / 1k tokens
	"chatglm_pro":               0.7143, // ￥0.01 / 1k tokens
	"chatglm_std":               0.3572, // ￥0.005 / 1k tokens
	"chatglm_lite":              0.1429, // ￥0.002 / 1k tokens
	"qwen-turbo":                0.8572, // ￥0.012 / 1k tokens
	"qwen-plus":                 10,     // ￥0.14 / 1k tokens
	"text-embedding-v1":         0.05,   // ￥0.0007 / 1k tokens
	"SparkDesk":                 1.2858, // ￥0.018 / 1k tokens
	"360GPT_S2_V9":              0.8572, // ¥0.012 / 1k tokens
	"embedding-bert-512-v1":     0.0715, // ¥0.001 / 1k tokens
	"embedding_s1_v1":           0.0715, // ¥0.001 / 1k tokens
	"semantic_similarity_s1_v1": 0.0715, // ¥0.001 / 1k tokens
	"hunyuan":                   7.143,  // ¥0.1 / 1k tokens  // https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
}

func ModelRatio2JSONString() string {
	jsonBytes, err := json.Marshal(ModelRatio)
	if err != nil {
		SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRatioByJSONString(jsonStr string) error {
	ModelRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &ModelRatio)
}

func GetModelRatio(name string) float64 {
	ratio, ok := ModelRatio[name]
	if !ok {
		SysError("model ratio not found: " + name)
		return 30
	}
	return ratio
}

func GetCompletionRatio(name string) float64 {
	if strings.HasPrefix(name, "gpt-3.5") {
		if strings.HasSuffix(name, "1106") {
			return 2
		}
		if name == "gpt-3.5-turbo" || name == "gpt-3.5-turbo-16k" {
			// TODO: clear this after 2023-12-11
			now := time.Now()
			// https://platform.openai.com/docs/models/continuous-model-upgrades
			// if after 2023-12-11, use 2
			if now.After(time.Date(2023, 12, 11, 0, 0, 0, 0, time.UTC)) {
				return 2
			}
		}
		return 1.333333
	}
	if strings.HasPrefix(name, "gpt-4") {
		if strings.HasSuffix(name, "preview") {
			return 3
		}
		return 2
	}
	if strings.HasPrefix(name, "claude-instant-1") {
		return 3.38
	}
	if strings.HasPrefix(name, "claude-2") {
		return 2.965517
	}
	return 1
}
