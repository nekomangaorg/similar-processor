package mangadex

import (
	"fmt"
	"testing"
)

func TestParameterToString(t *testing.T) {
	tests := []struct {
		obj              interface{}
		collectionFormat string
		want             string
	}{
		{"hello", "", "hello"},
		{123, "", "123"},
		{true, "", "true"},
		{[]string{"a", "b"}, "csv", "a,b"},
		{[]string{"a", "b"}, "ssv", "a b"},
		{[]string{"a", "b"}, "tsv", "a\tb"},
		{[]string{"a", "b"}, "pipes", "a|b"},
		{[]string{"a", "b"}, "", "ab"},
		{[]int{1, 2}, "csv", "1,2"},
		{[]string{"a b", "c"}, "csv", "a b,c"},
		{[]string{"a b", "c"}, "ssv", "a b c"},
		{nil, "", ""},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v/%s", tt.obj, tt.collectionFormat), func(t *testing.T) {
			if got := parameterToString(tt.obj, tt.collectionFormat); got != tt.want {
				t.Errorf("parameterToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
