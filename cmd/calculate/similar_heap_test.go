package calculate

import (
	"container/heap"
	"testing"
)

func TestMatchMinHeap(t *testing.T) {
	h := &MatchMinHeap{}
	heap.Init(h)

	// Test Len() on empty heap
	if h.Len() != 0 {
		t.Errorf("Expected Len 0, got %d", h.Len())
	}

	// Test Push and Len
	matches := []customMatch{
		{ID: 1, Distance: 0.5},
		{ID: 2, Distance: 0.1},
		{ID: 3, Distance: 0.8},
		{ID: 4, Distance: 0.3},
	}

	for _, m := range matches {
		heap.Push(h, m)
	}

	if h.Len() != len(matches) {
		t.Errorf("Expected Len %d, got %d", len(matches), h.Len())
	}

	// Test Pop order (Min-Heap: smallest distance first)
	expectedDistances := []float64{0.1, 0.3, 0.5, 0.8}
	for _, expected := range expectedDistances {
		if h.Len() == 0 {
			t.Fatal("Heap unexpectedly empty")
		}
		m := heap.Pop(h).(customMatch)
		if m.Distance != expected {
			t.Errorf("Expected distance %f, got %f", expected, m.Distance)
		}
	}

	if h.Len() != 0 {
		t.Errorf("Expected Len 0 after all pops, got %d", h.Len())
	}
}

func TestMatchMinHeap_Less(t *testing.T) {
	h := MatchMinHeap{
		{ID: 1, Distance: 0.5},
		{ID: 2, Distance: 0.1},
	}

	if !h.Less(1, 0) {
		t.Error("h[1].Distance (0.1) should be less than h[0].Distance (0.5)")
	}
	if h.Less(0, 1) {
		t.Error("h[0].Distance (0.5) should NOT be less than h[1].Distance (0.1)")
	}
}

func TestMatchMinHeap_Swap(t *testing.T) {
	h := MatchMinHeap{
		{ID: 1, Distance: 0.5},
		{ID: 2, Distance: 0.1},
	}

	h.Swap(0, 1)

	if h[0].ID != 2 || h[1].ID != 1 {
		t.Errorf("Swap failed: h[0].ID=%d, h[1].ID=%d", h[0].ID, h[1].ID)
	}
}
