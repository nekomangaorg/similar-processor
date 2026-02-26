package calculate

import (
	"regexp"
	"strings"
	"testing"
)

var tagRegexBench = regexp.MustCompile("[^a-zA-Z0-9]+")

func TestCorrectness(t *testing.T) {
	cases := []string{
		"Simple",
		"With spaces",
		"With-hyphens",
		"With!@#Symbols",
		"123Numbers",
		"Mixed123!@#Test",
		"",
	}

	for _, c := range cases {
		expected := tagRegexBench.ReplaceAllString(c, "")

		var b strings.Builder
		cleanTag(c, &b)
		actual := b.String()

		if expected != actual {
			t.Errorf("Mismatch for input '%s': expected '%s', got '%s'", c, expected, actual)
		}
	}
}

func BenchmarkRegex(b *testing.B) {
	tag := "Example-Tag with 123!@#"
	for i := 0; i < b.N; i++ {
		_ = tagRegexBench.ReplaceAllString(tag, "")
	}
}

func BenchmarkCleanTag(b *testing.B) {
	tag := "Example-Tag with 123!@#"
	for i := 0; i < b.N; i++ {
		var builder strings.Builder
		builder.Grow(len(tag))
		cleanTag(tag, &builder)
		_ = builder.String()
	}
}
