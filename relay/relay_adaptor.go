package relay

import (
	"one-api/relay/channel"
	"one-api/relay/channel/ali"
	"one-api/relay/channel/baidu"
	"one-api/relay/channel/claude"
	"one-api/relay/channel/gemini"
	"one-api/relay/channel/openai"
	"one-api/relay/channel/palm"
	"one-api/relay/channel/tencent"
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
	case constant.APITypeZhipu_v4:
		return &zhipu_4v.Adaptor{}
	}
	return nil
}
