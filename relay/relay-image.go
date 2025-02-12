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
	"one-api/model"
	relaycommon "one-api/relay/common"
	"one-api/service"
	"one-api/setting"
	"strings"
)

func getAndValidImageRequest(c *gin.Context, info *relaycommon.RelayInfo) (*dto.ImageRequest, error) {
	imageRequest := &dto.ImageRequest{}
	err := common.UnmarshalBodyReusable(c, imageRequest)
	if err != nil {
		return nil, err
	}
	if imageRequest.Prompt == "" {
		return nil, errors.New("prompt is required")
	}
	if strings.Contains(imageRequest.Size, "×") {
		return nil, errors.New("size an unexpected error occurred in the parameter, please use 'x' instead of the multiplication sign '×'")
	}
	if imageRequest.N == 0 {
		imageRequest.N = 1
	}
	if imageRequest.Size == "" {
		imageRequest.Size = "1024x1024"
	}
	if imageRequest.Model == "" {
		imageRequest.Model = "dall-e-2"
	}

	// Not "256x256", "512x512", or "1024x1024"
	if imageRequest.Model == "dall-e-2" || imageRequest.Model == "dall-e" {
		if imageRequest.Size != "" && imageRequest.Size != "256x256" && imageRequest.Size != "512x512" && imageRequest.Size != "1024x1024" {
			return nil, errors.New("size must be one of 256x256, 512x512, or 1024x1024, dall-e-3 1024x1792 or 1792x1024")
		}
	} else if imageRequest.Model == "dall-e-3" {
		if imageRequest.Size != "" && imageRequest.Size != "1024x1024" && imageRequest.Size != "1024x1792" && imageRequest.Size != "1792x1024" {
			return nil, errors.New("size must be one of 256x256, 512x512, or 1024x1024, dall-e-3 1024x1792 or 1792x1024")
		}
		if imageRequest.Quality == "" {
			imageRequest.Quality = "standard"
		}
		//if imageRequest.N != 1 {
		//	return nil, errors.New("n must be 1")
		//}
	}
	// N should between 1 and 10
	//if imageRequest.N != 0 && (imageRequest.N < 1 || imageRequest.N > 10) {
	//	return service.OpenAIErrorWrapper(errors.New("n must be between 1 and 10"), "invalid_field_value", http.StatusBadRequest)
	//}
	if setting.ShouldCheckPromptSensitive() {
		err := service.CheckSensitiveInput(imageRequest.Prompt)
		if err != nil {
			return nil, err
		}
	}
	return imageRequest, nil
}

func ImageHelper(c *gin.Context, relayMode int) *dto.OpenAIErrorWithStatusCode {
	relayInfo := relaycommon.GenRelayInfo(c)

	imageRequest, err := getAndValidImageRequest(c, relayInfo)
	if err != nil {
		common.LogError(c, fmt.Sprintf("getAndValidImageRequest failed: %s", err.Error()))
		return service.OpenAIErrorWrapper(err, "invalid_image_request", http.StatusBadRequest)
	}

	// map model name
	modelMapping := c.GetString("model_mapping")
	if modelMapping != "" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
		}
		if modelMap[imageRequest.Model] != "" {
			imageRequest.Model = modelMap[imageRequest.Model]
		}
	}
	relayInfo.UpstreamModelName = imageRequest.Model

	modelPrice, success := common.GetModelPrice(imageRequest.Model, true)
	if !success {
		modelRatio := common.GetModelRatio(imageRequest.Model)
		// modelRatio 16 = modelPrice $0.04
		// per 1 modelRatio = $0.04 / 16
		modelPrice = 0.0025 * modelRatio
	}

	groupRatio := setting.GetGroupRatio(relayInfo.Group)
	userQuota, err := model.GetUserQuota(relayInfo.UserId, false)

	sizeRatio := 1.0
	// Size
	if imageRequest.Size == "256x256" {
		sizeRatio = 0.4
	} else if imageRequest.Size == "512x512" {
		sizeRatio = 0.45
	} else if imageRequest.Size == "1024x1024" {
		sizeRatio = 1
	} else if imageRequest.Size == "1024x1792" || imageRequest.Size == "1792x1024" {
		sizeRatio = 2
	}

	qualityRatio := 1.0
	if imageRequest.Model == "dall-e-3" && imageRequest.Quality == "hd" {
		qualityRatio = 2.0
		if imageRequest.Size == "1024x1792" || imageRequest.Size == "1792x1024" {
			qualityRatio = 1.5
		}
	}

	imageRatio := modelPrice * sizeRatio * qualityRatio * float64(imageRequest.N)
	quota := int(imageRatio * groupRatio * common.QuotaPerUnit)

	if userQuota-quota < 0 {
		return service.OpenAIErrorWrapperLocal(errors.New(fmt.Sprintf("image pre-consumed quota failed, user quota: %d, need quota: %d", userQuota, quota)), "insufficient_user_quota", http.StatusBadRequest)
	}

	adaptor := GetAdaptor(relayInfo.ApiType)
	if adaptor == nil {
		return service.OpenAIErrorWrapperLocal(fmt.Errorf("invalid api type: %d", relayInfo.ApiType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(relayInfo)

	var requestBody io.Reader

	convertedRequest, err := adaptor.ConvertImageRequest(c, relayInfo, *imageRequest)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "convert_request_failed", http.StatusInternalServerError)
	}

	jsonData, err := json.Marshal(convertedRequest)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "json_marshal_failed", http.StatusInternalServerError)
	}
	requestBody = bytes.NewBuffer(jsonData)

	statusCodeMappingStr := c.GetString("status_code_mapping")

	resp, err := adaptor.DoRequest(c, relayInfo, requestBody)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}
	var httpResp *http.Response
	if resp != nil {
		httpResp = resp.(*http.Response)
		relayInfo.IsStream = relayInfo.IsStream || strings.HasPrefix(httpResp.Header.Get("Content-Type"), "text/event-stream")
		if httpResp.StatusCode != http.StatusOK {
			openaiErr := service.RelayErrorHandler(httpResp)
			// reset status code 重置状态码
			service.ResetStatusCode(openaiErr, statusCodeMappingStr)
			return openaiErr
		}
	}

	_, openaiErr := adaptor.DoResponse(c, httpResp, relayInfo)
	if openaiErr != nil {
		// reset status code 重置状态码
		service.ResetStatusCode(openaiErr, statusCodeMappingStr)
		return openaiErr
	}

	usage := &dto.Usage{
		PromptTokens: imageRequest.N,
		TotalTokens:  imageRequest.N,
	}

	quality := "standard"
	if imageRequest.Quality == "hd" {
		quality = "hd"
	}

	logContent := fmt.Sprintf("大小 %s, 品质 %s", imageRequest.Size, quality)
	postConsumeQuota(c, relayInfo, imageRequest.Model, usage, 0, 0, userQuota, 0, groupRatio, imageRatio, true, logContent)

	return nil
}
