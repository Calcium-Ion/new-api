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

func validateContextRequest(c *gin.Context, info *relaycommon.RelayInfo, contextRequest dto.ContextRequest) error {

	if contextRequest.Model == "" {
		return fmt.Errorf("model is empty")
	}
	if contextRequest.Ttl < 0 {
		return fmt.Errorf("ttl is negative")
	}
	if len(contextRequest.Messages) == 0 {
		return fmt.Errorf("messages is empty")
	}
	return nil
}

func ContextHelper(c *gin.Context) (openaiErr *dto.OpenAIErrorWithStatusCode) {
	relayInfo := relaycommon.GenRelayInfo(c)

	var contextRequest *dto.ContextRequest
	err := common.UnmarshalBodyReusable(c, &contextRequest)
	if err != nil {
		common.LogError(c, fmt.Sprintf("getAndValidateTextRequest failed: %s", err.Error()))
		return service.OpenAIErrorWrapperLocal(err, "invalid_text_request", http.StatusBadRequest)
	}

	err = validateContextRequest(c, relayInfo, *contextRequest)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "invalid_embedding_request", http.StatusBadRequest)
	}

	// map model name
	modelMapping := c.GetString("model_mapping")
	//isModelMapped := false
	if modelMapping != "" && modelMapping != "{}" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return service.OpenAIErrorWrapperLocal(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
		}
		if modelMap[contextRequest.Model] != "" {
			contextRequest.Model = modelMap[contextRequest.Model]
			// set upstream model name
			//isModelMapped = true
		}
	}

	relayInfo.UpstreamModelName = contextRequest.Model

	adaptor := GetAdaptor(relayInfo.ApiType)
	if adaptor == nil {
		return service.OpenAIErrorWrapperLocal(fmt.Errorf("invalid api type: %d", relayInfo.ApiType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(relayInfo)

	//convertedRequest, err := adaptor.ConvertEmbeddingRequest(c, relayInfo, *contextRequest)

	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "convert_request_failed", http.StatusInternalServerError)
	}
	jsonData, err := json.Marshal(contextRequest)
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
			openaiErr = service.RelayErrorHandler(httpResp)
			// reset status code 重置状态码
			service.ResetStatusCode(openaiErr, statusCodeMappingStr)
			return openaiErr
		}
	}

	_, openaiErr = adaptor.DoResponse(c, httpResp, relayInfo)
	if openaiErr != nil {
		// reset status code 重置状态码
		service.ResetStatusCode(openaiErr, statusCodeMappingStr)
		return openaiErr
	}
	return nil
}
