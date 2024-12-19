package common

// 定义不同分组的类型
type PriceGroup map[string]float64

// OpenAIPrices https://openai.com/api/pricing/
func OpenAIPrices() PriceGroup {
	return PriceGroup{
		// ---- GPT-4o & 4o Mini----
		// gpt-4o Point to 2024-08-06
		"gpt-4o": 1.25,
		// $5 / 1M tokens
		"gpt-4o-2024-05-13": 2.5,
		// $2.5 / 1M tokens
		"gpt-4o-2024-08-06": 1.25,
		// $2.5 / 1M tokens
		"gpt-4o-2024-11-20": 1.25,
		// gpt-4o-mini Point to 2024-07-18
		"gpt-4o-mini": 0.075,
		// $0.15 / 1K tokens
		"gpt-4o-mini-2024-07-18": 0.075,
		// chatgpt-4o-latest $5 / 1M tokens
		"chatgpt-4o-latest": 2.5,
		// ---- GPT-4o and GPT-4o-mini Realtime & Audio ----
		// Point to 2024-10-01
		"gpt-4o-realtime-preview":            2.5,
		"gpt-4o-realtime-preview-2024-10-01": 2.5,
		"gpt-4o-realtime-preview-2024-12-17": 2.5,
		// Point to 2024-12-17
		"gpt-4o-mini-realtime-preview": 0.3,
		// $0.6 / 1K tokens
		"gpt-4o-mini-realtime-preview-2024-12-17": 0.3,
		// Point to 2024-10-01
		"gpt-4o-audio-preview": 1.25,
		// $2.5 / 1M tokens
		"gpt-4o-audio-preview-2024-10-01": 1.25,
		// $2.5 / 1M tokens
		"gpt-4o-audio-preview-2024-12-17": 1.25,
		// ---- o1 ----
		"o1-mini":               1.5,
		"o1-mini-2024-09-12":    1.5,
		"o1-preview":            7.5,
		"o1-preview-2024-09-12": 7.5,
		"o1":                    7.5,
		"o1-2024-12-17":         7.5,
		// ---- GPT-4 Turbo ----
		"gpt-4-turbo":            5, // $0.01 / 1K tokens
		"gpt-4-turbo-2024-04-09": 5, // $0.01 / 1K tokens
		// ---- Chat2API ----
		"gpt-4-gizmo-*":  15,
		"gpt-4o-gizmo-*": 2.5,
		"gpt-4-all":      15,
		"gpt-4o-all":     15,
		// ---- GPT-4 ----
		"gpt-4": 15,
		//"gpt-4-0314":                   15, //deprecated
		"gpt-4-0613": 15,
		"gpt-4-32k":  30,
		//"gpt-4-32k-0314":               30, //deprecated
		"gpt-4-32k-0613":            30,
		"gpt-4-1106-preview":        5, // $10 / 1M tokens
		"gpt-4-0125-preview":        5, // $10 / 1M tokens
		"gpt-4-turbo-preview":       5, // $10 / 1M tokens
		"gpt-4-vision-preview":      5, // $10 / 1M tokens
		"gpt-4-1106-vision-preview": 5, // $10 / 1M tokens
		// ---- GPT-3.5 ----
		//"gpt-3.5-turbo-0301":           0.75, //deprecated
		"gpt-3.5-turbo":          0.25,
		"gpt-3.5-turbo-0613":     0.75,
		"gpt-3.5-turbo-16k":      1.5, // $0.003 / 1K tokens
		"gpt-3.5-turbo-16k-0613": 1.5,
		"gpt-3.5-turbo-instruct": 0.75, // $0.0015 / 1K tokens
		"gpt-3.5-turbo-1106":     0.5,  // $0.001 / 1K tokens
		"gpt-3.5-turbo-0125":     0.25,
		// Other Models
		"babbage-002":      0.2, // $0.0004 / 1K tokens
		"davinci-002":      1,   // $0.002 / 1K tokens
		"text-ada-001":     0.2,
		"text-babbage-001": 0.25,
		"text-curie-001":   1,
		//"text-davinci-002":               10,
		//"text-davinci-003":               10,
		"text-davinci-edit-001":   10,
		"code-davinci-edit-001":   10,
		"whisper-1":               15,  // $0.006 / minute -> $0.006 / 150 words -> $0.006 / 200 tokens -> $0.03 / 1k tokens
		"tts-1":                   7.5, // 1k characters -> $0.015
		"tts-1-1106":              7.5, // 1k characters -> $0.015
		"tts-1-hd":                15,  // 1k characters -> $0.03
		"tts-1-hd-1106":           15,  // 1k characters -> $0.03
		"davinci":                 10,
		"curie":                   10,
		"babbage":                 10,
		"ada":                     10,
		"text-embedding-3-small":  0.01,
		"text-embedding-3-large":  0.065,
		"text-embedding-ada-002":  0.05,
		"text-search-ada-doc-001": 10,
		"text-moderation-stable":  0.1,
		"text-moderation-latest":  0.1,
	}
}

// AnthropicPrices https://www.anthropic.com/pricing#anthropic-api
func AnthropicPrices() PriceGroup {
	return PriceGroup{
		"claude-instant-1":           0.4,   // $0.8 / 1M tokens
		"claude-2.0":                 4,     // $8 / 1M tokens
		"claude-2.1":                 4,     // $8 / 1M tokens
		"claude-3-haiku-20240307":    0.125, // $0.25 / 1M tokens
		"claude-3-5-haiku-20241022":  0.5,   // $1 / 1M tokens
		"claude-3-sonnet-20240229":   1.5,   // $3 / 1M tokens
		"claude-3-5-sonnet-20240620": 1.5,
		"claude-3-5-sonnet-20241022": 1.5,
		"claude-3-opus-20240229":     7.5, // $15 / 1M tokens
	}
}

func GeminiPrices() PriceGroup {
	return PriceGroup{
		// ---- Gemini AI studio MainStream ----
		// $0.0375 / 1M tokens
		"gemini-1.5-flash-8b": 0.01875,
		// $0.075 / 1M tokens
		"gemini-1.5-flash": 0.0375,
		// $1.25 / 1M tokens
		"gemini-1.5-pro": 0.625,
		// ---- Gemini Vertex & others ----
		"PaLM-2":                       1,
		"gemini-pro":                   1, // $0.00025 / 1k characters -> $0.001 / 1k tokens
		"gemini-pro-vision":            1, // $0.00025 / 1k characters -> $0.001 / 1k tokens
		"gemini-1.0-pro-001":           1,
		"gemini-1.0-pro-vision-001":    1,
		"gemini-1.0-pro-latest":        1,
		"gemini-1.0-pro-vision-latest": 1,
		"gemini-1.5-pro-latest":        1.75,
		"gemini-1.5-flash-latest":      1,
		"gemini-ultra":                 1,
	}
}

func CoherePrices() PriceGroup {
	return PriceGroup{
		"command":                0.5,
		"command-nightly":        0.5,
		"command-light":          0.5,
		"command-light-nightly":  0.5,
		"command-r":              0.25,
		"command-r-plus":         1.5,
		"command-r-08-2024":      0.075,
		"command-r-plus-08-2024": 1.25,
		"command-r7b-12-2024":    0.01875,
	}
}

var defaultModelRatio = map[string]float64{
	//"midjourney":                50,
	"ERNIE-4.0-8K":                   0.120 * RMB,
	"ERNIE-3.5-8K":                   0.012 * RMB,
	"ERNIE-3.5-8K-0205":              0.024 * RMB,
	"ERNIE-3.5-8K-1222":              0.012 * RMB,
	"ERNIE-Bot-8K":                   0.024 * RMB,
	"ERNIE-3.5-4K-0205":              0.012 * RMB,
	"ERNIE-Speed-8K":                 0.004 * RMB,
	"ERNIE-Speed-128K":               0.004 * RMB,
	"ERNIE-Lite-8K-0922":             0.008 * RMB,
	"ERNIE-Lite-8K-0308":             0.003 * RMB,
	"ERNIE-Tiny-8K":                  0.001 * RMB,
	"BLOOMZ-7B":                      0.004 * RMB,
	"Embedding-V1":                   0.002 * RMB,
	"bge-large-zh":                   0.002 * RMB,
	"bge-large-en":                   0.002 * RMB,
	"tao-8k":                         0.002 * RMB,
	"chatglm_turbo":                  0.3572,     // ￥0.005 / 1k tokens
	"chatglm_pro":                    0.7143,     // ￥0.01 / 1k tokens
	"chatglm_std":                    0.3572,     // ￥0.005 / 1k tokens
	"chatglm_lite":                   0.1429,     // ￥0.002 / 1k tokens
	"glm-4":                          7.143,      // ￥0.1 / 1k tokens
	"glm-4v":                         0.05 * RMB, // ￥0.05 / 1k tokens
	"glm-4-alltools":                 0.1 * RMB,  // ￥0.1 / 1k tokens
	"glm-3-turbo":                    0.3572,
	"glm-4-plus":                     0.05 * RMB,
	"glm-4-0520":                     0.1 * RMB,
	"glm-4-air":                      0.001 * RMB,
	"glm-4-airx":                     0.01 * RMB,
	"glm-4-long":                     0.001 * RMB,
	"glm-4-flash":                    0,
	"glm-4v-plus":                    0.01 * RMB,
	"qwen-turbo":                     0.8572, // ￥0.012 / 1k tokens
	"qwen-plus":                      10,     // ￥0.14 / 1k tokens
	"text-embedding-v1":              0.05,   // ￥0.0007 / 1k tokens
	"SparkDesk-v1.1":                 1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v2.1":                 1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.1":                 1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.5":                 1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v4.0":                 1.2858,
	"360GPT_S2_V9":                   0.8572, // ¥0.012 / 1k tokens
	"360gpt-turbo":                   0.0858, // ¥0.0012 / 1k tokens
	"360gpt-turbo-responsibility-8k": 0.8572, // ¥0.012 / 1k tokens
	"360gpt-pro":                     0.8572, // ¥0.012 / 1k tokens
	"360gpt2-pro":                    0.8572, // ¥0.012 / 1k tokens
	"embedding-bert-512-v1":          0.0715, // ¥0.001 / 1k tokens
	"embedding_s1_v1":                0.0715, // ¥0.001 / 1k tokens
	"semantic_similarity_s1_v1":      0.0715, // ¥0.001 / 1k tokens
	"hunyuan":                        7.143,  // ¥0.1 / 1k tokens  // https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
	// https://platform.lingyiwanwu.com/docs#-计费单元
	// 已经按照 7.2 来换算美元价格
	"yi-34b-chat-0205":     0.18,
	"yi-34b-chat-200k":     0.864,
	"yi-vl-plus":           0.432,
	"yi-large":             20.0 / 1000 * RMB,
	"yi-medium":            2.5 / 1000 * RMB,
	"yi-vision":            6.0 / 1000 * RMB,
	"yi-medium-200k":       12.0 / 1000 * RMB,
	"yi-spark":             1.0 / 1000 * RMB,
	"yi-large-rag":         25.0 / 1000 * RMB,
	"yi-large-turbo":       12.0 / 1000 * RMB,
	"yi-large-preview":     20.0 / 1000 * RMB,
	"yi-large-rag-preview": 25.0 / 1000 * RMB,
	// deepseek https://api-docs.deepseek.com/zh-cn/quick_start/pricing
	"deepseek-chat":  0.07,
	"deepseek-coder": 0.07,
	// Perplexity online 模型对搜索额外收费，有需要应自行调整，此处不计入搜索费用
	"llama-3-sonar-small-32k-chat":   0.2 / 1000 * USD,
	"llama-3-sonar-small-32k-online": 0.2 / 1000 * USD,
	"llama-3-sonar-large-32k-chat":   1 / 1000 * USD,
	"llama-3-sonar-large-32k-online": 1 / 1000 * USD,
}

// 合并所有分组到主Map
func init() {
	// 合并各个分组
	groups := []PriceGroup{
		OpenAIPrices(),
		AnthropicPrices(),
		GeminiPrices(),
		CoherePrices(),
	}

	// 将所有分组数据合并到主 map
	for _, group := range groups {
		for k, v := range group {
			defaultModelRatio[k] = v
		}
	}

}
