package similar_helpers

import (
	"github.com/caneroj1/stemmer"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"strings"
	"unicode"
)

var apostropheReplacer = strings.NewReplacer(
	"isn't", "is not",
	"aren't", "are not",
	"ain't", "am not",
	"won't", "will not",
	"didn't", "did not",
	"shan't", "shall not",
	"haven't", "have not",
	"hadn't", "had not",
	"hasn't", "has not",
	"don't", "do not",
	"wasn't", "was not",
	"weren't", "were not",
	"doesn't", "does not",
	"'s", " is",
	"'re", " are",
	"'m", " am",
	"'d", " would",
	"'ll", " will",
	"\r", " ",
	"\n", " ",
)

var transformer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

func fastCleanTitle(strRaw string) string {
	var b strings.Builder
	b.Grow(len(strRaw))
	var lastChar byte
	for i := 0; i < len(strRaw); i++ {
		char := strRaw[i]
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			b.WriteByte(char)
			lastChar = char
		} else if char == ' ' {
			if b.Len() == 0 || lastChar != ' ' {
				b.WriteByte(' ')
				lastChar = ' '
			}
		}
	}

	// preserve original trailing space behavior
	if strings.HasSuffix(strRaw, " ") && (b.Len() == 0 || lastChar != ' ') {
		b.WriteByte(' ')
	}
	return b.String()
}

func CleanTitle(strRaw string) string {
	strRaw = fastCleanTitle(strRaw)

	if strRaw == "" || strRaw == " " {
		return strRaw
	}

	// Stemming (porter)
	strRawArray := strings.Split(strRaw, " ")
	stemmer.StemMultipleMutate(&strRawArray)
	strRaw = strings.Join(strRawArray, " ")

	// Stemmer makes upper, so we need to lowercase again
	strRaw = strings.ToLower(strRaw)

	return strRaw
}

func fastCleanDescription(strRaw string) string {
	var b strings.Builder
	b.Grow(len(strRaw))
	var lastChar byte
	for i := 0; i < len(strRaw); i++ {
		char := strRaw[i]
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			b.WriteByte(char)
			lastChar = char
		} else if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
			if b.Len() == 0 || lastChar != ' ' {
				b.WriteByte(' ')
				lastChar = ' '
			}
		} else {
			// Other unicode characters are ignored just like regexp [^a-zA-Z0-9 ]+ -> "" ignores them
		}
	}

	// preserve original trailing space behavior
	if strings.HasSuffix(strRaw, " ") && (b.Len() == 0 || lastChar != ' ') {
		b.WriteByte(' ')
	}
	return b.String()
}

func isEmailChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '_'
}

func filterTextTags(strRaw string) string {
	if strings.HasPrefix(strRaw, "http://") || strings.HasPrefix(strRaw, "https://") {
		idx := strings.IndexByte(strRaw, '\n')
		if idx != -1 {
			strRaw = " " + strRaw[idx+1:]
		} else {
			strRaw = " "
		}
	}

	b := make([]byte, 0, len(strRaw))
	i := 0
	for i < len(strRaw) {
		// [.*?]
		if strRaw[i] == '[' {
			end := -1
			for j := i + 1; j < len(strRaw); j++ {
				if strRaw[j] == ']' {
					end = j
					break
				}
			}
			if end != -1 {
				i = end + 1
				continue
			}
		}

		// (source: ... ) or (from: ... )
		if strRaw[i] == '(' {
			if strings.HasPrefix(strRaw[i:], "(source: ") || strings.HasPrefix(strRaw[i:], "(from: ") {
				end := -1
				for j := i + 1; j < len(strRaw); j++ {
					if strRaw[j] == ')' {
						end = j
						break
					}
				}
				if end != -1 {
					b = append(b, ' ')
					i = end + 1
					continue
				}
			}
		}

		// <[^>]*>
		if strRaw[i] == '<' {
			end := -1
			for j := i + 1; j < len(strRaw); j++ {
				if strRaw[j] == '>' {
					end = j
					break
				}
			}
			if end != -1 {
				b = append(b, ' ')
				i = end + 1
				continue
			}
		}

		// Email [\w\.-]+@[\w\.-]+
		if strRaw[i] == '@' {
			start := len(b) - 1
			validEmail := false
			for start >= 0 {
				if isEmailChar(b[start]) {
					start--
					validEmail = true
				} else {
					break
				}
			}
			start++ // first valid char

			end := i + 1
			validEnd := false
			for end < len(strRaw) {
				if isEmailChar(strRaw[end]) {
					end++
					validEnd = true
				} else {
					break
				}
			}

			if validEmail && validEnd {
				b = b[:start]
				b = append(b, ' ')
				i = end
				continue
			}
		}

		b = append(b, strRaw[i])
		i++
	}

	return string(b)
}

func CleanDescription(strRaw string) string {
	// Remove all non-english descriptions
	for _, tag := range DescriptionLanguages {
		if idx := strings.Index(strRaw, tag); idx != -1 {
			strRaw = strRaw[:idx]
		}
	}

	// Remove "rune" / umlauts / diacritics
	strRaw, _, _ = transform.String(transformer, strRaw)

	// To lowercase
	strRaw = strings.ToLower(strRaw)

	// Replace new lines with space and standard lexicons
	strRaw = apostropheReplacer.Replace(strRaw)

	// Now remove all english tags which are no longer needed
	for _, tag := range EnglishDescriptionTags {
		strRaw = strings.ReplaceAll(strRaw, tag, "")
	}

	// Next clean the string from any bbcodes
	for _, tag := range BBCodes {
		strRaw = strings.ReplaceAll(strRaw, tag, "")
	}

	strRaw = filterTextTags(strRaw)

	// Remove all symbols (clean to normal english) and collapse spaces
	strRaw = fastCleanDescription(strRaw)

	if strRaw == "" || strRaw == " " {
		return strRaw
	}

	// Stemming (porter)
	strRawArray := strings.Split(strRaw, " ")
	stemmer.StemMultipleMutate(&strRawArray)
	strRaw = strings.Join(strRawArray, " ")

	// To lowercase (again since stemmer makes upper)
	strRaw = strings.ToLower(strRaw)

	return strRaw
}
