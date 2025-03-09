package helper

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"one-api/common"
	relaycommon "one-api/relay/common"
	"one-api/setting"
	"one-api/setting/operation_setting"
)

type PriceData struct {
	ModelPrice             float64
	ModelRatio             float64
	CompletionRatio        float64
	CacheRatio             float64
	GroupRatio             float64
	UsePrice               bool
	ShouldPreConsumedQuota int
}

func ModelPriceHelper(c *gin.Context, info *relaycommon.RelayInfo, promptTokens int, maxTokens int) (PriceData, error) {
	modelPrice, usePrice := operation_setting.GetModelPrice(info.OriginModelName, false)
	groupRatio := setting.GetGroupRatio(info.Group)
	var preConsumedQuota int
	var modelRatio float64
	var completionRatio float64
	var cacheRatio float64
	if !usePrice {
		preConsumedTokens := common.PreConsumedQuota
		if maxTokens != 0 {
			preConsumedTokens = promptTokens + maxTokens
		}
		var success bool
		modelRatio, success = operation_setting.GetModelRatio(info.OriginModelName)
		if !success {
			if info.UserId == 1 {
				return PriceData{}, fmt.Errorf("模型 %s 倍率或价格未配置，请设置或开始自用模式；Model %s ratio or price not set, please set or start self-use mode", info.OriginModelName, info.OriginModelName)
			} else {
				return PriceData{}, fmt.Errorf("模型 %s 倍率或价格未配置, 请联系管理员设置；Model %s ratio or price not set, please contact administrator to set", info.OriginModelName, info.OriginModelName)
			}
		}
		completionRatio = operation_setting.GetCompletionRatio(info.OriginModelName)
		cacheRatio, _ = operation_setting.GetCacheRatio(info.OriginModelName)
		ratio := modelRatio * groupRatio
		preConsumedQuota = int(float64(preConsumedTokens) * ratio)
	} else {
		preConsumedQuota = int(modelPrice * common.QuotaPerUnit * groupRatio)
	}
	return PriceData{
		ModelPrice:             modelPrice,
		ModelRatio:             modelRatio,
		CompletionRatio:        completionRatio,
		GroupRatio:             groupRatio,
		UsePrice:               usePrice,
		CacheRatio:             cacheRatio,
		ShouldPreConsumedQuota: preConsumedQuota,
	}, nil
}
