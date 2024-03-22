package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"one-api/common"
	"one-api/dto"
	"one-api/relay"
	"one-api/relay/constant"
	relayconstant "one-api/relay/constant"
	"one-api/service"
	"strconv"
)

func Relay(c *gin.Context) {
	relayMode := constant.Path2RelayMode(c.Request.URL.Path)
	var err *dto.OpenAIErrorWithStatusCode
	switch relayMode {
	case relayconstant.RelayModeImagesGenerations:
		err = relay.RelayImageHelper(c, relayMode)
	case relayconstant.RelayModeAudioSpeech:
		fallthrough
	case relayconstant.RelayModeAudioTranslation:
		fallthrough
	case relayconstant.RelayModeAudioTranscription:
		err = relay.AudioHelper(c, relayMode)
	default:
		err = relay.TextHelper(c)
	}
	if err != nil {
		requestId := c.GetString(common.RequestIdKey)
		retryTimesStr := c.Query("retry")
		retryTimes, _ := strconv.Atoi(retryTimesStr)
		if retryTimesStr == "" {
			retryTimes = common.RetryTimes
		}
		if retryTimes > 0 {
			c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?retry=%d", c.Request.URL.Path, retryTimes-1))
		} else {
			if err.StatusCode == http.StatusTooManyRequests {
				//err.Error.Message = "当前分组上游负载已饱和，请稍后再试"
			}
			err.Error.Message = common.MessageWithRequestId(err.Error.Message, requestId)
			c.JSON(err.StatusCode, gin.H{
				"error": err.Error,
			})
		}
		channelId := c.GetInt("channel_id")
		autoBan := c.GetBool("auto_ban")
		common.LogError(c.Request.Context(), fmt.Sprintf("relay error (channel #%d): %s", channelId, err.Error.Message))
		// https://platform.openai.com/docs/guides/error-codes/api-errors
		if service.ShouldDisableChannel(&err.Error, err.StatusCode) && autoBan {
			channelId := c.GetInt("channel_id")
			channelName := c.GetString("channel_name")
			service.DisableChannel(channelId, channelName, err.Error.Message)
		}
	}
}

func RelayMidjourney(c *gin.Context) {
	relayMode := c.GetInt("relay_mode")
	var err *dto.MidjourneyResponse
	switch relayMode {
	case relayconstant.RelayModeMidjourneyNotify:
		err = relay.RelayMidjourneyNotify(c)
	case relayconstant.RelayModeMidjourneyTaskFetch, relayconstant.RelayModeMidjourneyTaskFetchByCondition:
		err = relay.RelayMidjourneyTask(c, relayMode)
	case relayconstant.RelayModeMidjourneyTaskImageSeed:
		err = relay.RelayMidjourneyTaskImageSeed(c)
	case relayconstant.RelayModeSwapFace:
		err = relay.RelaySwapFace(c)
	default:
		err = relay.RelayMidjourneySubmit(c, relayMode)
	}
	//err = relayMidjourneySubmit(c, relayMode)
	log.Println(err)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Code == 30 {
			err.Result = "当前分组负载已饱和，请稍后再试，或升级账户以提升服务质量。"
			statusCode = http.StatusTooManyRequests
		}
		c.JSON(statusCode, gin.H{
			"description": fmt.Sprintf("%s %s", err.Description, err.Result),
			"type":        "upstream_error",
			"code":        err.Code,
		})
		channelId := c.GetInt("channel_id")
		common.SysError(fmt.Sprintf("relay error (channel #%d): %s", channelId, fmt.Sprintf("%s %s", err.Description, err.Result)))
	}
}

func RelayNotImplemented(c *gin.Context) {
	err := dto.OpenAIError{
		Message: "API not implemented",
		Type:    "tx_api_error",
		Param:   "",
		Code:    "api_not_implemented",
	}
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": err,
	})
}

func RelayNotFound(c *gin.Context) {
	err := dto.OpenAIError{
		Message: fmt.Sprintf("Invalid URL (%s %s)", c.Request.Method, c.Request.URL.Path),
		Type:    "invalid_request_error",
		Param:   "",
		Code:    "",
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": err,
	})
}
