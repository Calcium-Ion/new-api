package service

import (
	"net/http"
	"one-api/common"
	"time"
)

var httpClient *http.Client
var impatientHTTPClient *http.Client

func init() {
	if common.RelayTimeout == 0 {
		httpClient = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// 自定义重定向策略，例如限制重定向次数或修改请求头
				return nil // 允许重定向
			},
		}
	} else {
		httpClient = &http.Client{
			Timeout: time.Duration(common.RelayTimeout) * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// 自定义重定向策略，例如限制重定向次数或修改请求头
				return nil // 允许重定向
			},
		}
	}

	impatientHTTPClient = &http.Client{
		Timeout: 5 * time.Second,
	}
}

func GetHttpClient() *http.Client {
	return httpClient
}

func GetImpatientHttpClient() *http.Client {
	return impatientHTTPClient
}
