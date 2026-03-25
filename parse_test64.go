package main

import (
	"fmt"
	"regexp"
	"strings"
)

var reg02 = regexp.MustCompile(`\[.*?]`)

func main() {
	input := "Nested [a [b] c] text"

	// In real code, BBCodes are stripped BEFORE reg02.
	// BBCodes stripping:
	input = strings.ReplaceAll(input, "[b]", "")

	// So `Nested [a [b] c] text` -> `Nested [a  c] text`
	fmt.Printf("After BBCodes: %q\n", input)
	fmt.Printf("Regex: %q\n", reg02.ReplaceAllString(input, ""))
}
