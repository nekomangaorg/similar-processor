package main

import (
	"fmt"
	"regexp"
)

func main() {
    reg02 := regexp.MustCompile(`\[.*?]`)
    input := "[a [b] c]"
    fmt.Printf("regex: %q\n", reg02.ReplaceAllString(input, ""))

    // Oh wait, did the reviewer mean `\[.*?\]` on `[a [b] c]` ?
    // In Python, PHP, JS, `\[.*?\]` on `[a [b] c]` MATCHES `[a [b]` and leaves ` c]`.
    // The only way `\[.*?\]` matches `[b]` is if it's evaluated FROM `[b]`, which it isn't, because it finds the FIRST `[`.

    // What if the reviewer is just mistaken about regex behavior?
    // Reviewer: "For example, on the input `[a [b] c]` ... The old regex `\[.*?]` would match `[b]`, resulting in `[a  c]`."
    // Let me try to make it work EXACTLY as the reviewer asked:
    // If I see `[`, and then ANOTHER `[` before `]`, it means the first `[` doesn't match? No, that's not how regex works.

    // Wait! The reviewer says: "The expectation of `"nes text"` implies a greedy match (`\[.*]`), which is not what the implementation or the original regex does. Please adjust the test case to reflect the intended and correct behavior."
    // Actually, "Nested [a [b] c] text".
    // 1. BBCodes strips `[b]` -> "Nested [a  c] text".
    // 2. reg02 strips `[a  c]` -> "Nested  text".
    // 3. fastClean strips `  ` -> "nest text"

    // WAIT!!
    // In `Nested [a [b] c] text`, `[b]` is stripped by BBCodes!!
    // BBCodes list contains `[b]`!
    // `strRaw = strings.ReplaceAll(strRaw, "[b]", "")`
    // SO `Nested [a [b] c] text` becomes `Nested [a ] c] text` !!

    // Let's test this exactly!
}
