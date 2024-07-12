package dify

import "one-api/dto"

type DifyChatRequest struct {
	Inputs           map[string]interface{} `json:"inputs"`
	Query            string                 `json:"query"`
	ResponseMode     string                 `json:"response_mode"`
	User             string                 `json:"user"`
	AutoGenerateName bool                   `json:"auto_generate_name"`
}

type DifyMetaData struct {
	Usage dto.Usage `json:"usage"`
}

type DifyData struct {
	WorkflowId string `json:"workflow_id"`
	NodeId     string `json:"node_id"`
}

type DifyChatCompletionResponse struct {
	ConversationId string       `json:"conversation_id"`
	Answer         string       `json:"answer"`
	CreateAt       int64        `json:"create_at"`
	MetaData       DifyMetaData `json:"metadata"`
}

type DifyChunkChatCompletionResponse struct {
	Event          string       `json:"event"`
	ConversationId string       `json:"conversation_id"`
	Answer         string       `json:"answer"`
	Data           DifyData     `json:"data"`
	MetaData       DifyMetaData `json:"metadata"`
}
