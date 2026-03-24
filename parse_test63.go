package main

import (
	"fmt"
	"strings"
	"regexp"
)

var reg02 = regexp.MustCompile(`\[.*?]`)

func mycleanBytesFull(strRaw string) string {
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
    input := "[a [b] c]"
    fmt.Printf("regex: %q\n", reg02.ReplaceAllString(input, ""))
    fmt.Printf("mine: %q\n", mycleanBytesFull(input))
}
