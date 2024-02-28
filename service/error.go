package service

import (
	"fmt"
	"one-api/common"
	"one-api/dto"
	"strings"
)

// OpenAIErrorWrapper wraps an error into an OpenAIErrorWithStatusCode
func OpenAIErrorWrapper(err error, code string, statusCode int) *dto.OpenAIErrorWithStatusCode {
	text := err.Error()
	// 定义一个正则表达式匹配URL
	if strings.Contains(text, "Post") {
		common.SysLog(fmt.Sprintf("error: %s", text))
		text = "请求上游地址失败"
	}
	//避免暴露内部错误

	openAIError := dto.OpenAIError{
		Message: text,
		Type:    "new_api_error",
		Code:    code,
	}
	return &dto.OpenAIErrorWithStatusCode{
		OpenAIError: openAIError,
		StatusCode:  statusCode,
	}
}
