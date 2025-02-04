package common

import (
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	relayconstant "one-api/relay/constant"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type RelayInfo struct {
	ChannelType          int
	ChannelId            int
	TokenId              int
	TokenKey             string
	UserId               int
	Group                string
	TokenUnlimited       bool
	StartTime            time.Time
	FirstResponseTime    time.Time
	setFirstResponse     bool
	ApiType              int
	IsStream             bool
	IsPlayground         bool
	UsePrice             bool
	RelayMode            int
	UpstreamModelName    string
	OriginModelName      string
	RecodeModelName      string
	RequestURLPath       string
	ApiVersion           string
	PromptTokens         int
	ApiKey               string
	Organization         string
	BaseUrl              string
	SupportStreamOptions bool
	ShouldIncludeUsage   bool
	ClientWs             *websocket.Conn
	TargetWs             *websocket.Conn
	InputAudioFormat     string
	OutputAudioFormat    string
	RealtimeTools        []dto.RealTimeTool
	IsFirstRequest       bool
	AudioUsage           bool
	ReasoningEffort      string
	ChannelSetting       map[string]interface{}
}

func GenRelayInfoWs(c *gin.Context, ws *websocket.Conn) *RelayInfo {
	info := GenRelayInfo(c)
	info.ClientWs = ws
	info.InputAudioFormat = "pcm16"
	info.OutputAudioFormat = "pcm16"
	info.IsFirstRequest = true
	return info
}

func GenRelayInfo(c *gin.Context) *RelayInfo {
	channelType := c.GetInt("channel_type")
	channelId := c.GetInt("channel_id")
	channelSetting := c.GetStringMap("channel_setting")

	tokenId := c.GetInt("token_id")
	tokenKey := c.GetString("token_key")
	userId := c.GetInt("id")
	group := c.GetString("group")
	tokenUnlimited := c.GetBool("token_unlimited_quota")
	startTime := c.GetTime(constant.ContextKeyRequestStartTime)
	// firstResponseTime = time.Now() - 1 second

	apiType, _ := relayconstant.ChannelType2APIType(channelType)

	info := &RelayInfo{
		RelayMode:         relayconstant.Path2RelayMode(c.Request.URL.Path),
		BaseUrl:           c.GetString("base_url"),
		RequestURLPath:    c.Request.URL.String(),
		ChannelType:       channelType,
		ChannelId:         channelId,
		TokenId:           tokenId,
		TokenKey:          tokenKey,
		UserId:            userId,
		Group:             group,
		TokenUnlimited:    tokenUnlimited,
		StartTime:         startTime,
		FirstResponseTime: startTime.Add(-time.Second),
		OriginModelName:   c.GetString("original_model"),
		UpstreamModelName: c.GetString("original_model"),
		RecodeModelName:   c.GetString("recode_model"),
		ApiType:           apiType,
		ApiVersion:        c.GetString("api_version"),
		ApiKey:            strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer "),
		Organization:      c.GetString("channel_organization"),
		ChannelSetting:    channelSetting,
	}
	if strings.HasPrefix(c.Request.URL.Path, "/pg") {
		info.IsPlayground = true
		info.RequestURLPath = strings.TrimPrefix(info.RequestURLPath, "/pg")
		info.RequestURLPath = "/v1" + info.RequestURLPath
	}
	if info.BaseUrl == "" {
		info.BaseUrl = common.ChannelBaseURLs[channelType]
	}
	if info.ChannelType == common.ChannelTypeAzure {
		info.ApiVersion = GetAPIVersion(c)
	}
	if info.ChannelType == common.ChannelTypeVertexAi {
		info.ApiVersion = c.GetString("region")
	}
	if info.ChannelType == common.ChannelTypeOpenAI || info.ChannelType == common.ChannelTypeAnthropic ||
		info.ChannelType == common.ChannelTypeAws || info.ChannelType == common.ChannelTypeGemini ||
		info.ChannelType == common.ChannelCloudflare || info.ChannelType == common.ChannelTypeAzure {
		info.SupportStreamOptions = true
	}
	return info
}

func (info *RelayInfo) SetPromptTokens(promptTokens int) {
	info.PromptTokens = promptTokens
}

func (info *RelayInfo) SetIsStream(isStream bool) {
	info.IsStream = isStream
}

func (info *RelayInfo) SetFirstResponseTime() {
	if !info.setFirstResponse {
		info.FirstResponseTime = time.Now()
		info.setFirstResponse = true
	}
}

type TaskRelayInfo struct {
	ChannelType       int
	ChannelId         int
	TokenId           int
	UserId            int
	Group             string
	StartTime         time.Time
	ApiType           int
	RelayMode         int
	UpstreamModelName string
	RequestURLPath    string
	ApiKey            string
	BaseUrl           string

	Action       string
	OriginTaskID string

	ConsumeQuota bool
}

func GenTaskRelayInfo(c *gin.Context) *TaskRelayInfo {
	channelType := c.GetInt("channel_type")
	channelId := c.GetInt("channel_id")

	tokenId := c.GetInt("token_id")
	userId := c.GetInt("id")
	group := c.GetString("group")
	startTime := time.Now()

	apiType, _ := relayconstant.ChannelType2APIType(channelType)

	info := &TaskRelayInfo{
		RelayMode:      relayconstant.Path2RelayMode(c.Request.URL.Path),
		BaseUrl:        c.GetString("base_url"),
		RequestURLPath: c.Request.URL.String(),
		ChannelType:    channelType,
		ChannelId:      channelId,
		TokenId:        tokenId,
		UserId:         userId,
		Group:          group,
		StartTime:      startTime,
		ApiType:        apiType,
		ApiKey:         strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer "),
	}
	if info.BaseUrl == "" {
		info.BaseUrl = common.ChannelBaseURLs[channelType]
	}
	return info
}

func (info *TaskRelayInfo) ToRelayInfo() *RelayInfo {
	return &RelayInfo{
		ChannelType:       info.ChannelType,
		ChannelId:         info.ChannelId,
		TokenId:           info.TokenId,
		UserId:            info.UserId,
		Group:             info.Group,
		StartTime:         info.StartTime,
		ApiType:           info.ApiType,
		RelayMode:         info.RelayMode,
		UpstreamModelName: info.UpstreamModelName,
		RequestURLPath:    info.RequestURLPath,
		ApiKey:            info.ApiKey,
		BaseUrl:           info.BaseUrl,
	}
}
