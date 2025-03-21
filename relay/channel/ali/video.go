package ali

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/service"
)

func oaiVideo2Ali(request dto.VideoRequest) *AliVideoRequest {
	var videoRequest AliVideoRequest
	videoRequest.Model = request.Model
	videoRequest.Input.Prompt = request.Prompt
	videoRequest.Input.ImgUrl = request.ImgUrl
	videoRequest.Parameters.Duration = request.Duration
	videoRequest.Parameters.PromptExtend = request.PromptExtend
	videoRequest.Parameters.Seed = request.Seed
	videoRequest.Parameters.Size = request.Size

	return &videoRequest
}

func responseAli2OpenAIVideo(c *gin.Context, response *AliResponse, info *relaycommon.RelayInfo) *dto.VideoResponse {
	VideoResponse := dto.VideoResponse{
		TaskID: response.Output.TaskId,
	}
	return &VideoResponse
}

func aliVideoHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.VideoResponse) {
	var aliResponse AliResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &aliResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	if aliResponse.Message != "" {
		common.LogError(c, "ali_async_task_failed: "+aliResponse.Message)
		return service.OpenAIErrorWrapper(errors.New(aliResponse.Message), "ali_async_task_failed", http.StatusInternalServerError), nil
	}

	fullVideoResponse := responseAli2OpenAIVideo(c, &aliResponse, info)
	jsonResponse, err := json.Marshal(fullVideoResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, fullVideoResponse
}
