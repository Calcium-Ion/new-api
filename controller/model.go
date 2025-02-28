package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	"one-api/model"
	"one-api/relay"
	"one-api/relay/channel/ai360"
	"one-api/relay/channel/lingyiwanwu"
	"one-api/relay/channel/minimax"
	"one-api/relay/channel/moonshot"
	relaycommon "one-api/relay/common"
	relayconstant "one-api/relay/constant"
)

// https://platform.openai.com/docs/api-reference/models/list

var openAIModels []dto.OpenAIModels
var openAIModelsMap map[string]dto.OpenAIModels
var channelId2Models map[int][]string

func getPermission() []dto.OpenAIModelPermission {
	var permission []dto.OpenAIModelPermission
	permission = append(permission, dto.OpenAIModelPermission{
		Id:                 "modelperm-LwHkVFn8AcMItP432fKKDIKJ",
		Object:             "model_permission",
		Created:            1626777600,
		AllowCreateEngine:  true,
		AllowSampling:      true,
		AllowLogprobs:      true,
		AllowSearchIndices: false,
		AllowView:          true,
		AllowFineTuning:    false,
		Organization:       "*",
		Group:              nil,
		IsBlocking:         false,
	})
	return permission
}

func init() {
	// https://platform.openai.com/docs/models/model-endpoint-compatibility
	permission := getPermission()
	for i := 0; i < relayconstant.APITypeDummy; i++ {
		if i == relayconstant.APITypeAIProxyLibrary {
			continue
		}
		adaptor := relay.GetAdaptor(i)
		channelName := adaptor.GetChannelName()
		modelNames := adaptor.GetModelList()
		for _, modelName := range modelNames {
			openAIModels = append(openAIModels, dto.OpenAIModels{
				Id:         modelName,
				Object:     "model",
				Created:    1626777600,
				OwnedBy:    channelName,
				Permission: permission,
				Root:       modelName,
				Parent:     nil,
			})
		}
	}
	for _, modelName := range ai360.ModelList {
		openAIModels = append(openAIModels, dto.OpenAIModels{
			Id:         modelName,
			Object:     "model",
			Created:    1626777600,
			OwnedBy:    ai360.ChannelName,
			Permission: permission,
			Root:       modelName,
			Parent:     nil,
		})
	}
	for _, modelName := range moonshot.ModelList {
		openAIModels = append(openAIModels, dto.OpenAIModels{
			Id:         modelName,
			Object:     "model",
			Created:    1626777600,
			OwnedBy:    moonshot.ChannelName,
			Permission: permission,
			Root:       modelName,
			Parent:     nil,
		})
	}
	for _, modelName := range lingyiwanwu.ModelList {
		openAIModels = append(openAIModels, dto.OpenAIModels{
			Id:         modelName,
			Object:     "model",
			Created:    1626777600,
			OwnedBy:    lingyiwanwu.ChannelName,
			Permission: permission,
			Root:       modelName,
			Parent:     nil,
		})
	}
	for _, modelName := range minimax.ModelList {
		openAIModels = append(openAIModels, dto.OpenAIModels{
			Id:         modelName,
			Object:     "model",
			Created:    1626777600,
			OwnedBy:    minimax.ChannelName,
			Permission: permission,
			Root:       modelName,
			Parent:     nil,
		})
	}
	for modelName, _ := range constant.MidjourneyModel2Action {
		openAIModels = append(openAIModels, dto.OpenAIModels{
			Id:         modelName,
			Object:     "model",
			Created:    1626777600,
			OwnedBy:    "midjourney",
			Permission: permission,
			Root:       modelName,
			Parent:     nil,
		})
	}
	openAIModelsMap = make(map[string]dto.OpenAIModels)
	for _, aiModel := range openAIModels {
		openAIModelsMap[aiModel.Id] = aiModel
	}
	channelId2Models = make(map[int][]string)
	for i := 1; i <= common.ChannelTypeDummy; i++ {
		apiType, success := relayconstant.ChannelType2APIType(i)
		if !success || apiType == relayconstant.APITypeAIProxyLibrary {
			continue
		}
		meta := &relaycommon.RelayInfo{ChannelType: i}
		adaptor := relay.GetAdaptor(apiType)
		adaptor.Init(meta)
		channelId2Models[i] = adaptor.GetModelList()
	}
}

func ListModels(c *gin.Context) {
	userOpenAiModels := make([]dto.OpenAIModels, 0)
	permission := getPermission()

	modelLimitEnable := c.GetBool("token_model_limit_enabled")
	if modelLimitEnable {
		s, ok := c.Get("token_model_limit")
		var tokenModelLimit map[string]bool
		if ok {
			tokenModelLimit = s.(map[string]bool)
		} else {
			tokenModelLimit = map[string]bool{}
		}
		for allowModel, _ := range tokenModelLimit {
			if _, ok := openAIModelsMap[allowModel]; ok {
				userOpenAiModels = append(userOpenAiModels, openAIModelsMap[allowModel])
			} else {
				userOpenAiModels = append(userOpenAiModels, dto.OpenAIModels{
					Id:         allowModel,
					Object:     "model",
					Created:    1626777600,
					OwnedBy:    "custom",
					Permission: permission,
					Root:       allowModel,
					Parent:     nil,
				})
			}
		}
	} else {
		userId := c.GetInt("id")
		userGroup, err := model.GetUserGroup(userId, true)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "get user group failed",
			})
			return
		}
		group := userGroup
		tokenGroup := c.GetString("token_group")
		if tokenGroup != "" {
			group = tokenGroup
		}
		models := model.GetGroupModels(group)
		for _, s := range models {
			if _, ok := openAIModelsMap[s]; ok {
				userOpenAiModels = append(userOpenAiModels, openAIModelsMap[s])
			} else {
				userOpenAiModels = append(userOpenAiModels, dto.OpenAIModels{
					Id:         s,
					Object:     "model",
					Created:    1626777600,
					OwnedBy:    "custom",
					Permission: permission,
					Root:       s,
					Parent:     nil,
				})
			}
		}
	}
	c.JSON(200, gin.H{
		"success": true,
		"data":    userOpenAiModels,
	})
}

func ChannelListModels(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    openAIModels,
	})
}

func DashboardListModels(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    channelId2Models,
	})
}

func EnabledListModels(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    model.GetEnabledModels(),
	})
}

func RetrieveModel(c *gin.Context) {
	modelId := c.Param("model")
	if aiModel, ok := openAIModelsMap[modelId]; ok {
		c.JSON(200, aiModel)
	} else {
		openAIError := dto.OpenAIError{
			Message: fmt.Sprintf("The model '%s' does not exist", modelId),
			Type:    "invalid_request_error",
			Param:   "model",
			Code:    "model_not_found",
		}
		c.JSON(200, gin.H{
			"error": openAIError,
		})
	}
}
