package main

import (
	"fmt"
	"regexp"
)

var reg02 = regexp.MustCompile(`\[.*?]`)

func main() {
	// What if there is a tag NOT in BBCodes?
	// Like `[a [nested] c]`
	input := "Nested [a [nested] c] text"
	fmt.Printf("Regex: %q\n", reg02.ReplaceAllString(input, ""))

	// Output of Regex: "Nested  c] text"
	// Output of my code: "Nested  c] text"

	// SO MY CODE IS PERFECTLY MATCHING REGEX!!
}
