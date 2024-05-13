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
	"one-api/relay/channel/moonshot"
	relaycommon "one-api/relay/common"
	relayconstant "one-api/relay/constant"
)

// https://platform.openai.com/docs/api-reference/models/list

type OpenAIModelPermission struct {
	Id                 string  `json:"id"`
	Object             string  `json:"object"`
	Created            int     `json:"created"`
	AllowCreateEngine  bool    `json:"allow_create_engine"`
	AllowSampling      bool    `json:"allow_sampling"`
	AllowLogprobs      bool    `json:"allow_logprobs"`
	AllowSearchIndices bool    `json:"allow_search_indices"`
	AllowView          bool    `json:"allow_view"`
	AllowFineTuning    bool    `json:"allow_fine_tuning"`
	Organization       string  `json:"organization"`
	Group              *string `json:"group"`
	IsBlocking         bool    `json:"is_blocking"`
}

type OpenAIModels struct {
	Id         string                  `json:"id"`
	Object     string                  `json:"object"`
	Created    int                     `json:"created"`
	OwnedBy    string                  `json:"owned_by"`
	Permission []OpenAIModelPermission `json:"permission"`
	Root       string                  `json:"root"`
	Parent     *string                 `json:"parent"`
}

var openAIModels []OpenAIModels
var openAIModelsMap map[string]OpenAIModels
var channelId2Models map[int][]string

func getPermission() []OpenAIModelPermission {
	var permission []OpenAIModelPermission
	permission = append(permission, OpenAIModelPermission{
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
			openAIModels = append(openAIModels, OpenAIModels{
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
		openAIModels = append(openAIModels, OpenAIModels{
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
		openAIModels = append(openAIModels, OpenAIModels{
			Id:         modelName,
			Object:     "model",
			Created:    1626777600,
			OwnedBy:    "moonshot",
			Permission: permission,
			Root:       modelName,
			Parent:     nil,
		})
	}
	for _, modelName := range lingyiwanwu.ModelList {
		openAIModels = append(openAIModels, OpenAIModels{
			Id:         modelName,
			Object:     "model",
			Created:    1626777600,
			OwnedBy:    "lingyiwanwu",
			Permission: permission,
			Root:       modelName,
			Parent:     nil,
		})
	}
	for modelName, _ := range constant.MidjourneyModel2Action {
		openAIModels = append(openAIModels, OpenAIModels{
			Id:         modelName,
			Object:     "model",
			Created:    1626777600,
			OwnedBy:    "midjourney",
			Permission: permission,
			Root:       modelName,
			Parent:     nil,
		})
	}
	openAIModelsMap = make(map[string]OpenAIModels)
	for _, model := range openAIModels {
		openAIModelsMap[model.Id] = model
	}
	channelId2Models = make(map[int][]string)
	for i := 1; i <= common.ChannelTypeDummy; i++ {
		apiType := relayconstant.ChannelType2APIType(i)
		if apiType == -1 || apiType == relayconstant.APITypeAIProxyLibrary {
			continue
		}
		meta := &relaycommon.RelayInfo{ChannelType: i}
		adaptor := relay.GetAdaptor(apiType)
		adaptor.Init(meta, dto.GeneralOpenAIRequest{})
		channelId2Models[i] = adaptor.GetModelList()
	}
}

func ListModels(c *gin.Context) {
	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	models := model.GetGroupModels(user.Group)
	userOpenAiModels := make([]OpenAIModels, 0)
	permission := getPermission()
	for _, s := range models {
		if _, ok := openAIModelsMap[s]; ok {
			userOpenAiModels = append(userOpenAiModels, openAIModelsMap[s])
		} else {
			userOpenAiModels = append(userOpenAiModels, OpenAIModels{
				Id:         s,
				Object:     "model",
				Created:    1626777600,
				OwnedBy:    "openai",
				Permission: permission,
				Root:       s,
				Parent:     nil,
			})
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

func RetrieveModel(c *gin.Context) {
	modelId := c.Param("model")
	if model, ok := openAIModelsMap[modelId]; ok {
		c.JSON(200, model)
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
