package ali

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"io"
	"net/http"
	"one-api/common"
	"one-api/model"
	"one-api/relay/channel"
)

func HandleGetTask(baseUrl, key, taskId string, adaptor channel.TaskAdaptor) (aliResp *AliResponse, err error) {
	resp, err := adaptor.SingleTask(baseUrl, key, map[string]any{
		"task_id": taskId,
	})

	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get Task Do req error: %d", resp.StatusCode)
		return
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	err = json.Unmarshal(responseBody, &aliResp)
	if err != nil {
		return
	}

	return
}

func HandleUpdateTask(ctx context.Context, task *model.Task, aliResp *AliResponse) (err error) {
	newTask, err := convertTask(aliResp)
	if err != nil {
		return err
	}

	if !checkTaskNeedUpdate(task, newTask) {
		return
	}

	task.Status = lo.If(model.TaskStatus(newTask.Status) != "", newTask.Status).Else(task.Status)
	task.FailReason = lo.If(newTask.FailReason != "", newTask.FailReason).Else(task.FailReason)
	task.SubmitTime = lo.If(newTask.SubmitTime != 0, newTask.SubmitTime).Else(task.SubmitTime)
	task.StartTime = lo.If(newTask.StartTime != 0, newTask.StartTime).Else(task.StartTime)
	task.FinishTime = lo.If(newTask.FinishTime != 0, newTask.FinishTime).Else(task.FinishTime)

	if task.Status == model.TaskStatusFailure || task.Status == model.TaskStatusUnknown {
		common.LogInfo(ctx, task.TaskID+" 构建失败，"+task.FailReason)
		task.Progress = "100%"
		//err = model.CacheUpdateUserQuota(task.UserId) ?
		if err != nil {
			common.LogError(ctx, "error update user quota cache: "+err.Error())
		} else {
			//TODO The amount is not realized first

			//quota := task.Quota
			//if quota != 0 {
			//	err = model.IncreaseUserQuota(task.UserId, quota, false)
			//	if err != nil {
			//		common.LogError(ctx, "fail to increase user quota: "+err.Error())
			//	}
			//	logContent := fmt.Sprintf("异步任务执行失败 %s，补偿 %s", task.TaskID, common.LogQuota(quota))
			//	model.RecordLog(task.UserId, model.LogTypeSystem, logContent)
			//}
		}
	}

	if newTask.Status == model.TaskStatusSuccess {
		task.Progress = "100%"
	}

	task.SetData(aliResp)

	err = task.Update()
	if err != nil {
		common.SysError("Update Ali task error:  " + err.Error())
	}

	return
}
func checkTaskNeedUpdate(oldTask *model.Task, newTask *model.Task) bool {
	return oldTask.SubmitTime != newTask.SubmitTime || oldTask.StartTime != newTask.StartTime || oldTask.FinishTime != newTask.FinishTime ||
		oldTask.Status != newTask.Status || oldTask.FailReason != newTask.FailReason
}

func convertTask(aliResp *AliResponse) (task *model.Task, err error) {
	var taskStatus model.TaskStatus
	switch aliResp.Output.TaskStatus {
	case "PENDING":
		taskStatus = model.TaskStatusQueued
	case "RUNNING":
		taskStatus = model.TaskStatusInProgress
	case "SUSPENDED":
		taskStatus = model.TaskStatusNotStart
	case "SUCCEEDED":
		taskStatus = model.TaskStatusSuccess
	case "FAILED":
		taskStatus = model.TaskStatusFailure
	case "UNKNOWN":
		taskStatus = model.TaskStatusUnknown
	}

	submitTimeSec := int64(0)
	scheduledTimeSec := int64(0)
	endTimeSec := int64(0)
	if aliResp.Output.SubmitTime != "" {
		submitTime, _ := common.StrConvertTime0(aliResp.Output.SubmitTime)
		submitTimeSec = submitTime.Unix()
	}
	if aliResp.Output.ScheduledTime != "" {
		scheduledTime, _ := common.StrConvertTime0(aliResp.Output.ScheduledTime)
		scheduledTimeSec = scheduledTime.Unix()
	}
	if aliResp.Output.EndTime != "" {
		endTime, _ := common.StrConvertTime0(aliResp.Output.EndTime)
		endTimeSec = endTime.Unix()
	}

	failReasonStr := ""
	if (taskStatus == model.TaskStatusFailure || taskStatus == model.TaskStatusUnknown) && aliResp.Message != "" && aliResp.Code != "" {
		type failReason struct {
			Message string `json:"message,omitempty"`
			Code    string `json:"code,omitempty"`
		}

		reason := failReason{
			Message: aliResp.Message,
			Code:    aliResp.Code,
		}

		failReasonJson, err1 := json.Marshal(reason)
		if err1 != nil {
			return nil, err1
		}
		failReasonStr = string(failReasonJson)
	}

	task = &model.Task{
		Status:     taskStatus,
		FailReason: failReasonStr,
		SubmitTime: submitTimeSec,
		StartTime:  scheduledTimeSec,
		FinishTime: endTimeSec,
	}

	return
}
