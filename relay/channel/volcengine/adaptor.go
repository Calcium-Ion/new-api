package volcengine

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/dto"
	"one-api/relay/channel"
	"one-api/relay/channel/openai"
	relaycommon "one-api/relay/common"
	"one-api/relay/constant"
	"one-api/service"
	"strings"
)

type Adaptor struct {
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
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	switch info.RelayMode {
	case constant.RelayModeChatCompletions:
		if strings.HasPrefix(info.UpstreamModelName, "bot") {
			return fmt.Sprintf("%s/api/v3/bots/chat/completions", info.BaseUrl), nil
		}
		if info.Cache {
			return fmt.Sprintf("%s/api/v3/context/chat/completions", info.BaseUrl), nil
		}
		return fmt.Sprintf("%s/api/v3/chat/completions", info.BaseUrl), nil
	case constant.RelayModeEmbeddings:
		return fmt.Sprintf("%s/api/v3/embeddings", info.BaseUrl), nil
	case constant.RelayModeContext:
		return fmt.Sprintf("%s/api/v3/context/create", info.BaseUrl), nil
	default:
	}
	return "", fmt.Errorf("unsupported relay mode: %d", info.RelayMode)
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	req.Set("Authorization", "Bearer "+info.ApiKey)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *dto.OpenAIErrorWithStatusCode) {
	switch info.RelayMode {
	case constant.RelayModeChatCompletions:
		if info.IsStream {
			err, usage = openai.OaiStreamHandler(c, resp, info)
		} else {
			err, usage = openai.OpenaiHandler(c, resp, info.PromptTokens, info.UpstreamModelName)
		}
	case constant.RelayModeEmbeddings:
		err, usage = openai.OpenaiHandler(c, resp, info.PromptTokens, info.UpstreamModelName)
	case constant.RelayModeContext:
		err, usage = BaseHandler(c, resp, info.PromptTokens, info.UpstreamModelName)
	}
	return
}

func BaseHandler(c *gin.Context, resp *http.Response, promptTokens int, model string) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var simpleResponse = dto.SimpleResponse{
		Usage: dto.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: 0,
			TotalTokens:      promptTokens,
		},
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	//err = json.Unmarshal(responseBody, &simpleResponse)
	//if err != nil {
	//	return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	//}
	//if simpleResponse.Error.Type != "" {
	//	return &dto.OpenAIErrorWithStatusCode{
	//		Error:      simpleResponse.Error,
	//		StatusCode: resp.StatusCode,
	//	}, nil
	//}
	// Reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the httpClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	resp.Body.Close()
	//if simpleResponse.Usage.TotalTokens == 0 || (simpleResponse.Usage.PromptTokens == 0 && simpleResponse.Usage.CompletionTokens == 0) {
	//	completionTokens := 0
	//	for _, choice := range simpleResponse.Choices {
	//		ctkm, _ := service.CountTextToken(string(choice.Message.Content), model)
	//		completionTokens += ctkm
	//	}
	//	simpleResponse.Usage = dto.Usage{
	//		PromptTokens:     promptTokens,
	//		CompletionTokens: completionTokens,
	//		TotalTokens:      promptTokens + completionTokens,
	//	}
	//}
	return nil, &simpleResponse.Usage
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
