package service

import (
	"bytes"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/setting"
	"strings"
)

func DoDownloadRequest(originUrl string) (resp *http.Response, err error) {
	if setting.EnableWorker() {
		common.SysLog(fmt.Sprintf("downloading file from worker: %s", originUrl))
		if !strings.HasPrefix(originUrl, "https") {
			return nil, fmt.Errorf("only support https url")
		}
		workerUrl := setting.WorkerUrl
		if !strings.HasSuffix(workerUrl, "/") {
			workerUrl += "/"
		}
		// post request to worker
		data := []byte(`{"url":"` + originUrl + `","key":"` + setting.WorkerValidKey + `"}`)
		return http.Post(setting.WorkerUrl, "application/json", bytes.NewBuffer(data))
	} else {
		common.SysLog(fmt.Sprintf("downloading from origin: %s", originUrl))
		return http.Get(originUrl)
	}
}
