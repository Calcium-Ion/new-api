package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	"one-api/model"
	relaycommon "one-api/relay/common"
	relayconstant "one-api/relay/constant"
	"one-api/service"
	"strings"
	"time"
)

func AudioHelper(c *gin.Context, relayMode int) *dto.OpenAIErrorWithStatusCode {
	tokenId := c.GetInt("token_id")
	channelType := c.GetInt("channel")
	channelId := c.GetInt("channel_id")
	userId := c.GetInt("id")
	group := c.GetString("group")
	startTime := time.Now()

	var audioRequest dto.TextToSpeechRequest
	if !strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions") {
		err := common.UnmarshalBodyReusable(c, &audioRequest)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "bind_request_body_failed", http.StatusBadRequest)
		}
	} else {
		audioRequest = dto.TextToSpeechRequest{
			Model: "whisper-1",
		}
	}
	//err := common.UnmarshalBodyReusable(c, &audioRequest)

	// request validation
	if audioRequest.Model == "" {
		return service.OpenAIErrorWrapper(errors.New("model is required"), "required_field_missing", http.StatusBadRequest)
	}

	if strings.HasPrefix(audioRequest.Model, "tts-1") {
		if audioRequest.Voice == "" {
			return service.OpenAIErrorWrapper(errors.New("voice is required"), "required_field_missing", http.StatusBadRequest)
		}
	}
	var err error
	promptTokens := 0
	preConsumedTokens := common.PreConsumedQuota
	if strings.HasPrefix(audioRequest.Model, "tts-1") {
		if constant.ShouldCheckPromptSensitive() {
			err = service.CheckSensitiveInput(audioRequest.Input)
			if err != nil {
				return service.OpenAIErrorWrapper(err, "sensitive_words_detected", http.StatusBadRequest)
			}
		}
		promptTokens, err = service.CountAudioToken(audioRequest.Input, audioRequest.Model)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "count_audio_token_failed", http.StatusInternalServerError)
		}
		preConsumedTokens = promptTokens
	}
	modelRatio := common.GetModelRatio(audioRequest.Model)
	groupRatio := common.GetGroupRatio(group)
	ratio := modelRatio * groupRatio
	preConsumedQuota := int(float64(preConsumedTokens) * ratio)
	userQuota, err := model.CacheGetUserQuota(userId)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "get_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota-preConsumedQuota < 0 {
		return service.OpenAIErrorWrapperLocal(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}
	err = model.CacheDecreaseUserQuota(userId, preConsumedQuota)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota > 100*preConsumedQuota {
		// in this case, we do not pre-consume quota
		// because the user has enough quota
		preConsumedQuota = 0
	}
	if preConsumedQuota > 0 {
		userQuota, err = model.PreConsumeTokenQuota(tokenId, preConsumedQuota)
		if err != nil {
			return service.OpenAIErrorWrapperLocal(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}

	succeed := false
	defer func() {
		if succeed {
			return
		}
		if preConsumedQuota > 0 {
			// we need to roll back the pre-consumed quota
			defer func() {
				go func() {
					// negative means add quota back for token & user
					returnPreConsumedQuota(c, tokenId, userQuota, preConsumedQuota)
				}()
			}()
		}
	}()

	// map model name
	modelMapping := c.GetString("model_mapping")
	if modelMapping != "" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
		}
		if modelMap[audioRequest.Model] != "" {
			audioRequest.Model = modelMap[audioRequest.Model]
		}
	}

	baseURL := common.ChannelBaseURLs[channelType]
	requestURL := c.Request.URL.String()
	if c.GetString("base_url") != "" {
		baseURL = c.GetString("base_url")
	}

	fullRequestURL := relaycommon.GetFullRequestURL(baseURL, requestURL, channelType)
	if relayMode == relayconstant.RelayModeAudioTranscription && channelType == common.ChannelTypeAzure {
		// https://learn.microsoft.com/en-us/azure/ai-services/openai/whisper-quickstart?tabs=command-line#rest-api
		apiVersion := relaycommon.GetAPIVersion(c)
		fullRequestURL = fmt.Sprintf("%s/openai/deployments/%s/audio/transcriptions?api-version=%s", baseURL, audioRequest.Model, apiVersion)
	}

	requestBody := c.Request.Body

	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	if relayMode == relayconstant.RelayModeAudioTranscription && channelType == common.ChannelTypeAzure {
		// https://learn.microsoft.com/en-us/azure/ai-services/openai/whisper-quickstart?tabs=command-line#rest-api
		apiKey := c.Request.Header.Get("Authorization")
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		req.Header.Set("api-key", apiKey)
		req.ContentLength = c.Request.ContentLength
	} else {
		req.Header.Set("Authorization", c.Request.Header.Get("Authorization"))
	}

	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))

	resp, err := service.GetHttpClient().Do(req)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	err = req.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_request_body_failed", http.StatusInternalServerError)
	}
	err = c.Request.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_request_body_failed", http.StatusInternalServerError)
	}

	if resp.StatusCode != http.StatusOK {
		return relaycommon.RelayErrorHandler(resp)
	}
	succeed = true

	var audioResponse dto.AudioResponse

	defer func(ctx context.Context) {
		go func() {
			useTimeSeconds := time.Now().Unix() - startTime.Unix()
			quota := 0
			if strings.HasPrefix(audioRequest.Model, "tts-1") {
				quota = promptTokens
			} else {
				quota, err = service.CountAudioToken(audioResponse.Text, audioRequest.Model)
			}
			quota = int(float64(quota) * ratio)
			if ratio != 0 && quota <= 0 {
				quota = 1
			}
			quotaDelta := quota - preConsumedQuota
			err := model.PostConsumeTokenQuota(tokenId, userQuota, quotaDelta, preConsumedQuota, true)
			if err != nil {
				common.SysError("error consuming token remain quota: " + err.Error())
			}
			err = model.CacheUpdateUserQuota(userId)
			if err != nil {
				common.SysError("error update user quota cache: " + err.Error())
			}
			if quota != 0 {
				tokenName := c.GetString("token_name")
				logContent := fmt.Sprintf("模型倍率 %.2f，分组倍率 %.2f", modelRatio, groupRatio)
				other := make(map[string]interface{})
				other["model_ratio"] = modelRatio
				other["group_ratio"] = groupRatio
				model.RecordConsumeLog(ctx, userId, channelId, promptTokens, 0, audioRequest.Model, tokenName, quota, logContent, tokenId, userQuota, int(useTimeSeconds), false, other)
				model.UpdateUserUsedQuotaAndRequestCount(userId, quota)
				channelId := c.GetInt("channel_id")
				model.UpdateChannelUsedQuota(channelId, quota)
			}
		}()
	}(c.Request.Context())

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
	}
	if strings.HasPrefix(audioRequest.Model, "tts-1") {

	} else {
		err = json.Unmarshal(responseBody, &audioResponse)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
		}
		contains, words := service.SensitiveWordContains(audioResponse.Text)
		if contains {
			return service.OpenAIErrorWrapper(errors.New("response contains sensitive words: "+strings.Join(words, ", ")), "response_contains_sensitive_words", http.StatusBadRequest)
		}
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)

	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError)
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
	}
	return nil
}
