package constant

import (
	"one-api/common"
)

var StreamingTimeout int
var DifyDebug bool
var MaxFileDownloadMB int
var ForceStreamOption bool
var GetMediaToken bool
var GetMediaTokenNotStream bool
var UpdateTask bool
var AzureDefaultAPIVersion string
var GeminiVisionMaxImageNum int
var NotifyLimitCount int
var NotificationLimitDurationMinute int
var GenerateDefaultToken bool

//var GeminiModelMap = map[string]string{
//	"gemini-1.0-pro": "v1",
//}

func InitEnv() {
	StreamingTimeout = common.GetEnvOrDefault("STREAMING_TIMEOUT", 60)
	DifyDebug = common.GetEnvOrDefaultBool("DIFY_DEBUG", true)
	MaxFileDownloadMB = common.GetEnvOrDefault("MAX_FILE_DOWNLOAD_MB", 20)
	// ForceStreamOption 覆盖请求参数，强制返回usage信息
	ForceStreamOption = common.GetEnvOrDefaultBool("FORCE_STREAM_OPTION", true)
	GetMediaToken = common.GetEnvOrDefaultBool("GET_MEDIA_TOKEN", true)
	GetMediaTokenNotStream = common.GetEnvOrDefaultBool("GET_MEDIA_TOKEN_NOT_STREAM", true)
	UpdateTask = common.GetEnvOrDefaultBool("UPDATE_TASK", true)
	AzureDefaultAPIVersion = common.GetEnvOrDefaultString("AZURE_DEFAULT_API_VERSION", "2024-12-01-preview")
	GeminiVisionMaxImageNum = common.GetEnvOrDefault("GEMINI_VISION_MAX_IMAGE_NUM", 16)
	NotifyLimitCount = common.GetEnvOrDefault("NOTIFY_LIMIT_COUNT", 2)
	NotificationLimitDurationMinute = common.GetEnvOrDefault("NOTIFICATION_LIMIT_DURATION_MINUTE", 10)
	// GenerateDefaultToken 是否生成初始令牌，默认关闭。
	GenerateDefaultToken = common.GetEnvOrDefaultBool("GENERATE_DEFAULT_TOKEN", false)

	//modelVersionMapStr := strings.TrimSpace(os.Getenv("GEMINI_MODEL_MAP"))
	//if modelVersionMapStr == "" {
	//	return
	//}
	//for _, pair := range strings.Split(modelVersionMapStr, ",") {
	//	parts := strings.Split(pair, ":")
	//	if len(parts) == 2 {
	//		GeminiModelMap[parts[0]] = parts[1]
	//	} else {
	//		common.SysError(fmt.Sprintf("invalid model version map: %s", pair))
	//	}
	//}
}
