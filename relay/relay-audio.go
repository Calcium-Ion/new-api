package relay

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common"
	"one-api/dto"
	"one-api/metrics"
	relaycommon "one-api/relay/common"
	relayconstant "one-api/relay/constant"
	"one-api/relay/helper"
	"one-api/service"
	"one-api/setting"
	"strconv"
	"strings"
	"time"
)

func getAndValidAudioRequest(c *gin.Context, info *relaycommon.RelayInfo) (*dto.AudioRequest, error) {
	audioRequest := &dto.AudioRequest{}
	err := common.UnmarshalBodyReusable(c, audioRequest)
	if err != nil {
		return nil, err
	}
	switch info.RelayMode {
	case relayconstant.RelayModeAudioSpeech:
		if audioRequest.Model == "" {
			return nil, errors.New("model is required")
		}
		if setting.ShouldCheckPromptSensitive() {
			words, err := service.CheckSensitiveInput(audioRequest.Input)
			if err != nil {
				common.LogWarn(c, fmt.Sprintf("user sensitive words detected: %s", strings.Join(words, ",")))
				return nil, err
			}
		}
	default:
		err = c.Request.ParseForm()
		if err != nil {
			return nil, err
		}
		formData := c.Request.PostForm
		if audioRequest.Model == "" {
			audioRequest.Model = formData.Get("model")
		}

		if audioRequest.Model == "" {
			return nil, errors.New("model is required")
		}
		audioRequest.ResponseFormat = formData.Get("response_format")
		if audioRequest.ResponseFormat == "" {
			audioRequest.ResponseFormat = "json"
		}
	}
	return audioRequest, nil
}

func AudioInfo(c *gin.Context) (*relaycommon.RelayInfo, *dto.AudioRequest, *dto.OpenAIErrorWithStatusCode) {
	relayInfo := relaycommon.GenRelayInfo(c)
	audioRequest, err := getAndValidAudioRequest(c, relayInfo)
	if err != nil {
		common.LogError(c, fmt.Sprintf("getAndValidAudioRequest failed: %s", err.Error()))
		return nil, nil, service.OpenAIErrorWrapper(err, "invalid_audio_request", http.StatusBadRequest)
	}

	return relayInfo, audioRequest, nil
}

func AudioHelper(c *gin.Context, relayInfo *relaycommon.RelayInfo, audioRequest *dto.AudioRequest) (openaiErr *dto.OpenAIErrorWithStatusCode) {
	startTime := time.Now()
	var funcErr *dto.OpenAIErrorWithStatusCode
	metrics.IncrementRelayRequestTotalCounter(strconv.Itoa(relayInfo.ChannelId), audioRequest.Model, relayInfo.Group, 1)
	defer func() {
		if funcErr != nil {
			metrics.IncrementRelayRequestFailedCounter(strconv.Itoa(relayInfo.ChannelId), audioRequest.Model, relayInfo.Group, strconv.Itoa(funcErr.StatusCode), 1)
		} else {
			metrics.IncrementRelayRequestSuccessCounter(strconv.Itoa(relayInfo.ChannelId), audioRequest.Model, relayInfo.Group, 1)
			metrics.ObserveRelayRequestDuration(strconv.Itoa(relayInfo.ChannelId), audioRequest.Model, relayInfo.Group, time.Since(startTime).Seconds())
		}
	}()
	var (
		err          error
		promptTokens = 0
	)
	preConsumedTokens := common.PreConsumedQuota
	if relayInfo.RelayMode == relayconstant.RelayModeAudioSpeech {
		promptTokens, err = service.CountTTSToken(audioRequest.Input, audioRequest.Model)
		if err != nil {
			funcErr = service.OpenAIErrorWrapper(err, "count_audio_token_failed", http.StatusInternalServerError)
			return funcErr
		}
		preConsumedTokens = promptTokens
		relayInfo.PromptTokens = promptTokens
	}

	priceData, err := helper.ModelPriceHelper(c, relayInfo, preConsumedTokens, 0)
	if err != nil {
		funcErr = service.OpenAIErrorWrapperLocal(err, "model_price_error", http.StatusInternalServerError)
		return funcErr
	}

	preConsumedQuota, userQuota, openaiErr := preConsumeQuota(c, priceData.ShouldPreConsumedQuota, relayInfo)
	if openaiErr != nil {
		funcErr = openaiErr
		return openaiErr
	}
	defer func() {
		if openaiErr != nil {
			returnPreConsumedQuota(c, relayInfo, userQuota, preConsumedQuota)
		}
	}()

	err = helper.ModelMappedHelper(c, relayInfo)
	if err != nil {
		funcErr = service.OpenAIErrorWrapperLocal(err, "model_mapped_error", http.StatusInternalServerError)
		return funcErr
	}

	audioRequest.Model = relayInfo.UpstreamModelName

	adaptor := GetAdaptor(relayInfo.ApiType)
	if adaptor == nil {
		funcErr = service.OpenAIErrorWrapperLocal(fmt.Errorf("invalid api type: %d", relayInfo.ApiType), "invalid_api_type", http.StatusBadRequest)
		return funcErr
	}
	adaptor.Init(relayInfo)

	ioReader, err := adaptor.ConvertAudioRequest(c, relayInfo, *audioRequest)
	if err != nil {
		funcErr = service.OpenAIErrorWrapperLocal(err, "convert_request_failed", http.StatusInternalServerError)
		return funcErr
	}

	resp, err := adaptor.DoRequest(c, relayInfo, ioReader)
	if err != nil {
		funcErr = service.OpenAIErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
		return funcErr
	}
	statusCodeMappingStr := c.GetString("status_code_mapping")

	var httpResp *http.Response
	if resp != nil {
		httpResp = resp.(*http.Response)
		if httpResp.StatusCode != http.StatusOK {
			openaiErr = service.RelayErrorHandler(httpResp)
			funcErr = openaiErr
			// reset status code 重置状态码
			service.ResetStatusCode(openaiErr, statusCodeMappingStr)
			return openaiErr
		}
	}

	usage, openaiErr := adaptor.DoResponse(c, httpResp, relayInfo)
	if openaiErr != nil {
		funcErr = openaiErr
		// reset status code 重置状态码
		service.ResetStatusCode(openaiErr, statusCodeMappingStr)
		return openaiErr
	}

	postConsumeQuota(c, relayInfo, usage.(*dto.Usage), preConsumedQuota, userQuota, priceData, "")

	return nil
}
