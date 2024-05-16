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
	model.UpdateChannelStatusById(channelId, common.ChannelStatusAutoDisabled)
	subject := fmt.Sprintf("通道「%s」（#%d）已被禁用", channelName, channelId)
	content := fmt.Sprintf("通道「%s」（#%d）已被禁用，原因：%s", channelName, channelId, reason)
	notifyRootUser(subject, content)
}

func EnableChannel(channelId int, channelName string) {
	model.UpdateChannelStatusById(channelId, common.ChannelStatusEnabled)
	subject := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
	content := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
	notifyRootUser(subject, content)
}

func ShouldDisableChannel(err *relaymodel.OpenAIError, statusCode int) bool {
	if !common.AutomaticDisableChannelEnabled {
		return false
	}
	if err == nil {
		return false
	}
	if statusCode == http.StatusUnauthorized {
		return true
	}
	switch err.Code {
	case "invalid_api_key":
		return true
	case "account_deactivated":
		return true
	case "billing_not_active":
		return true
	}
	switch err.Type {
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
	if strings.HasPrefix(err.Message, "Your credit balance is too low") { // anthropic
		return true
	} else if strings.HasPrefix(err.Message, "This organization has been disabled.") {
		return true
	} else if strings.HasPrefix(err.Message, "You exceeded your current quota") {
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

// CalculateChannelErrorRate calculates the error rate for a given channel based on recent error logs.
func CalculateChannelErrorRate(channelId int) (float64, error) {
	// Fetch error logs for the channel within a specific timeframe.
	errorLogs, err := model.FetchChannelErrorLogs(channelId, common.ErrorLogTimeframe)
	if err != nil {
		return 0, err
	}

	// Calculate error rate.
	totalRequests := len(errorLogs)
	if totalRequests == 0 {
		return 0, nil
	}

	errorCount := 0
	for _, log := range errorLogs {
		if log.Type == model.LogError {
			errorCount++
		}
	}

	errorRate := float64(errorCount) / float64(totalRequests)
	return errorRate, nil
}

// UpdateChannelErrorRate updates the error rate of a channel in the database.
func UpdateChannelErrorRate(channelId int, errorRate float64) error {
	return model.UpdateChannelErrorRate(channelId, errorRate)
}
