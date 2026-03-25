package main

import (
	"fmt"
	"strings"
)

var BBCodes = []string{
	"[list]",
	"[/list]",
	"[*]",
	"[hr]",
	"[u]",
	"[/u]",
	"[b]",
	"[/b]",
}

func filterTextTags(strRaw string) string {
	b := make([]byte, 0, len(strRaw))
	i := 0

	for i < len(strRaw) {
		// [.*?]
		if strRaw[i] == '[' {
			end := strings.IndexByte(strRaw[i+1:], ']')
			if end != -1 {
				newlineIdx := strings.IndexByte(strRaw[i+1:i+1+end], '\n')
				if newlineIdx == -1 {
					i = i + 1 + end + 1
					continue
				}
			}
		}
		b = append(b, strRaw[i])
		i++
	}

	return string(b)
}

func main() {
	strRaw := "Nested [a [b] c] text"

	// BBCodes stripping
	for _, tag := range BBCodes {
		strRaw = strings.ReplaceAll(strRaw, tag, "")
	}

	fmt.Printf("After BBCodes: %q\n", strRaw)

	// Then my code:
	strRaw = filterTextTags(strRaw)

	fmt.Printf("After filterTextTags: %q\n", strRaw)
}
