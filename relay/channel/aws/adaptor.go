package aws

import (
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/dto"
	"one-api/relay/channel/claude"
	relaycommon "one-api/relay/common"
	"strings"
)

const (
	RequestModeCompletion = 1
	RequestModeMessage    = 2
)

type Adaptor struct {
	RequestMode int
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo, request dto.GeneralOpenAIRequest) {
	if strings.HasPrefix(info.UpstreamModelName, "claude-3") {
		a.RequestMode = RequestModeMessage
	} else {
		a.RequestMode = RequestModeCompletion
	}
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	return "", nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	var claudeReq *claude.ClaudeRequest
	var err error
	if a.RequestMode == RequestModeCompletion {
		claudeReq = claude.RequestOpenAI2ClaudeComplete(*request)
	} else {
		claudeReq, err = claude.RequestOpenAI2ClaudeMessage(*request)
	}
	c.Set("request_model", request.Model)
	c.Set("converted_request", claudeReq)
	return claudeReq, err
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return nil, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage *dto.Usage, err *dto.OpenAIErrorWithStatusCode) {
	if info.IsStream {
		err, usage = awsStreamHandler(c, info, a.RequestMode)
	} else {
		err, usage = awsHandler(c, info, a.RequestMode)
	}
	return
}

func (a *Adaptor) GetModelList() (models []string) {
	for n := range awsModelIDMap {
		models = append(models, n)
	}

	return
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
