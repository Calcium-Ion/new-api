package constant

var WebHookEnabled = false
var WebHookUrl = ""
var WebHookHeaders = make(map[string]string)
var WebHookDataMapStr = ""
var WebHookQueueSize = 5 // if 0, no queue
