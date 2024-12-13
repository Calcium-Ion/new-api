package relay

import (
	commonconstant "one-api/constant"
	"one-api/relay/channel"
	"one-api/relay/channel/ali"
	"one-api/relay/channel/aws"
	"one-api/relay/channel/baidu"
	"one-api/relay/channel/claude"
	"one-api/relay/channel/cloudflare"
	"one-api/relay/channel/cohere"
	"one-api/relay/channel/deepseek"
	"one-api/relay/channel/dify"
	"one-api/relay/channel/gemini"
	"one-api/relay/channel/jina"
	"one-api/relay/channel/mistral"
	"one-api/relay/channel/ollama"
	"one-api/relay/channel/openai"
	"one-api/relay/channel/palm"
	"one-api/relay/channel/perplexity"
	"one-api/relay/channel/siliconflow"
	"one-api/relay/channel/task/suno"
	"one-api/relay/channel/tencent"
	"one-api/relay/channel/vertex"
	"one-api/relay/channel/xunfei"
	"one-api/relay/channel/zhipu"
	"one-api/relay/channel/zhipu_4v"
	"one-api/relay/constant"
)

func GetAdaptor(apiType int) channel.Adaptor {
	switch apiType {
	//case constant.APITypeAIProxyLibrary:
	//	return &aiproxy.Adaptor{}
	case constant.APITypeAli:
		return &ali.Adaptor{}
	case constant.APITypeAnthropic:
		return &claude.Adaptor{}
	case constant.APITypeBaidu:
		return &baidu.Adaptor{}
	case constant.APITypeGemini:
		return &gemini.Adaptor{}
	case constant.APITypeOpenAI:
		return &openai.Adaptor{}
	case constant.APITypePaLM:
		return &palm.Adaptor{}
	case constant.APITypeTencent:
		return &tencent.Adaptor{}
	case constant.APITypeXunfei:
		return &xunfei.Adaptor{}
	case constant.APITypeZhipu:
		return &zhipu.Adaptor{}
	case constant.APITypeZhipuV4:
		return &zhipu_4v.Adaptor{}
	case constant.APITypeOllama:
		return &ollama.Adaptor{}
	case constant.APITypePerplexity:
		return &perplexity.Adaptor{}
	case constant.APITypeAws:
		return &aws.Adaptor{}
	case constant.APITypeCohere:
		return &cohere.Adaptor{}
	case constant.APITypeDify:
		return &dify.Adaptor{}
	case constant.APITypeJina:
		return &jina.Adaptor{}
	case constant.APITypeCloudflare:
		return &cloudflare.Adaptor{}
	case constant.APITypeSiliconFlow:
		return &siliconflow.Adaptor{}
	case constant.APITypeVertexAi:
		return &vertex.Adaptor{}
	case constant.APITypeMistral:
		return &mistral.Adaptor{}
	case constant.APITypeDeepSeek:
		return &deepseek.Adaptor{}
	}
	return nil
}

func GetTaskAdaptor(platform commonconstant.TaskPlatform) channel.TaskAdaptor {
	switch platform {
	//case constant.APITypeAIProxyLibrary:
	//	return &aiproxy.Adaptor{}
	case commonconstant.TaskPlatformSuno:
		return &suno.TaskAdaptor{}
	}
	return nil
}
