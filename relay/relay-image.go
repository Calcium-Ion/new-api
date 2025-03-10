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
	"one-api/metrics"
	"one-api/model"
	relaycommon "one-api/relay/common"
	"one-api/relay/helper"
	"one-api/service"
	"one-api/setting"
	"strconv"
	"strings"
	"time"
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
		words, err := service.CheckSensitiveInput(imageRequest.Prompt)
		if err != nil {
			common.LogWarn(c, fmt.Sprintf("user sensitive words detected: %s", strings.Join(words, ",")))
			return nil, err
		}
	}
	return imageRequest, nil
}

func ImageInfo(c *gin.Context) (*relaycommon.RelayInfo, *dto.ImageRequest, *dto.OpenAIErrorWithStatusCode) {
	relayInfo := relaycommon.GenRelayInfo(c)

	imageRequest, err := getAndValidImageRequest(c, relayInfo)
	if err != nil {
		common.LogError(c, fmt.Sprintf("getAndValidImageRequest failed: %s", err.Error()))
		return nil, nil, service.OpenAIErrorWrapper(err, "invalid_image_request", http.StatusBadRequest)
	}
	return relayInfo, imageRequest, nil
}

func ImageHelper(c *gin.Context, relayInfo *relaycommon.RelayInfo, imageRequest *dto.ImageRequest) (openaiErr *dto.OpenAIErrorWithStatusCode) {
	startTime := time.Now()
	var funcErr *dto.OpenAIErrorWithStatusCode
	metrics.IncrementRelayRequestTotalCounter(strconv.Itoa(relayInfo.ChannelId), imageRequest.Model, relayInfo.Group, 1)
	defer func() {
		if funcErr != nil {
			metrics.IncrementRelayRequestFailedCounter(strconv.Itoa(relayInfo.ChannelId), imageRequest.Model, relayInfo.Group, strconv.Itoa(funcErr.StatusCode), 1)
		} else {
			metrics.IncrementRelayRequestSuccessCounter(strconv.Itoa(relayInfo.ChannelId), imageRequest.Model, relayInfo.Group, 1)
			metrics.ObserveRelayRequestDuration(strconv.Itoa(relayInfo.ChannelId), imageRequest.Model, relayInfo.Group, time.Since(startTime).Seconds())
		}
	}()
	err := helper.ModelMappedHelper(c, relayInfo)
	if err != nil {
		funcErr = service.OpenAIErrorWrapperLocal(err, "model_mapped_error", http.StatusInternalServerError)
		return funcErr
	}

	imageRequest.Model = relayInfo.UpstreamModelName

	priceData, err := helper.ModelPriceHelper(c, relayInfo, 0, 0)
	if err != nil {
		funcErr = service.OpenAIErrorWrapperLocal(err, "model_price_error", http.StatusInternalServerError)
		return funcErr
	}
	if !priceData.UsePrice {
		// modelRatio 16 = modelPrice $0.04
		// per 1 modelRatio = $0.04 / 16
		priceData.ModelPrice = 0.0025 * priceData.ModelRatio
	}

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

	priceData.ModelPrice *= sizeRatio * qualityRatio * float64(imageRequest.N)
	quota := int(priceData.ModelPrice * priceData.GroupRatio * common.QuotaPerUnit)

	if userQuota-quota < 0 {
		funcErr = service.OpenAIErrorWrapperLocal(fmt.Errorf("image pre-consumed quota failed, user quota: %s, need quota: %s", common.FormatQuota(userQuota), common.FormatQuota(quota)), "insufficient_user_quota", http.StatusForbidden)
		return funcErr
	}

	adaptor := GetAdaptor(relayInfo.ApiType)
	if adaptor == nil {
		funcErr = service.OpenAIErrorWrapperLocal(fmt.Errorf("invalid api type: %d", relayInfo.ApiType), "invalid_api_type", http.StatusBadRequest)
		return funcErr
	}
	adaptor.Init(relayInfo)

	var requestBody io.Reader

	convertedRequest, err := adaptor.ConvertImageRequest(c, relayInfo, *imageRequest)
	if err != nil {
		funcErr = service.OpenAIErrorWrapperLocal(err, "convert_request_failed", http.StatusInternalServerError)
		return funcErr
	}

	jsonData, err := json.Marshal(convertedRequest)
	if err != nil {
		funcErr = service.OpenAIErrorWrapperLocal(err, "json_marshal_failed", http.StatusInternalServerError)
		return funcErr
	}
	requestBody = bytes.NewBuffer(jsonData)

	statusCodeMappingStr := c.GetString("status_code_mapping")

	resp, err := adaptor.DoRequest(c, relayInfo, requestBody)
	if err != nil {
		funcErr = service.OpenAIErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
		return funcErr
	}
	var httpResp *http.Response
	if resp != nil {
		httpResp = resp.(*http.Response)
		relayInfo.IsStream = relayInfo.IsStream || strings.HasPrefix(httpResp.Header.Get("Content-Type"), "text/event-stream")
		if httpResp.StatusCode != http.StatusOK {
			openaiErr := service.RelayErrorHandler(httpResp)
			funcErr = openaiErr
			// reset status code 重置状态码
			service.ResetStatusCode(openaiErr, statusCodeMappingStr)
			return openaiErr
		}
	}

	_, openaiErr = adaptor.DoResponse(c, httpResp, relayInfo)
	if openaiErr != nil {
		funcErr = openaiErr
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
	postConsumeQuota(c, relayInfo, usage, 0, userQuota, priceData, logContent)
	return nil
}
