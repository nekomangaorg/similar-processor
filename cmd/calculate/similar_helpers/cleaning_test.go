package similar_helpers

import (
	"testing"
)

func TestCleanDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic Clean",
			input:    "This is a test description.",
			expected: "thi is a test descript",
		},
		{
			name:     "Foreign Language Truncation - Spanish",
			input:    "English part. Spanish:[spoiler] Spanish part.",
			expected: "english part ",
		},
		{
			name:     "Foreign Language Truncation - French",
			input:    "English part. \n\nFrench part starts here.",
			expected: "english part ",
		},
		{
			name:     "Unicode Normalization",
			input:    "Café, jalapeño, résumé.",
			expected: "cafe jalapeno resum",
		},
		{
			name:     "Newline Handling",
			input:    "Line 1.\nLine 2.\r\nLine 3.",
			expected: "line 1 line 2 line 3",
		},
		{
			name:     "English Tag Removal",
			input:    "[b][u]English: This is the description.",
			expected: " thi is the descript",
		},
		{
			name:     "BBCode Removal",
			input:    "This is [b]bold[/b] and [u]underlined[/u].",
			expected: "thi is bold and underlin",
		},
		{
			name:     "Bracket Removal",
			input:    "Description with [useless info].",
			expected: "descript with ",
		},
		{
			name:     "Source Removal",
			input:    "Description (source: mangaupdates).",
			expected: "descript ",
		},
		{
			name:     "HTML Removal",
			input:    "<p>Paragraph 1</p><br>Paragraph 2",
			expected: " paragraph 1 paragraph 2",
		},
		{
			name:     "URL Removal - Middle",
			input:    "Check http://example.com and https://test.org for more.",
			expected: "check httpexamplecom and httpstestorg for more",
		},
		{
			name:     "URL Removal - Start",
			input:    "http://example.com is a link.",
			expected: " ", // regex ^http.* removes everything
		},
		{
			name:     "Email Removal",
			input:    "Contact us at support@example.com.",
			expected: "contact us at ",
		},
		{
			name:     "Contraction Expansion",
			input:    "I don't think it's fair.",
			expected: "i do not think it is fair",
		},
		{
			name:     "Symbol Removal",
			input:    "Hello! How are you? @#$%^",
			expected: "hello how ar you ",
		},
		{
			name:     "Multiple Spaces",
			input:    "Too    many     spaces.",
			expected: "too mani space",
		},
		{
			name:     "Stemming Check",
			input:    "Running walking played categories",
			expected: "run wal plai categori",
		},
		{
			name:     "Mixed Case",
			input:    "MiXeD CaSe TeXt",
			expected: "mix case text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanDescription(tt.input)
			if got != tt.expected {
				t.Errorf("CleanDescription(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
