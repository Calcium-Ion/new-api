package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"net/http"
	"one-api/common"
	"one-api/dto"
	"one-api/relay/channel"
	"one-api/relay/channel/ai360"
	"one-api/relay/channel/lingyiwanwu"
	"one-api/relay/channel/minimax"
	"one-api/relay/channel/moonshot"
	relaycommon "one-api/relay/common"
	"one-api/relay/constant"
	"strings"
)

type Adaptor struct {
	ChannelType    int
	ResponseFormat string
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
	a.ChannelType = info.ChannelType
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	switch info.ChannelType {
	case common.ChannelTypeAzure:
		// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/chatgpt-quickstart?pivots=rest-api&tabs=command-line#rest-api
		requestURL := strings.Split(info.RequestURLPath, "?")[0]
		requestURL = fmt.Sprintf("%s?api-version=%s", requestURL, info.ApiVersion)
		task := strings.TrimPrefix(requestURL, "/v1/")
		model_ := info.UpstreamModelName
		model_ = strings.Replace(model_, ".", "", -1)
		// https://github.com/songquanpeng/one-api/issues/67

		requestURL = fmt.Sprintf("/openai/deployments/%s/%s", model_, task)
		return relaycommon.GetFullRequestURL(info.BaseUrl, requestURL, info.ChannelType), nil
	case common.ChannelTypeMiniMax:
		return minimax.GetRequestURL(info)
	case common.ChannelTypeCustom:
		url := info.BaseUrl
		url = strings.Replace(url, "{model}", info.UpstreamModelName, -1)
		return url, nil
	default:
		return relaycommon.GetFullRequestURL(info.BaseUrl, info.RequestURLPath, info.ChannelType), nil
	}
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	if info.ChannelType == common.ChannelTypeAzure {
		req.Header.Set("api-key", info.ApiKey)
		return nil
	}
	if info.ChannelType == common.ChannelTypeOpenAI && "" != info.Organization {
		req.Header.Set("OpenAI-Organization", info.Organization)
	}
	req.Header.Set("Authorization", "Bearer "+info.ApiKey)
	//if info.ChannelType == common.ChannelTypeOpenRouter {
	//	req.Header.Set("HTTP-Referer", "https://github.com/songquanpeng/one-api")
	//	req.Header.Set("X-Title", "One API")
	//}
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	if info.ChannelType != common.ChannelTypeOpenAI {
		request.StreamOptions = nil
	}
	return request, nil
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	a.ResponseFormat = request.ResponseFormat
	if info.RelayMode == constant.RelayModeAudioSpeech {
		jsonData, err := json.Marshal(request)
		if err != nil {
			return nil, fmt.Errorf("error marshalling object: %w", err)
		}
		return bytes.NewReader(jsonData), nil
	} else {
		var requestBody bytes.Buffer
		writer := multipart.NewWriter(&requestBody)

		writer.WriteField("model", request.Model)

		// 添加文件字段
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			return nil, errors.New("file is required")
		}
		defer file.Close()

		part, err := writer.CreateFormFile("file", header.Filename)
		if err != nil {
			return nil, errors.New("create form file failed")
		}
		if _, err := io.Copy(part, file); err != nil {
			return nil, errors.New("copy file failed")
		}

		// 关闭 multipart 编写器以设置分界线
		writer.Close()
		c.Request.Header.Set("Content-Type", writer.FormDataContentType())
		return &requestBody, nil
	}
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	if info.RelayMode == constant.RelayModeAudioTranscription || info.RelayMode == constant.RelayModeAudioTranslation {
		return channel.DoFormRequest(a, c, info, requestBody)
	} else {
		return channel.DoApiRequest(a, c, info, requestBody)
	}
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage *dto.Usage, err *dto.OpenAIErrorWithStatusCode) {
	switch info.RelayMode {
	case constant.RelayModeAudioSpeech:
		err, usage = OpenaiTTSHandler(c, resp, info)
	case constant.RelayModeAudioTranslation:
		fallthrough
	case constant.RelayModeAudioTranscription:
		err, usage = OpenaiSTTHandler(c, resp, info, a.ResponseFormat)
	case constant.RelayModeImagesGenerations:
		err, usage = OpenaiTTSHandler(c, resp, info)
	default:
		if info.IsStream {
			err, usage = OaiStreamHandler(c, resp, info)
		} else {
			err, usage = OpenaiHandler(c, resp, info.PromptTokens, info.UpstreamModelName)
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	switch a.ChannelType {
	case common.ChannelType360:
		return ai360.ModelList
	case common.ChannelTypeMoonshot:
		return moonshot.ModelList
	case common.ChannelTypeLingYiWanWu:
		return lingyiwanwu.ModelList
	case common.ChannelTypeMiniMax:
		return minimax.ModelList
	default:
		return ModelList
	}
}

func (a *Adaptor) GetChannelName() string {
	switch a.ChannelType {
	case common.ChannelType360:
		return ai360.ChannelName
	case common.ChannelTypeMoonshot:
		return moonshot.ChannelName
	case common.ChannelTypeLingYiWanWu:
		return lingyiwanwu.ChannelName
	case common.ChannelTypeMiniMax:
		return minimax.ChannelName
	default:
		return ChannelName
	}
}
