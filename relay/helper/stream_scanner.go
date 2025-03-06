package helper

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"one-api/common"
	"one-api/constant"
	relaycommon "one-api/relay/common"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func StreamScannerHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo, dataHandler func(data string) bool) {

	if resp == nil {
		return
	}

	defer resp.Body.Close()

	streamingTimeout := time.Duration(constant.StreamingTimeout) * time.Second
	if strings.HasPrefix(info.UpstreamModelName, "o1") || strings.HasPrefix(info.UpstreamModelName, "o3") {
		// twice timeout for thinking model
		streamingTimeout *= 2
	}

	var (
		stopChan = make(chan bool, 2)
		scanner  = bufio.NewScanner(resp.Body)
		ticker   = time.NewTicker(streamingTimeout)
	)

	defer func() {
		ticker.Stop()
		close(stopChan)
	}()

	scanner.Split(bufio.ScanLines)
	SetEventStreamHeaders(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = context.WithValue(ctx, "stop_chan", stopChan)
	common.RelayCtxGo(ctx, func() {
		for scanner.Scan() {
			ticker.Reset(streamingTimeout)
			data := scanner.Text()
			if common.DebugEnabled {
				println(data)
			}

			if len(data) < 6 {
				continue
			}
			if data[:5] != "data:" && data[:6] != "[DONE]" {
				continue
			}
			data = data[5:]
			data = strings.TrimLeft(data, " ")
			data = strings.TrimSuffix(data, "\"")
			if !strings.HasPrefix(data, "[DONE]") {
				info.SetFirstResponseTime()
				success := dataHandler(data)
				if !success {
					break
				}
			}
		}

		if err := scanner.Err(); err != nil {
			if err != io.EOF {
				common.LogError(c, "scanner error: "+err.Error())
			}
		}

		common.SafeSendBool(stopChan, true)
	})

	select {
	case <-ticker.C:
		// 超时处理逻辑
		common.LogError(c, "streaming timeout")
	case <-stopChan:
		// 正常结束
	}
}
