package mokaai

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	// "one-api/relay/adaptor"
	// "one-api/relay/meta"
	// "one-api/relay/model"
	// "one-api/relay/constant"
	"one-api/dto"
	"one-api/relay/channel"
	relaycommon "one-api/relay/common"
	"one-api/relay/constant"
)

type Adaptor struct {
}

// ConvertImageRequest implements adaptor.Adaptor.
func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}
func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
}


func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo)  (string, error) {
	
	var urlPrefix = info.BaseUrl
	
	switch info.RelayMode {
	case constant.RelayModeChatCompletions:
		return fmt.Sprintf("%s/chat/completions", urlPrefix), nil
	case constant.RelayModeEmbeddings:
		return fmt.Sprintf("%s/embeddings", urlPrefix), nil
	default:
		return fmt.Sprintf("%s/run/%s", urlPrefix, info.UpstreamModelName), nil
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
	case constant.RelayModeChatCompletions:
		return nil, errors.New("not implemented")
	case  constant.RelayModeEmbeddings:
		// return ConvertCompletionsRequest(*request), nil
		return ConvertEmbeddingRequest(*request), nil
	default:
		return nil, errors.New("not implemented")
	}
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *dto.OpenAIErrorWithStatusCode) {
	switch info.RelayMode {
		
	case constant.RelayModeAudioTranscription:
	case constant.RelayModeAudioTranslation:
	case constant.RelayModeChatCompletions:
		fallthrough
	case constant.RelayModeEmbeddings:
		if info.IsStream {
			err, usage = StreamHandler(c, resp, info)
		} else {
			err, usage = Handler(c, resp, info)
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
