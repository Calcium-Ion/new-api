package dto

type VideoRequest struct {
	Model        string `json:"model"`
	Prompt       string `json:"prompt" binding:"required"`
	ImgUrl       string `json:"img_url"`
	Duration     int64  `json:"parameters"`
	PromptExtend bool   `json:"prompt_extend"`
	Seed         int64  `json:"seed,omitempty"`
	Size         string `json:"size"`
}

type VideoResponse struct {
	RequestID     string `json:"request_id"`
	TaskID        string `json:"task_id"`
	TaskStatus    string `json:"task_status"`
	SubmitTime    string `json:"submit_time"`
	ScheduledTime string `json:"scheduled_time"`
	EndTime       string `json:"end_time"`
	VideoURL      string `json:"video_url"`
	VideoDuration int    `json:"video_duration"`
	VideoRatio    string `json:"video_ratio"`
	VideoCount    int    `json:"video_count"`
}
