package vbasedata

import (
	"testing"
)

func TestIdgenerator_Unique(t *testing.T) {
	g := NewIdgenerator(1)

	const n = 5000
	seen := make(map[int64]struct{}, n)
	for i := 0; i < n; i++ {
		id := g.NextId()
		if id <= 0 {
			t.Fatalf("NextId returned non-positive id: %d", id)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate id generated: %d", id)
		}
		seen[id] = struct{}{}
	}
}

func TestIdgenerator_Monotonic(t *testing.T) {
	g := NewIdgenerator(2)
	prev := g.NextId()
	for i := 0; i < 1000; i++ {
		id := g.NextId()
		if id <= prev {
			t.Fatalf("ids not strictly increasing: prev=%d cur=%d", prev, id)
		}
		prev = id
	}
}
