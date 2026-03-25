package main

import (
	"fmt"
)

func filterTextTagsOrig(strRaw string) string {
	b := make([]byte, 0, len(strRaw))
	i := 0

	for i < len(strRaw) {
		// [.*?] with nested tracking
		if strRaw[i] == '[' {
			depth := 1
			j := i + 1
			newlineFound := false
			for ; j < len(strRaw); j++ {
				if strRaw[j] == '[' {
					depth++
				} else if strRaw[j] == ']' {
					depth--
					if depth == 0 {
						break
					}
				} else if strRaw[j] == '\n' {
					newlineFound = true
					break
				}
			}
			if depth == 0 && !newlineFound {
				i = j + 1
				continue
			}
		}
		b = append(b, strRaw[i])
		i++
	}

	return string(b)
}

func main() {
	input := "[a [b] c]"
	fmt.Printf("nested counter: %q\n", filterTextTagsOrig(input))

	// Wait, the comment says: "The old regex `\\[.*?]` would match `[b]`, resulting in `[a  c]`."
	// If I use a counter, `[a [b] c]` has depth=1 at `[`, then depth=2 at `[b`, then depth=1 at `]`, then depth=0 at `]`.
	// It would remove THE ENTIRE `[a [b] c]` string! Resulting in `""`.
	// But the comment specifically expects `[a  c]` !!
}
