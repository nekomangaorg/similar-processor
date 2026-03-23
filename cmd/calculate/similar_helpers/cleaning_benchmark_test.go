package similar_helpers

import (
	"testing"
)

func BenchmarkCleanDescription(b *testing.B) {
	input := "This is a typical description for a manga. It has several sentences. (source: mangadex) [b]bold text[/b] http://example.com"
	for i := 0; i < b.N; i++ {
		CleanDescription(input)
	}
}

func BenchmarkCleanTitle(b *testing.B) {
	input := "This is a typical description for a manga. It has several sentences. (source: mangadex) [b]bold text[/b] http://example.com"
	for i := 0; i < b.N; i++ {
		CleanTitle(input)
	}
}
