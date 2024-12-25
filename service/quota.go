package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"one-api/common"
	"one-api/dto"
	"one-api/model"
	relaycommon "one-api/relay/common"
	"one-api/setting"
	"strings"
	"time"
)

func PreWssConsumeQuota(ctx *gin.Context, relayInfo *relaycommon.RelayInfo, usage *dto.RealtimeUsage) error {
	if relayInfo.UsePrice {
		return nil
	}
	userQuota, err := model.GetUserQuota(relayInfo.UserId)
	if err != nil {
		return err
	}

	token, err := model.CacheGetTokenByKey(strings.TrimLeft(relayInfo.TokenKey, "sk-"))
	if err != nil {
		return err
	}

	modelName := relayInfo.UpstreamModelName
	textInputTokens := usage.InputTokenDetails.TextTokens
	textOutTokens := usage.OutputTokenDetails.TextTokens
	audioInputTokens := usage.InputTokenDetails.AudioTokens
	audioOutTokens := usage.OutputTokenDetails.AudioTokens

	completionRatio := common.GetCompletionRatio(modelName)
	audioRatio := common.GetAudioRatio(relayInfo.UpstreamModelName)
	audioCompletionRatio := common.GetAudioCompletionRatio(modelName)
	groupRatio := setting.GetGroupRatio(relayInfo.Group)
	modelRatio := common.GetModelRatio(modelName)

	ratio := groupRatio * modelRatio

	quota := textInputTokens + int(math.Round(float64(textOutTokens)*completionRatio))
	quota += int(math.Round(float64(audioInputTokens)*audioRatio)) + int(math.Round(float64(audioOutTokens)*audioRatio*audioCompletionRatio))

	quota = int(math.Round(float64(quota) * ratio))
	if ratio != 0 && quota <= 0 {
		quota = 1
	}

	if userQuota < quota {
		return errors.New(fmt.Sprintf("用户额度不足，剩余额度为 %d", userQuota))
	}

	if !token.UnlimitedQuota && token.RemainQuota < quota {
		return errors.New(fmt.Sprintf("令牌额度不足，剩余额度为 %d", token.RemainQuota))
	}

	err = model.PostConsumeTokenQuota(relayInfo, 0, quota, 0, false)
	if err != nil {
		return err
	}
	common.LogInfo(ctx, "realtime streaming consume quota success, quota: "+fmt.Sprintf("%d", quota))
	err = model.CacheUpdateUserQuota(relayInfo.UserId)
	if err != nil {
		return err
	}
	return nil
}

func PostWssConsumeQuota(ctx *gin.Context, relayInfo *relaycommon.RelayInfo, modelName string,
	usage *dto.RealtimeUsage, ratio float64, preConsumedQuota int, userQuota int, modelRatio float64,
	groupRatio float64,
	modelPrice float64, usePrice bool, extraContent string) {

	useTimeSeconds := time.Now().Unix() - relayInfo.StartTime.Unix()
	textInputTokens := usage.InputTokenDetails.TextTokens
	textOutTokens := usage.OutputTokenDetails.TextTokens

	audioInputTokens := usage.InputTokenDetails.AudioTokens
	audioOutTokens := usage.OutputTokenDetails.AudioTokens

	tokenName := ctx.GetString("token_name")
	completionRatio := common.GetCompletionRatio(modelName)
	audioRatio := common.GetAudioRatio(relayInfo.UpstreamModelName)
	audioCompletionRatio := common.GetAudioCompletionRatio(modelName)

	quota := 0
	if !usePrice {
		quota = int(math.Round(float64(textInputTokens) + float64(textOutTokens)*completionRatio))
		quota += int(math.Round(float64(audioInputTokens)*audioRatio + float64(audioOutTokens)*audioRatio*audioCompletionRatio))
		quota = int(math.Round(float64(quota) * ratio))
		if ratio != 0 && quota <= 0 {
			quota = 1
		}
	} else {
		quota = int(modelPrice * common.QuotaPerUnit * groupRatio)
	}
	totalTokens := usage.TotalTokens
	var logContent string
	if !usePrice {
		logContent = fmt.Sprintf("模型倍率 %.2f，补全倍率 %.2f，音频倍率 %.2f，音频补全倍率 %.2f，分组倍率 %.2f", modelRatio, completionRatio, audioRatio, audioCompletionRatio, groupRatio)
	} else {
		logContent = fmt.Sprintf("模型价格 %.2f，分组倍率 %.2f", modelPrice, groupRatio)
	}

	// record all the consume log even if quota is 0
	if totalTokens == 0 {
		// in this case, must be some error happened
		// we cannot just return, because we may have to return the pre-consumed quota
		quota = 0
		logContent += fmt.Sprintf("（可能是上游超时）")
		common.LogError(ctx, fmt.Sprintf("total tokens is 0, cannot consume quota, userId %d, channelId %d, "+
			"tokenId %d, model %s， pre-consumed quota %d", relayInfo.UserId, relayInfo.ChannelId, relayInfo.TokenId, modelName, preConsumedQuota))
	} else {
		//if sensitiveResp != nil {
		//	logContent += fmt.Sprintf("，敏感词：%s", strings.Join(sensitiveResp.SensitiveWords, ", "))
		//}
		//quotaDelta := quota - preConsumedQuota
		//if quotaDelta != 0 {
		//	err := model.PostConsumeTokenQuota(relayInfo, userQuota, quotaDelta, preConsumedQuota, true)
		//	if err != nil {
		//		common.LogError(ctx, "error consuming token remain quota: "+err.Error())
		//	}
		//}

		//err := model.CacheUpdateUserQuota(relayInfo.UserId)
		//if err != nil {
		//	common.LogError(ctx, "error update user quota cache: "+err.Error())
		//}
		model.UpdateUserUsedQuotaAndRequestCount(relayInfo.UserId, quota)
		model.UpdateChannelUsedQuota(relayInfo.ChannelId, quota)
	}

	logModel := modelName
	if extraContent != "" {
		logContent += ", " + extraContent
	}
	other := GenerateWssOtherInfo(ctx, relayInfo, usage, modelRatio, groupRatio, completionRatio, audioRatio, audioCompletionRatio, modelPrice)
	model.RecordConsumeLog(ctx, relayInfo.UserId, relayInfo.ChannelId, usage.InputTokens, usage.OutputTokens, logModel,
		tokenName, quota, logContent, relayInfo.TokenId, userQuota, int(useTimeSeconds), relayInfo.IsStream, relayInfo.Group, other)
}

func PostAudioConsumeQuota(ctx *gin.Context, relayInfo *relaycommon.RelayInfo,
	usage *dto.Usage, ratio float64, preConsumedQuota int, userQuota int, modelRatio float64,
	groupRatio float64,
	modelPrice float64, usePrice bool, extraContent string) {

	useTimeSeconds := time.Now().Unix() - relayInfo.StartTime.Unix()
	textInputTokens := usage.PromptTokensDetails.TextTokens
	textOutTokens := usage.CompletionTokenDetails.TextTokens

	audioInputTokens := usage.PromptTokensDetails.AudioTokens
	audioOutTokens := usage.CompletionTokenDetails.AudioTokens

	tokenName := ctx.GetString("token_name")
	completionRatio := common.GetCompletionRatio(relayInfo.UpstreamModelName)
	audioRatio := common.GetAudioRatio(relayInfo.UpstreamModelName)
	audioCompletionRatio := common.GetAudioCompletionRatio(relayInfo.UpstreamModelName)

	quota := 0
	if !usePrice {
		quota = int(math.Round(float64(textInputTokens) + float64(textOutTokens)*completionRatio))
		quota += int(math.Round(float64(audioInputTokens)*audioRatio + float64(audioOutTokens)*audioRatio*audioCompletionRatio))
		quota = int(math.Round(float64(quota) * ratio))
		if ratio != 0 && quota <= 0 {
			quota = 1
		}
	} else {
		quota = int(modelPrice * common.QuotaPerUnit * groupRatio)
	}
	totalTokens := usage.TotalTokens
	var logContent string
	if !usePrice {
		logContent = fmt.Sprintf("模型倍率 %.2f，补全倍率 %.2f，音频倍率 %.2f，音频补全倍率 %.2f，分组倍率 %.2f", modelRatio, completionRatio, audioRatio, audioCompletionRatio, groupRatio)
	} else {
		logContent = fmt.Sprintf("模型价格 %.2f，分组倍率 %.2f", modelPrice, groupRatio)
	}

	// record all the consume log even if quota is 0
	if totalTokens == 0 {
		// in this case, must be some error happened
		// we cannot just return, because we may have to return the pre-consumed quota
		quota = 0
		logContent += fmt.Sprintf("（可能是上游超时）")
		common.LogError(ctx, fmt.Sprintf("total tokens is 0, cannot consume quota, userId %d, channelId %d, "+
			"tokenId %d, model %s， pre-consumed quota %d", relayInfo.UserId, relayInfo.ChannelId, relayInfo.TokenId, relayInfo.UpstreamModelName, preConsumedQuota))
	} else {
		quotaDelta := quota - preConsumedQuota
		if quotaDelta != 0 {
			err := model.PostConsumeTokenQuota(relayInfo, userQuota, quotaDelta, preConsumedQuota, true)
			if err != nil {
				common.LogError(ctx, "error consuming token remain quota: "+err.Error())
			}
		}
		err := model.CacheUpdateUserQuota(relayInfo.UserId)
		if err != nil {
			common.LogError(ctx, "error update user quota cache: "+err.Error())
		}
		model.UpdateUserUsedQuotaAndRequestCount(relayInfo.UserId, quota)
		model.UpdateChannelUsedQuota(relayInfo.ChannelId, quota)
	}

	logModel := relayInfo.UpstreamModelName
	if extraContent != "" {
		logContent += ", " + extraContent
	}
	other := GenerateAudioOtherInfo(ctx, relayInfo, usage, modelRatio, groupRatio, completionRatio, audioRatio, audioCompletionRatio, modelPrice)
	model.RecordConsumeLog(ctx, relayInfo.UserId, relayInfo.ChannelId, usage.PromptTokens, usage.CompletionTokens, logModel,
		tokenName, quota, logContent, relayInfo.TokenId, userQuota, int(useTimeSeconds), relayInfo.IsStream, relayInfo.Group, other)
}
