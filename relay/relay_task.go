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
	relayconstant "one-api/relay/constant"
	"one-api/service"
	"one-api/setting"
	"one-api/setting/operation_setting"
)

/*
Task 任务通过平台、Action 区分任务
*/
func RelayTaskSubmit(c *gin.Context, relayMode int) (taskErr *dto.TaskError) {
	platform := constant.TaskPlatform(c.GetString("platform"))
	relayInfo := relaycommon.GenTaskRelayInfo(c)

	adaptor := GetTaskAdaptor(platform)
	if adaptor == nil {
		return service.TaskErrorWrapperLocal(fmt.Errorf("invalid api platform: %s", platform), "invalid_api_platform", http.StatusBadRequest)
	}
	adaptor.Init(relayInfo)
	// get & validate taskRequest 获取并验证文本请求
	taskErr = adaptor.ValidateRequestAndSetAction(c, relayInfo)
	if taskErr != nil {
		return
	}

	modelName := service.CoverTaskActionToModelName(platform, relayInfo.Action)
	modelPrice, success := operation_setting.GetModelPrice(modelName, true)
	if !success {
		defaultPrice, ok := operation_setting.GetDefaultModelRatioMap()[modelName]
		if !ok {
			modelPrice = 0.1
		} else {
			modelPrice = defaultPrice
		}
	}

	// 预扣
	groupRatio := setting.GetGroupRatio(relayInfo.Group)
	ratio := modelPrice * groupRatio
	userQuota, err := model.GetUserQuota(relayInfo.UserId, false)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
		return
	}
	quota := int(ratio * common.QuotaPerUnit)
	if userQuota-quota < 0 {
		taskErr = service.TaskErrorWrapperLocal(errors.New("user quota is not enough"), "quota_not_enough", http.StatusForbidden)
		return
	}

	if relayInfo.OriginTaskID != "" {
		originTask, exist, err := model.GetByTaskId(relayInfo.UserId, relayInfo.OriginTaskID)
		if err != nil {
			taskErr = service.TaskErrorWrapper(err, "get_origin_task_failed", http.StatusInternalServerError)
			return
		}
		if !exist {
			taskErr = service.TaskErrorWrapperLocal(errors.New("task_origin_not_exist"), "task_not_exist", http.StatusBadRequest)
			return
		}
		if originTask.ChannelId != relayInfo.ChannelId {
			channel, err := model.GetChannelById(originTask.ChannelId, true)
			if err != nil {
				taskErr = service.TaskErrorWrapperLocal(err, "channel_not_found", http.StatusBadRequest)
				return
			}
			if channel.Status != common.ChannelStatusEnabled {
				return service.TaskErrorWrapperLocal(errors.New("该任务所属渠道已被禁用"), "task_channel_disable", http.StatusBadRequest)
			}
			c.Set("base_url", channel.GetBaseURL())
			c.Set("channel_id", originTask.ChannelId)
			c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))

			relayInfo.BaseUrl = channel.GetBaseURL()
			relayInfo.ChannelId = originTask.ChannelId
		}
	}

	// build body
	requestBody, err := adaptor.BuildRequestBody(c, relayInfo)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "build_request_failed", http.StatusInternalServerError)
		return
	}
	// do request
	resp, err := adaptor.DoRequest(c, relayInfo, requestBody)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
		return
	}
	// handle response
	if resp != nil && resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		taskErr = service.TaskErrorWrapper(fmt.Errorf(string(responseBody)), "fail_to_fetch_task", resp.StatusCode)
		return
	}

	defer func() {
		// release quota
		if relayInfo.ConsumeQuota && taskErr == nil {

			err := service.PostConsumeQuota(relayInfo.RelayInfo, quota, 0, true)
			if err != nil {
				common.SysError("error consuming token remain quota: " + err.Error())
			}
			if quota != 0 {
				tokenName := c.GetString("token_name")
				logContent := fmt.Sprintf("模型固定价格 %.2f，分组倍率 %.2f，操作 %s", modelPrice, groupRatio, relayInfo.Action)
				other := make(map[string]interface{})
				other["model_price"] = modelPrice
				other["group_ratio"] = groupRatio
				model.RecordConsumeLog(c, relayInfo.UserId, relayInfo.ChannelId, 0, 0,
					modelName, tokenName, quota, logContent, relayInfo.TokenId, userQuota, 0, false, relayInfo.Group, other)
				model.UpdateUserUsedQuotaAndRequestCount(relayInfo.UserId, quota)
				model.UpdateChannelUsedQuota(relayInfo.ChannelId, quota)
			}
		}
	}()

	taskID, taskData, taskErr := adaptor.DoResponse(c, resp, relayInfo)
	if taskErr != nil {
		return
	}
	relayInfo.ConsumeQuota = true
	// insert task
	task := model.InitTask(constant.TaskPlatformSuno, relayInfo)
	task.TaskID = taskID
	task.Quota = quota
	task.Data = taskData
	err = task.Insert()
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "insert_task_failed", http.StatusInternalServerError)
		return
	}
	return nil
}

var fetchRespBuilders = map[int]func(c *gin.Context) (respBody []byte, taskResp *dto.TaskError){
	relayconstant.RelayModeSunoFetchByID: sunoFetchByIDRespBodyBuilder,
	relayconstant.RelayModeSunoFetch:     sunoFetchRespBodyBuilder,
}

func RelayTaskFetch(c *gin.Context, relayMode int) (taskResp *dto.TaskError) {
	respBuilder, ok := fetchRespBuilders[relayMode]
	if !ok {
		taskResp = service.TaskErrorWrapperLocal(errors.New("invalid_relay_mode"), "invalid_relay_mode", http.StatusBadRequest)
	}

	respBody, taskErr := respBuilder(c)
	if taskErr != nil {
		return taskErr
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	_, err := io.Copy(c.Writer, bytes.NewBuffer(respBody))
	if err != nil {
		taskResp = service.TaskErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError)
		return
	}
	return
}

func sunoFetchRespBodyBuilder(c *gin.Context) (respBody []byte, taskResp *dto.TaskError) {
	userId := c.GetInt("id")
	var condition = struct {
		IDs    []any  `json:"ids"`
		Action string `json:"action"`
	}{}
	err := c.BindJSON(&condition)
	if err != nil {
		taskResp = service.TaskErrorWrapper(err, "invalid_request", http.StatusBadRequest)
		return
	}
	var tasks []any
	if len(condition.IDs) > 0 {
		taskModels, err := model.GetByTaskIds(userId, condition.IDs)
		if err != nil {
			taskResp = service.TaskErrorWrapper(err, "get_tasks_failed", http.StatusInternalServerError)
			return
		}
		for _, task := range taskModels {
			tasks = append(tasks, TaskModel2Dto(task))
		}
	} else {
		tasks = make([]any, 0)
	}
	respBody, err = json.Marshal(dto.TaskResponse[[]any]{
		Code: "success",
		Data: tasks,
	})
	return
}

func sunoFetchByIDRespBodyBuilder(c *gin.Context) (respBody []byte, taskResp *dto.TaskError) {
	taskId := c.Param("id")
	userId := c.GetInt("id")

	originTask, exist, err := model.GetByTaskId(userId, taskId)
	if err != nil {
		taskResp = service.TaskErrorWrapper(err, "get_task_failed", http.StatusInternalServerError)
		return
	}
	if !exist {
		taskResp = service.TaskErrorWrapperLocal(errors.New("task_not_exist"), "task_not_exist", http.StatusBadRequest)
		return
	}

	respBody, err = json.Marshal(dto.TaskResponse[any]{
		Code: "success",
		Data: TaskModel2Dto(originTask),
	})
	return
}

func TaskModel2Dto(task *model.Task) *dto.TaskDto {
	return &dto.TaskDto{
		TaskID:     task.TaskID,
		Action:     task.Action,
		Status:     string(task.Status),
		FailReason: task.FailReason,
		SubmitTime: task.SubmitTime,
		StartTime:  task.StartTime,
		FinishTime: task.FinishTime,
		Progress:   task.Progress,
		Data:       task.Data,
	}
}
