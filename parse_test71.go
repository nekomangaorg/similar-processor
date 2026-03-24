package main

import (
	"fmt"
	"strings"
	"regexp"
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

func main() {
    strRaw := "Nested [a [b] c] text"

    // BBCodes stripping
    for _, tag := range BBCodes {
		strRaw = strings.ReplaceAll(strRaw, tag, "")
	}

    fmt.Printf("After BBCodes: %q\n", strRaw)

    // Then reg02:
    reg02 := regexp.MustCompile(`\[.*?]`)
    strRaw = reg02.ReplaceAllString(strRaw, "")

    fmt.Printf("After reg02: %q\n", strRaw)
}
