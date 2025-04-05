package helper

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"one-api/common"
	constant2 "one-api/constant"
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
	CacheCreationRatio     float64
	ShouldPreConsumedQuota int
}

func (p PriceData) ToSetting() string {
	return fmt.Sprintf("ModelPrice: %f, ModelRatio: %f, CompletionRatio: %f, CacheRatio: %f, GroupRatio: %f, UsePrice: %t, CacheCreationRatio: %f, ShouldPreConsumedQuota: %d", p.ModelPrice, p.ModelRatio, p.CompletionRatio, p.CacheRatio, p.GroupRatio, p.UsePrice, p.CacheCreationRatio, p.ShouldPreConsumedQuota)
}

func ModelPriceHelper(c *gin.Context, info *relaycommon.RelayInfo, promptTokens int, maxTokens int) (PriceData, error) {
	modelPrice, usePrice := operation_setting.GetModelPrice(info.OriginModelName, false)
	groupRatio := setting.GetGroupRatio(info.Group)
	var preConsumedQuota int
	var modelRatio float64
	var completionRatio float64
	var cacheRatio float64
	var cacheCreationRatio float64
	if !usePrice {
		preConsumedTokens := common.PreConsumedQuota
		if maxTokens != 0 {
			preConsumedTokens = promptTokens + maxTokens
		}
		var success bool
		modelRatio, success = operation_setting.GetModelRatio(info.OriginModelName)
		if !success {
			acceptUnsetRatio := false
			if accept, ok := info.UserSetting[constant2.UserAcceptUnsetRatioModel]; ok {
				b, ok := accept.(bool)
				if ok {
					acceptUnsetRatio = b
				}
			}
			if !acceptUnsetRatio {
				if info.UserId == 1 {
					return PriceData{}, fmt.Errorf("模型 %s 倍率或价格未配置，请设置或开始自用模式；Model %s ratio or price not set, please set or start self-use mode", info.OriginModelName, info.OriginModelName)
				} else {
					return PriceData{}, fmt.Errorf("模型 %s 倍率或价格未配置, 请联系管理员设置；Model %s ratio or price not set, please contact administrator to set", info.OriginModelName, info.OriginModelName)
				}
			}
		}
		completionRatio = operation_setting.GetCompletionRatio(info.OriginModelName)
		cacheRatio, _ = operation_setting.GetCacheRatio(info.OriginModelName)
		cacheCreationRatio, _ = operation_setting.GetCreateCacheRatio(info.OriginModelName)
		ratio := modelRatio * groupRatio
		preConsumedQuota = int(float64(preConsumedTokens) * ratio)
	} else {
		preConsumedQuota = int(modelPrice * common.QuotaPerUnit * groupRatio)
	}

	priceData := PriceData{
		ModelPrice:             modelPrice,
		ModelRatio:             modelRatio,
		CompletionRatio:        completionRatio,
		GroupRatio:             groupRatio,
		UsePrice:               usePrice,
		CacheRatio:             cacheRatio,
		CacheCreationRatio:     cacheCreationRatio,
		ShouldPreConsumedQuota: preConsumedQuota,
	}

	if common.DebugEnabled {
		println(fmt.Sprintf("model_price_helper result: %s", priceData.ToSetting()))
	}

	return priceData, nil
}
