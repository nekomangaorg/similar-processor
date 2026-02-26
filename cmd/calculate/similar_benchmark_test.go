package calculate

import (
	"fmt"
	"iter"
	"testing"
	"github.com/similar-manga/similar/internal"
)

// Helper to create a large manga list
func createLargeMangaList(n int) []internal.Manga {
	list := make([]internal.Manga, n)
	for i := 0; i < n; i++ {
		title := map[string]string{"en": fmt.Sprintf("Manga Title %d", i)}
		desc := map[string]string{"en": fmt.Sprintf("Description for manga %d. It has some words.", i)}
		tagNames := map[string]string{"en": "Tag1"}
		list[i] = internal.Manga{
			Id:          fmt.Sprintf("uuid-%d", i),
			Title:       &title,
			Description: &desc,
			Tags:        []internal.Tag{{Name: &tagNames}},
		}
	}
	return list
}

func BenchmarkFilterAndBuildCorpus(b *testing.B) {
	// Create data once
	mangaList := createLargeMangaList(10000)

	var iterator iter.Seq[internal.Manga] = func(yield func(internal.Manga) bool) {
		for _, m := range mangaList {
			if !yield(m) {
				return
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filterAndBuildCorpus(iterator)
	}
}
