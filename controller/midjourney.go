package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"
	"strings"
	"time"
)

/*func UpdateMidjourneyTask() {
	//revocer
	//imageModel := "midjourney"
	ctx := context.TODO()
	imageModel := "midjourney"
	defer func() {
		if err := recover(); err != nil {
			log.Printf("UpdateMidjourneyTask panic: %v", err)
		}
	}()
	for {
		time.Sleep(time.Duration(15) * time.Second)
		tasks := model.GetAllUnFinishTasks()
		if len(tasks) != 0 {
			common.LogInfo(ctx, fmt.Sprintf("检测到未完成的任务数有: %v", len(tasks)))
			for _, task := range tasks {
				common.LogInfo(ctx, fmt.Sprintf("未完成的任务信息: %v", task))
				midjourneyChannel, err := model.GetChannelById(task.ChannelId, true)
				if err != nil {
					common.LogError(ctx, fmt.Sprintf("UpdateMidjourneyTask: %v", err))
					task.FailReason = fmt.Sprintf("获取渠道信息失败，请联系管理员，渠道ID：%d", task.ChannelId)
					task.Status = "FAILURE"
					task.Progress = "100%"
					err := task.Update()
					if err != nil {
						common.LogInfo(ctx, fmt.Sprintf("UpdateMidjourneyTask error: %v", err))
						continue
					}
					continue
				}
				requestUrl := fmt.Sprintf("%s/mj/task/%s/fetch", *midjourneyChannel.BaseURL, task.MjId)
				common.LogInfo(ctx, fmt.Sprintf("requestUrl: %s", requestUrl))

				req, err := http.NewRequest("GET", requestUrl, bytes.NewBuffer([]byte("")))
				if err != nil {
					common.LogInfo(ctx, fmt.Sprintf("Get Task error: %v", err))
					continue
				}

				// 设置超时时间
				timeout := time.Second * 5
				ctx, cancel := context.WithTimeout(context.Background(), timeout)

				// 使用带有超时的 context 创建新的请求
				req = req.WithContext(ctx)

				req.Header.Set("Content-Type", "application/json")
				//req.Header.Set("Authorization", "Bearer midjourney-proxy")
				req.Header.Set("mj-api-secret", midjourneyChannel.Key)
				resp, err := httpClient.Do(req)
				if err != nil {
					log.Printf("UpdateMidjourneyTask error: %v", err)
					continue
				}
				responseBody, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				log.Printf("responseBody: %s", string(responseBody))
				var responseItem Midjourney
				// err = json.NewDecoder(resp.Body).Decode(&responseItem)
				err = json.Unmarshal(responseBody, &responseItem)
				if err != nil {
					if strings.Contains(err.Error(), "cannot unmarshal number into Go struct field Midjourney.status of type string") {
						var responseWithoutStatus MidjourneyWithoutStatus
						var responseStatus MidjourneyStatus
						err1 := json.Unmarshal(responseBody, &responseWithoutStatus)
						err2 := json.Unmarshal(responseBody, &responseStatus)
						if err1 == nil && err2 == nil {
							jsonData, err3 := json.Marshal(responseWithoutStatus)
							if err3 != nil {
								log.Printf("UpdateMidjourneyTask error1: %v", err3)
								continue
							}
							err4 := json.Unmarshal(jsonData, &responseStatus)
							if err4 != nil {
								log.Printf("UpdateMidjourneyTask error2: %v", err4)
								continue
							}
							responseItem.Status = strconv.Itoa(responseStatus.Status)
						} else {
							log.Printf("UpdateMidjourneyTask error3: %v", err)
							continue
						}
					} else {
						log.Printf("UpdateMidjourneyTask error4: %v", err)
						continue
					}
				}
				task.Code = 1
				task.Progress = responseItem.Progress
				task.PromptEn = responseItem.PromptEn
				task.State = responseItem.State
				task.SubmitTime = responseItem.SubmitTime
				task.StartTime = responseItem.StartTime
				task.FinishTime = responseItem.FinishTime
				task.ImageUrl = responseItem.ImageUrl
				task.Status = responseItem.Status
				task.FailReason = responseItem.FailReason
				if task.Progress != "100%" && responseItem.FailReason != "" {
					common.LogWarn(task.MjId + " 构建失败，" + task.FailReason)
					task.Progress = "100%"
					err = model.CacheUpdateUserQuota(task.UserId)
					if err != nil {
						log.Println("error update user quota cache: " + err.Error())
					} else {
						modelRatio := common.GetModelRatio(imageModel)
						groupRatio := common.GetGroupRatio("default")
						ratio := modelRatio * groupRatio
						quota := int(ratio * 1 * 1000)
						if quota != 0 {
							err := model.IncreaseUserQuota(task.UserId, quota)
							if err != nil {
								log.Println("fail to increase user quota")
							}
							logContent := fmt.Sprintf("构图失败 %s，补偿 %s", task.MjId, common.LogQuota(quota))
							model.RecordLog(task.UserId, model.LogTypeSystem, logContent)
						}
					}
				}

				err = task.Update()
				if err != nil {
					log.Printf("UpdateMidjourneyTask error5: %v", err)
				}
				log.Printf("UpdateMidjourneyTask success: %v", task)
				cancel()
			}
		}
	}
}
*/

func UpdateMidjourneyTaskBulk() {
	//revocer
	defer func() {
		if err := recover(); err != nil {
			log.Printf("UpdateMidjourneyTask panic: %v", err)
		}
	}()
	imageModel := "midjourney"
	ctx := context.TODO()
	for {
		time.Sleep(time.Duration(15) * time.Second)

		tasks := model.GetAllUnFinishTasks()
		if len(tasks) == 0 {
			continue
		}

		common.LogInfo(ctx, fmt.Sprintf("检测到未完成的任务数有: %v", len(tasks)))
		taskChannelM := make(map[int][]string)
		taskM := make(map[string]*model.Midjourney)
		for _, task := range tasks {
			if task.MjId == "" {
				continue
			}
			taskM[task.MjId] = task
			taskChannelM[task.ChannelId] = append(taskChannelM[task.ChannelId], task.MjId)
		}
		if len(taskChannelM) == 0 {
			continue
		}

		for channelId, taskIds := range taskChannelM {
			common.LogInfo(ctx, fmt.Sprintf("渠道 #%d 未完成的任务有: %d", channelId, len(taskIds)))
			if len(taskIds) == 0 {
				continue
			}
			midjourneyChannel, err := model.CacheGetChannel(channelId)
			if err != nil {
				common.LogError(ctx, fmt.Sprintf("CacheGetChannel: %v", err))
				err := model.MjBulkUpdate(taskIds, map[string]any{
					"fail_reason": fmt.Sprintf("获取渠道信息失败，请联系管理员，渠道ID：%d", channelId),
					"status":      "FAILURE",
					"progress":    "100%",
				})
				if err != nil {
					common.LogInfo(ctx, fmt.Sprintf("UpdateMidjourneyTask error: %v", err))
				}
				continue
			}
			requestUrl := fmt.Sprintf("%s/mj/task/list-by-condition", *midjourneyChannel.BaseURL)

			body, _ := json.Marshal(map[string]any{
				"ids": taskIds,
			})
			req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(body))
			if err != nil {
				common.LogError(ctx, fmt.Sprintf("Get Task error: %v", err))
				continue
			}
			// 设置超时时间
			timeout := time.Second * 5
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			// 使用带有超时的 context 创建新的请求
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("mj-api-secret", midjourneyChannel.Key)
			resp, err := httpClient.Do(req)
			if err != nil {
				common.LogError(ctx, fmt.Sprintf("Get Task Do req error: %v", err))
				continue
			}
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				common.LogError(ctx, fmt.Sprintf("Get Task parse body error: %v", err))
				continue
			}
			var responseItems []Midjourney
			err = json.Unmarshal(responseBody, &responseItems)
			if err != nil {
				common.LogError(ctx, fmt.Sprintf("Get Task parse body error2: %v", err))
				continue
			}
			resp.Body.Close()
			req.Body.Close()
			cancel()

			for _, responseItem := range responseItems {
				task := taskM[responseItem.MjId]
				if !checkMjTaskNeedUpdate(task, responseItem) {
					continue
				}

				task.Code = 1
				task.Progress = responseItem.Progress
				task.PromptEn = responseItem.PromptEn
				task.State = responseItem.State
				task.SubmitTime = responseItem.SubmitTime
				task.StartTime = responseItem.StartTime
				task.FinishTime = responseItem.FinishTime
				task.ImageUrl = responseItem.ImageUrl
				task.Status = responseItem.Status
				task.FailReason = responseItem.FailReason
				if task.Progress != "100%" && responseItem.FailReason != "" {
					common.LogInfo(ctx, task.MjId+" 构建失败，"+task.FailReason)
					task.Progress = "100%"
					err = model.CacheUpdateUserQuota(task.UserId)
					if err != nil {
						common.LogError(ctx, "error update user quota cache: "+err.Error())
					} else {
						modelRatio := common.GetModelRatio(imageModel)
						groupRatio := common.GetGroupRatio("default")
						ratio := modelRatio * groupRatio
						quota := int(ratio * 1 * 1000)
						if quota != 0 {
							err = model.IncreaseUserQuota(task.UserId, quota)
							if err != nil {
								common.LogError(ctx, "fail to increase user quota: "+err.Error())
							}
							logContent := fmt.Sprintf("构图失败 %s，补偿 %s", task.MjId, common.LogQuota(quota))
							model.RecordLog(task.UserId, model.LogTypeSystem, logContent)
						}
					}
				}
				err = task.Update()
				if err != nil {
					common.LogError(ctx, "UpdateMidjourneyTask task error: "+err.Error())
				}
			}
		}
	}
}

func checkMjTaskNeedUpdate(oldTask *model.Midjourney, newTask Midjourney) bool {
	if oldTask.Code != 1 {
		return true
	}
	if oldTask.Progress != newTask.Progress {
		return true
	}
	if oldTask.PromptEn != newTask.PromptEn {
		return true
	}
	if oldTask.State != newTask.State {
		return true
	}
	if oldTask.SubmitTime != newTask.SubmitTime {
		return true
	}
	if oldTask.StartTime != newTask.StartTime {
		return true
	}
	if oldTask.FinishTime != newTask.FinishTime {
		return true
	}
	if oldTask.ImageUrl != newTask.ImageUrl {
		return true
	}
	if oldTask.Status != newTask.Status {
		return true
	}
	if oldTask.FailReason != newTask.FailReason {
		return true
	}
	if oldTask.FinishTime != newTask.FinishTime {
		return true
	}
	if oldTask.Progress != "100%" && newTask.FailReason != "" {
		return true
	}

	return false
}

func GetAllMidjourney(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	// 解析其他查询参数
	queryParams := model.TaskQueryParams{
		ChannelID:      c.Query("channel_id"),
		MjID:           c.Query("mj_id"),
		StartTimestamp: c.Query("start_timestamp"),
		EndTimestamp:   c.Query("end_timestamp"),
	}

	logs := model.GetAllTasks(p*common.ItemsPerPage, common.ItemsPerPage, queryParams)
	if logs == nil {
		logs = make([]*model.Midjourney, 0)
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "",
		"data":    logs,
	})
}

func GetUserMidjourney(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	userId := c.GetInt("id")
	log.Printf("userId = %d \n", userId)

	queryParams := model.TaskQueryParams{
		MjID:           c.Query("mj_id"),
		StartTimestamp: c.Query("start_timestamp"),
		EndTimestamp:   c.Query("end_timestamp"),
	}

	logs := model.GetAllUserTask(userId, p*common.ItemsPerPage, common.ItemsPerPage, queryParams)
	if logs == nil {
		logs = make([]*model.Midjourney, 0)
	}
	if !strings.Contains(common.ServerAddress, "localhost") {
		for i, midjourney := range logs {
			midjourney.ImageUrl = common.ServerAddress + "/mj/image/" + midjourney.MjId
			logs[i] = midjourney
		}
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "",
		"data":    logs,
	})
}
