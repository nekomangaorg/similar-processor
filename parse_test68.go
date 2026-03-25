package main

import (
	"fmt"
	"regexp"
)

var reg02 = regexp.MustCompile(`\[.*?]`)

func main() {
	input := "[a [b] c]"
	fmt.Printf("regex on nested: %q\n", reg02.ReplaceAllString(input, ""))

	// Ah! Wait, if the reviewer explicitly said:
	// "The old regex \[.*?] would match [b], resulting in [a  c]."
	// The reviewer is mathematically wrong about Go's regex engine.
	// However, I should probably implement the "correct" logical behavior they WANT:
	// "To fix this, you might need a more sophisticated parsing logic, perhaps using a counter to handle nested brackets correctly to replicate the original regex behavior."
	// Wait, the comment says: "replicate the original regex behavior."
	// BUT the original regex behavior IS skipping from the first `[` to the first `]` !

	// Maybe they MEAN to say "replicate the intended non-greedy behavior" which means finding inner-most brackets or handling nesting?
	// Wait, if they say "using a counter to handle nested brackets correctly", they WANT nested bracket handling!
	// So I should implement a nested bracket counter.
}
