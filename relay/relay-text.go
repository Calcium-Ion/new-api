package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
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

	"github.com/gin-gonic/gin"
)

func getAndValidateTextRequest(c *gin.Context, relayInfo *relaycommon.RelayInfo) (*dto.GeneralOpenAIRequest, error) {
	textRequest := &dto.GeneralOpenAIRequest{}
	err := common.UnmarshalBodyReusable(c, textRequest)
	if err != nil {
		return nil, err
	}
	if relayInfo.RelayMode == relayconstant.RelayModeModerations && textRequest.Model == "" {
		textRequest.Model = "text-moderation-latest"
	}
	if relayInfo.RelayMode == relayconstant.RelayModeEmbeddings && textRequest.Model == "" {
		textRequest.Model = c.Param("model")
	}

	if textRequest.MaxTokens < 0 || textRequest.MaxTokens > math.MaxInt32/2 {
		return nil, errors.New("max_tokens is invalid")
	}
	if textRequest.Model == "" {
		return nil, errors.New("model is required")
	}
	switch relayInfo.RelayMode {
	case relayconstant.RelayModeCompletions:
		if textRequest.Prompt == "" {
			return nil, errors.New("field prompt is required")
		}
	case relayconstant.RelayModeChatCompletions:
		if textRequest.Messages == nil || len(textRequest.Messages) == 0 {
			return nil, errors.New("field messages is required")
		}
	case relayconstant.RelayModeEmbeddings:
	case relayconstant.RelayModeModerations:
		if textRequest.Input == "" {
			return nil, errors.New("field input is required")
		}
	case relayconstant.RelayModeEdits:
		if textRequest.Instruction == "" {
			return nil, errors.New("field instruction is required")
		}
	}
	relayInfo.IsStream = textRequest.Stream
	return textRequest, nil
}

func TextHelper(c *gin.Context) *dto.OpenAIErrorWithStatusCode {

	relayInfo := relaycommon.GenRelayInfo(c)

	// get & validate textRequest 获取并验证文本请求
	textRequest, err := getAndValidateTextRequest(c, relayInfo)
	if err != nil {
		common.LogError(c, fmt.Sprintf("getAndValidateTextRequest failed: %s", err.Error()))
		return service.OpenAIErrorWrapperLocal(err, "invalid_text_request", http.StatusBadRequest)
	}

	// map model name
	modelMapping := c.GetString("model_mapping")
	isModelMapped := false
	if modelMapping != "" && modelMapping != "{}" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return service.OpenAIErrorWrapperLocal(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
		}
		if modelMap[textRequest.Model] != "" {
			textRequest.Model = modelMap[textRequest.Model]
			// set upstream model name
			isModelMapped = true
		}
	}
	relayInfo.UpstreamModelName = textRequest.Model
	modelPrice := common.GetModelPrice(textRequest.Model, false)
	groupRatio := common.GetGroupRatio(relayInfo.Group)

	var preConsumedQuota int
	var ratio float64
	var modelRatio float64
	//err := service.SensitiveWordsCheck(textRequest)
	promptTokens, err, sensitiveTrigger := getPromptTokens(textRequest, relayInfo)

	// count messages token error 计算promptTokens错误
	if err != nil {
		if sensitiveTrigger {
			return service.OpenAIErrorWrapperLocal(err, "sensitive_words_detected", http.StatusBadRequest)
		}
		return service.OpenAIErrorWrapper(err, "count_token_messages_failed", http.StatusInternalServerError)
	}

	if modelPrice == -1 {
		preConsumedTokens := common.PreConsumedQuota
		if textRequest.MaxTokens != 0 {
			preConsumedTokens = promptTokens + int(textRequest.MaxTokens)
		}
		modelRatio = common.GetModelRatio(textRequest.Model)
		ratio = modelRatio * groupRatio
		preConsumedQuota = int(float64(preConsumedTokens) * ratio)
	} else {
		preConsumedQuota = int(modelPrice * common.QuotaPerUnit * groupRatio)
	}

	// pre-consume quota 预消耗配额
	preConsumedQuota, userQuota, openaiErr := preConsumeQuota(c, preConsumedQuota, relayInfo)
	if openaiErr != nil {
		return openaiErr
	}

	adaptor := GetAdaptor(relayInfo.ApiType)
	if adaptor == nil {
		return service.OpenAIErrorWrapper(fmt.Errorf("invalid api type: %d", relayInfo.ApiType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(relayInfo, *textRequest)
	var requestBody io.Reader
	if relayInfo.ApiType == relayconstant.APITypeOpenAI {
		if isModelMapped {
			jsonStr, err := json.Marshal(textRequest)
			if err != nil {
				return service.OpenAIErrorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
			}
			requestBody = bytes.NewBuffer(jsonStr)
		} else {
			requestBody = c.Request.Body
		}
	} else {
		convertedRequest, err := adaptor.ConvertRequest(c, relayInfo.RelayMode, textRequest)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "convert_request_failed", http.StatusInternalServerError)
		}
		jsonData, err := json.Marshal(convertedRequest)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}

	resp, err := adaptor.DoRequest(c, relayInfo, requestBody)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}
	relayInfo.IsStream = relayInfo.IsStream || strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream")

	if resp.StatusCode != http.StatusOK {
		returnPreConsumedQuota(c, relayInfo.TokenId, userQuota, preConsumedQuota)
		return service.RelayErrorHandler(resp)
	}

	usage, openaiErr := adaptor.DoResponse(c, resp, relayInfo)
	if openaiErr != nil {
		returnPreConsumedQuota(c, relayInfo.TokenId, userQuota, preConsumedQuota)
		return openaiErr
	}
	postConsumeQuota(c, relayInfo, *textRequest, usage, ratio, preConsumedQuota, userQuota, modelRatio, groupRatio, modelPrice)
	return nil
}

func getPromptTokens(textRequest *dto.GeneralOpenAIRequest, info *relaycommon.RelayInfo) (int, error, bool) {
	var promptTokens int
	var err error
	var sensitiveTrigger bool
	checkSensitive := constant.ShouldCheckPromptSensitive()
	switch info.RelayMode {
	case relayconstant.RelayModeChatCompletions:
		promptTokens, err, sensitiveTrigger = service.CountTokenMessages(textRequest.Messages, textRequest.Model, checkSensitive)
	case relayconstant.RelayModeCompletions:
		promptTokens, err, sensitiveTrigger = service.CountTokenInput(textRequest.Prompt, textRequest.Model, checkSensitive)
	case relayconstant.RelayModeModerations:
		promptTokens, err, sensitiveTrigger = service.CountTokenInput(textRequest.Input, textRequest.Model, checkSensitive)
	case relayconstant.RelayModeEmbeddings:
		promptTokens, err, sensitiveTrigger = service.CountTokenInput(textRequest.Input, textRequest.Model, checkSensitive)
	default:
		err = errors.New("unknown relay mode")
		promptTokens = 0
	}
	info.PromptTokens = promptTokens
	return promptTokens, err, sensitiveTrigger
}

// 预扣费并返回用户剩余配额
func preConsumeQuota(c *gin.Context, preConsumedQuota int, relayInfo *relaycommon.RelayInfo) (int, int, *dto.OpenAIErrorWithStatusCode) {
	userQuota, err := model.CacheGetUserQuota(relayInfo.UserId)
	if err != nil {
		return 0, 0, service.OpenAIErrorWrapperLocal(err, "get_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota <= 0 || userQuota-preConsumedQuota < 0 {
		return 0, 0, service.OpenAIErrorWrapperLocal(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}
	err = model.CacheDecreaseUserQuota(relayInfo.UserId, preConsumedQuota)
	if err != nil {
		return 0, 0, service.OpenAIErrorWrapperLocal(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota > 100*preConsumedQuota {
		// 用户额度充足，判断令牌额度是否充足
		if !relayInfo.TokenUnlimited {
			// 非无限令牌，判断令牌额度是否充足
			tokenQuota := c.GetInt("token_quota")
			if tokenQuota > 100*preConsumedQuota {
				// 令牌额度充足，信任令牌
				preConsumedQuota = 0
				common.LogInfo(c.Request.Context(), fmt.Sprintf("user %d quota %d and token %d quota %d are enough, trusted and no need to pre-consume", relayInfo.UserId, userQuota, relayInfo.TokenId, tokenQuota))
			}
		} else {
			// in this case, we do not pre-consume quota
			// because the user has enough quota
			preConsumedQuota = 0
			common.LogInfo(c.Request.Context(), fmt.Sprintf("user %d with unlimited token has enough quota %d, trusted and no need to pre-consume", relayInfo.UserId, userQuota))
		}
	}
	if preConsumedQuota > 0 {
		userQuota, err = model.PreConsumeTokenQuota(relayInfo.TokenId, preConsumedQuota)
		if err != nil {
			return 0, 0, service.OpenAIErrorWrapperLocal(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}
	return preConsumedQuota, userQuota, nil
}

func returnPreConsumedQuota(c *gin.Context, tokenId int, userQuota int, preConsumedQuota int) {
	if preConsumedQuota != 0 {
		go func(ctx context.Context) {
			// return pre-consumed quota
			err := model.PostConsumeTokenQuota(tokenId, userQuota, -preConsumedQuota, 0, false)
			if err != nil {
				common.SysError("error return pre-consumed quota: " + err.Error())
			}
		}(c)
	}
}

func postConsumeQuota(ctx *gin.Context, relayInfo *relaycommon.RelayInfo, textRequest dto.GeneralOpenAIRequest,
	usage *dto.Usage, ratio float64, preConsumedQuota int, userQuota int, modelRatio float64, groupRatio float64,
	modelPrice float64) {

	useTimeSeconds := time.Now().Unix() - relayInfo.StartTime.Unix()
	promptTokens := usage.PromptTokens
	completionTokens := usage.CompletionTokens

	tokenName := ctx.GetString("token_name")

	quota := 0
	if modelPrice == -1 {
		completionRatio := common.GetCompletionRatio(textRequest.Model)
		quota = promptTokens + int(float64(completionTokens)*completionRatio)
		quota = int(float64(quota) * ratio)
		if ratio != 0 && quota <= 0 {
			quota = 1
		}
	} else {
		quota = int(modelPrice * common.QuotaPerUnit * groupRatio)
	}
	totalTokens := promptTokens + completionTokens
	var logContent string
	if modelPrice == -1 {
		logContent = fmt.Sprintf("模型倍率 %.2f，分组倍率 %.2f", modelRatio, groupRatio)
	} else {
		logContent = fmt.Sprintf("模型价格 %.2f，分组倍率 %.2f", modelPrice, groupRatio)
	}

	// record all the consume log even if quota is 0
	if totalTokens == 0 {
		// in this case, must be some error happened
		// we cannot just return, because we may have to return the pre-consumed quota
		quota = 0
		logContent += fmt.Sprintf("（可能是上游超时）")
		common.LogError(ctx, fmt.Sprintf("total tokens is 0, cannot consume quota, userId %d, channelId %d, tokenId %d, model %s， pre-consumed quota %d", relayInfo.UserId, relayInfo.ChannelId, relayInfo.TokenId, textRequest.Model, preConsumedQuota))
	} else {
		//if sensitiveResp != nil {
		//	logContent += fmt.Sprintf("，敏感词：%s", strings.Join(sensitiveResp.SensitiveWords, ", "))
		//}
		quotaDelta := quota - preConsumedQuota
		if quotaDelta != 0 {
			err := model.PostConsumeTokenQuota(relayInfo.TokenId, userQuota, quotaDelta, preConsumedQuota, true)
			if err != nil {
				common.LogError(ctx, "error consuming token remain quota: "+err.Error())
			}
		}
		err := model.CacheUpdateUserQuota(relayInfo.UserId)
		if err != nil {
			common.LogError(ctx, "error update user quota cache: "+err.Error())
		}
		model.UpdateUserUsedQuotaAndRequestCount(relayInfo.UserId, quota)
		model.UpdateChannelUsedQuota(relayInfo.ChannelId, quota)
	}

	logModel := textRequest.Model
	if strings.HasPrefix(logModel, "gpt-4-gizmo") {
		logModel = "gpt-4-gizmo-*"
		logContent += fmt.Sprintf("，模型 %s", textRequest.Model)
	}
	model.RecordConsumeLog(ctx, relayInfo.UserId, relayInfo.ChannelId, promptTokens, completionTokens, logModel, tokenName, quota, logContent, relayInfo.TokenId, userQuota, int(useTimeSeconds), relayInfo.IsStream)

	//if quota != 0 {
	//
	//}
}
