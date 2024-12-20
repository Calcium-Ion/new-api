package dto

import "fmt"

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    any    `json:"code"`
}

func (e OpenAIError) ToMessageString() string {
	return fmt.Sprintf("type %s, code %v, message %s", e.Type, e.Code, e.Message)
}

type OpenAIErrorWithStatusCode struct {
	Error      OpenAIError `json:"error"`
	StatusCode int         `json:"status_code"`
	LocalError bool
}

func (e OpenAIErrorWithStatusCode) ToMessageString() string {
	return fmt.Sprintf("status code %d, %s", e.StatusCode, e.Error.ToMessageString())
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
