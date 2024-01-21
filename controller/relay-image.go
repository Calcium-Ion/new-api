package controller

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
	"one-api/model"
	"strings"
	"time"
)

func relayImageHelper(c *gin.Context, relayMode int) *OpenAIErrorWithStatusCode {
	tokenId := c.GetInt("token_id")
	channelType := c.GetInt("channel")
	channelId := c.GetInt("channel_id")
	userId := c.GetInt("id")
	consumeQuota := c.GetBool("consume_quota")
	group := c.GetString("group")
	startTime := time.Now()

	var imageRequest ImageRequest
	if consumeQuota {
		err := common.UnmarshalBodyReusable(c, &imageRequest)
		if err != nil {
			return errorWrapper(err, "bind_request_body_failed", http.StatusBadRequest)
		}
	}

	if imageRequest.Model == "" {
		imageRequest.Model = "dall-e-2"
	}
	if imageRequest.Size == "" {
		imageRequest.Size = "1024x1024"
	}
	if imageRequest.N == 0 {
		imageRequest.N = 1
	}
	// Prompt validation
	if imageRequest.Prompt == "" {
		return errorWrapper(errors.New("prompt is required"), "required_field_missing", http.StatusBadRequest)
	}

	if strings.Contains(imageRequest.Size, "×") {
		return errorWrapper(errors.New("size an unexpected error occurred in the parameter, please use 'x' instead of the multiplication sign '×'"), "invalid_field_value", http.StatusBadRequest)
	}
	// Not "256x256", "512x512", or "1024x1024"
	if imageRequest.Model == "dall-e-2" || imageRequest.Model == "dall-e" {
		if imageRequest.Size != "" && imageRequest.Size != "256x256" && imageRequest.Size != "512x512" && imageRequest.Size != "1024x1024" {
			return errorWrapper(errors.New("size must be one of 256x256, 512x512, or 1024x1024, dall-e-3 1024x1792 or 1792x1024"), "invalid_field_value", http.StatusBadRequest)
		}
	} else if imageRequest.Model == "dall-e-3" {
		if imageRequest.Size != "" && imageRequest.Size != "1024x1024" && imageRequest.Size != "1024x1792" && imageRequest.Size != "1792x1024" {
			return errorWrapper(errors.New("size must be one of 256x256, 512x512, or 1024x1024, dall-e-3 1024x1792 or 1792x1024"), "invalid_field_value", http.StatusBadRequest)
		}
		if imageRequest.N != 1 {
			return errorWrapper(errors.New("n must be 1"), "invalid_field_value", http.StatusBadRequest)
		}
	}

	// N should between 1 and 10
	if imageRequest.N != 0 && (imageRequest.N < 1 || imageRequest.N > 10) {
		return errorWrapper(errors.New("n must be between 1 and 10"), "invalid_field_value", http.StatusBadRequest)
	}

	// map model name
	modelMapping := c.GetString("model_mapping")
	isModelMapped := false
	if modelMapping != "" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return errorWrapper(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
		}
		if modelMap[imageRequest.Model] != "" {
			imageRequest.Model = modelMap[imageRequest.Model]
			isModelMapped = true
		}
	}
	baseURL := common.ChannelBaseURLs[channelType]
	requestURL := c.Request.URL.String()
	if c.GetString("base_url") != "" {
		baseURL = c.GetString("base_url")
	}
	fullRequestURL := getFullRequestURL(baseURL, requestURL, channelType)
	if channelType == common.ChannelTypeAzure && relayMode == RelayModeImagesGenerations {
		// https://learn.microsoft.com/en-us/azure/ai-services/openai/dall-e-quickstart?tabs=dalle3%2Ccommand-line&pivots=rest-api
		apiVersion := GetAPIVersion(c)
		// https://{resource_name}.openai.azure.com/openai/deployments/dall-e-3/images/generations?api-version=2023-06-01-preview
		fullRequestURL = fmt.Sprintf("%s/openai/deployments/%s/images/generations?api-version=%s", baseURL, imageRequest.Model, apiVersion)
	}
	var requestBody io.Reader
	if isModelMapped || channelType == common.ChannelTypeAzure { // make Azure channel request body
		jsonStr, err := json.Marshal(imageRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	} else {
		requestBody = c.Request.Body
	}

	modelRatio := common.GetModelRatio(imageRequest.Model)
	groupRatio := common.GetGroupRatio(group)
	ratio := modelRatio * groupRatio
	userQuota, err := model.CacheGetUserQuota(userId)

	sizeRatio := 1.0
	// Size
	if imageRequest.Size == "256x256" {
		sizeRatio = 1
	} else if imageRequest.Size == "512x512" {
		sizeRatio = 1.125
	} else if imageRequest.Size == "1024x1024" {
		sizeRatio = 1.25
	} else if imageRequest.Size == "1024x1792" || imageRequest.Size == "1792x1024" {
		sizeRatio = 2.5
	}

	qualityRatio := 1.0
	if imageRequest.Model == "dall-e-3" && imageRequest.Quality == "hd" {
		qualityRatio = 2.0
		if imageRequest.Size == "1024×1792" || imageRequest.Size == "1792×1024" {
			qualityRatio = 1.5
		}
	}

	quota := int(ratio*sizeRatio*qualityRatio*1000) * imageRequest.N

	if consumeQuota && userQuota-quota < 0 {
		return errorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}

	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return errorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	token := c.Request.Header.Get("Authorization")
	if channelType == common.ChannelTypeAzure { // Azure authentication
		token = strings.TrimPrefix(token, "Bearer ")
		req.Header.Set("api-key", token)
	} else {
		req.Header.Set("Authorization", token)
	}
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))

	resp, err := httpClient.Do(req)
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

	if resp.StatusCode != http.StatusOK {
		return relayErrorHandler(resp)
	}

	var textResponse ImageResponse
	defer func(ctx context.Context) {
		useTimeSeconds := time.Now().Unix() - startTime.Unix()
		if consumeQuota {
			if resp.StatusCode != http.StatusOK {
				return
			}
			err := model.PostConsumeTokenQuota(tokenId, userQuota, quota, 0, true)
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
				model.RecordConsumeLog(ctx, userId, channelId, 0, 0, imageRequest.Model, tokenName, quota, logContent, tokenId, userQuota, int(useTimeSeconds), false)
				model.UpdateUserUsedQuotaAndRequestCount(userId, quota)
				channelId := c.GetInt("channel_id")
				model.UpdateChannelUsedQuota(channelId, quota)
			}
		}
	}(c.Request.Context())

	if consumeQuota {
		responseBody, err := io.ReadAll(resp.Body)

		if err != nil {
			return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		}
		err = resp.Body.Close()
		if err != nil {
			return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
		}
		err = json.Unmarshal(responseBody, &textResponse)
		if err != nil {
			return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
		}

		resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}

	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)

	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return errorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError)
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
	}
	return nil
}
