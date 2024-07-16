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

func MidjourneyErrorWrapper(code int, desc string) *dto.MidjourneyResponse {
	return &dto.MidjourneyResponse{
		Code:        code,
		Description: desc,
	}
}

func MidjourneyErrorWithStatusCodeWrapper(code int, desc string, statusCode int) *dto.MidjourneyResponseWithStatusCode {
	return &dto.MidjourneyResponseWithStatusCode{
		StatusCode: statusCode,
		Response:   *MidjourneyErrorWrapper(code, desc),
	}
}

// OpenAIErrorWrapper wraps an error into an OpenAIErrorWithStatusCode
func OpenAIErrorWrapper(err error, code string, statusCode int) *dto.OpenAIErrorWithStatusCode {
	text := err.Error()
	// 定义一个正则表达式匹配URL
	if strings.Contains(text, "Post") || strings.Contains(text, "dial") {
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

func OpenAIErrorWrapperLocal(err error, code string, statusCode int) *dto.OpenAIErrorWithStatusCode {
	openaiErr := OpenAIErrorWrapper(err, code, statusCode)
	openaiErr.LocalError = true
	return openaiErr
}

func RelayErrorHandler(resp *http.Response) (errWithStatusCode *dto.OpenAIErrorWithStatusCode) {
	errWithStatusCode = &dto.OpenAIErrorWithStatusCode{
		StatusCode: resp.StatusCode,
		Error: dto.OpenAIError{
			Type:  "upstream_error",
			Code:  "bad_response_status_code",
			Param: strconv.Itoa(resp.StatusCode),
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

func ResetStatusCode(openaiErr *dto.OpenAIErrorWithStatusCode, statusCodeMappingStr string) {
	if statusCodeMappingStr == "" || statusCodeMappingStr == "{}" {
		return
	}
	statusCodeMapping := make(map[string]string)
	err := json.Unmarshal([]byte(statusCodeMappingStr), &statusCodeMapping)
	if err != nil {
		return
	}
	if openaiErr.StatusCode == http.StatusOK {
		return
	}
	codeStr := strconv.Itoa(openaiErr.StatusCode)
	if _, ok := statusCodeMapping[codeStr]; ok {
		intCode, _ := strconv.Atoi(statusCodeMapping[codeStr])
		openaiErr.StatusCode = intCode
	}
}

func TaskErrorWrapperLocal(err error, code string, statusCode int) *dto.TaskError {
	openaiErr := TaskErrorWrapper(err, code, statusCode)
	openaiErr.LocalError = true
	return openaiErr
}

func TaskErrorWrapper(err error, code string, statusCode int) *dto.TaskError {
	text := err.Error()

	// 定义一个正则表达式匹配URL
	if strings.Contains(text, "Post") || strings.Contains(text, "dial") {
		common.SysLog(fmt.Sprintf("error: %s", text))
		text = "请求上游地址失败"
	}
	//避免暴露内部错误

	taskError := &dto.TaskError{
		Code:       code,
		Message:    text,
		StatusCode: statusCode,
		Error:      err,
	}

	return taskError
}
