package naive

import (
	"errors"
	"testing"
)

func TestCache(t *testing.T) {
	cache, err := New(1)
	if err != nil {
		t.Fatalf("cache creating should be ok. got %s", err)
	}
	if value, ok := cache.Get("missing"); ok || value != "" {
		t.Fatalf("missing key: got (%s, %t)", value, ok)
	}
	_ = cache.Set("key", "0")
	if value, ok := cache.Get("key"); !ok || value != "0" {
		t.Fatalf("stored zero: got (%s, %t)", value, ok)
	}
	_ = cache.Set("key", "42")
	if value, ok := cache.Get("key"); !ok || value != "42" {
		t.Fatalf("updated key: got (%s, %t)", value, ok)
	}

	capE := cache.Set("cap-error", "87");
	if capE == nil {
		t.Fatalf("cache capacity should get error: %e", capE)
	}
	if !errors.Is(capE, ErrFullCapacity) {
		t.Fatalf("wrong error type: got (%e)", capE)
	}

	cache.Delete("key")
	cache.Delete("missing")
	if _, ok := cache.Get("key"); ok {
		t.Fatal("deleted key still present")
	}


}


func TestCacheZeroCap(t *testing.T) {
	_, err := New(0)
	if err == nil  {
		t.Fatalf("cache creating should got err.")
	}
	if !errors.Is(err, ErrInvalidCapacity) {
		t.Fatalf("wrong error type: got (%e)", err)
	}

}
