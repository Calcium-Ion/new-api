package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func testChannel(channel *model.Channel, request ChatRequest) (err error, openaiErr *OpenAIError) {
	common.SysLog(fmt.Sprintf("testing channel %d with model %s", channel.Id, request.Model))
	switch channel.Type {
	case common.ChannelTypePaLM:
		fallthrough
	case common.ChannelTypeAnthropic:
		fallthrough
	case common.ChannelTypeBaidu:
		fallthrough
	case common.ChannelTypeZhipu:
		fallthrough
	case common.ChannelTypeAli:
		fallthrough
	case common.ChannelType360:
		fallthrough
	case common.ChannelTypeGemini:
		fallthrough
	case common.ChannelTypeXunfei:
		return errors.New("该渠道类型当前版本不支持测试，请手动测试"), nil
	case common.ChannelTypeAzure:
		if request.Model == "" {
			request.Model = "gpt-35-turbo"
		}
		defer func() {
			if err != nil {
				err = errors.New("请确保已在 Azure 上创建了 gpt-35-turbo 模型，并且 apiVersion 已正确填写！")
			}
		}()
	default:
		if request.Model == "" {
			request.Model = "gpt-3.5-turbo"
		}
	}
	baseUrl := common.ChannelBaseURLs[channel.Type]
	if channel.GetBaseURL() != "" {
		baseUrl = channel.GetBaseURL()
	}
	requestURL := getFullRequestURL(baseUrl, "/v1/chat/completions", channel.Type)

	if channel.Type == common.ChannelTypeAzure {
		requestURL = getFullRequestURL(channel.GetBaseURL(), fmt.Sprintf("/openai/deployments/%s/chat/completions?api-version=2023-03-15-preview", request.Model), channel.Type)
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return err, nil
	}
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err, nil
	}
	if channel.Type == common.ChannelTypeAzure {
		req.Header.Set("api-key", channel.Key)
	} else {
		req.Header.Set("Authorization", "Bearer "+channel.Key)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err, nil
	}
	defer resp.Body.Close()
	var response TextResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err, nil
	}
	if response.Usage.CompletionTokens == 0 {
		if response.Error.Message == "" {
			response.Error.Message = "补全 tokens 非预期返回 0"
		}
		return errors.New(fmt.Sprintf("type %s, code %v, message %s", response.Error.Type, response.Error.Code, response.Error.Message)), &response.Error
	}
	return nil, nil
}

func buildTestRequest() *ChatRequest {
	testRequest := &ChatRequest{
		Model:     "", // this will be set later
		MaxTokens: 1,
	}
	content, _ := json.Marshal("hi")
	testMessage := Message{
		Role:    "user",
		Content: content,
	}
	testRequest.Messages = append(testRequest.Messages, testMessage)
	return testRequest
}

func TestChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	testModel := c.Query("model")
	channel, err := model.GetChannelById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	testRequest := buildTestRequest()
	if testModel != "" {
		testRequest.Model = testModel
	}
	tik := time.Now()
	err, _ = testChannel(channel, *testRequest)
	tok := time.Now()
	milliseconds := tok.Sub(tik).Milliseconds()
	go channel.UpdateResponseTime(milliseconds)
	consumedTime := float64(milliseconds) / 1000.0
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
			"time":    consumedTime,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"time":    consumedTime,
	})
	return
}

var testAllChannelsLock sync.Mutex
var testAllChannelsRunning bool = false

// disable & notify
func disableChannel(channelId int, channelName string, reason string) {
	model.UpdateChannelStatusById(channelId, common.ChannelStatusAutoDisabled)
	subject := fmt.Sprintf("通道「%s」（#%d）已被禁用", channelName, channelId)
	content := fmt.Sprintf("通道「%s」（#%d）已被禁用，原因：%s", channelName, channelId, reason)
	notifyRootUser(subject, content)
}

func enableChannel(channelId int, channelName string) {
	model.UpdateChannelStatusById(channelId, common.ChannelStatusEnabled)
	subject := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
	content := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
	notifyRootUser(subject, content)
}

func notifyRootUser(subject string, content string) {
	if common.RootUserEmail == "" {
		common.RootUserEmail = model.GetRootUserEmail()
	}
	err := common.SendEmail(subject, common.RootUserEmail, content)
	if err != nil {
		common.SysError(fmt.Sprintf("failed to send email: %s", err.Error()))
	}
}

func testAllChannels(notify bool) error {
	if common.RootUserEmail == "" {
		common.RootUserEmail = model.GetRootUserEmail()
	}
	testAllChannelsLock.Lock()
	if testAllChannelsRunning {
		testAllChannelsLock.Unlock()
		return errors.New("测试已在运行中")
	}
	testAllChannelsRunning = true
	testAllChannelsLock.Unlock()
	channels, err := model.GetAllChannels(0, 0, true, false)
	if err != nil {
		return err
	}
	testRequest := buildTestRequest()
	var disableThreshold = int64(common.ChannelDisableThreshold * 1000)
	if disableThreshold == 0 {
		disableThreshold = 10000000 // a impossible value
	}
	go func() {
		for _, channel := range channels {
			isChannelEnabled := channel.Status == common.ChannelStatusEnabled
			tik := time.Now()
			err, openaiErr := testChannel(channel, *testRequest)
			tok := time.Now()
			milliseconds := tok.Sub(tik).Milliseconds()

			ban := false
			if milliseconds > disableThreshold {
				err = errors.New(fmt.Sprintf("响应时间 %.2fs 超过阈值 %.2fs", float64(milliseconds)/1000.0, float64(disableThreshold)/1000.0))
				ban = true
			}
			if openaiErr != nil {
				err = errors.New(fmt.Sprintf("type %s, code %v, message %s", openaiErr.Type, openaiErr.Code, openaiErr.Message))
				ban = true
			}
			// parse *int to bool
			if channel.AutoBan != nil && *channel.AutoBan == 0 {
				ban = false
			}
			if isChannelEnabled && shouldDisableChannel(openaiErr, -1) && ban {
				disableChannel(channel.Id, channel.Name, err.Error())
			}
			if !isChannelEnabled && shouldEnableChannel(err, openaiErr) {
				enableChannel(channel.Id, channel.Name)
			}
			channel.UpdateResponseTime(milliseconds)
			time.Sleep(common.RequestInterval)
		}
		testAllChannelsLock.Lock()
		testAllChannelsRunning = false
		testAllChannelsLock.Unlock()
		if notify {
			err := common.SendEmail("通道测试完成", common.RootUserEmail, "通道测试完成，如果没有收到禁用通知，说明所有通道都正常")
			if err != nil {
				common.SysError(fmt.Sprintf("failed to send email: %s", err.Error()))
			}
		}
	}()
	return nil
}

func TestAllChannels(c *gin.Context) {
	err := testAllChannels(true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func AutomaticallyTestChannels(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Minute)
		common.SysLog("testing all channels")
		_ = testAllChannels(false)
		common.SysLog("channel test finished")
	}
}
