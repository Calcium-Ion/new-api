package helper

import (
	"fmt"
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

func ModelPriceHelper(c *gin.Context, info *relaycommon.RelayInfo, promptTokens int, maxTokens int) (PriceData, error) {
	modelPrice, usePrice := common.GetModelPrice(info.OriginModelName, false)
	groupRatio := setting.GetGroupRatio(info.Group)
	var preConsumedQuota int
	var modelRatio float64
	if !usePrice {
		preConsumedTokens := common.PreConsumedQuota
		if maxTokens != 0 {
			preConsumedTokens = promptTokens + maxTokens
		}
		var success bool
		modelRatio, success = common.GetModelRatio(info.OriginModelName)
		if !success {
			return PriceData{}, fmt.Errorf("model %s ratio or price not found, please contact admin", info.OriginModelName)
		}
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
	}, nil
}
