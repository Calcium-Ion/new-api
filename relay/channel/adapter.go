package channel

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/dto"
	relaycommon "one-api/relay/common"
)

type Adaptor interface {
	// Init IsStream bool
	Init(info *relaycommon.RelayInfo, request dto.GeneralOpenAIRequest)
	GetRequestURL(info *relaycommon.RelayInfo) (string, error)
	SetupRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error
	ConvertRequest(c *gin.Context, relayMode int, request *dto.GeneralOpenAIRequest) (any, error)
	DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error)
	DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage *dto.Usage, err *dto.OpenAIErrorWithStatusCode, sensitiveResp *dto.SensitiveResponse)
	GetModelList() []string
	GetChannelName() string
}
