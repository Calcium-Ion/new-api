package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	APITypeOpenAI = iota
	APITypeClaude
	APITypePaLM
	APITypeBaidu
	APITypeZhipu
	APITypeAli
	APITypeXunfei
	APITypeAIProxyLibrary
	APITypeTencent
	APITypeGemini
)

var httpClient *http.Client
var impatientHTTPClient *http.Client

func init() {
	if common.RelayTimeout == 0 {
		httpClient = &http.Client{}
	} else {
		httpClient = &http.Client{
			Timeout: time.Duration(common.RelayTimeout) * time.Second,
		}
	}

	impatientHTTPClient = &http.Client{
		Timeout: 5 * time.Second,
	}
}

func relayTextHelper(c *gin.Context, relayMode int) *OpenAIErrorWithStatusCode {
	channelType := c.GetInt("channel")
	channelId := c.GetInt("channel_id")
	tokenId := c.GetInt("token_id")
	userId := c.GetInt("id")
	group := c.GetString("group")
	tokenUnlimited := c.GetBool("token_unlimited_quota")
	startTime := time.Now()
	var textRequest GeneralOpenAIRequest

	err := common.UnmarshalBodyReusable(c, &textRequest)
	if err != nil {
		return errorWrapper(err, "bind_request_body_failed", http.StatusBadRequest)
	}
	if relayMode == RelayModeModerations && textRequest.Model == "" {
		textRequest.Model = "text-moderation-latest"
	}
	if relayMode == RelayModeEmbeddings && textRequest.Model == "" {
		textRequest.Model = c.Param("model")
	}
	// request validation
	if textRequest.Model == "" {
		return errorWrapper(errors.New("model is required"), "required_field_missing", http.StatusBadRequest)
	}
	switch relayMode {
	case RelayModeCompletions:
		if textRequest.Prompt == "" {
			return errorWrapper(errors.New("field prompt is required"), "required_field_missing", http.StatusBadRequest)
		}
	case RelayModeChatCompletions:
		if textRequest.Messages == nil || len(textRequest.Messages) == 0 {
			return errorWrapper(errors.New("field messages is required"), "required_field_missing", http.StatusBadRequest)
		}
	case RelayModeEmbeddings:
	case RelayModeModerations:
		if textRequest.Input == "" {
			return errorWrapper(errors.New("field input is required"), "required_field_missing", http.StatusBadRequest)
		}
	case RelayModeEdits:
		if textRequest.Instruction == "" {
			return errorWrapper(errors.New("field instruction is required"), "required_field_missing", http.StatusBadRequest)
		}
	}
	// map model name
	modelMapping := c.GetString("model_mapping")
	isModelMapped := false
	if modelMapping != "" && modelMapping != "{}" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return errorWrapper(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
		}
		if modelMap[textRequest.Model] != "" {
			textRequest.Model = modelMap[textRequest.Model]
			isModelMapped = true
		}
	}
	apiType := APITypeOpenAI
	switch channelType {
	case common.ChannelTypeAnthropic:
		apiType = APITypeClaude
	case common.ChannelTypeBaidu:
		apiType = APITypeBaidu
	case common.ChannelTypePaLM:
		apiType = APITypePaLM
	case common.ChannelTypeZhipu:
		apiType = APITypeZhipu
	case common.ChannelTypeAli:
		apiType = APITypeAli
	case common.ChannelTypeXunfei:
		apiType = APITypeXunfei
	case common.ChannelTypeAIProxyLibrary:
		apiType = APITypeAIProxyLibrary
	case common.ChannelTypeTencent:
		apiType = APITypeTencent
	case common.ChannelTypeGemini:
		apiType = APITypeGemini
	}
	baseURL := common.ChannelBaseURLs[channelType]
	requestURL := c.Request.URL.String()
	if c.GetString("base_url") != "" {
		baseURL = c.GetString("base_url")
	}
	fullRequestURL := getFullRequestURL(baseURL, requestURL, channelType)
	switch apiType {
	case APITypeOpenAI:
		if channelType == common.ChannelTypeAzure {
			// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/chatgpt-quickstart?pivots=rest-api&tabs=command-line#rest-api
			query := c.Request.URL.Query()
			apiVersion := query.Get("api-version")
			if apiVersion == "" {
				apiVersion = c.GetString("api_version")
			}
			requestURL := strings.Split(requestURL, "?")[0]
			requestURL = fmt.Sprintf("%s?api-version=%s", requestURL, apiVersion)
			baseURL = c.GetString("base_url")
			task := strings.TrimPrefix(requestURL, "/v1/")
			model_ := textRequest.Model
			model_ = strings.Replace(model_, ".", "", -1)
			// https://github.com/songquanpeng/one-api/issues/67
			model_ = strings.TrimSuffix(model_, "-0301")
			model_ = strings.TrimSuffix(model_, "-0314")
			model_ = strings.TrimSuffix(model_, "-0613")
			fullRequestURL = fmt.Sprintf("%s/openai/deployments/%s/%s", baseURL, model_, task)
		}
	case APITypeClaude:
		fullRequestURL = "https://api.anthropic.com/v1/complete"
		if baseURL != "" {
			fullRequestURL = fmt.Sprintf("%s/v1/complete", baseURL)
		}
	case APITypeBaidu:
		switch textRequest.Model {
		case "ERNIE-Bot":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions"
		case "ERNIE-Bot-turbo":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/eb-instant"
		case "ERNIE-Bot-4":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions_pro"
		case "BLOOMZ-7B":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/bloomz_7b1"
		case "Embedding-V1":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/embeddings/embedding-v1"
		}
		apiKey := c.Request.Header.Get("Authorization")
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		var err error
		if apiKey, err = getBaiduAccessToken(apiKey); err != nil {
			return errorWrapper(err, "invalid_baidu_config", http.StatusInternalServerError)
		}
		fullRequestURL += "?access_token=" + apiKey
	case APITypePaLM:
		fullRequestURL = "https://generativelanguage.googleapis.com/v1beta2/models/chat-bison-001:generateMessage"
		if baseURL != "" {
			fullRequestURL = fmt.Sprintf("%s/v1beta2/models/chat-bison-001:generateMessage", baseURL)
		}
		apiKey := c.Request.Header.Get("Authorization")
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		fullRequestURL += "?key=" + apiKey
	case APITypeGemini:
		requestBaseURL := "https://generativelanguage.googleapis.com"
		if baseURL != "" {
			requestBaseURL = baseURL
		}
		version := "v1beta"
		if c.GetString("api_version") != "" {
			version = c.GetString("api_version")
		}
		action := "generateContent"
		if textRequest.Stream {
			action = "streamGenerateContent"
		}
		fullRequestURL = fmt.Sprintf("%s/%s/models/%s:%s", requestBaseURL, version, textRequest.Model, action)
		apiKey := c.Request.Header.Get("Authorization")
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		fullRequestURL += "?key=" + apiKey
		//log.Println(fullRequestURL)

	case APITypeZhipu:
		method := "invoke"
		if textRequest.Stream {
			method = "sse-invoke"
		}
		fullRequestURL = fmt.Sprintf("https://open.bigmodel.cn/api/paas/v3/model-api/%s/%s", textRequest.Model, method)
	case APITypeAli:
		fullRequestURL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
		if relayMode == RelayModeEmbeddings {
			fullRequestURL = "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding"
		}
	case APITypeTencent:
		fullRequestURL = "https://hunyuan.cloud.tencent.com/hyllm/v1/chat/completions"
	case APITypeAIProxyLibrary:
		fullRequestURL = fmt.Sprintf("%s/api/library/ask", baseURL)
	}
	var promptTokens int
	var completionTokens int
	switch relayMode {
	case RelayModeChatCompletions:
		promptTokens, err = countTokenMessages(textRequest.Messages, textRequest.Model)
		if err != nil {
			return errorWrapper(err, "count_token_messages_failed", http.StatusInternalServerError)
		}
	case RelayModeCompletions:
		promptTokens = countTokenInput(textRequest.Prompt, textRequest.Model)
	case RelayModeModerations:
		promptTokens = countTokenInput(textRequest.Input, textRequest.Model)
	}
	modelPrice := common.GetModelPrice(textRequest.Model)
	groupRatio := common.GetGroupRatio(group)

	var preConsumedQuota int
	var ratio float64
	var modelRatio float64
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

	userQuota, err := model.CacheGetUserQuota(userId)
	if err != nil {
		return errorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota < 0 || userQuota-preConsumedQuota < 0 {
		return errorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}
	err = model.CacheDecreaseUserQuota(userId, preConsumedQuota)
	if err != nil {
		return errorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota > 100*preConsumedQuota {
		// 用户额度充足，判断令牌额度是否充足
		if !tokenUnlimited {
			// 非无限令牌，判断令牌额度是否充足
			tokenQuota := c.GetInt("token_quota")
			if tokenQuota > 100*preConsumedQuota {
				// 令牌额度充足，信任令牌
				preConsumedQuota = 0
				common.LogInfo(c.Request.Context(), fmt.Sprintf("user %d quota %d and token %d quota %d are enough, trusted and no need to pre-consume", userId, userQuota, tokenId, tokenQuota))
			}
		} else {
			// in this case, we do not pre-consume quota
			// because the user has enough quota
			preConsumedQuota = 0
			common.LogInfo(c.Request.Context(), fmt.Sprintf("user %d with unlimited token has enough quota %d, trusted and no need to pre-consume", userId, userQuota))
		}
	}
	if preConsumedQuota > 0 {
		userQuota, err = model.PreConsumeTokenQuota(tokenId, preConsumedQuota)
		if err != nil {
			return errorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}
	var requestBody io.Reader
	if isModelMapped {
		jsonStr, err := json.Marshal(textRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	} else {
		requestBody = c.Request.Body
	}
	switch apiType {
	case APITypeClaude:
		claudeRequest := requestOpenAI2Claude(textRequest)
		jsonStr, err := json.Marshal(claudeRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeBaidu:
		var jsonData []byte
		var err error
		switch relayMode {
		case RelayModeEmbeddings:
			baiduEmbeddingRequest := embeddingRequestOpenAI2Baidu(textRequest)
			jsonData, err = json.Marshal(baiduEmbeddingRequest)
		default:
			baiduRequest := requestOpenAI2Baidu(textRequest)
			jsonData, err = json.Marshal(baiduRequest)
		}
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonData)
	case APITypePaLM:
		palmRequest := requestOpenAI2PaLM(textRequest)
		jsonStr, err := json.Marshal(palmRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeGemini:
		geminiChatRequest := requestOpenAI2Gemini(textRequest)
		jsonStr, err := json.Marshal(geminiChatRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeZhipu:
		zhipuRequest := requestOpenAI2Zhipu(textRequest)
		jsonStr, err := json.Marshal(zhipuRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeAli:
		var jsonStr []byte
		var err error
		switch relayMode {
		case RelayModeEmbeddings:
			aliEmbeddingRequest := embeddingRequestOpenAI2Ali(textRequest)
			jsonStr, err = json.Marshal(aliEmbeddingRequest)
		default:
			aliRequest := requestOpenAI2Ali(textRequest)
			jsonStr, err = json.Marshal(aliRequest)
		}
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeTencent:
		apiKey := c.Request.Header.Get("Authorization")
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		appId, secretId, secretKey, err := parseTencentConfig(apiKey)
		if err != nil {
			return errorWrapper(err, "invalid_tencent_config", http.StatusInternalServerError)
		}
		tencentRequest := requestOpenAI2Tencent(textRequest)
		tencentRequest.AppId = appId
		tencentRequest.SecretId = secretId
		jsonStr, err := json.Marshal(tencentRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		sign := getTencentSign(*tencentRequest, secretKey)
		c.Request.Header.Set("Authorization", sign)
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeAIProxyLibrary:
		aiProxyLibraryRequest := requestOpenAI2AIProxyLibrary(textRequest)
		aiProxyLibraryRequest.LibraryId = c.GetString("library_id")
		jsonStr, err := json.Marshal(aiProxyLibraryRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	}

	var req *http.Request
	var resp *http.Response
	isStream := textRequest.Stream

	if apiType != APITypeXunfei { // cause xunfei use websocket
		req, err = http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
		// 设置GetBody函数，该函数返回一个新的io.ReadCloser，该io.ReadCloser返回与原始请求体相同的数据
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(requestBody), nil
		}
		if err != nil {
			return errorWrapper(err, "new_request_failed", http.StatusInternalServerError)
		}
		apiKey := c.Request.Header.Get("Authorization")
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		switch apiType {
		case APITypeOpenAI:
			if channelType == common.ChannelTypeAzure {
				req.Header.Set("api-key", apiKey)
			} else {
				req.Header.Set("Authorization", c.Request.Header.Get("Authorization"))
				if c.Request.Header.Get("OpenAI-Organization") != "" {
					req.Header.Set("OpenAI-Organization", c.Request.Header.Get("OpenAI-Organization"))
				}
				if channelType == common.ChannelTypeOpenRouter {
					req.Header.Set("HTTP-Referer", "https://github.com/songquanpeng/one-api")
					req.Header.Set("X-Title", "One API")
				}
			}
		case APITypeClaude:
			req.Header.Set("x-api-key", apiKey)
			anthropicVersion := c.Request.Header.Get("anthropic-version")
			if anthropicVersion == "" {
				anthropicVersion = "2023-06-01"
			}
			req.Header.Set("anthropic-version", anthropicVersion)
		case APITypeZhipu:
			token := getZhipuToken(apiKey)
			req.Header.Set("Authorization", token)
		case APITypeAli:
			req.Header.Set("Authorization", "Bearer "+apiKey)
			if textRequest.Stream {
				req.Header.Set("X-DashScope-SSE", "enable")
			}
		case APITypeTencent:
			req.Header.Set("Authorization", apiKey)
		case APITypeGemini:
			req.Header.Set("Content-Type", "application/json")
		default:
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
		if apiType != APITypeGemini {
			// 设置公共头部...
			req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
			req.Header.Set("Accept", c.Request.Header.Get("Accept"))
			if isStream && c.Request.Header.Get("Accept") == "" {
				req.Header.Set("Accept", "text/event-stream")
			}
		}
		//req.HeaderBar.Set("Connection", c.Request.HeaderBar.Get("Connection"))
		resp, err = httpClient.Do(req)
		if err != nil {
			return errorWrapper(err, "do_request_failed", http.StatusInternalServerError)
		}
		err = req.Body.Close()
		if err != nil {
			return errorWrapper(err, "close_request_body_failed", http.StatusInternalServerError)
		}
		err = c.Request.Body.Close()
		if err != nil {
			return errorWrapper(err, "close_request_body_failed", http.StatusInternalServerError)
		}
		isStream = isStream || strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream")

		if resp.StatusCode != http.StatusOK {
			if preConsumedQuota != 0 {
				go func(ctx context.Context) {
					// return pre-consumed quota
					err := model.PostConsumeTokenQuota(tokenId, userQuota, -preConsumedQuota, 0, false)
					if err != nil {
						common.LogError(ctx, "error return pre-consumed quota: "+err.Error())
					}
				}(c.Request.Context())
			}
			return relayErrorHandler(resp)
		}
	}

	var textResponse TextResponse
	tokenName := c.GetString("token_name")

	defer func(ctx context.Context) {
		// c.Writer.Flush()
		go func() {
			useTimeSeconds := time.Now().Unix() - startTime.Unix()
			promptTokens = textResponse.Usage.PromptTokens
			completionTokens = textResponse.Usage.CompletionTokens

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
				logContent += fmt.Sprintf("（有疑问请联系管理员）")
				common.LogError(ctx, fmt.Sprintf("total tokens is 0, cannot consume quota, userId %d, channelId %d, tokenId %d, model %s， pre-consumed quota %d", userId, channelId, tokenId, textRequest.Model, preConsumedQuota))
			} else {
				quotaDelta := quota - preConsumedQuota
				err := model.PostConsumeTokenQuota(tokenId, userQuota, quotaDelta, preConsumedQuota, true)
				if err != nil {
					common.LogError(ctx, "error consuming token remain quota: "+err.Error())
				}
				err = model.CacheUpdateUserQuota(userId)
				if err != nil {
					common.LogError(ctx, "error update user quota cache: "+err.Error())
				}
				model.UpdateUserUsedQuotaAndRequestCount(userId, quota)
				model.UpdateChannelUsedQuota(channelId, quota)
			}

			logModel := textRequest.Model
			if strings.HasPrefix(logModel, "gpt-4-gizmo") {
				logModel = "gpt-4-gizmo-*"
				logContent += fmt.Sprintf("，模型 %s", textRequest.Model)
			}
			model.RecordConsumeLog(ctx, userId, channelId, promptTokens, completionTokens, logModel, tokenName, quota, logContent, tokenId, userQuota, int(useTimeSeconds), isStream)

			//if quota != 0 {
			//
			//}
		}()
	}(c.Request.Context())
	switch apiType {
	case APITypeOpenAI:
		if isStream {
			err, responseText := openaiStreamHandler(c, resp, relayMode)
			if err != nil {
				return err
			}
			textResponse.Usage.PromptTokens = promptTokens
			textResponse.Usage.CompletionTokens = countTokenText(responseText, textRequest.Model)
			return nil
		} else {
			err, usage := openaiHandler(c, resp, promptTokens, textRequest.Model)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypeClaude:
		if isStream {
			err, responseText := claudeStreamHandler(c, resp)
			if err != nil {
				return err
			}
			textResponse.Usage.PromptTokens = promptTokens
			textResponse.Usage.CompletionTokens = countTokenText(responseText, textRequest.Model)
			return nil
		} else {
			err, usage := claudeHandler(c, resp, promptTokens, textRequest.Model)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypeBaidu:
		if isStream {
			err, usage := baiduStreamHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		} else {
			var err *OpenAIErrorWithStatusCode
			var usage *Usage
			switch relayMode {
			case RelayModeEmbeddings:
				err, usage = baiduEmbeddingHandler(c, resp)
			default:
				err, usage = baiduHandler(c, resp)
			}
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypePaLM:
		if textRequest.Stream { // PaLM2 API does not support stream
			err, responseText := palmStreamHandler(c, resp)
			if err != nil {
				return err
			}
			textResponse.Usage.PromptTokens = promptTokens
			textResponse.Usage.CompletionTokens = countTokenText(responseText, textRequest.Model)
			return nil
		} else {
			err, usage := palmHandler(c, resp, promptTokens, textRequest.Model)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypeGemini:
		if textRequest.Stream {
			err, responseText := geminiChatStreamHandler(c, resp)
			if err != nil {
				return err
			}
			textResponse.Usage.PromptTokens = promptTokens
			textResponse.Usage.CompletionTokens = countTokenText(responseText, textRequest.Model)
			return nil
		} else {
			err, usage := geminiChatHandler(c, resp, promptTokens, textRequest.Model)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypeZhipu:
		if isStream {
			err, usage := zhipuStreamHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			// zhipu's API does not return prompt tokens & completion tokens
			textResponse.Usage.PromptTokens = textResponse.Usage.TotalTokens
			return nil
		} else {
			err, usage := zhipuHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			// zhipu's API does not return prompt tokens & completion tokens
			textResponse.Usage.PromptTokens = textResponse.Usage.TotalTokens
			return nil
		}
	case APITypeAli:
		if isStream {
			err, usage := aliStreamHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		} else {
			var err *OpenAIErrorWithStatusCode
			var usage *Usage
			switch relayMode {
			case RelayModeEmbeddings:
				err, usage = aliEmbeddingHandler(c, resp)
			default:
				err, usage = aliHandler(c, resp)
			}
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypeXunfei:
		auth := c.Request.Header.Get("Authorization")
		auth = strings.TrimPrefix(auth, "Bearer ")
		splits := strings.Split(auth, "|")
		if len(splits) != 3 {
			return errorWrapper(errors.New("invalid auth"), "invalid_auth", http.StatusBadRequest)
		}
		var err *OpenAIErrorWithStatusCode
		var usage *Usage
		if isStream {
			err, usage = xunfeiStreamHandler(c, textRequest, splits[0], splits[1], splits[2])
		} else {
			err, usage = xunfeiHandler(c, textRequest, splits[0], splits[1], splits[2])
		}
		if err != nil {
			return err
		}
		if usage != nil {
			textResponse.Usage = *usage
		}
		return nil
	case APITypeAIProxyLibrary:
		if isStream {
			err, usage := aiProxyLibraryStreamHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		} else {
			err, usage := aiProxyLibraryHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypeTencent:
		if isStream {
			err, responseText := tencentStreamHandler(c, resp)
			if err != nil {
				return err
			}
			textResponse.Usage.PromptTokens = promptTokens
			textResponse.Usage.CompletionTokens = countTokenText(responseText, textRequest.Model)
			return nil
		} else {
			err, usage := tencentHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	default:
		return errorWrapper(errors.New("unknown api type"), "unknown_api_type", http.StatusInternalServerError)
	}
}
