package service

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"sync"
)

var logs = make([]map[string]interface{}, 0)
var lockLogs = sync.Mutex{}

func getCommonWebHookData(ctx *gin.Context, info *relaycommon.RelayInfo, webHookData, toMap map[string]interface{}) map[string]interface{} {
	for k, v := range toMap {
		switch v {
		case "{time}":
			webHookData[k] = info.StartTime
		case "{user_id}":
			webHookData[k] = info.UserId
		case "{token_id}":
			webHookData[k] = info.TokenId
		case "{channel_id}":
			webHookData[k] = info.ChannelId
		case "{response_id}":
			webHookData[k] = GetResponseID(ctx)
		case "{request_model}":
			webHookData[k] = info.UpstreamModelName
		case "{stream}":
			webHookData[k] = info.IsStream
		}
	}
	return webHookData
}

func GenerateInfoWebHookData(ctx *gin.Context, info *relaycommon.RelayInfo, usage dto.Usage, quota int) map[string]interface{} {
	webHookData := make(map[string]interface{})
	toMap := common.StrToMap(constant.WebHookDataMapStr)
	webHookData = getCommonWebHookData(ctx, info, webHookData, toMap)
	for k, v := range toMap {
		switch v {
		case "{prompt_token}":
			webHookData[k] = usage.PromptTokens
		case "{completion_token}":
			webHookData[k] = usage.CompletionTokens
		case "{quota}":
			webHookData[k] = float64(quota) / common.QuotaPerUnit
		}
	}
	webHookData["level"] = "INFO"
	return webHookData
}

func GenerateErrorWebHookData(ctx *gin.Context, info *relaycommon.RelayInfo, err dto.OpenAIErrorWithStatusCode) map[string]interface{} {
	webHookData := make(map[string]interface{})
	toMap := common.StrToMap(constant.WebHookDataMapStr)
	webHookData = getCommonWebHookData(ctx, info, webHookData, toMap)
	for k, v := range toMap {
		switch v {
		case "{prompt_token}":
			webHookData[k] = info.PromptTokens
		case "{status_code}":
			webHookData[k] = err.StatusCode
		case "{err}":
			webHookData[k] = err.Error.ToMessageString()
		}
	}
	webHookData["level"] = "ERROR"
	return webHookData
}

func webHook(data []map[string]interface{}) error {
	jsonBytes, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", constant.WebHookUrl, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	for k, v := range constant.WebHookHeaders {
		req.Header.Set(k, v)
	}

	client := GetHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
