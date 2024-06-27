package constant

import "one-api/common"

var ServerAddress = "http://localhost:3000"
var WorkerUrl = ""
var WorkerValidKey = ""

var StreamingTimeout = common.GetOrDefault("STREAMING_TIMEOUT", 30)

func EnableWorker() bool {
	return WorkerUrl != ""
}
