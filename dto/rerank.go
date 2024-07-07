package dto

type RerankRequest struct {
	Documents []any  `json:"documents"`
	Query     string `json:"query"`
	Model     string `json:"model"`
	TopN      int    `json:"top_n"`
}

type RerankResponseDocument struct {
	Document       any     `json:"document"`
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}

type RerankResponse struct {
	Results []RerankResponseDocument `json:"results"`
	Usage   Usage                    `json:"usage"`
}
