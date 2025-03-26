package ali

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/service"
)

type TaskAdaptor struct {
	ChannelType int
}

func (a *TaskAdaptor) Init(info *relaycommon.TaskRelayInfo) {
	a.ChannelType = info.ChannelType
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.TaskRelayInfo) (taskErr *dto.TaskError) {
	return nil
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.TaskRelayInfo) (string, error) {
	return "", nil
}

func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.TaskRelayInfo) error {
	return nil
}

func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.TaskRelayInfo) (io.Reader, error) {
	return nil, nil
}

func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.TaskRelayInfo, requestBody io.Reader) (*http.Response, error) {
	return nil, nil
}

func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.TaskRelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	return "", nil, nil
}

func (a *TaskAdaptor) GetModelList() []string {
	return ModelList
}

func (a *TaskAdaptor) GetChannelName() string {
	return ChannelName
}

func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any) (*http.Response, error) {
	return nil, nil
}

func (a *TaskAdaptor) SingleTask(baseUrl, key string, body map[string]any) (*http.Response, error) {
	requestUrl := fmt.Sprintf("%s/api/v1/tasks/%s", baseUrl, body["task_id"])
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		common.SysError(fmt.Sprintf("Get Task error: %v", err))
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := service.GETTransportHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
