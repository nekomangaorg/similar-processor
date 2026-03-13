package internal

import (
	"testing"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint64
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "Zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "One",
			input:    "1",
			expected: 1,
		},
		{
			name:     "Nine",
			input:    "9",
			expected: 9,
		},
		{
			name:     "Ten (A)",
			input:    "A",
			expected: 10,
		},
		{
			name:     "Thirty-five (Z)",
			input:    "Z",
			expected: 35,
		},
		{
			name:     "Lowercase a",
			input:    "a",
			expected: 10,
		},
		{
			name:     "Lowercase z",
			input:    "z",
			expected: 35,
		},
		{
			name:     "Base36 10",
			input:    "10",
			expected: 36,
		},
		{
			name:     "Base36 ZZ",
			input:    "ZZ",
			expected: 1295, // 35*36 + 35
		},
		{
			name:     "Length 12 (1 followed by 11 zeros)",
			input:    "100000000000",
			expected: 131621703842267136, // 36^11 from pow36Index[11]
		},
		{
			name:     "Length 13 truncated to 12 (1 followed by 12 zeros)",
			input:    "1000000000000",
			expected: 131621703842267136, // effectively "100000000000" -> 36^11
		},
		{
			name:     "Length 14 truncated to 12 (1 followed by 13 zeros)",
			input:    "10000000000000",
			expected: 131621703842267136, // effectively "100000000000" -> 36^11
		},
		{
			name:     "Invalid character (space)",
			input:    " ",
			expected: 0,
		},
		{
			name:     "Invalid character mixed with valid",
			input:    "1!",
			expected: 36, // '1' -> 1*36, '!' -> 0
		},
		{
			name:     "Invalid character mixed with valid 2",
			input:    "!1",
			expected: 1, // '!' -> 0*36, '1' -> 1
		},
		{
			name:     "Invalid characters (+)",
			input:    "+",
			expected: 0,
		},
		{
			name:     "Invalid characters (/)",
			input:    "/",
			expected: 0,
		},
		{
			name:     "Invalid characters (=)",
			input:    "=",
			expected: 0,
		},
		{
			name:     "Multiple invalid characters",
			input:    "!@#$",
			expected: 0,
		},
		{
			name:     "Non-ASCII characters (ñ)",
			input:    "ñ",
			expected: 0, // The implementation iterates over bytes, so multi-byte characters are processed byte-by-byte.
		},
		{
			name:     "Non-ASCII characters (本)",
			input:    "本",
			expected: 0,
		},
		{
			name:     "Non-ASCII characters (😊)",
			input:    "😊",
			expected: 0,
		},
		{
			name:     "Control characters (newline)",
			input:    "\n",
			expected: 0,
		},
		{
			name:     "Control characters (tab)",
			input:    "\t",
			expected: 0,
		},
		{
			name:     "Length 15 truncated to 12",
			input:    "1234567890AB CD", // length 15
			expected: 139250390350663091, // "1234567890AB" in base36
		},
		{
			name:     "Length 13 truncated to 12 (potential overflow)",
			input:    "5000000000000",
			expected: 658108519211335680, // "500000000000" in base36
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decode(tt.input)
			if got != tt.expected {
				t.Errorf("Decode(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
