package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/relay/helper"
	"one-api/service"
	"os"
	"strings"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

func sendStreamData(c *gin.Context, info *relaycommon.RelayInfo, data string, forceFormat bool, thinkToContent bool) error {
	if data == "" {
		return nil
	}

	if !forceFormat && !thinkToContent {
		return helper.StringData(c, data)
	}

	var lastStreamResponse dto.ChatCompletionsStreamResponse
	if err := common.DecodeJsonStr(data, &lastStreamResponse); err != nil {
		return err
	}

	if !thinkToContent {
		return helper.ObjectData(c, lastStreamResponse)
	}

	hasThinkingContent := false
	hasContent := false
	var thinkingContent strings.Builder
	for _, choice := range lastStreamResponse.Choices {
		if len(choice.Delta.GetReasoningContent()) > 0 {
			hasThinkingContent = true
			thinkingContent.WriteString(choice.Delta.GetReasoningContent())
		}
		if len(choice.Delta.GetContentString()) > 0 {
			hasContent = true
		}
	}

	// Handle think to content conversion
	if info.ThinkingContentInfo.IsFirstThinkingContent {
		if hasThinkingContent {
			response := lastStreamResponse.Copy()
			for i := range response.Choices {
				// send `think` tag with thinking content
				response.Choices[i].Delta.SetContentString("<think>\n" + thinkingContent.String())
				response.Choices[i].Delta.ReasoningContent = nil
				response.Choices[i].Delta.Reasoning = nil
			}
			info.ThinkingContentInfo.IsFirstThinkingContent = false
			info.ThinkingContentInfo.HasSentThinkingContent = true
			return helper.ObjectData(c, response)
		}
	}

	if lastStreamResponse.Choices == nil || len(lastStreamResponse.Choices) == 0 {
		return helper.ObjectData(c, lastStreamResponse)
	}

	// Process each choice
	for i, choice := range lastStreamResponse.Choices {
		// Handle transition from thinking to content
		// only send `</think>` tag when previous thinking content has been sent
		if hasContent && !info.ThinkingContentInfo.SendLastThinkingContent && info.ThinkingContentInfo.HasSentThinkingContent {
			response := lastStreamResponse.Copy()
			for j := range response.Choices {
				response.Choices[j].Delta.SetContentString("\n</think>\n")
				response.Choices[j].Delta.ReasoningContent = nil
				response.Choices[j].Delta.Reasoning = nil
			}
			info.ThinkingContentInfo.SendLastThinkingContent = true
			helper.ObjectData(c, response)
		}

		// Convert reasoning content to regular content if any
		if len(choice.Delta.GetReasoningContent()) > 0 {
			lastStreamResponse.Choices[i].Delta.SetContentString(choice.Delta.GetReasoningContent())
			lastStreamResponse.Choices[i].Delta.ReasoningContent = nil
			lastStreamResponse.Choices[i].Delta.Reasoning = nil
		} else if !hasThinkingContent && !hasContent {
			// flush thinking content
			lastStreamResponse.Choices[i].Delta.ReasoningContent = nil
			lastStreamResponse.Choices[i].Delta.Reasoning = nil
		}
	}

	return helper.ObjectData(c, lastStreamResponse)
}

func OaiStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	if resp == nil || resp.Body == nil {
		common.LogError(c, "invalid response or response body")
		return service.OpenAIErrorWrapper(fmt.Errorf("invalid response"), "invalid_response", http.StatusInternalServerError), nil
	}

	containStreamUsage := false
	var responseId string
	var createAt int64 = 0
	var systemFingerprint string
	model := info.UpstreamModelName

	var responseTextBuilder strings.Builder
	var usage = &dto.Usage{}
	var streamItems []string // store stream items
	var forceFormat bool
	var thinkToContent bool

	if forceFmt, ok := info.ChannelSetting[constant.ForceFormat].(bool); ok {
		forceFormat = forceFmt
	}

	if think2Content, ok := info.ChannelSetting[constant.ChannelSettingThinkingToContent].(bool); ok {
		thinkToContent = think2Content
	}

	toolCount := 0

	var (
		lastStreamData string
	)

	helper.StreamScannerHandler(c, resp, info, func(data string) bool {
		if lastStreamData != "" {
			err := handleStreamFormat(c, info, lastStreamData, forceFormat, thinkToContent)
			if err != nil {
				common.SysError("error handling stream format: " + err.Error())
			}
			info.SetFirstResponseTime()
		}
		lastStreamData = data
		streamItems = append(streamItems, data)
		return true
	})

	shouldSendLastResp := true
	var lastStreamResponse dto.ChatCompletionsStreamResponse
	err := common.DecodeJsonStr(lastStreamData, &lastStreamResponse)
	if err == nil {
		responseId = lastStreamResponse.Id
		createAt = lastStreamResponse.Created
		systemFingerprint = lastStreamResponse.GetSystemFingerprint()
		model = lastStreamResponse.Model
		if service.ValidUsage(lastStreamResponse.Usage) {
			containStreamUsage = true
			usage = lastStreamResponse.Usage
			if !info.ShouldIncludeUsage {
				shouldSendLastResp = false
			}
		}
		for _, choice := range lastStreamResponse.Choices {
			if choice.FinishReason != nil {
				shouldSendLastResp = true
			}
		}
	}
	if shouldSendLastResp {
		sendStreamData(c, info, lastStreamData, forceFormat, thinkToContent)
	}

	// 处理token计算
	if err := processTokens(info.RelayMode, streamItems, &responseTextBuilder, &toolCount); err != nil {
		common.SysError("error processing tokens: " + err.Error())
	}

	if !containStreamUsage {
		usage, _ = service.ResponseText2Usage(responseTextBuilder.String(), info.UpstreamModelName, info.PromptTokens)
		usage.CompletionTokens += toolCount * 7
	} else {
		if info.ChannelType == common.ChannelTypeDeepSeek {
			if usage.PromptCacheHitTokens != 0 {
				usage.PromptTokensDetails.CachedTokens = usage.PromptCacheHitTokens
			}
		}
	}

	handleFinalResponse(c, info, lastStreamData, responseId, createAt, model, systemFingerprint, usage, containStreamUsage)

	return nil, usage
}

func OpenaiHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var simpleResponse dto.OpenAITextResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = common.DecodeJson(responseBody, &simpleResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if simpleResponse.Error != nil && simpleResponse.Error.Type != "" {
		return &dto.OpenAIErrorWithStatusCode{
			Error:      *simpleResponse.Error,
			StatusCode: resp.StatusCode,
		}, nil
	}

	switch info.RelayFormat {
	case relaycommon.RelayFormatOpenAI:
		break
	case relaycommon.RelayFormatClaude:
		claudeResp := service.ResponseOpenAI2Claude(&simpleResponse, info)
		claudeRespStr, err := json.Marshal(claudeResp)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
		}
		responseBody = claudeRespStr
	}

	// Reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the httpClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		//return service.OpenAIErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
		common.SysError("error copying response body: " + err.Error())
	}
	resp.Body.Close()
	if simpleResponse.Usage.TotalTokens == 0 || (simpleResponse.Usage.PromptTokens == 0 && simpleResponse.Usage.CompletionTokens == 0) {
		completionTokens := 0
		for _, choice := range simpleResponse.Choices {
			ctkm, _ := service.CountTextToken(choice.Message.StringContent()+choice.Message.ReasoningContent+choice.Message.Reasoning, info.UpstreamModelName)
			completionTokens += ctkm
		}
		simpleResponse.Usage = dto.Usage{
			PromptTokens:     info.PromptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      info.PromptTokens + completionTokens,
		}
	}
	return nil, &simpleResponse.Usage
}

func OpenaiTTSHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	// Reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the httpClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	usage := &dto.Usage{}
	usage.PromptTokens = info.PromptTokens
	usage.TotalTokens = info.PromptTokens
	return nil, usage
}

func OpenaiSTTHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo, responseFormat string) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	// count tokens by audio file duration
	audioTokens, err := countAudioTokens(c)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "count_audio_tokens_failed", http.StatusInternalServerError), nil
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	// Reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the httpClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	resp.Body.Close()

	usage := &dto.Usage{}
	usage.PromptTokens = audioTokens
	usage.CompletionTokens = 0
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	return nil, usage
}

func countAudioTokens(c *gin.Context) (int, error) {
	body, err := common.GetRequestBody(c)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	var reqBody struct {
		File *multipart.FileHeader `form:"file" binding:"required"`
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if err = c.ShouldBind(&reqBody); err != nil {
		return 0, errors.WithStack(err)
	}

	reqFp, err := reqBody.File.Open()
	if err != nil {
		return 0, errors.WithStack(err)
	}

	tmpFp, err := os.CreateTemp("", "audio-*")
	if err != nil {
		return 0, errors.WithStack(err)
	}
	defer os.Remove(tmpFp.Name())

	_, err = io.Copy(tmpFp, reqFp)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	if err = tmpFp.Close(); err != nil {
		return 0, errors.WithStack(err)
	}

	duration, err := common.GetAudioDuration(c.Request.Context(), tmpFp.Name())
	if err != nil {
		return 0, errors.WithStack(err)
	}

	return int(math.Round(math.Ceil(duration) / 60.0 * 1000)), nil // 1 minute 相当于 1k tokens
}

func OpenaiRealtimeHandler(c *gin.Context, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.RealtimeUsage) {
	if info == nil || info.ClientWs == nil || info.TargetWs == nil {
		return service.OpenAIErrorWrapper(fmt.Errorf("invalid websocket connection"), "invalid_connection", http.StatusBadRequest), nil
	}

	info.IsStream = true
	clientConn := info.ClientWs
	targetConn := info.TargetWs

	clientClosed := make(chan struct{})
	targetClosed := make(chan struct{})
	sendChan := make(chan []byte, 100)
	receiveChan := make(chan []byte, 100)
	errChan := make(chan error, 2)

	usage := &dto.RealtimeUsage{}
	localUsage := &dto.RealtimeUsage{}
	sumUsage := &dto.RealtimeUsage{}

	gopool.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic in client reader: %v", r)
			}
		}()
		for {
			select {
			case <-c.Done():
				return
			default:
				_, message, err := clientConn.ReadMessage()
				if err != nil {
					if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
						errChan <- fmt.Errorf("error reading from client: %v", err)
					}
					close(clientClosed)
					return
				}

				realtimeEvent := &dto.RealtimeEvent{}
				err = json.Unmarshal(message, realtimeEvent)
				if err != nil {
					errChan <- fmt.Errorf("error unmarshalling message: %v", err)
					return
				}

				if realtimeEvent.Type == dto.RealtimeEventTypeSessionUpdate {
					if realtimeEvent.Session != nil {
						if realtimeEvent.Session.Tools != nil {
							info.RealtimeTools = realtimeEvent.Session.Tools
						}
					}
				}

				textToken, audioToken, err := service.CountTokenRealtime(info, *realtimeEvent, info.UpstreamModelName)
				if err != nil {
					errChan <- fmt.Errorf("error counting text token: %v", err)
					return
				}
				common.LogInfo(c, fmt.Sprintf("type: %s, textToken: %d, audioToken: %d", realtimeEvent.Type, textToken, audioToken))
				localUsage.TotalTokens += textToken + audioToken
				localUsage.InputTokens += textToken + audioToken
				localUsage.InputTokenDetails.TextTokens += textToken
				localUsage.InputTokenDetails.AudioTokens += audioToken

				err = helper.WssString(c, targetConn, string(message))
				if err != nil {
					errChan <- fmt.Errorf("error writing to target: %v", err)
					return
				}

				select {
				case sendChan <- message:
				default:
				}
			}
		}
	})

	gopool.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic in target reader: %v", r)
			}
		}()
		for {
			select {
			case <-c.Done():
				return
			default:
				_, message, err := targetConn.ReadMessage()
				if err != nil {
					if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
						errChan <- fmt.Errorf("error reading from target: %v", err)
					}
					close(targetClosed)
					return
				}
				info.SetFirstResponseTime()
				realtimeEvent := &dto.RealtimeEvent{}
				err = json.Unmarshal(message, realtimeEvent)
				if err != nil {
					errChan <- fmt.Errorf("error unmarshalling message: %v", err)
					return
				}

				if realtimeEvent.Type == dto.RealtimeEventTypeResponseDone {
					realtimeUsage := realtimeEvent.Response.Usage
					if realtimeUsage != nil {
						usage.TotalTokens += realtimeUsage.TotalTokens
						usage.InputTokens += realtimeUsage.InputTokens
						usage.OutputTokens += realtimeUsage.OutputTokens
						usage.InputTokenDetails.AudioTokens += realtimeUsage.InputTokenDetails.AudioTokens
						usage.InputTokenDetails.CachedTokens += realtimeUsage.InputTokenDetails.CachedTokens
						usage.InputTokenDetails.TextTokens += realtimeUsage.InputTokenDetails.TextTokens
						usage.OutputTokenDetails.AudioTokens += realtimeUsage.OutputTokenDetails.AudioTokens
						usage.OutputTokenDetails.TextTokens += realtimeUsage.OutputTokenDetails.TextTokens
						err := preConsumeUsage(c, info, usage, sumUsage)
						if err != nil {
							errChan <- fmt.Errorf("error consume usage: %v", err)
							return
						}
						// 本次计费完成，清除
						usage = &dto.RealtimeUsage{}

						localUsage = &dto.RealtimeUsage{}
					} else {
						textToken, audioToken, err := service.CountTokenRealtime(info, *realtimeEvent, info.UpstreamModelName)
						if err != nil {
							errChan <- fmt.Errorf("error counting text token: %v", err)
							return
						}
						common.LogInfo(c, fmt.Sprintf("type: %s, textToken: %d, audioToken: %d", realtimeEvent.Type, textToken, audioToken))
						localUsage.TotalTokens += textToken + audioToken
						info.IsFirstRequest = false
						localUsage.InputTokens += textToken + audioToken
						localUsage.InputTokenDetails.TextTokens += textToken
						localUsage.InputTokenDetails.AudioTokens += audioToken
						err = preConsumeUsage(c, info, localUsage, sumUsage)
						if err != nil {
							errChan <- fmt.Errorf("error consume usage: %v", err)
							return
						}
						// 本次计费完成，清除
						localUsage = &dto.RealtimeUsage{}
						// print now usage
					}
					//common.LogInfo(c, fmt.Sprintf("realtime streaming sumUsage: %v", sumUsage))
					//common.LogInfo(c, fmt.Sprintf("realtime streaming localUsage: %v", localUsage))
					//common.LogInfo(c, fmt.Sprintf("realtime streaming localUsage: %v", localUsage))

				} else if realtimeEvent.Type == dto.RealtimeEventTypeSessionUpdated || realtimeEvent.Type == dto.RealtimeEventTypeSessionCreated {
					realtimeSession := realtimeEvent.Session
					if realtimeSession != nil {
						// update audio format
						info.InputAudioFormat = common.GetStringIfEmpty(realtimeSession.InputAudioFormat, info.InputAudioFormat)
						info.OutputAudioFormat = common.GetStringIfEmpty(realtimeSession.OutputAudioFormat, info.OutputAudioFormat)
					}
				} else {
					textToken, audioToken, err := service.CountTokenRealtime(info, *realtimeEvent, info.UpstreamModelName)
					if err != nil {
						errChan <- fmt.Errorf("error counting text token: %v", err)
						return
					}
					common.LogInfo(c, fmt.Sprintf("type: %s, textToken: %d, audioToken: %d", realtimeEvent.Type, textToken, audioToken))
					localUsage.TotalTokens += textToken + audioToken
					localUsage.OutputTokens += textToken + audioToken
					localUsage.OutputTokenDetails.TextTokens += textToken
					localUsage.OutputTokenDetails.AudioTokens += audioToken
				}

				err = helper.WssString(c, clientConn, string(message))
				if err != nil {
					errChan <- fmt.Errorf("error writing to client: %v", err)
					return
				}

				select {
				case receiveChan <- message:
				default:
				}
			}
		}
	})

	select {
	case <-clientClosed:
	case <-targetClosed:
	case err := <-errChan:
		//return service.OpenAIErrorWrapper(err, "realtime_error", http.StatusInternalServerError), nil
		common.LogError(c, "realtime error: "+err.Error())
	case <-c.Done():
	}

	if usage.TotalTokens != 0 {
		_ = preConsumeUsage(c, info, usage, sumUsage)
	}

	if localUsage.TotalTokens != 0 {
		_ = preConsumeUsage(c, info, localUsage, sumUsage)
	}

	// check usage total tokens, if 0, use local usage

	return nil, sumUsage
}

func preConsumeUsage(ctx *gin.Context, info *relaycommon.RelayInfo, usage *dto.RealtimeUsage, totalUsage *dto.RealtimeUsage) error {
	if usage == nil || totalUsage == nil {
		return fmt.Errorf("invalid usage pointer")
	}

	totalUsage.TotalTokens += usage.TotalTokens
	totalUsage.InputTokens += usage.InputTokens
	totalUsage.OutputTokens += usage.OutputTokens
	totalUsage.InputTokenDetails.CachedTokens += usage.InputTokenDetails.CachedTokens
	totalUsage.InputTokenDetails.TextTokens += usage.InputTokenDetails.TextTokens
	totalUsage.InputTokenDetails.AudioTokens += usage.InputTokenDetails.AudioTokens
	totalUsage.OutputTokenDetails.TextTokens += usage.OutputTokenDetails.TextTokens
	totalUsage.OutputTokenDetails.AudioTokens += usage.OutputTokenDetails.AudioTokens
	// clear usage
	err := service.PreWssConsumeQuota(ctx, info, usage)
	return err
}
