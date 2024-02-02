package controller

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	GeminiVisionMaxImageNum = 16
)

type GeminiChatRequest struct {
	Contents         []GeminiChatContent        `json:"contents"`
	SafetySettings   []GeminiChatSafetySettings `json:"safety_settings,omitempty"`
	GenerationConfig GeminiChatGenerationConfig `json:"generation_config,omitempty"`
	Tools            []GeminiChatTools          `json:"tools,omitempty"`
}

type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type GeminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *GeminiInlineData `json:"inlineData,omitempty"`
}

type GeminiChatContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiChatSafetySettings struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type GeminiChatTools struct {
	FunctionDeclarations any `json:"functionDeclarations,omitempty"`
}

type GeminiChatGenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            float64  `json:"topK,omitempty"`
	MaxOutputTokens uint     `json:"maxOutputTokens,omitempty"`
	CandidateCount  int      `json:"candidateCount,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

// Setting safety to the lowest possible values since Gemini is already powerless enough
func requestOpenAI2Gemini(textRequest GeneralOpenAIRequest) *GeminiChatRequest {
	geminiRequest := GeminiChatRequest{
		Contents: make([]GeminiChatContent, 0, len(textRequest.Messages)),
		SafetySettings: []GeminiChatSafetySettings{
			{
				Category:  "HARM_CATEGORY_HARASSMENT",
				Threshold: common.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_HATE_SPEECH",
				Threshold: common.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				Threshold: common.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: common.GeminiSafetySetting,
			},
		},
		GenerationConfig: GeminiChatGenerationConfig{
			Temperature:     textRequest.Temperature,
			TopP:            textRequest.TopP,
			MaxOutputTokens: textRequest.MaxTokens,
		},
	}
	if textRequest.Functions != nil {
		geminiRequest.Tools = []GeminiChatTools{
			{
				FunctionDeclarations: textRequest.Functions,
			},
		}
	}
	shouldAddDummyModelMessage := false
	for _, message := range textRequest.Messages {
		content := GeminiChatContent{
			Role: message.Role,
			Parts: []GeminiPart{
				{
					Text: message.StringContent(),
				},
			},
		}
		openaiContent := message.ParseContent()
		var parts []GeminiPart
		imageNum := 0
		for _, part := range openaiContent {

			if part.Type == ContentTypeText {
				parts = append(parts, GeminiPart{
					Text: part.Text,
				})
			} else if part.Type == ContentTypeImageURL {
				imageNum += 1
				if imageNum > GeminiVisionMaxImageNum {
					continue
				}
				mimeType, data, _ := common.GetImageFromUrl(part.ImageUrl.(MessageImageUrl).Url)
				parts = append(parts, GeminiPart{
					InlineData: &GeminiInlineData{
						MimeType: mimeType,
						Data:     data,
					},
				})
			}
		}
		content.Parts = parts

		// there's no assistant role in gemini and API shall vomit if Role is not user or model
		if content.Role == "assistant" {
			content.Role = "model"
		}
		// Converting system prompt to prompt from user for the same reason
		if content.Role == "system" {
			content.Role = "user"
			shouldAddDummyModelMessage = true
		}
		geminiRequest.Contents = append(geminiRequest.Contents, content)

		// If a system message is the last message, we need to add a dummy model message to make gemini happy
		if shouldAddDummyModelMessage {
			geminiRequest.Contents = append(geminiRequest.Contents, GeminiChatContent{
				Role: "model",
				Parts: []GeminiPart{
					{
						Text: "Okay",
					},
				},
			})
			shouldAddDummyModelMessage = false
		}
	}

	return &geminiRequest
}

type GeminiChatResponse struct {
	Candidates     []GeminiChatCandidate    `json:"candidates"`
	PromptFeedback GeminiChatPromptFeedback `json:"promptFeedback"`
}

func (g *GeminiChatResponse) GetResponseText() string {
	if g == nil {
		return ""
	}
	if len(g.Candidates) > 0 && len(g.Candidates[0].Content.Parts) > 0 {
		return g.Candidates[0].Content.Parts[0].Text
	}
	return ""
}

type GeminiChatCandidate struct {
	Content       GeminiChatContent        `json:"content"`
	FinishReason  string                   `json:"finishReason"`
	Index         int64                    `json:"index"`
	SafetyRatings []GeminiChatSafetyRating `json:"safetyRatings"`
}

type GeminiChatSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type GeminiChatPromptFeedback struct {
	SafetyRatings []GeminiChatSafetyRating `json:"safetyRatings"`
}

func responseGeminiChat2OpenAI(response *GeminiChatResponse) *OpenAITextResponse {
	fullTextResponse := OpenAITextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: make([]OpenAITextResponseChoice, 0, len(response.Candidates)),
	}
	content, _ := json.Marshal("")
	for i, candidate := range response.Candidates {
		choice := OpenAITextResponseChoice{
			Index: i,
			Message: Message{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: stopFinishReason,
		}
		content, _ = json.Marshal(candidate.Content.Parts[0].Text)
		if len(candidate.Content.Parts) > 0 {
			choice.Message.Content = content
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	}
	return &fullTextResponse
}

func streamResponseGeminiChat2OpenAI(geminiResponse *GeminiChatResponse) *ChatCompletionsStreamResponse {
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = geminiResponse.GetResponseText()
	choice.FinishReason = &stopFinishReason
	var response ChatCompletionsStreamResponse
	response.Object = "chat.completion.chunk"
	response.Model = "gemini"
	response.Choices = []ChatCompletionsStreamResponseChoice{choice}
	return &response
}

func geminiChatStreamHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, string) {
	responseText := ""
	dataChan := make(chan string)
	stopChan := make(chan bool)
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "\n"); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	go func() {
		for scanner.Scan() {
			data := scanner.Text()
			data = strings.TrimSpace(data)
			if !strings.HasPrefix(data, "\"text\": \"") {
				continue
			}
			data = strings.TrimPrefix(data, "\"text\": \"")
			data = strings.TrimSuffix(data, "\"")
			dataChan <- data
		}
		stopChan <- true
	}()
	setEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			// this is used to prevent annoying \ related format bug
			data = fmt.Sprintf("{\"content\": \"%s\"}", data)
			type dummyStruct struct {
				Content string `json:"content"`
			}
			var dummy dummyStruct
			err := json.Unmarshal([]byte(data), &dummy)
			responseText += dummy.Content
			var choice ChatCompletionsStreamResponseChoice
			choice.Delta.Content = dummy.Content
			response := ChatCompletionsStreamResponse{
				Id:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
				Object:  "chat.completion.chunk",
				Created: common.GetTimestamp(),
				Model:   "gemini-pro",
				Choices: []ChatCompletionsStreamResponseChoice{choice},
			}
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonResponse)})
			return true
		case <-stopChan:
			c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}
	})
	err := resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), ""
	}
	return nil, responseText
}

func geminiChatHandler(c *gin.Context, resp *http.Response, promptTokens int, model string) (*OpenAIErrorWithStatusCode, *Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var geminiResponse GeminiChatResponse
	err = json.Unmarshal(responseBody, &geminiResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if len(geminiResponse.Candidates) == 0 {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
				Message: "No candidates returned",
				Type:    "server_error",
				Param:   "",
				Code:    500,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseGeminiChat2OpenAI(&geminiResponse)
	completionTokens := countTokenText(geminiResponse.GetResponseText(), model)
	usage := Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
	fullTextResponse.Usage = usage
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &usage
}
