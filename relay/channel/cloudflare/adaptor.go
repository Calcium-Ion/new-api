package cloudflare

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/dto"
	"one-api/relay/channel"
	relaycommon "one-api/relay/common"
	"one-api/relay/constant"
)

type Adaptor struct {
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	switch info.RelayMode {
	case constant.RelayModeChatCompletions:
		return fmt.Sprintf("%s/client/v4/accounts/%s/ai/v1/chat/completions", info.BaseUrl, info.ApiVersion), nil
	case constant.RelayModeEmbeddings:
		return fmt.Sprintf("%s/client/v4/accounts/%s/ai/v1/embeddings", info.BaseUrl, info.ApiVersion), nil
	default:
		return fmt.Sprintf("%s/client/v4/accounts/%s/ai/run/%s", info.BaseUrl, info.ApiVersion, info.UpstreamModelName), nil
	}
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	req.Set("Authorization", fmt.Sprintf("Bearer %s", info.ApiKey))
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	switch info.RelayMode {
	case constant.RelayModeCompletions:
		return convertCf2CompletionsRequest(*request), nil
	default:
		return request, nil
	}
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return request, nil
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	// 添加文件字段
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		return nil, errors.New("file is required")
	}
	defer file.Close()
	// 打开临时文件用于保存上传的文件内容
	requestBody := &bytes.Buffer{}

	// 将上传的文件内容复制到临时文件
	if _, err := io.Copy(requestBody, file); err != nil {
		return nil, err
	}
	return requestBody, nil
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *dto.OpenAIErrorWithStatusCode) {
	switch info.RelayMode {
	case constant.RelayModeEmbeddings:
		fallthrough
	case constant.RelayModeChatCompletions:
		if info.IsStream {
			err, usage = cfStreamHandler(c, resp, info)
		} else {
			err, usage = cfHandler(c, resp, info)
		}
	case constant.RelayModeAudioTranslation:
		fallthrough
	case constant.RelayModeAudioTranscription:
		err, usage = cfSTTHandler(c, resp, info)
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
