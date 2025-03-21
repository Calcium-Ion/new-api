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
	"one-api/constant"
	"one-api/dto"
	"one-api/model"
	relaycommon "one-api/relay/common"
	"one-api/relay/helper"
	"one-api/service"
	"strings"
	"time"
)

func getAndValidVideoRequest(c *gin.Context, info *relaycommon.RelayInfo) (*dto.VideoRequest, error) {
	VideoRequest := &dto.VideoRequest{}

	err := common.UnmarshalBodyReusable(c, VideoRequest)
	if err != nil {
		return nil, err
	}

	if VideoRequest.Prompt == "" {
		return nil, errors.New("prompt is required")
	}
	if VideoRequest.Model == "wanx2.1-i2v-turbo" || VideoRequest.Model == "wanx2.1-i2v-plus" { // tu
		if VideoRequest.ImgUrl == "" {
			return nil, errors.New("image_url is required")
		}

		if VideoRequest.Duration != 0 || VideoRequest.Duration < 3 && VideoRequest.Duration > 5 {
			return nil, errors.New("duration is error")
		}

	} else if VideoRequest.Model == "wanx2.1-t2v-turbo" || VideoRequest.Model == "wanx2.1-t2v-plus" { //wen
		switch VideoRequest.Size {
		case "":
		case "1280 * 720":
		case "960 * 960":
		case "720 * 1280":
		case "1088 * 832":
		case "832 * 1088":
		default:
			//异常
			return nil, errors.New("size is error")
		}
		if VideoRequest.Duration != 0 || VideoRequest.Duration != 5 {
			return nil, errors.New("duration is error")
		}
	}

	return VideoRequest, nil
}

func VideoHelper(c *gin.Context) *dto.OpenAIErrorWithStatusCode {
	relayInfo := relaycommon.GenRelayInfo(c)

	VideoRequest, err := getAndValidVideoRequest(c, relayInfo)
	if err != nil {
		common.LogError(c, fmt.Sprintf("getAndValidVideoRequest failed: %s", err.Error()))
		return service.OpenAIErrorWrapper(err, "invalid_Video_request", http.StatusBadRequest)
	}

	err = helper.ModelMappedHelper(c, relayInfo)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "model_mapped_error", http.StatusInternalServerError)
	}

	VideoRequest.Model = relayInfo.UpstreamModelName

	adaptor := GetAdaptor(relayInfo.ApiType)
	if adaptor == nil {
		return service.OpenAIErrorWrapperLocal(fmt.Errorf("invalid api type: %d", relayInfo.ApiType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(relayInfo)

	var requestBody io.Reader

	convertedRequest, err := adaptor.ConvertVideoRequest(c, relayInfo, *VideoRequest)
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
			openaiErr := service.RelayErrorHandler(httpResp, false)
			// reset status code 重置状态码
			service.ResetStatusCode(openaiErr, statusCodeMappingStr)
			return openaiErr
		}
	}

	data, openaiErr := adaptor.DoResponse(c, httpResp, relayInfo)
	if openaiErr != nil && data != nil {
		// reset status code 重置状态码
		service.ResetStatusCode(openaiErr, statusCodeMappingStr)
		return openaiErr
	}

	videoResponse := data.(*dto.VideoResponse)

	taskRelayInfo := relaycommon.GenTaskRelayInfo(c)
	taskRelayInfo.ConsumeQuota = true
	// insert task
	task := model.InitTask(constant.TaskPlatformAli, taskRelayInfo)
	task.TaskID = videoResponse.TaskID
	//task.Quota =
	var err1 error
	for i := 0; i < 5; i++ {
		if err1 = task.Insert(); err1 == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if err1 != nil {
		return service.OpenAIErrorWrapperLocal(err, "sql_insert_failed", http.StatusInternalServerError)
	}
	return nil
}
