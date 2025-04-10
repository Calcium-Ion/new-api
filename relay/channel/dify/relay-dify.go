package dify

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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

	"github.com/gin-gonic/gin"
)

func uploadDifyFile(c *gin.Context, info *relaycommon.RelayInfo, user string, media dto.MediaContent) *DifyFile {
	uploadUrl := fmt.Sprintf("%s/v1/files/upload", info.BaseUrl)
	switch media.Type {
	case dto.ContentTypeImageURL:
		// Decode base64 data
		imageMedia := media.GetImageMedia()
		base64Data := imageMedia.Url
		// Remove base64 prefix if exists (e.g., "data:image/jpeg;base64,")
		if idx := strings.Index(base64Data, ","); idx != -1 {
			base64Data = base64Data[idx+1:]
		}

		// Decode base64 string
		decodedData, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			common.SysError("failed to decode base64: " + err.Error())
			return nil
		}

		// Create temporary file
		tempFile, err := os.CreateTemp("", "dify-upload-*")
		if err != nil {
			common.SysError("failed to create temp file: " + err.Error())
			return nil
		}
		defer tempFile.Close()
		defer os.Remove(tempFile.Name())

		// Write decoded data to temp file
		if _, err := tempFile.Write(decodedData); err != nil {
			common.SysError("failed to write to temp file: " + err.Error())
			return nil
		}

		// Create multipart form
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add user field
		if err := writer.WriteField("user", user); err != nil {
			common.SysError("failed to add user field: " + err.Error())
			return nil
		}

		// Create form file with proper mime type
		mimeType := imageMedia.MimeType
		if mimeType == "" {
			mimeType = "image/jpeg" // default mime type
		}

		// Create form file
		part, err := writer.CreateFormFile("file", fmt.Sprintf("image.%s", strings.TrimPrefix(mimeType, "image/")))
		if err != nil {
			common.SysError("failed to create form file: " + err.Error())
			return nil
		}

		// Copy file content to form
		if _, err = io.Copy(part, bytes.NewReader(decodedData)); err != nil {
			common.SysError("failed to copy file content: " + err.Error())
			return nil
		}
		writer.Close()

		// Create HTTP request
		req, err := http.NewRequest("POST", uploadUrl, body)
		if err != nil {
			common.SysError("failed to create request: " + err.Error())
			return nil
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", info.ApiKey))

		// Send request
		client := service.GetImpatientHttpClient()
		resp, err := client.Do(req)
		if err != nil {
			common.SysError("failed to send request: " + err.Error())
			return nil
		}
		defer resp.Body.Close()

		// Parse response
		var result struct {
			Id string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			common.SysError("failed to decode response: " + err.Error())
			return nil
		}

		return &DifyFile{
			UploadFileId: result.Id,
			Type:         "image",
			TransferMode: "local_file",
		}
	}
	return nil
}

func requestOpenAI2Dify(c *gin.Context, info *relaycommon.RelayInfo, request dto.GeneralOpenAIRequest) *DifyChatRequest {
	difyReq := DifyChatRequest{
		Inputs:           make(map[string]interface{}),
		AutoGenerateName: false,
	}

	user := request.User
	if user == "" {
		user = helper.GetResponseID(c)
	}
	difyReq.User = user

	files := make([]DifyFile, 0)
	var content strings.Builder
	for _, message := range request.Messages {
		if message.Role == "system" {
			content.WriteString("SYSTEM: \n" + message.StringContent() + "\n")
		} else if message.Role == "assistant" {
			content.WriteString("ASSISTANT: \n" + message.StringContent() + "\n")
		} else {
			parseContent := message.ParseContent()
			for _, mediaContent := range parseContent {
				switch mediaContent.Type {
				case dto.ContentTypeText:
					content.WriteString("USER: \n" + mediaContent.Text + "\n")
				case dto.ContentTypeImageURL:
					media := mediaContent.GetImageMedia()
					var file *DifyFile
					if media.IsRemoteImage() {
						file.Type = media.MimeType
						file.TransferMode = "remote_url"
						file.URL = media.Url
					} else {
						file = uploadDifyFile(c, info, difyReq.User, mediaContent)
					}
					if file != nil {
						files = append(files, *file)
					}
				}
			}
		}
	}
	difyReq.Query = content.String()
	difyReq.Files = files
	mode := "blocking"
	if request.Stream {
		mode = "streaming"
	}
	difyReq.ResponseMode = mode
	return &difyReq
}

func streamResponseDify2OpenAI(difyResponse DifyChunkChatCompletionResponse) *dto.ChatCompletionsStreamResponse {
	response := dto.ChatCompletionsStreamResponse{
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "dify",
	}
	var choice dto.ChatCompletionsStreamResponseChoice
	if strings.HasPrefix(difyResponse.Event, "workflow_") {
		if constant.DifyDebug {
			text := "Workflow: " + difyResponse.Data.WorkflowId
			if difyResponse.Event == "workflow_finished" {
				text += " " + difyResponse.Data.Status
			}
			choice.Delta.SetReasoningContent(text + "\n")
		}
	} else if strings.HasPrefix(difyResponse.Event, "node_") {
		if constant.DifyDebug {
			text := "Node: " + difyResponse.Data.NodeType
			if difyResponse.Event == "node_finished" {
				text += " " + difyResponse.Data.Status
			}
			choice.Delta.SetReasoningContent(text + "\n")
		}
	} else if difyResponse.Event == "message" || difyResponse.Event == "agent_message" {
		if difyResponse.Answer == "<details style=\"color:gray;background-color: #f8f8f8;padding: 8px;border-radius: 4px;\" open> <summary> Thinking... </summary>\n" {
			difyResponse.Answer = "<think>"
		} else if difyResponse.Answer == "</details>" {
			difyResponse.Answer = "</think>"
		}

		choice.Delta.SetContentString(difyResponse.Answer)
	}
	response.Choices = append(response.Choices, choice)
	return &response
}

func difyStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var responseText string
	usage := &dto.Usage{}
	var nodeToken int
	helper.SetEventStreamHeaders(c)
	helper.StreamScannerHandler(c, resp, info, func(data string) bool {
		var difyResponse DifyChunkChatCompletionResponse
		err := json.Unmarshal([]byte(data), &difyResponse)
		if err != nil {
			common.SysError("error unmarshalling stream response: " + err.Error())
			return true
		}
		var openaiResponse dto.ChatCompletionsStreamResponse
		if difyResponse.Event == "message_end" {
			usage = &difyResponse.MetaData.Usage
			return false
		} else if difyResponse.Event == "error" {
			return false
		} else {
			openaiResponse = *streamResponseDify2OpenAI(difyResponse)
			if len(openaiResponse.Choices) != 0 {
				responseText += openaiResponse.Choices[0].Delta.GetContentString()
				if openaiResponse.Choices[0].Delta.ReasoningContent != nil {
					nodeToken += 1
				}
			}
		}
		err = helper.ObjectData(c, openaiResponse)
		if err != nil {
			common.SysError(err.Error())
		}
		return true
	})
	helper.Done(c)
	err := resp.Body.Close()
	if err != nil {
		// return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
		common.SysError("close_response_body_failed: " + err.Error())
	}
	if usage.TotalTokens == 0 {
		usage.PromptTokens = info.PromptTokens
		usage.CompletionTokens, _ = service.CountTextToken("gpt-3.5-turbo", responseText)
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	usage.CompletionTokens += nodeToken
	return nil, usage
}

func difyHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var difyResponse DifyChatCompletionResponse
	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &difyResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	fullTextResponse := dto.OpenAITextResponse{
		Id:      difyResponse.ConversationId,
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Usage:   difyResponse.MetaData.Usage,
	}
	content, _ := json.Marshal(difyResponse.Answer)
	choice := dto.OpenAITextResponseChoice{
		Index: 0,
		Message: dto.Message{
			Role:    "assistant",
			Content: content,
		},
		FinishReason: "stop",
	}
	fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &difyResponse.MetaData.Usage
}
