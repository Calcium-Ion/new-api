package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	"strconv"
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
		Error:      openAIError,
		StatusCode: statusCode,
	}
}

func RelayErrorHandler(resp *http.Response) (errWithStatusCode *dto.OpenAIErrorWithStatusCode) {
	errWithStatusCode = &dto.OpenAIErrorWithStatusCode{
		StatusCode: resp.StatusCode,
		Error: dto.OpenAIError{
			Message: "",
			Type:    "upstream_error",
			Code:    "bad_response_status_code",
			Param:   strconv.Itoa(resp.StatusCode),
		},
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = resp.Body.Close()
	if err != nil {
		return
	}
	var errResponse dto.GeneralErrorResponse
	err = json.Unmarshal(responseBody, &errResponse)
	if err != nil {
		return
	}
	if errResponse.Error.Message != "" {
		// OpenAI format error, so we override the default one
		errWithStatusCode.Error = errResponse.Error
	} else {
		errWithStatusCode.Error.Message = errResponse.ToMessage()
	}
	if errWithStatusCode.Error.Message == "" {
		errWithStatusCode.Error.Message = fmt.Sprintf("bad response status code %d", resp.StatusCode)
	}
	return
}
