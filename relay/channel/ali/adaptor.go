package ali

import (
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

func (a *Adaptor) InitRerank(info *relaycommon.RelayInfo, request dto.RerankRequest) {
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo, request dto.GeneralOpenAIRequest) {

}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	fullRequestURL := fmt.Sprintf("%s/api/v1/services/aigc/text-generation/generation", info.BaseUrl)
	if info.RelayMode == constant.RelayModeEmbeddings {
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/embeddings/text-embedding/text-embedding", info.BaseUrl)
	}
	return fullRequestURL, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	req.Header.Set("Authorization", "Bearer "+info.ApiKey)
	if info.IsStream {
		req.Header.Set("X-DashScope-SSE", "enable")
	}
	if c.GetString("plugin") != "" {
		req.Header.Set("X-DashScope-Plugin", c.GetString("plugin"))
	}
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	switch relayMode {
	case constant.RelayModeEmbeddings:
		baiduEmbeddingRequest := embeddingRequestOpenAI2Ali(*request)
		return baiduEmbeddingRequest, nil
	default:
		baiduRequest := requestOpenAI2Ali(*request)
		return baiduRequest, nil
	}
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage *dto.Usage, err *dto.OpenAIErrorWithStatusCode) {
	if info.IsStream {
		err, usage = aliStreamHandler(c, resp)
	} else {
		switch info.RelayMode {
		case constant.RelayModeEmbeddings:
			err, usage = aliEmbeddingHandler(c, resp)
		default:
			err, usage = aliHandler(c, resp)
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
