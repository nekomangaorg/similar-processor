package calculate

import (
	"math/rand"
	"testing"
	"time"

	"github.com/similar-manga/similar/internal"
)

// Helper to mimic the old logic exactly
func isInvalidOld(current, target internal.Manga) bool {
	common := false
	for _, l1 := range current.AvailableTranslatedLanguages {
		for _, l2 := range target.AvailableTranslatedLanguages {
			if l1 == l2 {
				common = true
				break
			}
		}
		if common {
			break
		}
	}
	if !common && len(current.AvailableTranslatedLanguages) > 0 {
		return true // INVALID
	}
	return false // VALID
}

// Helper to mimic the new logic exactly
func isInvalidNew(currentMask, targetMask uint64) bool {
	if currentMask != 0 && (currentMask&targetMask) == 0 {
		return true // INVALID (SKIP)
	}
	return false // VALID (DON'T SKIP)
}

func TestBitmaskOptimizationCorrectness(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// Test case scenarios
	scenarios := []struct {
		desc    string
		current []string
		target  []string
	}{
		{"Both Empty", []string{}, []string{}},
		{"Current Empty", []string{}, []string{"en"}},
		{"Target Empty", []string{"en"}, []string{}},
		{"Match Single", []string{"en"}, []string{"en"}},
		{"Match Multiple", []string{"en", "fr"}, []string{"fr", "de"}},
		{"No Match Disjoint", []string{"en"}, []string{"fr"}},
		{"No Match Complex", []string{"en", "es"}, []string{"fr", "de"}},
		{"Match Overflow", []string{"rare1"}, []string{"rare1"}}, // Assuming overflow logic handles this conservatively
		{"No Match Overflow vs Normal", []string{"rare1"}, []string{"en"}},
	}

	// Dynamic test with random languages to cover overflow
	languages := []string{}
	for i := 0; i < 100; i++ {
		languages = append(languages, string(rune('a'+(i%26)))+string(rune('a'+(i/26))))
	}

	// Setup bitmasks (global context)
	uniqueLangs := make(map[string]uint64)
	nextBit := 0

	// Add scenario languages to unique map first
	for _, sc := range scenarios {
		for _, l := range sc.current {
			if _, exists := uniqueLangs[l]; !exists {
				if nextBit < 63 {
					uniqueLangs[l] = 1 << nextBit
					nextBit++
				} else {
					uniqueLangs[l] = 1 << 63
				}
			}
		}
		for _, l := range sc.target {
			if _, exists := uniqueLangs[l]; !exists {
				if nextBit < 63 {
					uniqueLangs[l] = 1 << nextBit
					nextBit++
				} else {
					uniqueLangs[l] = 1 << 63
				}
			}
		}
	}

	// Add all random languages to unique map
	for _, l := range languages {
		if _, exists := uniqueLangs[l]; !exists {
			if nextBit < 63 {
				uniqueLangs[l] = 1 << nextBit
				nextBit++
			} else {
				uniqueLangs[l] = 1 << 63
			}
		}
	}

	// Helper to get mask
	getMask := func(langs []string) uint64 {
		var mask uint64
		for _, l := range langs {
			if val, ok := uniqueLangs[l]; ok {
				mask |= val
			}
		}
		return mask
	}

	// Run scenarios
	for _, sc := range scenarios {
		currManga := internal.Manga{AvailableTranslatedLanguages: sc.current}
		targManga := internal.Manga{AvailableTranslatedLanguages: sc.target}

		currMask := getMask(sc.current)
		targMask := getMask(sc.target)

		oldInv := isInvalidOld(currManga, targManga)
		newInv := isInvalidNew(currMask, targMask)

		// Verification Rule:
		// If newInv is true (SKIP), oldInv MUST be true (INVALID).
		// If newInv is false (DON'T SKIP), oldInv CAN be true (INVALID) [False Negative for Optimization, safe]
		// The error case is: newInv=True (SKIP) but oldInv=False (VALID). This means we skipped a valid match.

		if newInv && !oldInv {
			t.Errorf("Scenario '%s' failed! Optimization skipped a valid match.\nCurrent: %v\nTarget: %v\nMasks: %b vs %b",
				sc.desc, sc.current, sc.target, currMask, targMask)
		}
	}

	// Run random fuzzy tests
	for i := 0; i < 10000; i++ {
		// Generate random subsets of languages
		currLangs := getRandomSubset(languages)
		targLangs := getRandomSubset(languages)

		currManga := internal.Manga{AvailableTranslatedLanguages: currLangs}
		targManga := internal.Manga{AvailableTranslatedLanguages: targLangs}

		currMask := getMask(currLangs)
		targMask := getMask(targLangs)

		oldInv := isInvalidOld(currManga, targManga)
		newInv := isInvalidNew(currMask, targMask)

		if newInv && !oldInv {
			t.Errorf("Random test failed! Optimization skipped a valid match.\nCurrent: %v\nTarget: %v\nMasks: %b vs %b",
				currLangs, targLangs, currMask, targMask)
		}
	}
}

func getRandomSubset(langs []string) []string {
	n := rand.Intn(5) // 0 to 4 langs
	subset := make([]string, 0, n)
	for i := 0; i < n; i++ {
		subset = append(subset, langs[rand.Intn(len(langs))])
	}
	return subset
}
