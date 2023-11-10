package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
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

func init() {
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
	// https://platform.openai.com/docs/models/model-endpoint-compatibility
	openAIModels = []OpenAIModels{
		{
			Id:         "dall-e-2",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "dall-e-2",
			Parent:     nil,
		},
		{
			Id:         "dall-e-3",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "dall-e-3",
			Parent:     nil,
		},
		{
			Id:         "whisper-1",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "whisper-1",
			Parent:     nil,
		},
		{
			Id:         "gpt-3.5-turbo",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-3.5-turbo",
			Parent:     nil,
		},
		{
			Id:         "gpt-3.5-turbo-0301",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-3.5-turbo-0301",
			Parent:     nil,
		},
		{
			Id:         "gpt-3.5-turbo-0613",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-3.5-turbo-0613",
			Parent:     nil,
		},
		{
			Id:         "gpt-3.5-turbo-16k",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-3.5-turbo-16k",
			Parent:     nil,
		},
		{
			Id:         "gpt-3.5-turbo-16k-0613",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-3.5-turbo-16k-0613",
			Parent:     nil,
		},
		{
			Id:         "gpt-3.5-turbo-1106",
			Object:     "model",
			Created:    1699593571,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-3.5-turbo-1106",
			Parent:     nil,
		},
		{
			Id:         "gpt-3.5-turbo-instruct",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-3.5-turbo-instruct",
			Parent:     nil,
		},
		{
			Id:         "gpt-4",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-4",
			Parent:     nil,
		},
		{
			Id:         "gpt-4-0314",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-4-0314",
			Parent:     nil,
		},
		{
			Id:         "gpt-4-0613",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-4-0613",
			Parent:     nil,
		},
		{
			Id:         "gpt-4-32k",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-4-32k",
			Parent:     nil,
		},
		{
			Id:         "gpt-4-32k-0314",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-4-32k-0314",
			Parent:     nil,
		},
		{
			Id:         "gpt-4-32k-0613",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-4-32k-0613",
			Parent:     nil,
		},
		{
			Id:         "gpt-4-1106-preview",
			Object:     "model",
			Created:    1699593571,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-4-1106-preview",
			Parent:     nil,
		},
		{
			Id:         "gpt-4-vision-preview",
			Object:     "model",
			Created:    1699593571,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-4-vision-preview",
			Parent:     nil,
		},
		{
			Id:         "text-embedding-ada-002",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "text-embedding-ada-002",
			Parent:     nil,
		},
		{
			Id:         "text-davinci-003",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "text-davinci-003",
			Parent:     nil,
		},
		{
			Id:         "text-davinci-002",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "text-davinci-002",
			Parent:     nil,
		},
		{
			Id:         "text-curie-001",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "text-curie-001",
			Parent:     nil,
		},
		{
			Id:         "text-babbage-001",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "text-babbage-001",
			Parent:     nil,
		},
		{
			Id:         "text-ada-001",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "text-ada-001",
			Parent:     nil,
		},
		{
			Id:         "text-moderation-latest",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "text-moderation-latest",
			Parent:     nil,
		},
		{
			Id:         "text-moderation-stable",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "text-moderation-stable",
			Parent:     nil,
		},
		{
			Id:         "text-davinci-edit-001",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "text-davinci-edit-001",
			Parent:     nil,
		},
		{
			Id:         "code-davinci-edit-001",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "code-davinci-edit-001",
			Parent:     nil,
		},
		{
			Id:         "claude-instant-1",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "anthropic",
			Permission: permission,
			Root:       "claude-instant-1",
			Parent:     nil,
		},
		{
			Id:         "claude-2",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "anthropic",
			Permission: permission,
			Root:       "claude-2",
			Parent:     nil,
		},
		{
			Id:         "ERNIE-Bot",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "baidu",
			Permission: permission,
			Root:       "ERNIE-Bot",
			Parent:     nil,
		},
		{
			Id:         "ERNIE-Bot-turbo",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "baidu",
			Permission: permission,
			Root:       "ERNIE-Bot-turbo",
			Parent:     nil,
		},
		{
			Id:         "ERNIE-Bot-4",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "baidu",
			Permission: permission,
			Root:       "ERNIE-Bot-4",
			Parent:     nil,
		},
		{
			Id:         "Embedding-V1",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "baidu",
			Permission: permission,
			Root:       "Embedding-V1",
			Parent:     nil,
		},
		{
			Id:         "PaLM-2",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "google",
			Permission: permission,
			Root:       "PaLM-2",
			Parent:     nil,
		},
		{
			Id:         "chatglm_turbo",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "zhipu",
			Permission: permission,
			Root:       "chatglm_turbo",
			Parent:     nil,
		},
		{
			Id:         "chatglm_pro",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "zhipu",
			Permission: permission,
			Root:       "chatglm_pro",
			Parent:     nil,
		},
		{
			Id:         "chatglm_std",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "zhipu",
			Permission: permission,
			Root:       "chatglm_std",
			Parent:     nil,
		},
		{
			Id:         "chatglm_lite",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "zhipu",
			Permission: permission,
			Root:       "chatglm_lite",
			Parent:     nil,
		},
		{
			Id:         "qwen-turbo",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "ali",
			Permission: permission,
			Root:       "qwen-turbo",
			Parent:     nil,
		},
		{
			Id:         "qwen-plus",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "ali",
			Permission: permission,
			Root:       "qwen-plus",
			Parent:     nil,
		},
		{
			Id:         "text-embedding-v1",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "ali",
			Permission: permission,
			Root:       "text-embedding-v1",
			Parent:     nil,
		},
		{
			Id:         "SparkDesk",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "xunfei",
			Permission: permission,
			Root:       "SparkDesk",
			Parent:     nil,
		},
		{
			Id:         "360GPT_S2_V9",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "360",
			Permission: permission,
			Root:       "360GPT_S2_V9",
			Parent:     nil,
		},
		{
			Id:         "embedding-bert-512-v1",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "360",
			Permission: permission,
			Root:       "embedding-bert-512-v1",
			Parent:     nil,
		},
		{
			Id:         "embedding_s1_v1",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "360",
			Permission: permission,
			Root:       "embedding_s1_v1",
			Parent:     nil,
		},
		{
			Id:         "semantic_similarity_s1_v1",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "360",
			Permission: permission,
			Root:       "semantic_similarity_s1_v1",
			Parent:     nil,
		},
		{
			Id:         "hunyuan",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "tencent",
			Permission: permission,
			Root:       "hunyuan",
			Parent:     nil,
		},
	}
	openAIModelsMap = make(map[string]OpenAIModels)
	for _, model := range openAIModels {
		openAIModelsMap[model.Id] = model
	}
}

func ListModels(c *gin.Context) {
	c.JSON(200, gin.H{
		"object": "list",
		"data":   openAIModels,
	})
}

func RetrieveModel(c *gin.Context) {
	modelId := c.Param("model")
	if model, ok := openAIModelsMap[modelId]; ok {
		c.JSON(200, model)
	} else {
		openAIError := OpenAIError{
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
