package baidu

import (
	"one-api/dto"
	"time"
)

type BaiduMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type BaiduChatRequest struct {
	Messages []BaiduMessage `json:"messages"`
	Stream   bool           `json:"stream"`
	UserId   string         `json:"user_id,omitempty"`
}

type Error struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

type BaiduChatResponse struct {
	Id               string    `json:"id"`
	Object           string    `json:"object"`
	Created          int64     `json:"created"`
	Result           string    `json:"result"`
	IsTruncated      bool      `json:"is_truncated"`
	NeedClearHistory bool      `json:"need_clear_history"`
	Usage            dto.Usage `json:"usage"`
	Error
}

type BaiduChatStreamResponse struct {
	BaiduChatResponse
	SentenceId int  `json:"sentence_id"`
	IsEnd      bool `json:"is_end"`
}

type BaiduEmbeddingRequest struct {
	Input []string `json:"input"`
}

type BaiduEmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type BaiduEmbeddingResponse struct {
	Id      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Data    []BaiduEmbeddingData `json:"data"`
	Usage   dto.Usage            `json:"usage"`
	Error
}

type BaiduAccessToken struct {
	AccessToken      string    `json:"access_token"`
	Error            string    `json:"error,omitempty"`
	ErrorDescription string    `json:"error_description,omitempty"`
	ExpiresIn        int64     `json:"expires_in,omitempty"`
	ExpiresAt        time.Time `json:"-"`
}

type BaiduTokenResponse struct {
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}
