package relay

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"one-api/common"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/service"
	"one-api/setting"
	"one-api/setting/operation_setting"
)

func WssHelper(c *gin.Context, ws *websocket.Conn) (openaiErr *dto.OpenAIErrorWithStatusCode) {
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
	modelPrice, getModelPriceSuccess := operation_setting.GetModelPrice(relayInfo.UpstreamModelName, false)
	groupRatio := setting.GetGroupRatio(relayInfo.Group)

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
		modelRatio, _ = operation_setting.GetModelRatio(relayInfo.UpstreamModelName)
		ratio = modelRatio * groupRatio
		preConsumedQuota = int(float64(preConsumedTokens) * ratio)
	} else {
		preConsumedQuota = int(modelPrice * common.QuotaPerUnit * groupRatio)
		relayInfo.UsePrice = true
	}

	// pre-consume quota 预消耗配额
	preConsumedQuota, userQuota, openaiErr := preConsumeQuota(c, preConsumedQuota, relayInfo)
	if openaiErr != nil {
		return openaiErr
	}

	defer func() {
		if openaiErr != nil {
			returnPreConsumedQuota(c, relayInfo, userQuota, preConsumedQuota)
		}
	}()

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
		// reset status code 重置状态码
		service.ResetStatusCode(openaiErr, statusCodeMappingStr)
		return openaiErr
	}
	service.PostWssConsumeQuota(c, relayInfo, relayInfo.UpstreamModelName, usage.(*dto.RealtimeUsage), preConsumedQuota,
		userQuota, modelRatio, groupRatio, modelPrice, getModelPriceSuccess, "")
	return nil
}
