package vbasedata

import (
	"testing"
	"time"
)

func TestLruCache_SetGet(t *testing.T) {
	c := NewLruCache(10, time.Minute)
	if err := c.Set("k", "v"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	// Get without clear keeps the entry.
	if got := c.Get("k", false); got != "v" {
		t.Errorf("Get = %q, want v", got)
	}
	if got := c.Get("k", false); got != "v" {
		t.Errorf("second Get = %q, want v (should not be cleared)", got)
	}
	// Get with clear removes it.
	if got := c.Get("k", true); got != "v" {
		t.Errorf("Get(clear) = %q, want v", got)
	}
	if got := c.Get("k", false); got != "" {
		t.Errorf("after clear Get = %q, want empty", got)
	}
	if got := c.Get("missing", false); got != "" {
		t.Errorf("missing key Get = %q, want empty", got)
	}
}

func TestLruCache_Incr(t *testing.T) {
	c := NewLruCache(10, time.Minute)

	// First Incr on a missing key initializes to "1".
	if err := c.Incr("n"); err != nil {
		t.Fatalf("Incr: %v", err)
	}
	if got := c.Get("n", false); got != "1" {
		t.Fatalf("after 1st Incr = %q, want 1", got)
	}
	// Subsequent Incr actually increments.
	_ = c.Incr("n")
	_ = c.Incr("n")
	if got := c.Get("n", false); got != "3" {
		t.Fatalf("after 3 Incr = %q, want 3", got)
	}

	// Incr on a non-numeric existing value resets to "1".
	_ = c.Set("bad", "abc")
	if err := c.Incr("bad"); err != nil {
		t.Fatalf("Incr(bad): %v", err)
	}
	if got := c.Get("bad", false); got != "1" {
		t.Errorf("Incr on non-numeric = %q, want 1", got)
	}
}

func TestLruCache_Verify(t *testing.T) {
	c := NewLruCache(10, time.Minute)
	_ = c.Set("id", "answer")

	if c.Verify("id", "wrong", false) {
		t.Error("Verify with wrong answer should be false")
	}
	// Wrong answer with clear=false must not consume the entry.
	if !c.Verify("id", "answer", false) {
		t.Error("Verify with correct answer should be true")
	}
	// Correct answer with clear=true consumes the entry.
	if !c.Verify("id", "answer", true) {
		t.Error("Verify(clear) with correct answer should be true")
	}
	if c.Verify("id", "answer", false) {
		t.Error("entry should have been cleared after Verify(clear)")
	}
	if c.Verify("missing", "x", false) {
		t.Error("Verify on missing key should be false")
	}
}

func TestLruCache_Expiry(t *testing.T) {
	c := NewLruCache(10, 30*time.Millisecond)
	_ = c.Set("k", "v")
	time.Sleep(60 * time.Millisecond)
	if got := c.Get("k", false); got != "" {
		t.Errorf("expected entry to expire, got %q", got)
	}
}

func TestLruCache_Eviction(t *testing.T) {
	// Capacity of 1: adding a second key evicts the first.
	c := NewLruCache(1, time.Minute)
	_ = c.Set("a", "1")
	_ = c.Set("b", "2")
	if got := c.Get("a", false); got != "" {
		t.Errorf("expected 'a' to be evicted, got %q", got)
	}
	if got := c.Get("b", false); got != "2" {
		t.Errorf("expected 'b' to remain, got %q", got)
	}
}
