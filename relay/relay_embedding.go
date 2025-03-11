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
	relayconstant "one-api/relay/constant"
	"one-api/relay/helper"
	"one-api/service"
)

func getEmbeddingPromptToken(embeddingRequest dto.EmbeddingRequest) int {
	token, _ := service.CountTokenInput(embeddingRequest.Input, embeddingRequest.Model)
	return token
}

func validateEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, embeddingRequest dto.EmbeddingRequest) error {
	if embeddingRequest.Input == nil {
		return fmt.Errorf("input is empty")
	}
	if info.RelayMode == relayconstant.RelayModeModerations && embeddingRequest.Model == "" {
		embeddingRequest.Model = "omni-moderation-latest"
	}
	if info.RelayMode == relayconstant.RelayModeEmbeddings && embeddingRequest.Model == "" {
		embeddingRequest.Model = c.Param("model")
	}
	return nil
}

func EmbeddingHelper(c *gin.Context) (openaiErr *dto.OpenAIErrorWithStatusCode) {
	relayInfo := relaycommon.GenRelayInfo(c)

	var embeddingRequest *dto.EmbeddingRequest
	err := common.UnmarshalBodyReusable(c, &embeddingRequest)
	if err != nil {
		common.LogError(c, fmt.Sprintf("getAndValidateTextRequest failed: %s", err.Error()))
		return service.OpenAIErrorWrapperLocal(err, "invalid_text_request", http.StatusBadRequest)
	}

	err = validateEmbeddingRequest(c, relayInfo, *embeddingRequest)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "invalid_embedding_request", http.StatusBadRequest)
	}

	err = helper.ModelMappedHelper(c, relayInfo)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "model_mapped_error", http.StatusInternalServerError)
	}

	embeddingRequest.Model = relayInfo.UpstreamModelName

	promptToken := getEmbeddingPromptToken(*embeddingRequest)
	relayInfo.PromptTokens = promptToken

	priceData, err := helper.ModelPriceHelper(c, relayInfo, promptToken, 0)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "model_price_error", http.StatusInternalServerError)
	}
	// pre-consume quota 预消耗配额
	preConsumedQuota, userQuota, openaiErr := preConsumeQuota(c, priceData.ShouldPreConsumedQuota, relayInfo)
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

	convertedRequest, err := adaptor.ConvertEmbeddingRequest(c, relayInfo, *embeddingRequest)

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

	var httpResp *http.Response
	if resp != nil {
		httpResp = resp.(*http.Response)
		if httpResp.StatusCode != http.StatusOK {
			openaiErr = service.RelayErrorHandler(httpResp, false)
			// reset status code 重置状态码
			service.ResetStatusCode(openaiErr, statusCodeMappingStr)
			return openaiErr
		}
	}

	usage, openaiErr := adaptor.DoResponse(c, httpResp, relayInfo)
	if openaiErr != nil {
		// reset status code 重置状态码
		service.ResetStatusCode(openaiErr, statusCodeMappingStr)
		return openaiErr
	}
	postConsumeQuota(c, relayInfo, usage.(*dto.Usage), preConsumedQuota, userQuota, priceData, "")
	return nil
}
