package helper

import (
	"github.com/gin-gonic/gin"
	"one-api/common"
	relaycommon "one-api/relay/common"
	"one-api/setting"
)

type PriceData struct {
	ModelPrice             float64
	ModelRatio             float64
	GroupRatio             float64
	UsePrice               bool
	ShouldPreConsumedQuota int
}

func ModelPriceHelper(c *gin.Context, info *relaycommon.RelayInfo, promptTokens int, maxTokens int) PriceData {
	modelPrice, usePrice := common.GetModelPrice(info.OriginModelName, false)
	groupRatio := setting.GetGroupRatio(info.Group)
	var preConsumedQuota int
	var modelRatio float64
	if !usePrice {
		preConsumedTokens := common.PreConsumedQuota
		if maxTokens != 0 {
			preConsumedTokens = promptTokens + maxTokens
		}
		modelRatio = common.GetModelRatio(info.OriginModelName)
		ratio := modelRatio * groupRatio
		preConsumedQuota = int(float64(preConsumedTokens) * ratio)
	} else {
		preConsumedQuota = int(modelPrice * common.QuotaPerUnit * groupRatio)
	}
	return PriceData{
		ModelPrice:             modelPrice,
		ModelRatio:             modelRatio,
		GroupRatio:             groupRatio,
		UsePrice:               usePrice,
		ShouldPreConsumedQuota: preConsumedQuota,
	}
}
