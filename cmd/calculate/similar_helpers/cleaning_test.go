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

func TestCleanDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic Description",
			input:    "This is a typical description for a manga. It has several sentences. (source: mangadex) [b]bold text[/b] http://example.com",
			expected: "thi is a typic descript for a manga it ha sever sentenc bold text httpexamplecom",
		},
		{
			name:     "Nested Brackets",
			input:    "Nested [a [b] c] text",
			expected: "nes text",
		},
		{
			name:     "HTTP Without Newline",
			input:    "http://example.com",
			expected: " ",
		},
		{
			name:     "HTTPS With Newline",
			input:    "https://example.com\nhello world",
			expected: " ",
		},
		{
			name:     "Email Removal",
			input:    "Contact us at admin@example.com for more info",
			expected: "contact us at for more info",
		},
		{
			name:     "HTML Tags",
			input:    "Hello <b>bold</b> and <i>italic</i>",
			expected: "hello bold and ital ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanDescription(tt.input)
			if got != tt.expected {
				t.Errorf("CleanDescription(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
