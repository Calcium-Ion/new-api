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

func GetPricing(user *User, openAIModels []dto.OpenAIModels) []dto.ModelPricing {
	updatePricingLock.Lock()
	defer updatePricingLock.Unlock()

	if time.Since(lastGetPricingTime) > time.Minute*1 || len(pricingMap) == 0 {
		updatePricing(openAIModels)
	}
	if user != nil {
		userPricingMap := make([]dto.ModelPricing, 0)
		models := GetGroupModels(user.Group)
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

func updatePricing(openAIModels []dto.OpenAIModels) {
	modelRatios := common.GetModelRatios()
	enabledModels := GetEnabledModels()
	allModels := make(map[string]string)
	for _, openAIModel := range openAIModels {
		if common.StringsContains(enabledModels, openAIModel.Id) {
			allModels[openAIModel.Id] = openAIModel.OwnedBy
		}
	}
	for model, _ := range modelRatios {
		if common.StringsContains(enabledModels, model) {
			if _, ok := allModels[model]; !ok {
				allModels[model] = "custom"
			}
		}
	}
	pricingMap = make([]dto.ModelPricing, 0)
	for model, ownerBy := range allModels {
		pricing := dto.ModelPricing{
			Available: true,
			ModelName: model,
			OwnerBy:   ownerBy,
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
