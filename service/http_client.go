package service

import (
	"context"
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
	"one-api/common"
	"time"
	"math/rand"
	"strings"
	"strconv"
)

var httpClient *http.Client
var impatientHTTPClient *http.Client

func init() {
	if common.RelayTimeout == 0 {
		httpClient = &http.Client{}
	} else {
		httpClient = &http.Client{
			Timeout: time.Duration(common.RelayTimeout) * time.Second,
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

// NewProxyHttpClient 创建支持代理的 HTTP 客户端
func NewProxyHttpClient(proxyURLs string) (*http.Client, error) {
	if proxyURLs == "" {
		return http.DefaultClient, nil
	}

	proxyList := strings.Split(proxyURLs, ",")
	rand.Seed(time.Now().UnixNano())

	var client *http.Client
	var err error

	for i := 0; i < len(proxyList); i++ {
		proxyTimeoutPair := strings.Fields(proxyList[rand.Intn(len(proxyList))])
		if len(proxyTimeoutPair) == 0 {
			continue
		}
		proxyURL := proxyTimeoutPair[0]
		timeout := 10000 * time.Millisecond

		if len(proxyTimeoutPair) > 1 {
			if customTimeout, parseErr := strconv.Atoi(proxyTimeoutPair[1]); parseErr == nil {
				timeout = time.Duration(customTimeout) * time.Millisecond
			} else {
				fmt.Printf("Error parsing timeout value: %v\n", parseErr)
			}
		}

		client, err = createHttpClientWithProxy(proxyURL, timeout)
		if err == nil {
			return client, nil
		}
	}

	return nil, fmt.Errorf("all proxies failed")
}

func createHttpClientWithProxy(proxyURL string, timeout time.Duration) (*http.Client, error) {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	switch parsedURL.Scheme {
	case "http", "https":
		return &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(parsedURL),
			},
			Timeout: timeout,
		}, nil

	case "socks5":
		var auth *proxy.Auth
		if parsedURL.User != nil {
			auth = &proxy.Auth{
				User:     parsedURL.User.Username(),
				Password: "",
			}
			if password, ok := parsedURL.User.Password(); ok {
				auth.Password = password
			}
		}

		dialer, err := proxy.SOCKS5("tcp", parsedURL.Host, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}

		return &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dialer.Dial(network, addr)
				},
			},
			Timeout: timeout,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %s", parsedURL.Scheme)
	}
}
