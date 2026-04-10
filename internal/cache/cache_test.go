package cache

import (
	"testing"
	"time"
)

func TestNew_StoresAndRetrieves(t *testing.T) {
	c := New(time.Second, 0)
	c.Set("key", "value")
	v, ok := c.Get("key")
	if !ok {
		t.Fatal("expected key to be present")
	}
	if v.(string) != "value" {
		t.Fatalf("expected 'value', got %v", v)
	}
}

func TestGet_MissingKey(t *testing.T) {
	c := New(time.Second, 0)
	_, ok := c.Get("missing")
	if ok {
		t.Fatal("expected miss for unknown key")
	}
}

func TestGet_ExpiredEntry(t *testing.T) {
	c := New(10*time.Millisecond, 0)
	c.Set("k", 42)
	time.Sleep(20 * time.Millisecond)
	_, ok := c.Get("k")
	if ok {
		t.Fatal("expected expired entry to be a miss")
	}
}

func TestDelete_RemovesKey(t *testing.T) {
	c := New(time.Second, 0)
	c.Set("k", "v")
	c.Delete("k")
	_, ok := c.Get("k")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestLen_ReturnsCount(t *testing.T) {
	c := New(time.Second, 0)
	if c.Len() != 0 {
		t.Fatalf("expected 0, got %d", c.Len())
	}
	c.Set("a", 1)
	c.Set("b", 2)
	if c.Len() != 2 {
		t.Fatalf("expected 2, got %d", c.Len())
	}
}

func TestGC_EvictsExpiredEntries(t *testing.T) {
	c := New(10*time.Millisecond, 20*time.Millisecond)
	defer c.Stop()
	c.Set("x", "y")
	time.Sleep(50 * time.Millisecond)
	// After GC the internal map should be empty.
	if c.Len() != 0 {
		t.Fatalf("expected 0 after GC, got %d", c.Len())
	}
}

func TestSet_OverwritesExisting(t *testing.T) {
	c := New(time.Second, 0)
	c.Set("k", "first")
	c.Set("k", "second")
	v, ok := c.Get("k")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if v.(string) != "second" {
		t.Fatalf("expected 'second', got %v", v)
	}
}
