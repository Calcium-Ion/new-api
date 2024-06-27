package constant

import (
	"one-api/common"
)

var StreamingTimeout = common.GetEnvOrDefault("STREAMING_TIMEOUT", 30)
