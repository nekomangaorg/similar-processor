package similar_helpers

import (
	"github.com/caneroj1/stemmer"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"regexp"
	"strings"
	"unicode"
)

var (
	reg02 = regexp.MustCompile(`\[.*?]`)
	reg03 = regexp.MustCompile(`\(source: [^)]*\)`)
	reg04 = regexp.MustCompile(`\(from: [^)]*\)`)
	reg05 = regexp.MustCompile(`<[^>]*>`)
	reg06 = regexp.MustCompile(`^https?:\/\/.*[\r\n]*`)
	reg07 = regexp.MustCompile(`^http?:\/\/.*[\r\n]*`)
	reg08 = regexp.MustCompile(`[\w\.-]+@[\w\.-]+`)
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
	for i := 0; i < len(strRaw); i++ {
		char := strRaw[i]
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			b.WriteByte(char)
		} else if char == ' ' {
			if b.Len() > 0 && b.String()[b.Len()-1] != ' ' {
				b.WriteByte(' ')
			} else if b.Len() == 0 {
				b.WriteByte(' ')
			}
		}
	}

	res := b.String()
	// preserve original trailing space behavior
	if strings.HasSuffix(strRaw, " ") && !strings.HasSuffix(res, " ") {
		res += " "
	}
	return res
}

func CleanTitle(strRaw string) string {
    // Note: Do NOT use transform.String here, original code did not!
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
	for i := 0; i < len(strRaw); i++ {
		char := strRaw[i]
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			b.WriteByte(char)
		} else if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
			if b.Len() > 0 && b.String()[b.Len()-1] != ' ' {
				b.WriteByte(' ')
			} else if b.Len() == 0 {
				b.WriteByte(' ')
			}
		} else {
		    // Other unicode characters are ignored just like regexp [^a-zA-Z0-9 ]+ -> "" ignores them
		}
	}

	res := b.String()
	// preserve original trailing space behavior
	if strings.HasSuffix(strRaw, " ") && !strings.HasSuffix(res, " ") {
		res += " "
	}
	return res
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

	strRaw = reg02.ReplaceAllString(strRaw, "")
	strRaw = reg03.ReplaceAllString(strRaw, " ")
	strRaw = reg04.ReplaceAllString(strRaw, " ")
	strRaw = reg05.ReplaceAllString(strRaw, " ")

	// Remove emails and urls
	strRaw = reg06.ReplaceAllString(strRaw, " ")
	strRaw = reg07.ReplaceAllString(strRaw, " ")
	strRaw = reg08.ReplaceAllString(strRaw, " ")

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
