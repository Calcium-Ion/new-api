package constant

import (
	"one-api/common"
)

var StreamingTimeout = common.GetEnvOrDefault("STREAMING_TIMEOUT", 30)
var DifyDebug = common.GetEnvOrDefaultBool("DIFY_DEBUG", true)

// ForceStreamOption 覆盖请求参数，强制返回usage信息
var ForceStreamOption = common.GetEnvOrDefaultBool("FORCE_STREAM_OPTION", true)

var GetMediaToken = common.GetEnvOrDefaultBool("GET_MEDIA_TOKEN", true)

var GetMediaTokenNotStream = common.GetEnvOrDefaultBool("GET_MEDIA_TOKEN_NOT_STREAM", true)

var UpdateTask = common.GetEnvOrDefaultBool("UPDATE_TASK", true)
