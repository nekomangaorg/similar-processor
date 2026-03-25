package main

import (
	"fmt"
	"regexp"
)

var reg02 = regexp.MustCompile(`\[.*?]`)

func main() {
	input := "[a [b] c]"
	fmt.Printf("regex on nested: %q\n", reg02.ReplaceAllString(input, ""))

	// Is it possible the original regex was `\[[^\]]*\]` ?
	// In `cleaning.go`:
	// reg02 = regexp.MustCompile(`\[.*?]`)
	// Yes, it was `\[.*?]` !!
}
