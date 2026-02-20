package internal

import (
	"testing"
	"time"
)

func TestProgressBar_Update(t *testing.T) {
	total := 100
	bar := NewProgressBar(total, "items")

	// Basic check that Update doesn't panic
	for i := 0; i <= total; i += 10 {
		bar.Update(i)
		// Sleep a tiny bit to allow rate calculation to be non-zero (though not strictly required for no panic)
		time.Sleep(1 * time.Millisecond)
	}
	bar.Finish()

	if bar.Current != total {
		t.Errorf("Expected current to be %d, got %d", total, bar.Current)
	}
}

func TestProgressBar_Formatting(t *testing.T) {
    // This test just ensures no runtime errors during formatting with edge cases
    bar := NewProgressBar(100, "items")
    bar.Update(0)
    bar.Update(50)
    bar.Update(100)
    bar.Finish()
}
