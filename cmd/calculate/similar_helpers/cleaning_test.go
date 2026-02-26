package similar_helpers

import (
	"testing"
)

func TestCleanTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic Alphanumeric",
			input:    "Hello World",
			expected: "hello world",
		},
		{
			name:     "With Symbols",
			input:    "Hello! @World#",
			expected: "hello world",
		},
		{
			name:     "Multiple Spaces",
			input:    "Hello   World",
			expected: "hello world",
		},
		{
			name:     "Stemming Plural",
			input:    "Cats Dogs",
			expected: "cat dog",
		},
		{
			name:     "Stemming Gerund",
			input:    "Running Walking",
			expected: "run wal",
		},
		{
			name:     "Mixed Case and Symbols",
			input:    "RUNNING!!! fast??",
			expected: "run fast",
		},
		{
			name:     "Empty String",
			input:    "",
			expected: "",
		},
		{
			name:     "Numbers",
			input:    "Chapter 123",
			expected: "chapter 123",
		},
		{
			name:     "With Hyphens",
			input:    "spider-man",
			expected: "spiderman", // regex removes hyphen
		},
		{
			name:     "Tabs and Newlines",
			input:    "Hello\tWorld\nHere",
			expected: "helloworldher", // regex removes non-space whitespace, stemmer truncates 'e'
		},
		{
			name:     "Foreign Characters",
			input:    "Manga Café",
			expected: "manga caf", // accents are removed
		},
		{
			name:     "Only Symbols",
			input:    "!@#$%^&*()",
			expected: "",
		},
		{
			name:     "Only Spaces",
			input:    "   ",
			expected: " ",
		},
		{
			name:     "Leading and Trailing Spaces",
			input:    "  Hello  ",
			expected: " hello ",
		},
		{
			name:     "Non-ASCII Characters",
			input:    "Héllo Wörld",
			expected: "hllo wrld",
		},
		{
			name:     "Only Non-ASCII",
			input:    "日本語",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanTitle(tt.input)
			if got != tt.expected {
				t.Errorf("CleanTitle(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
