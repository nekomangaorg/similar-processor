package calculate

type customMatch struct {
	ID                                  int
	Distance, DistanceTag, DistanceDesc float64
}

type MatchMinHeap []customMatch

func (h MatchMinHeap) Len() int           { return len(h) }
func (h MatchMinHeap) Less(i, j int) bool { return h[i].Distance < h[j].Distance }
func (h MatchMinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *MatchMinHeap) Push(x any)        { *h = append(*h, x.(customMatch)) }
func (h *MatchMinHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
