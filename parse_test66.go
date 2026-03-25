package main

import (
	"fmt"
	"regexp"
)

var reg02 = regexp.MustCompile(`\[.*?]`)

func main() {
	input := "[a [b] c]"
	fmt.Printf("regex on original: %q\n", reg02.ReplaceAllString(input, ""))
}
