package middleware

import (
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ModelRequest struct {
	Model string `json:"model"`
}

func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		userId := c.GetInt("id")
		var channel *model.Channel
		channelId, ok := c.Get("channelId")
		if ok {
			id, err := strconv.Atoi(channelId.(string))
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			channel, err = model.GetChannelById(id, true)
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			if channel.Status != common.ChannelStatusEnabled {
				abortWithMessage(c, http.StatusForbidden, "该渠道已被禁用")
				return
			}
		} else {
			// Select a channel for the user
			var modelRequest ModelRequest
			var err error
			if strings.HasPrefix(c.Request.URL.Path, "/mj") {
				// Midjourney
				if modelRequest.Model == "" {
					modelRequest.Model = "midjourney"
				}
			} else if !strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions") {
				err = common.UnmarshalBodyReusable(c, &modelRequest)
			}
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的请求, "+err.Error())
				return
			}
			if strings.HasPrefix(c.Request.URL.Path, "/v1/moderations") {
				if modelRequest.Model == "" {
					modelRequest.Model = "text-moderation-stable"
				}
			}
			if strings.HasSuffix(c.Request.URL.Path, "embeddings") {
				if modelRequest.Model == "" {
					modelRequest.Model = c.Param("model")
				}
			}
			if strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations") {
				if modelRequest.Model == "" {
					modelRequest.Model = "dall-e"
				}
			}
			if strings.HasPrefix(c.Request.URL.Path, "/v1/audio") {
				if modelRequest.Model == "" {
					if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/speech") {
						modelRequest.Model = "tts-1"
					} else {
						modelRequest.Model = "whisper-1"
					}
				}
			}
			// check token model mapping
			modelLimitEnable := c.GetBool("token_model_limit_enabled")
			if modelLimitEnable {
				s, ok := c.Get("token_model_limit")
				var tokenModelLimit map[string]bool
				if ok {
					tokenModelLimit = s.(map[string]bool)
				} else {
					tokenModelLimit = map[string]bool{}
				}
				if tokenModelLimit != nil {
					if _, ok := tokenModelLimit[modelRequest.Model]; !ok {
						abortWithMessage(c, http.StatusForbidden, "该令牌无权访问模型 "+modelRequest.Model)
						return
					}
				} else {
					// token model limit is empty, all models are not allowed
					abortWithMessage(c, http.StatusForbidden, "该令牌无权访问任何模型")
					return
				}
			}

			userGroup, _ := model.CacheGetUserGroup(userId)
			c.Set("group", userGroup)

			channel, err = model.CacheGetRandomSatisfiedChannel(userGroup, modelRequest.Model)
			if err != nil {
				message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", userGroup, modelRequest.Model)
				// 如果错误，但是渠道不为空，说明是数据库一致性问题
				if channel != nil {
					common.SysError(fmt.Sprintf("渠道不存在：%d", channel.Id))
					message = "数据库一致性已被破坏，请联系管理员"
				}
				// 如果错误，而且渠道为空，说明是没有可用渠道
				abortWithMessage(c, http.StatusServiceUnavailable, message)
				return
			}
			if channel == nil {
				abortWithMessage(c, http.StatusServiceUnavailable, fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道（数据库一致性已被破坏）", userGroup, modelRequest.Model))
				return
			}
		}
		c.Set("channel", channel.Type)
		c.Set("channel_id", channel.Id)
		c.Set("channel_name", channel.Name)
		ban := true
		// parse *int to bool
		if channel.AutoBan != nil && *channel.AutoBan == 0 {
			ban = false
		}
		c.Set("auto_ban", ban)
		c.Set("model_mapping", channel.GetModelMapping())
		c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))
		c.Set("base_url", channel.GetBaseURL())
		switch channel.Type {
		case common.ChannelTypeAzure:
			c.Set("api_version", channel.Other)
		case common.ChannelTypeXunfei:
			c.Set("api_version", channel.Other)
		case common.ChannelTypeAIProxyLibrary:
			c.Set("library_id", channel.Other)
		case common.ChannelTypeGemini:
			c.Set("api_version", channel.Other)
		}
		c.Next()
	}
}
