package dto

import (
	"fmt"
)

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    any    `json:"code"`
}

type OpenAIErrorWithStatusCode struct {
	Error      OpenAIError `json:"error"`
	StatusCode int         `json:"status_code"`
	LocalError bool
}

// ChannelMailInfo channel test error to mail info
type ChannelMailInfo struct {
	ChannelId   int    `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	Reason      string `json:"reason"`
}

func (m *ChannelMailInfo) ToChannelMailInfo(msg string) (subject, content string) {
	subject = fmt.Sprintf("通道「%s」（#%d）%s", m.ChannelName, m.ChannelId, msg)
	if m.Reason == "" {
		content = fmt.Sprintf("通道「%s」（#%d）%s", m.ChannelName, m.ChannelId, msg)
	} else {
		content = fmt.Sprintf("通道「%s」（#%d）%s，原因：%s", m.ChannelName, m.ChannelId, m.Reason, msg)
	}
	return
}

func NewChannelMailInfo(channelId int, channelName, reason string) *ChannelMailInfo {
	return &ChannelMailInfo{
		ChannelId:   channelId,
		ChannelName: channelName,
		Reason:      reason,
	}
}

type GeneralErrorResponse struct {
	Error    OpenAIError `json:"error"`
	Message  string      `json:"message"`
	Msg      string      `json:"msg"`
	Err      string      `json:"err"`
	ErrorMsg string      `json:"error_msg"`
	Header   struct {
		Message string `json:"message"`
	} `json:"header"`
	Response struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	} `json:"response"`
}

func (e GeneralErrorResponse) ToMessage() string {
	if e.Error.Message != "" {
		return e.Error.Message
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Msg != "" {
		return e.Msg
	}
	if e.Err != "" {
		return e.Err
	}
	if e.ErrorMsg != "" {
		return e.ErrorMsg
	}
	if e.Header.Message != "" {
		return e.Header.Message
	}
	if e.Response.Error.Message != "" {
		return e.Response.Error.Message
	}
	return ""
}
