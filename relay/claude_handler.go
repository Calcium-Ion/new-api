package relay

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/relay/helper"
	"one-api/service"
	"one-api/setting/model_setting"
	"strings"
)

func getAndValidateClaudeRequest(c *gin.Context) (textRequest *dto.ClaudeRequest, err error) {
	textRequest = &dto.ClaudeRequest{}
	err = c.ShouldBindJSON(textRequest)
	if err != nil {
		return nil, err
	}
	if textRequest.Messages == nil || len(textRequest.Messages) == 0 {
		return nil, errors.New("field messages is required")
	}
	if textRequest.Model == "" {
		return nil, errors.New("field model is required")
	}
	return textRequest, nil
}

func ClaudeHelper(c *gin.Context) (claudeError *dto.ClaudeErrorWithStatusCode) {

	relayInfo := relaycommon.GenRelayInfoClaude(c)

	// get & validate textRequest 获取并验证文本请求
	textRequest, err := getAndValidateClaudeRequest(c)
	if err != nil {
		return service.ClaudeErrorWrapperLocal(err, "invalid_claude_request", http.StatusBadRequest)
	}

	if textRequest.Stream {
		relayInfo.IsStream = true
	}

	err = helper.ModelMappedHelper(c, relayInfo)
	if err != nil {
		return service.ClaudeErrorWrapperLocal(err, "model_mapped_error", http.StatusInternalServerError)
	}

	textRequest.Model = relayInfo.UpstreamModelName

	promptTokens, err := getClaudePromptTokens(textRequest, relayInfo)
	// count messages token error 计算promptTokens错误
	if err != nil {
		return service.ClaudeErrorWrapperLocal(err, "count_token_messages_failed", http.StatusInternalServerError)
	}

	priceData, err := helper.ModelPriceHelper(c, relayInfo, promptTokens, int(textRequest.MaxTokens))
	if err != nil {
		return service.ClaudeErrorWrapperLocal(err, "model_price_error", http.StatusInternalServerError)
	}

	// pre-consume quota 预消耗配额
	preConsumedQuota, userQuota, openaiErr := preConsumeQuota(c, priceData.ShouldPreConsumedQuota, relayInfo)

	if openaiErr != nil {
		return service.OpenAIErrorToClaudeError(openaiErr)
	}
	defer func() {
		if openaiErr != nil {
			returnPreConsumedQuota(c, relayInfo, userQuota, preConsumedQuota)
		}
	}()

	adaptor := GetAdaptor(relayInfo.ApiType)
	if adaptor == nil {
		return service.ClaudeErrorWrapperLocal(fmt.Errorf("invalid api type: %d", relayInfo.ApiType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(relayInfo)
	var requestBody io.Reader

	if textRequest.MaxTokens == 0 {
		textRequest.MaxTokens = uint(model_setting.GetClaudeSettings().GetDefaultMaxTokens(textRequest.Model))
	}

	if model_setting.GetClaudeSettings().ThinkingAdapterEnabled &&
		strings.HasSuffix(textRequest.Model, "-thinking") {
		if textRequest.Thinking == nil {
			// 因为BudgetTokens 必须大于1024
			if textRequest.MaxTokens < 1280 {
				textRequest.MaxTokens = 1280
			}

			// BudgetTokens 为 max_tokens 的 80%
			textRequest.Thinking = &dto.Thinking{
				Type:         "enabled",
				BudgetTokens: int(float64(textRequest.MaxTokens) * model_setting.GetClaudeSettings().ThinkingAdapterBudgetTokensPercentage),
			}
			// TODO: 临时处理
			// https://docs.anthropic.com/en/docs/build-with-claude/extended-thinking#important-considerations-when-using-extended-thinking
			textRequest.TopP = 0
			textRequest.Temperature = common.GetPointer[float64](1.0)
		}
		textRequest.Model = strings.TrimSuffix(textRequest.Model, "-thinking")
		relayInfo.UpstreamModelName = textRequest.Model
	}

	convertedRequest, err := adaptor.ConvertClaudeRequest(c, relayInfo, textRequest)
	if err != nil {
		return service.ClaudeErrorWrapperLocal(err, "convert_request_failed", http.StatusInternalServerError)
	}
	jsonData, err := json.Marshal(convertedRequest)
	if common.DebugEnabled {
		println("requestBody: ", string(jsonData))
	}
	if err != nil {
		return service.ClaudeErrorWrapperLocal(err, "json_marshal_failed", http.StatusInternalServerError)
	}
	requestBody = bytes.NewBuffer(jsonData)

	statusCodeMappingStr := c.GetString("status_code_mapping")
	var httpResp *http.Response
	resp, err := adaptor.DoRequest(c, relayInfo, requestBody)
	if err != nil {
		return service.ClaudeErrorWrapperLocal(err, "do_request_failed", http.StatusInternalServerError)
	}

	if resp != nil {
		httpResp = resp.(*http.Response)
		relayInfo.IsStream = relayInfo.IsStream || strings.HasPrefix(httpResp.Header.Get("Content-Type"), "text/event-stream")
		if httpResp.StatusCode != http.StatusOK {
			openaiErr = service.RelayErrorHandler(httpResp, false)
			// reset status code 重置状态码
			service.ResetStatusCode(openaiErr, statusCodeMappingStr)
			return service.OpenAIErrorToClaudeError(openaiErr)
		}
	}

	usage, openaiErr := adaptor.DoResponse(c, httpResp, relayInfo)
	//log.Printf("usage: %v", usage)
	if openaiErr != nil {
		// reset status code 重置状态码
		service.ResetStatusCode(openaiErr, statusCodeMappingStr)
		return service.OpenAIErrorToClaudeError(openaiErr)
	}
	service.PostClaudeConsumeQuota(c, relayInfo, usage.(*dto.Usage), preConsumedQuota, userQuota, priceData, "")
	return nil
}

func getClaudePromptTokens(textRequest *dto.ClaudeRequest, info *relaycommon.RelayInfo) (int, error) {
	var promptTokens int
	var err error
	switch info.RelayMode {
	default:
		promptTokens, err = service.CountTokenClaudeRequest(*textRequest, info.UpstreamModelName)
	}
	info.PromptTokens = promptTokens
	return promptTokens, err
}
