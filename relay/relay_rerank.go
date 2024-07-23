package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/service"
)

func getRerankPromptToken(rerankRequest dto.RerankRequest) int {
	token, _ := service.CountTokenInput(rerankRequest.Query, rerankRequest.Model)
	for _, document := range rerankRequest.Documents {
		tkm, err := service.CountTokenInput(document, rerankRequest.Model)
		if err == nil {
			token += tkm
		}
	}
	return token
}

func RerankHelper(c *gin.Context, relayMode int) *dto.OpenAIErrorWithStatusCode {
	relayInfo := relaycommon.GenRelayInfo(c)

	var rerankRequest *dto.RerankRequest
	err := common.UnmarshalBodyReusable(c, &rerankRequest)
	if err != nil {
		common.LogError(c, fmt.Sprintf("getAndValidateTextRequest failed: %s", err.Error()))
		return service.OpenAIErrorWrapperLocal(err, "invalid_text_request", http.StatusBadRequest)
	}
	if rerankRequest.Query == "" {
		return service.OpenAIErrorWrapperLocal(fmt.Errorf("query is empty"), "invalid_query", http.StatusBadRequest)
	}
	if len(rerankRequest.Documents) == 0 {
		return service.OpenAIErrorWrapperLocal(fmt.Errorf("documents is empty"), "invalid_documents", http.StatusBadRequest)
	}
	relayInfo.UpstreamModelName = rerankRequest.Model
	modelPrice, success := common.GetModelPrice(rerankRequest.Model, false)
	groupRatio := common.GetGroupRatio(relayInfo.Group)

	var preConsumedQuota int
	var ratio float64
	var modelRatio float64

	promptToken := getRerankPromptToken(*rerankRequest)
	if !success {
		preConsumedTokens := promptToken
		modelRatio = common.GetModelRatio(rerankRequest.Model)
		ratio = modelRatio * groupRatio
		preConsumedQuota = int(float64(preConsumedTokens) * ratio)
	} else {
		preConsumedQuota = int(modelPrice * common.QuotaPerUnit * groupRatio)
	}
	relayInfo.PromptTokens = promptToken

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

	convertedRequest, err := adaptor.ConvertRerankRequest(c, relayInfo.RelayMode, *rerankRequest)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "convert_request_failed", http.StatusInternalServerError)
	}
	jsonData, err := json.Marshal(convertedRequest)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "json_marshal_failed", http.StatusInternalServerError)
	}
	requestBody := bytes.NewBuffer(jsonData)
	statusCodeMappingStr := c.GetString("status_code_mapping")
	resp, err := adaptor.DoRequest(c, relayInfo, requestBody)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}
	if resp != nil {
		if resp.StatusCode != http.StatusOK {
			returnPreConsumedQuota(c, relayInfo.TokenId, userQuota, preConsumedQuota)
			openaiErr := service.RelayErrorHandler(resp)
			// reset status code 重置状态码
			service.ResetStatusCode(openaiErr, statusCodeMappingStr)
			return openaiErr
		}
	}

	usage, openaiErr := adaptor.DoResponse(c, resp, relayInfo)
	if openaiErr != nil {
		returnPreConsumedQuota(c, relayInfo.TokenId, userQuota, preConsumedQuota)
		// reset status code 重置状态码
		service.ResetStatusCode(openaiErr, statusCodeMappingStr)
		return openaiErr
	}
	postConsumeQuota(c, relayInfo, rerankRequest.Model, usage, ratio, preConsumedQuota, userQuota, modelRatio, groupRatio, modelPrice, success, "")
	return nil
}
