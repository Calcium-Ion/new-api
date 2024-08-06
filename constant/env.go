package constant

import (
	"fmt"
	"one-api/common"
	"os"
	"strings"
)

var StreamingTimeout = common.GetEnvOrDefault("STREAMING_TIMEOUT", 30)
var DifyDebug = common.GetEnvOrDefaultBool("DIFY_DEBUG", true)

// ForceStreamOption 覆盖请求参数，强制返回usage信息
var ForceStreamOption = common.GetEnvOrDefaultBool("FORCE_STREAM_OPTION", true)

var GetMediaToken = common.GetEnvOrDefaultBool("GET_MEDIA_TOKEN", true)

var GetMediaTokenNotStream = common.GetEnvOrDefaultBool("GET_MEDIA_TOKEN_NOT_STREAM", true)

var UpdateTask = common.GetEnvOrDefaultBool("UPDATE_TASK", true)

var GeminiModelMap = map[string]string{
	"gemini-1.5-pro-latest":   "v1beta",
	"gemini-1.5-pro-001":      "v1beta",
	"gemini-1.5-pro":          "v1beta",
	"gemini-1.5-pro-exp-0801": "v1beta",
	"gemini-1.5-flash-latest": "v1beta",
	"gemini-1.5-flash-001":    "v1beta",
	"gemini-1.5-flash":        "v1beta",
	"gemini-ultra":            "v1beta",
}

func InitEnv() {
	modelVersionMapStr := strings.TrimSpace(os.Getenv("GEMINI_MODEL_MAP"))
	if modelVersionMapStr == "" {
		return
	}
	for _, pair := range strings.Split(modelVersionMapStr, ",") {
		parts := strings.Split(pair, ":")
		if len(parts) == 2 {
			GeminiModelMap[parts[0]] = parts[1]
		} else {
			common.SysError(fmt.Sprintf("invalid model version map: %s", pair))
		}
	}
}
