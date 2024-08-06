package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bytedance/gopkg/util/gopool"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"one-api/common"
	"one-api/dto"
	"one-api/middleware"
	"one-api/model"
	"one-api/relay"
	relaycommon "one-api/relay/common"
	"one-api/relay/constant"
	"one-api/service"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func testChannel(channel *model.Channel, testModel string) (err error, openAIErrorWithStatusCode *dto.OpenAIErrorWithStatusCode) {
	tik := time.Now()
	if channel.Type == common.ChannelTypeMidjourney {
		return errors.New("midjourney channel test is not supported"), nil
	}
	if channel.Type == common.ChannelTypeSunoAPI {
		return errors.New("suno channel test is not supported"), nil
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Method: "POST",
		URL:    &url.URL{Path: "/v1/chat/completions"},
		Body:   nil,
		Header: make(http.Header),
	}

	if testModel == "" {
		if channel.TestModel != nil && *channel.TestModel != "" {
			testModel = *channel.TestModel
		} else {
			if len(channel.GetModels()) > 0 {
				testModel = channel.GetModels()[0]
			} else {
				testModel = "gpt-3.5-turbo"
			}
		}
	} else {
		modelMapping := *channel.ModelMapping
		if modelMapping != "" && modelMapping != "{}" {
			modelMap := make(map[string]string)
			err := json.Unmarshal([]byte(modelMapping), &modelMap)
			if err != nil {
				return err, service.OpenAIErrorWrapperLocal(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
			}
			if modelMap[testModel] != "" {
				testModel = modelMap[testModel]
			}
		}
	}

	c.Request.Header.Set("Authorization", "Bearer "+channel.Key)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("channel", channel.Type)
	c.Set("base_url", channel.GetBaseURL())

	middleware.SetupContextForSelectedChannel(c, channel, testModel)

	meta := relaycommon.GenRelayInfo(c)
	apiType, _ := constant.ChannelType2APIType(channel.Type)
	adaptor := relay.GetAdaptor(apiType)
	if adaptor == nil {
		return fmt.Errorf("invalid api type: %d, adaptor is nil", apiType), nil
	}

	request := buildTestRequest()
	request.Model = testModel
	meta.UpstreamModelName = testModel
	common.SysLog(fmt.Sprintf("testing channel %d with model %s", channel.Id, testModel))

	adaptor.Init(meta)

	convertedRequest, err := adaptor.ConvertRequest(c, meta, request)
	if err != nil {
		return err, nil
	}
	jsonData, err := json.Marshal(convertedRequest)
	if err != nil {
		return err, nil
	}
	requestBody := bytes.NewBuffer(jsonData)
	c.Request.Body = io.NopCloser(requestBody)
	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		return err, nil
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		err := service.RelayErrorHandler(resp)
		return fmt.Errorf("status code %d: %s", resp.StatusCode, err.Error.Message), err
	}
	usage, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		return fmt.Errorf("%s", respErr.Error.Message), respErr
	}
	if usage == nil {
		return errors.New("usage is nil"), nil
	}
	result := w.Result()
	respBody, err := io.ReadAll(result.Body)
	if err != nil {
		return err, nil
	}
	modelPrice, usePrice := common.GetModelPrice(testModel, false)
	modelRatio := common.GetModelRatio(testModel)
	completionRatio := common.GetCompletionRatio(testModel)
	ratio := modelRatio
	quota := 0
	if !usePrice {
		quota = usage.PromptTokens + int(math.Round(float64(usage.CompletionTokens)*completionRatio))
		quota = int(math.Round(float64(quota) * ratio))
		if ratio != 0 && quota <= 0 {
			quota = 1
		}
	} else {
		quota = int(modelPrice * common.QuotaPerUnit)
	}
	tok := time.Now()
	milliseconds := tok.Sub(tik).Milliseconds()
	consumedTime := float64(milliseconds) / 1000.0
	other := service.GenerateTextOtherInfo(c, meta, modelRatio, 1, completionRatio, modelPrice)
	model.RecordConsumeLog(c, 1, channel.Id, usage.PromptTokens, usage.CompletionTokens, testModel, "模型测试", quota, "模型测试", 0, quota, int(consumedTime), false, other)
	common.SysLog(fmt.Sprintf("testing channel #%d, response: \n%s", channel.Id, string(respBody)))
	return nil, nil
}

func buildTestRequest() *dto.GeneralOpenAIRequest {
	testRequest := &dto.GeneralOpenAIRequest{
		Model:     "", // this will be set later
		MaxTokens: 1,
		Stream:    false,
	}
	content, _ := json.Marshal("hi")
	testMessage := dto.Message{
		Role:    "user",
		Content: content,
	}
	testRequest.Messages = append(testRequest.Messages, testMessage)
	return testRequest
}

func TestChannel(c *gin.Context) {
	channelId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	channel, err := model.GetChannelById(channelId, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	testModel := c.Query("model")
	tik := time.Now()
	err, _ = testChannel(channel, testModel)
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
	var disableThreshold = int64(common.ChannelDisableThreshold * 1000)
	if disableThreshold == 0 {
		disableThreshold = 10000000 // a impossible value
	}
	gopool.Go(func() {
		for _, channel := range channels {
			isChannelEnabled := channel.Status == common.ChannelStatusEnabled
			tik := time.Now()
			err, openaiWithStatusErr := testChannel(channel, "")
			tok := time.Now()
			milliseconds := tok.Sub(tik).Milliseconds()

			ban := false
			if milliseconds > disableThreshold {
				err = errors.New(fmt.Sprintf("响应时间 %.2fs 超过阈值 %.2fs", float64(milliseconds)/1000.0, float64(disableThreshold)/1000.0))
				ban = true
			}

			// request error disables the channel
			if openaiWithStatusErr != nil {
				oaiErr := openaiWithStatusErr.Error
				err = errors.New(fmt.Sprintf("type %s, httpCode %d, code %v, message %s", oaiErr.Type, openaiWithStatusErr.StatusCode, oaiErr.Code, oaiErr.Message))
				ban = service.ShouldDisableChannel(channel.Type, openaiWithStatusErr)
			}

			// parse *int to bool
			if !channel.GetAutoBan() {
				ban = false
			}

			// disable channel
			if ban && isChannelEnabled {
				service.DisableChannel(channel.Id, channel.Name, err.Error())
			}

			// enable channel
			if !isChannelEnabled && service.ShouldEnableChannel(err, openaiWithStatusErr, channel.Status) {
				service.EnableChannel(channel.Id, channel.Name)
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
	})
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
