package tencent

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	"one-api/relay/channel"
	relaycommon "one-api/relay/common"
	"one-api/service"
	"strconv"
	"strings"
)

type Adaptor struct {
	Sign      string
	AppID     int64
	Action    string
	Version   string
	Timestamp int64
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
	a.Action = "ChatCompletions"
	a.Version = "2023-09-01"
	a.Timestamp = common.GetTimestamp()
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	return fmt.Sprintf("%s/", info.BaseUrl), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	req.Set("Authorization", a.Sign)
	req.Set("X-TC-Action", a.Action)
	req.Set("X-TC-Version", a.Version)
	req.Set("X-TC-Timestamp", strconv.FormatInt(a.Timestamp, 10))
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	apiKey := c.Request.Header.Get("Authorization")
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")
	appId, secretId, secretKey, err := parseTencentConfig(apiKey)
	a.AppID = appId
	if err != nil {
		return nil, err
	}
	tencentRequest := requestOpenAI2Tencent(a, *request)
	// we have to calculate the sign here
	a.Sign = getTencentSign(*tencentRequest, a, secretId, secretKey)
	return tencentRequest, nil
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *dto.OpenAIErrorWithStatusCode) {
	if info.IsStream {
		var responseText string
		err, responseText = tencentStreamHandler(c, resp)
		usage, _ = service.ResponseText2Usage(responseText, info.UpstreamModelName, info.PromptTokens)
	} else {
		err, usage = tencentHandler(c, resp)
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
