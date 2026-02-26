package mangadex

import (
	"testing"
)

func TestParameterToString(t *testing.T) {
	tests := []struct {
		name             string
		obj              interface{}
		collectionFormat string
		want             string
	}{
		{
			name: "String input",
			obj:  "hello",
			want: "hello",
		},
		{
			name: "Integer input",
			obj:  123,
			want: "123",
		},
		{
			name: "Boolean input",
			obj:  true,
			want: "true",
		},
		{
			name:             "Slice CSV",
			obj:              []string{"a", "b", "c"},
			collectionFormat: "csv",
			want:             "a,b,c",
		},
		{
			name:             "Slice SSV",
			obj:              []string{"a", "b", "c"},
			collectionFormat: "ssv",
			want:             "a b c",
		},
		{
			name:             "Slice TSV",
			obj:              []string{"a", "b", "c"},
			collectionFormat: "tsv",
			want:             "a\tb\tc",
		},
		{
			name:             "Slice Pipes",
			obj:              []string{"a", "b", "c"},
			collectionFormat: "pipes",
			want:             "a|b|c",
		},
		{
			name:             "Slice Default (empty format)",
			obj:              []string{"a", "b"},
			collectionFormat: "",
			want:             "ab", // Spaces removed
		},
		{
			name:             "Slice with spaces in elements (CSV)",
			obj:              []string{"hello world", "foo"},
			collectionFormat: "csv",
			want:             "hello,world,foo", // Known limitation: spaces within elements are replaced by delimiter
		},
		{
			name:             "Empty Slice",
			obj:              []string{},
			collectionFormat: "csv",
			want:             "",
		},
		{
			name:             "Int Slice CSV",
			obj:              []int{1, 2, 3},
			collectionFormat: "csv",
			want:             "1,2,3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parameterToString(tt.obj, tt.collectionFormat); got != tt.want {
				t.Errorf("parameterToString() = %q, want %q", got, tt.want)
			}
		})
	}
}
