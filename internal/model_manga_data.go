package internal

type Manga struct {
	Id                           string              `json:"id,omitempty"`
	Title                        *map[string]string  `json:"title,omitempty"`
	AltTitles                    []map[string]string `json:"altTitles,omitempty"`
	LastChapter                  string              `json:"lastChapter,omitempty"`
	AvailableTranslatedLanguages []string            `json:"availableTranslatedLanguages,omitempty"`
	RelatedIds                   []string            `json:"relatedIdds,omitempty"`
	Description                  *map[string]string  `json:"description,omitempty"`
	Links                        map[string]string   `json:"links,omitempty"`
	OriginalLanguage             string              `json:"originalLanguage,omitempty"`
	PublicationDemographic       string              `json:"publicationDemographic,omitempty"`
	ContentRating                string              `json:"contentRating,omitempty"`
	Tags                         []Tag               `json:"tags,omitempty"`
}

type Tag struct {
	Id   string             `json:"id,omitempty"`
	Name *map[string]string `json:"name,omitempty"`
}
