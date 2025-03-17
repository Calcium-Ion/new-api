package siliconflow

import "one-api/dto"

type SFTokens struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type SFMeta struct {
	Tokens SFTokens `json:"tokens"`
}

type SFRerankResponse struct {
	Results []dto.RerankResponseResult `json:"results"`
	Meta    SFMeta                     `json:"meta"`
}
