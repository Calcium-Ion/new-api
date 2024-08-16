package dto

type RerankRequest struct {
	Documents       []any  `json:"documents"`
	Query           string `json:"query"`
	Model           string `json:"model"`
	TopN            int    `json:"top_n"`
	ReturnDocuments bool   `json:"return_documents,omitempty"`
	MaxChunkPerDoc  int    `json:"max_chunk_per_doc,omitempty"`
	OverLapTokens   int    `json:"overlap_tokens,omitempty"`
}

type RerankResponseDocument struct {
	Document       any     `json:"document,omitempty"`
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}

type RerankResponse struct {
	Results []RerankResponseDocument `json:"results"`
	Usage   Usage                    `json:"usage"`
}
