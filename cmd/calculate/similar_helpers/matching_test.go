package similar_helpers

import (
	"testing"

	"github.com/similar-manga/similar/internal"
)

func TestNotValidMatch(t *testing.T) {
	// Helper to create a manga with common fields
	createManga := func(id string, title string, contentRating string, demo string, relatedIds []string, tags []internal.Tag) internal.Manga {
		titleMap := map[string]string{"en": title}
		return internal.Manga{
			Id:                     id,
			Title:                  &titleMap,
			ContentRating:          contentRating,
			PublicationDemographic: demo,
			RelatedIds:             relatedIds,
			Tags:                   tags,
		}
	}

	// Helper for creating tags
	createTag := func(id string) internal.Tag {
		return internal.Tag{Id: id}
	}

	tests := []struct {
		name       string
		manga      internal.Manga
		mangaOther internal.Manga
		want       bool // true means INVALID, false means VALID
	}{
		{
			name:       "Valid Match: No conflicts",
			manga:      createManga("1", "Manga A", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "safe", "shounen", nil, nil),
			want:       false,
		},
		{
			name:       "Invalid Match: Related ID (Forward)",
			manga:      createManga("1", "Manga A", "safe", "shounen", []string{"2"}, nil),
			mangaOther: createManga("2", "Manga B", "safe", "shounen", nil, nil),
			want:       true,
		},
		{
			name:       "Invalid Match: Related ID (Backward)",
			manga:      createManga("1", "Manga A", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "safe", "shounen", []string{"1"}, nil),
			want:       true,
		},
		{
			name:       "Invalid Match: Content Rating Mismatch",
			manga:      createManga("1", "Manga A", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "suggestive", "shounen", nil, nil),
			want:       true,
		},
		{
			name:       "Valid Match: Content Rating Same (Empty)",
			manga:      createManga("1", "Manga A", "", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "", "shounen", nil, nil),
			want:       false,
		},
		{
			name:       "Valid Match: Content Rating Same (Non-Empty)",
			manga:      createManga("1", "Manga A", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "safe", "shounen", nil, nil),
			want:       false,
		},
		{
			name:       "Invalid Match: Content Rating Source Set, Target Empty",
			manga:      createManga("1", "Manga A", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "", "shounen", nil, nil),
			want:       true,
		},
		{
			name:       "Valid Match: Content Rating Source Empty, Target Set",
			manga:      createManga("1", "Manga A", "", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "safe", "shounen", nil, nil),
			want:       false,
		},
		{
			name:       "Invalid Match: Promo Check (Source Not Promo, Target Promo)",
			manga:      createManga("1", "Normal Title", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Promo Title (Promo)", "safe", "shounen", nil, nil),
			want:       true,
		},
		{
			name:       "Valid Match: Promo Check (Source Promo, Target Not Promo)",
			manga:      createManga("1", "Promo Title (Promo)", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Normal Title", "safe", "shounen", nil, nil),
			want:       false,
		},
		{
			name:       "Valid Match: Promo Check (Both Promo)",
			manga:      createManga("1", "Promo Title (Promo)", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Promo Title (Promo)", "safe", "shounen", nil, nil),
			want:       false,
		},
		{
			name:       "Valid Match: Promo Check (Neither Promo)",
			manga:      createManga("1", "Normal Title", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Normal Title", "safe", "shounen", nil, nil),
			want:       false,
		},
		{
			name:       "Invalid Match: Demographic Mismatch",
			manga:      createManga("1", "Manga A", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "safe", "seinen", nil, nil),
			want:       true,
		},
		{
			name:       "Valid Match: Demographic Same (Non-Empty)",
			manga:      createManga("1", "Manga A", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "safe", "shounen", nil, nil),
			want:       false,
		},
		{
			name:       "Invalid Match: Demographic Source Set, Target Empty",
			manga:      createManga("1", "Manga A", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "safe", "", nil, nil),
			want:       true,
		},
		{
			name:       "Valid Match: Demographic Source Empty, Target Set",
			manga:      createManga("1", "Manga A", "safe", "", nil, nil),
			mangaOther: createManga("2", "Manga B", "safe", "shounen", nil, nil),
			want:       false,
		},
		{
			name:       "Valid Match: Erotica Bypass (Tag Check Skipped)",
			manga:      createManga("1", "Manga A", "erotica", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "erotica", "shounen", nil, []internal.Tag{createTag(oneWayTags[0])}),
			want:       false,
		},
		{
			name:       "Valid Match: Pornographic Bypass (Tag Check Skipped)",
			manga:      createManga("1", "Manga A", "pornographic", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "pornographic", "shounen", nil, []internal.Tag{createTag(oneWayTags[0])}),
			want:       false,
		},
		{
			name:       "Invalid Match: One-Way Tag (Source Missing, Target Has)",
			manga:      createManga("1", "Manga A", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "safe", "shounen", nil, []internal.Tag{createTag(oneWayTags[0])}),
			want:       true,
		},
		{
			name:       "Valid Match: One-Way Tag (Source Has, Target Has)",
			manga:      createManga("1", "Manga A", "safe", "shounen", nil, []internal.Tag{createTag(oneWayTags[0])}),
			mangaOther: createManga("2", "Manga B", "safe", "shounen", nil, []internal.Tag{createTag(oneWayTags[0])}),
			want:       false,
		},
		{
			name:       "Valid Match: One-Way Tag (Source Has, Target Missing)",
			manga:      createManga("1", "Manga A", "safe", "shounen", nil, []internal.Tag{createTag(oneWayTags[0])}),
			mangaOther: createManga("2", "Manga B", "safe", "shounen", nil, nil),
			want:       false,
		},
		{
			name:       "Valid Match: One-Way Tag (Neither Has)",
			manga:      createManga("1", "Manga A", "safe", "shounen", nil, nil),
			mangaOther: createManga("2", "Manga B", "safe", "shounen", nil, nil),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NotValidMatch(tt.manga, tt.mangaOther)
			if got != tt.want {
				t.Errorf("NotValidMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
