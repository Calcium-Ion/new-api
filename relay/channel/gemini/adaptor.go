package gemini

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"one-api/dto"
	"one-api/relay/channel"
	"strings"
	relaycommon "one-api/relay/common"
)

type Adaptor struct {
	modelVersionMap map[string]string
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
	modelVersionMapStr := os.Getenv("GEMINI_MODEL_API")
	if modelVersionMapStr == "" {
		a.modelVersionMap = map[string]string{ 
			"gemini-1.5-pro-latest":   "v1beta",
			"gemini-1.5-pro-001":      "v1beta",
			"gemini-1.5-pro":          "v1beta",
			"gemini-1.5-pro-exp-0801": "v1beta",
			"gemini-1.5-flash-latest": "v1beta",
			"gemini-1.5-flash-001":    "v1beta",
			"gemini-1.5-flash":        "v1beta",
			"gemini-ultra":            "v1beta",
		}
		return
	}
	a.modelVersionMap = make(map[string]string) 
	for _, pair := range strings.Split(modelVersionMapStr, ",") {
		parts := strings.Split(pair, ":")
		if len(parts) == 2 {
			a.modelVersionMap[parts[0]] = parts[1] 
		}
	}
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	// 从映射中获取模型名称对应的版本，如果找不到就使用 info.ApiVersion 或默认的版本 "v1"
	version, beta := a.modelVersionMap[info.UpstreamModelName]
	if !beta {
		if info.ApiVersion != "" {
			version = info.ApiVersion
		} else {
			version = "v1"
		}
	}

	action := "generateContent"
	if info.IsStream {
		action = "streamGenerateContent?alt=sse"
	}
	return fmt.Sprintf("%s/%s/models/%s:%s", info.BaseUrl, version, info.UpstreamModelName, action), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	req.Header.Set("x-goog-api-key", info.ApiKey)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return CovertGemini2OpenAI(*request), nil
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage *dto.Usage, err *dto.OpenAIErrorWithStatusCode) {
	if info.IsStream {
		err, usage = geminiChatStreamHandler(c, resp, info)
	} else {
		err, usage = geminiChatHandler(c, resp, info.PromptTokens, info.UpstreamModelName)
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
