package service

import (
	"bytes"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"strings"
)

func DoImageRequest(originUrl string) (resp *http.Response, err error) {
	if constant.EnableWorker() {
		common.SysLog(fmt.Sprintf("downloading image from worker: %s", originUrl))
		workerUrl := constant.WorkerUrl
		if !strings.HasSuffix(workerUrl, "/") {
			workerUrl += "/"
		}
		// post request to worker
		data := []byte(`{"url":"` + originUrl + `","key":"` + constant.WorkerValidKey + `"}`)
		return http.Post(constant.WorkerUrl, "application/json", bytes.NewBuffer(data))
	} else {
		common.SysLog(fmt.Sprintf("downloading image from origin: %s", originUrl))
		return http.Get(originUrl)
	}
}
