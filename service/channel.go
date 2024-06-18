package service

import (
	"fmt"
	"net/http"
	"one-api/common"
	relaymodel "one-api/dto"
	"one-api/model"
	"strings"
)

// disable & notify
func DisableChannel(channelId int, channelName string, reason string) {
	model.UpdateChannelStatusById(channelId, common.ChannelStatusAutoDisabled, reason)
	subject := fmt.Sprintf("通道「%s」（#%d）已被禁用", channelName, channelId)
	content := fmt.Sprintf("通道「%s」（#%d）已被禁用，原因：%s", channelName, channelId, reason)
	notifyRootUser(subject, content)
}

func EnableChannel(channelId int, channelName string) {
	model.UpdateChannelStatusById(channelId, common.ChannelStatusEnabled, "")
	subject := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
	content := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
	notifyRootUser(subject, content)
}

func ShouldDisableChannel(err *relaymodel.OpenAIErrorWithStatusCode) bool {
	if !common.AutomaticDisableChannelEnabled {
		return false
	}
	if err == nil {
		return false
	}
	if err.LocalError {
		return false
	}
	if err.StatusCode == http.StatusUnauthorized || err.StatusCode == http.StatusForbidden {
		return true
	}
	switch err.Error.Code {
	case "invalid_api_key":
		return true
	case "account_deactivated":
		return true
	case "billing_not_active":
		return true
	}
	switch err.Error.Type {
	case "insufficient_quota":
		return true
	// https://docs.anthropic.com/claude/reference/errors
	case "authentication_error":
		return true
	case "permission_error":
		return true
	case "forbidden":
		return true
	}
	if strings.HasPrefix(err.Error.Message, "Your credit balance is too low") { // anthropic
		return true
	} else if strings.HasPrefix(err.Error.Message, "This organization has been disabled.") {
		return true
	} else if strings.HasPrefix(err.Error.Message, "You exceeded your current quota") {
		return true
	} else if strings.HasPrefix(err.Error.Message, "Permission denied") {
		return true
	}
	return false
}

func ShouldEnableChannel(err error, openAIErr *relaymodel.OpenAIError, status int) bool {
	if !common.AutomaticEnableChannelEnabled {
		return false
	}
	if err != nil {
		return false
	}
	if openAIErr != nil {
		return false
	}
	if status != common.ChannelStatusAutoDisabled {
		return false
	}
	return true
}
