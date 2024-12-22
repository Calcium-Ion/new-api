package service

import (
	"bytes"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/setting"
	"strings"
)

func DoImageRequest(originUrl string) (resp *http.Response, err error) {
	if setting.EnableWorker() {
		common.SysLog(fmt.Sprintf("downloading image from worker: %s", originUrl))
		workerUrl := setting.WorkerUrl
		if !strings.HasSuffix(workerUrl, "/") {
			workerUrl += "/"
		}
		// post request to worker
		data := []byte(`{"url":"` + originUrl + `","key":"` + setting.WorkerValidKey + `"}`)
		return http.Post(setting.WorkerUrl, "application/json", bytes.NewBuffer(data))
	} else {
		common.SysLog(fmt.Sprintf("downloading image from origin: %s", originUrl))
		return http.Get(originUrl)
	}
}
