package internal

type SimilarManga struct {
	Id             string            `json:"id,omitempty"`
	Title          map[string]string `json:"title,omitempty"`
	ContentRating  string            `json:"contentRating,omitempty"`
	SimilarMatches []SimilarMatch    `json:"matches,omitempty"`
	UpdatedAt      string            `json:"updatedAt,omitempty"`
}

type SimilarMatch struct {
	Id            string            `json:"id,omitempty"`
	Title         map[string]string `json:"title,omitempty"`
	ContentRating string            `json:"contentRating,omitempty"`
	Score         float32           `json:"score,omitempty"`
	Languages     []string          `json:"languages,omitempty"`
}
