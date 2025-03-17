package xinference

type XinRerankResponseDocument struct {
	Document       string  `json:"document,omitempty"`
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}

type XinRerankResponse struct {
	Results []XinRerankResponseDocument `json:"results"`
}
