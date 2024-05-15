package model

import (
	"one-api/common"
	"one-api/dto"
	"sync"
	"time"
)

var (
	pricingMap         []dto.ModelPricing
	lastGetPricingTime time.Time
	updatePricingLock  sync.Mutex
)

func GetPricing(group string) []dto.ModelPricing {
	updatePricingLock.Lock()
	defer updatePricingLock.Unlock()

	if time.Since(lastGetPricingTime) > time.Minute*1 || len(pricingMap) == 0 {
		updatePricing()
	}
	if group != "" {
		userPricingMap := make([]dto.ModelPricing, 0)
		models := GetGroupModels(group)
		for _, pricing := range pricingMap {
			if !common.StringsContains(models, pricing.ModelName) {
				pricing.Available = false
			}
			userPricingMap = append(userPricingMap, pricing)
		}
		return userPricingMap
	}
	return pricingMap
}

func updatePricing() {
	//modelRatios := common.GetModelRatios()
	enabledModels := GetEnabledModels()
	allModels := make(map[string]int)
	for i, model := range enabledModels {
		allModels[model] = i
	}

	pricingMap = make([]dto.ModelPricing, 0)
	for model, _ := range allModels {
		pricing := dto.ModelPricing{
			Available: true,
			ModelName: model,
		}
		modelPrice, findPrice := common.GetModelPrice(model, false)
		if findPrice {
			pricing.ModelPrice = modelPrice
			pricing.QuotaType = 1
		} else {
			pricing.ModelRatio = common.GetModelRatio(model)
			pricing.CompletionRatio = common.GetCompletionRatio(model)
			pricing.QuotaType = 0
		}
		pricingMap = append(pricingMap, pricing)
	}
	lastGetPricingTime = time.Now()
}
