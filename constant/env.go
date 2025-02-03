package constant

import (
	"fmt"
	"one-api/common"
	"os"
	"strings"
)

var StreamingTimeout = common.GetEnvOrDefault("STREAMING_TIMEOUT", 60)
var DifyDebug = common.GetEnvOrDefaultBool("DIFY_DEBUG", true)

var MaxFileDownloadMB = common.GetEnvOrDefault("MAX_FILE_DOWNLOAD_MB", 20)

// ForceStreamOption 覆盖请求参数，强制返回usage信息
var ForceStreamOption = common.GetEnvOrDefaultBool("FORCE_STREAM_OPTION", true)

var GetMediaToken = common.GetEnvOrDefaultBool("GET_MEDIA_TOKEN", true)

var GetMediaTokenNotStream = common.GetEnvOrDefaultBool("GET_MEDIA_TOKEN_NOT_STREAM", true)

var UpdateTask = common.GetEnvOrDefaultBool("UPDATE_TASK", true)

var AzureDefaultAPIVersion = common.GetEnvOrDefaultString("AZURE_DEFAULT_API_VERSION", "2024-12-01-preview")

var GeminiModelMap = map[string]string{
	"gemini-1.0-pro": "v1",
}

var GeminiVisionMaxImageNum = common.GetEnvOrDefault("GEMINI_VISION_MAX_IMAGE_NUM", 16)

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

// 是否生成初始令牌，默认关闭。
var GenerateDefaultToken = common.GetEnvOrDefaultBool("GENERATE_DEFAULT_TOKEN", false)
