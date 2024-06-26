package dto

import (
	"encoding/json"
)

type TaskData interface {
	SunoDataResponse | []SunoDataResponse | string | any
}

type SunoSubmitReq struct {
	GptDescriptionPrompt string  `json:"gpt_description_prompt,omitempty"`
	Prompt               string  `json:"prompt,omitempty"`
	Mv                   string  `json:"mv,omitempty"`
	Title                string  `json:"title,omitempty"`
	Tags                 string  `json:"tags,omitempty"`
	ContinueAt           float64 `json:"continue_at,omitempty"`
	TaskID               string  `json:"task_id,omitempty"`
	ContinueClipId       string  `json:"continue_clip_id,omitempty"`
	MakeInstrumental     bool    `json:"make_instrumental"`
}

type FetchReq struct {
	IDs []string `json:"ids"`
}

type SunoDataResponse struct {
	TaskID     string          `json:"task_id" gorm:"type:varchar(50);index"`
	Action     string          `json:"action" gorm:"type:varchar(40);index"` // 任务类型, song, lyrics, description-mode
	Status     string          `json:"status" gorm:"type:varchar(20);index"` // 任务状态, submitted, queueing, processing, success, failed
	FailReason string          `json:"fail_reason"`
	SubmitTime int64           `json:"submit_time" gorm:"index"`
	StartTime  int64           `json:"start_time" gorm:"index"`
	FinishTime int64           `json:"finish_time" gorm:"index"`
	Data       json.RawMessage `json:"data" gorm:"type:json"`
}

type SunoSong struct {
	ID                string       `json:"id"`
	VideoURL          string       `json:"video_url"`
	AudioURL          string       `json:"audio_url"`
	ImageURL          string       `json:"image_url"`
	ImageLargeURL     string       `json:"image_large_url"`
	MajorModelVersion string       `json:"major_model_version"`
	ModelName         string       `json:"model_name"`
	Status            string       `json:"status"`
	Title             string       `json:"title"`
	Text              string       `json:"text"`
	Metadata          SunoMetadata `json:"metadata"`
}

type SunoMetadata struct {
	Tags                 string      `json:"tags"`
	Prompt               string      `json:"prompt"`
	GPTDescriptionPrompt interface{} `json:"gpt_description_prompt"`
	AudioPromptID        interface{} `json:"audio_prompt_id"`
	Duration             interface{} `json:"duration"`
	ErrorType            interface{} `json:"error_type"`
	ErrorMessage         interface{} `json:"error_message"`
}

type SunoLyrics struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Title  string `json:"title"`
	Text   string `json:"text"`
}

const TaskSuccessCode = "success"

type TaskResponse[T TaskData] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func (t *TaskResponse[T]) IsSuccess() bool {
	return t.Code == TaskSuccessCode
}

type TaskDto struct {
	TaskID     string          `json:"task_id"` // 第三方id，不一定有/ song id\ Task id
	Action     string          `json:"action"`  // 任务类型, song, lyrics, description-mode
	Status     string          `json:"status"`  // 任务状态, submitted, queueing, processing, success, failed
	FailReason string          `json:"fail_reason"`
	SubmitTime int64           `json:"submit_time"`
	StartTime  int64           `json:"start_time"`
	FinishTime int64           `json:"finish_time"`
	Progress   string          `json:"progress"`
	Data       json.RawMessage `json:"data"`
}

type SunoGoAPISubmitReq struct {
	CustomMode bool `json:"custom_mode"`

	Input SunoGoAPISubmitReqInput `json:"input"`

	NotifyHook string `json:"notify_hook,omitempty"`
}

type SunoGoAPISubmitReqInput struct {
	GptDescriptionPrompt string  `json:"gpt_description_prompt"`
	Prompt               string  `json:"prompt"`
	Mv                   string  `json:"mv"`
	Title                string  `json:"title"`
	Tags                 string  `json:"tags"`
	ContinueAt           float64 `json:"continue_at"`
	TaskID               string  `json:"task_id"`
	ContinueClipId       string  `json:"continue_clip_id"`
	MakeInstrumental     bool    `json:"make_instrumental"`
}

type GoAPITaskResponse[T any] struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	Data         T      `json:"data"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type GoAPITaskResponseData struct {
	TaskID string `json:"task_id"`
}

type GoAPIFetchResponseData struct {
	TaskID string              `json:"task_id"`
	Status string              `json:"status"`
	Input  string              `json:"input"`
	Clips  map[string]SunoSong `json:"clips"`
}
