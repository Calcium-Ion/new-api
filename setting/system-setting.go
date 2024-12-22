package setting

var ServerAddress = "http://localhost:3000"
var WorkerUrl = ""
var WorkerValidKey = ""

func EnableWorker() bool {
	return WorkerUrl != ""
}
