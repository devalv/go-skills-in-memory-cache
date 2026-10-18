package naive

import "testing"

func TestCache(t *testing.T) {
	var cache Cache[string, int]
	if value, ok := cache.Get("missing"); ok || value != 0 {
		t.Fatalf("missing key: got (%d, %t)", value, ok)
	}
	cache.Set("key", 0)
	if value, ok := cache.Get("key"); !ok || value != 0 {
		t.Fatalf("stored zero: got (%d, %t)", value, ok)
	}
	cache.Set("key", 42)
	if value, ok := cache.Get("key"); !ok || value != 42 {
		t.Fatalf("updated key: got (%d, %t)", value, ok)
	}
	cache.Delete("key")
	cache.Delete("missing")
	if _, ok := cache.Get("key"); ok {
		t.Fatal("deleted key still present")
	}
}
