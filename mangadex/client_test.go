package mangadex

import (
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		haystack []string
		needle   string
		want     bool
	}{
		{
			name:     "Exact match",
			haystack: []string{"foo", "bar"},
			needle:   "bar",
			want:     true,
		},
		{
			name:     "Case-insensitive match",
			haystack: []string{"foo", "BAR"},
			needle:   "bar",
			want:     true,
		},
		{
			name:     "No match",
			haystack: []string{"foo", "bar"},
			needle:   "baz",
			want:     false,
		},
		{
			name:     "Empty haystack",
			haystack: []string{},
			needle:   "foo",
			want:     false,
		},
		{
			name:     "Empty needle in haystack",
			haystack: []string{"foo", ""},
			needle:   "",
			want:     true,
		},
		{
			name:     "Empty needle in empty haystack",
			haystack: []string{},
			needle:   "",
			want:     false,
		},
		{
			name:     "Nil haystack",
			haystack: nil,
			needle:   "foo",
			want:     false,
		},
		{
			name:     "Unicode match",
			haystack: []string{"Ångström", "Go"},
			needle:   "ångström",
			want:     true,
		},
		{
			name:     "Match with whitespace",
			haystack: []string{" foo ", "bar"},
			needle:   " FOO ",
			want:     true,
		},
		{
			name:     "Partial match fail",
			haystack: []string{"foobar"},
			needle:   "foo",
			want:     false,
		},
		{
			name:     "Whitespace mismatch",
			haystack: []string{" foo "},
			needle:   "foo",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.haystack, tt.needle); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
