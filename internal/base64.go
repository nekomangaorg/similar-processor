package internal

/*
Base36 implementation in golang
https://github.com/martinlindhe/base36

The MIT License (MIT)

Copyright (c) 2015-2021 Martin Lindhe

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

var (
	base36 = []byte{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
		'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T',
		'U', 'V', 'W', 'X', 'Y', 'Z'}

	//index = map[byte]int{
	//	'0': 0, '1': 1, '2': 2, '3': 3, '4': 4,
	//	'5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
	//	'A': 10, 'B': 11, 'C': 12, 'D': 13, 'E': 14,
	//	'F': 15, 'G': 16, 'H': 17, 'I': 18, 'J': 19,
	//	'K': 20, 'L': 21, 'M': 22, 'N': 23, 'O': 24,
	//	'P': 25, 'Q': 26, 'R': 27, 'S': 28, 'T': 29,
	//	'U': 30, 'V': 31, 'W': 32, 'X': 33, 'Y': 34,
	//	'Z': 35,
	//	'a': 10, 'b': 11, 'c': 12, 'd': 13, 'e': 14,
	//	'f': 15, 'g': 16, 'h': 17, 'i': 18, 'j': 19,
	//	'k': 20, 'l': 21, 'm': 22, 'n': 23, 'o': 24,
	//	'p': 25, 'q': 26, 'r': 27, 's': 28, 't': 29,
	//	'u': 30, 'v': 31, 'w': 32, 'x': 33, 'y': 34,
	//	'z': 35,
	//}
	uint8Index = []uint64{
		0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 1, 2,
		3, 4, 5, 6, 7, 8, 9, 0, 0, 0,
		0, 0, 0, 0, 10, 11, 12, 13, 14,
		15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
		25, 26, 27, 28, 29, 30, 31, 32, 33, 34,
		35, 0, 0, 0, 0, 0, 0, 10, 11, 12, 13,
		14, 15, 16, 17, 18, 19, 20, 21, 22, 23,
		24, 25, 26, 27, 28, 29, 30, 31, 32, 33,
		34, 35, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, // 256
	}
	pow36Index = []uint64{
		1, 36, 1296, 46656, 1679616, 60466176,
		2176782336, 78364164096, 2821109907456,
		101559956668416, 3656158440062976,
		131621703842267136, 4738381338321616896,
		9223372036854775808,
	}
)

// Decode decodes a base36-encoded string.
func Decode(s string) uint64 {
	if len(s) > 13 {
		s = s[:12]
	}
	res := uint64(0)
	l := len(s) - 1
	for idx := 0; idx < len(s); idx++ {
		c := s[l-idx]
		res += uint8Index[c] * pow36Index[idx]
	}
	return res
}
