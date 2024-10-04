package relay

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"math"
	"net/http"
	"one-api/common"
	"one-api/dto"
	"one-api/model"
	relaycommon "one-api/relay/common"
	"one-api/service"
	"strings"
	"time"
)

//func getAndValidateWssRequest(c *gin.Context, ws *websocket.Conn) (*dto.RealtimeEvent, error) {
//	_, p, err := ws.ReadMessage()
//	if err != nil {
//		return nil, err
//	}
//	realtimeEvent := &dto.RealtimeEvent{}
//	err = json.Unmarshal(p, realtimeEvent)
//	if err != nil {
//		return nil, err
//	}
//	// save the original request
//	if realtimeEvent.Session == nil {
//		return nil, errors.New("session object is nil")
//	}
//	c.Set("first_wss_request", p)
//	return realtimeEvent, nil
//}

func WssHelper(c *gin.Context, ws *websocket.Conn) *dto.OpenAIErrorWithStatusCode {
	relayInfo := relaycommon.GenRelayInfoWs(c, ws)

	// get & validate textRequest 获取并验证文本请求
	//realtimeEvent, err := getAndValidateWssRequest(c, ws)
	//if err != nil {
	//	common.LogError(c, fmt.Sprintf("getAndValidateWssRequest failed: %s", err.Error()))
	//	return service.OpenAIErrorWrapperLocal(err, "invalid_text_request", http.StatusBadRequest)
	//}

	// map model name
	modelMapping := c.GetString("model_mapping")
	//isModelMapped := false
	if modelMapping != "" && modelMapping != "{}" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return service.OpenAIErrorWrapperLocal(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
		}
		if modelMap[relayInfo.OriginModelName] != "" {
			relayInfo.UpstreamModelName = modelMap[relayInfo.OriginModelName]
			// set upstream model name
			//isModelMapped = true
		}
	}
	//relayInfo.UpstreamModelName = textRequest.Model
	modelPrice, getModelPriceSuccess := common.GetModelPrice(relayInfo.UpstreamModelName, false)
	groupRatio := common.GetGroupRatio(relayInfo.Group)

	var preConsumedQuota int
	var ratio float64
	var modelRatio float64
	//err := service.SensitiveWordsCheck(textRequest)

	//if constant.ShouldCheckPromptSensitive() {
	//	err = checkRequestSensitive(textRequest, relayInfo)
	//	if err != nil {
	//		return service.OpenAIErrorWrapperLocal(err, "sensitive_words_detected", http.StatusBadRequest)
	//	}
	//}

	//promptTokens, err := getWssPromptTokens(realtimeEvent, relayInfo)
	//// count messages token error 计算promptTokens错误
	//if err != nil {
	//	return service.OpenAIErrorWrapper(err, "count_token_messages_failed", http.StatusInternalServerError)
	//}
	//
	if !getModelPriceSuccess {
		preConsumedTokens := common.PreConsumedQuota
		//if realtimeEvent.Session.MaxResponseOutputTokens != 0 {
		//	preConsumedTokens = promptTokens + int(realtimeEvent.Session.MaxResponseOutputTokens)
		//}
		modelRatio = common.GetModelRatio(relayInfo.UpstreamModelName)
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
		return service.OpenAIErrorWrapperLocal(fmt.Errorf("invalid api type: %d", relayInfo.ApiType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(relayInfo)
	//var requestBody io.Reader
	//firstWssRequest, _ := c.Get("first_wss_request")
	//requestBody = bytes.NewBuffer(firstWssRequest.([]byte))

	statusCodeMappingStr := c.GetString("status_code_mapping")
	resp, err := adaptor.DoRequest(c, relayInfo, nil)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	if resp != nil {
		relayInfo.TargetWs = resp.(*websocket.Conn)
		defer relayInfo.TargetWs.Close()
	}

	usage, openaiErr := adaptor.DoResponse(c, nil, relayInfo)
	if openaiErr != nil {
		returnPreConsumedQuota(c, relayInfo, userQuota, preConsumedQuota)
		// reset status code 重置状态码
		service.ResetStatusCode(openaiErr, statusCodeMappingStr)
		return openaiErr
	}
	postWssConsumeQuota(c, relayInfo, relayInfo.UpstreamModelName, usage.(*dto.RealtimeUsage), ratio, preConsumedQuota, userQuota, modelRatio, groupRatio, modelPrice, getModelPriceSuccess, "")
	return nil
}

func postWssConsumeQuota(ctx *gin.Context, relayInfo *relaycommon.RelayInfo, modelName string,
	usage *dto.RealtimeUsage, ratio float64, preConsumedQuota int, userQuota int, modelRatio float64,
	groupRatio float64,
	modelPrice float64, usePrice bool, extraContent string) {

	useTimeSeconds := time.Now().Unix() - relayInfo.StartTime.Unix()
	textInputTokens := usage.InputTokenDetails.TextTokens
	textOutTokens := usage.OutputTokenDetails.TextTokens

	audioInputTokens := usage.InputTokenDetails.AudioTokens
	audioOutTokens := usage.OutputTokenDetails.AudioTokens

	tokenName := ctx.GetString("token_name")
	completionRatio := common.GetCompletionRatio(modelName)
	audioRatio := common.GetAudioRatio(relayInfo.UpstreamModelName)
	audioCompletionRatio := common.GetAudioCompletionRatio(modelName)

	quota := 0
	if !usePrice {
		quota = textInputTokens + int(math.Round(float64(textOutTokens)*completionRatio))
		quota += int(math.Round(float64(audioInputTokens)*audioRatio)) + int(math.Round(float64(audioOutTokens)*completionRatio*audioCompletionRatio))

		quota = int(math.Round(float64(quota) * ratio))
		if ratio != 0 && quota <= 0 {
			quota = 1
		}
	} else {
		quota = int(modelPrice * common.QuotaPerUnit * groupRatio)
	}
	totalTokens := usage.TotalTokens
	var logContent string
	if !usePrice {
		logContent = fmt.Sprintf("模型倍率 %.2f，补全倍率 %.2f，音频倍率 %.2f，音频补全倍率 %.2f，分组倍率 %.2f", modelRatio, completionRatio, audioRatio, audioCompletionRatio, groupRatio)
	} else {
		logContent = fmt.Sprintf("模型价格 %.2f，分组倍率 %.2f", modelPrice, groupRatio)
	}

	// record all the consume log even if quota is 0
	if totalTokens == 0 {
		// in this case, must be some error happened
		// we cannot just return, because we may have to return the pre-consumed quota
		quota = 0
		logContent += fmt.Sprintf("（可能是上游超时）")
		common.LogError(ctx, fmt.Sprintf("total tokens is 0, cannot consume quota, userId %d, channelId %d, "+
			"tokenId %d, model %s， pre-consumed quota %d", relayInfo.UserId, relayInfo.ChannelId, relayInfo.TokenId, modelName, preConsumedQuota))
	} else {
		//if sensitiveResp != nil {
		//	logContent += fmt.Sprintf("，敏感词：%s", strings.Join(sensitiveResp.SensitiveWords, ", "))
		//}
		quotaDelta := quota - preConsumedQuota
		if quotaDelta != 0 {
			err := model.PostConsumeTokenQuota(relayInfo, userQuota, quotaDelta, preConsumedQuota, true)
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

	logModel := modelName
	if strings.HasPrefix(logModel, "gpt-4-gizmo") {
		logModel = "gpt-4-gizmo-*"
		logContent += fmt.Sprintf("，模型 %s", modelName)
	}
	if strings.HasPrefix(logModel, "gpt-4o-gizmo") {
		logModel = "gpt-4o-gizmo-*"
		logContent += fmt.Sprintf("，模型 %s", modelName)
	}
	if extraContent != "" {
		logContent += ", " + extraContent
	}
	other := service.GenerateWssOtherInfo(ctx, relayInfo, usage, modelRatio, groupRatio, completionRatio, modelPrice)
	model.RecordConsumeLog(ctx, relayInfo.UserId, relayInfo.ChannelId, usage.InputTokens, usage.OutputTokens, logModel,
		tokenName, quota, logContent, relayInfo.TokenId, userQuota, int(useTimeSeconds), relayInfo.IsStream, other)

	//if quota != 0 {
	//
	//}
}

func getWssPromptTokens(textRequest *dto.RealtimeEvent, info *relaycommon.RelayInfo) (int, error) {
	var promptTokens int
	var err error
	switch info.RelayMode {
	default:
		promptTokens, err = service.CountTokenRealtime(*textRequest, info.UpstreamModelName)
	}
	info.PromptTokens = promptTokens
	return promptTokens, err
}

//func checkWssRequestSensitive(textRequest *dto.GeneralOpenAIRequest, info *relaycommon.RelayInfo) error {
//	var err error
//	switch info.RelayMode {
//	case relayconstant.RelayModeChatCompletions:
//		err = service.CheckSensitiveMessages(textRequest.Messages)
//	case relayconstant.RelayModeCompletions:
//		err = service.CheckSensitiveInput(textRequest.Prompt)
//	case relayconstant.RelayModeModerations:
//		err = service.CheckSensitiveInput(textRequest.Input)
//	case relayconstant.RelayModeEmbeddings:
//		err = service.CheckSensitiveInput(textRequest.Input)
//	}
//	return err
//}
