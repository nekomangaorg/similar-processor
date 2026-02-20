package internal

import (
	"fmt"
	"strings"
	"time"
)

// ProgressBar displays a progress bar in the terminal
type ProgressBar struct {
	Total      int
	Current    int
	StartTime  time.Time
	Unit       string
	Width      int
	lastUpdate time.Time
}

// NewProgressBar creates a new ProgressBar
func NewProgressBar(total int, unit string) *ProgressBar {
	return &ProgressBar{
		Total:      total,
		Unit:       unit,
		StartTime:  time.Now(),
		Width:      40,
		lastUpdate: time.Time{},
	}
}

// Update updates the progress bar with the current count
func (p *ProgressBar) Update(current int) {
	p.Current = current
	now := time.Now()

	// Throttle updates to avoid flickering, unless it's the last one
	if current < p.Total && now.Sub(p.lastUpdate) < 100*time.Millisecond {
		return
	}
	p.lastUpdate = now

	percent := 0.0
	if p.Total > 0 {
		percent = float64(current) / float64(p.Total) * 100
	}
	if percent > 100 {
		percent = 100
	}

	completed := 0
	if p.Total > 0 {
		completed = int(float64(p.Width) * (float64(current) / float64(p.Total)))
	}
	if completed > p.Width {
		completed = p.Width
	}
	if completed < 0 {
		completed = 0
	}

	bar := strings.Repeat("=", completed) + strings.Repeat("-", p.Width-completed)

	elapsed := now.Sub(p.StartTime)
	rate := 0.0
	if elapsed.Seconds() > 0 {
		rate = float64(current) / elapsed.Seconds()
	}

	var eta time.Duration
	if current > 0 && rate > 0 {
		remaining := p.Total - current
		eta = time.Duration(float64(remaining)/rate) * time.Second
	}

	// Clear line and print
	// Format: [=====>....] 25.0% (250/1000) | 12.5 manga/s | ETA: 1m0s
	fmt.Printf("\r[%s] %.1f%% (%d/%d) | %.2f %s/s | ETA: %v   ", bar, percent, current, p.Total, rate, p.Unit, eta.Round(time.Second))
}

// Finish completes the progress bar output
func (p *ProgressBar) Finish() {
	fmt.Println()
}
