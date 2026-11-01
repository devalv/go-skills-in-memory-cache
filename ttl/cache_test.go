package ttl

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

const (
	shortTTL = 50 * time.Millisecond
	longTTL  = 5 * time.Second
)

// newCache creates a cache through its public constructor,
// then tests it through the public Cache interface.
func newCache[K comparable, V any](
	t *testing.T,
	capacity int,
	defaultTTL time.Duration,
) Cache[K, V] {
	t.Helper()

	cache, err := New[K, V](capacity, defaultTTL)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	t.Cleanup(cache.Close)
	return cache
}

// waitUntilMissing waits for a key to become unavailable.
// It avoids assuming exactly when the expiration is observed.
func waitUntilMissing[K comparable, V any](
	t *testing.T,
	cache Cache[K, V],
	key K,
	timeout time.Duration,
) {
	t.Helper()

	deadline := time.Now().Add(timeout)

	for {
		if _, ok := cache.Get(key); !ok {
			return
		}

		if !time.Now().Before(deadline) {
			t.Fatalf("key %v did not expire within %v", key, timeout)
		}

		time.Sleep(5 * time.Millisecond)
	}
}

func TestCacheBasicOperations(t *testing.T) {
	cache := newCache[string, int](t, 2, 0)

	if value, ok := cache.Get("missing"); ok || value != 0 {
		t.Fatalf("missing key: got (%d, %t)", value, ok)
	}

	if err := cache.Set("one", 10); err != nil {
		t.Fatal(err)
	}

	if value, ok := cache.Get("one"); !ok || value != 10 {
		t.Fatalf("expected (10, true), got (%d, %t)", value, ok)
	}

	if err := cache.Set("one", 20); err != nil {
		t.Fatalf("updating key failed: %v", err)
	}

	if value, ok := cache.Get("one"); !ok || value != 20 {
		t.Fatalf("expected (20, true), got (%d, %t)", value, ok)
	}

	if err := cache.Set("zero", 0); err != nil {
		t.Fatal(err)
	}

	if value, ok := cache.Get("zero"); !ok || value != 0 {
		t.Fatalf("zero value must be distinguishable from missing: (%d, %t)", value, ok)
	}
}

func TestCacheCapacity(t *testing.T) {
	cache := newCache[int, string](t, 2, 0)

	for key, value := range map[int]string{
		1: "one",
		2: "two",
	} {
		if err := cache.Set(key, value); err != nil {
			t.Fatal(err)
		}
	}

	if err := cache.Set(3, "three"); !errors.Is(err, ErrFullCapacity) {
		t.Fatalf("expected ErrFullCapacity, got %v", err)
	}

	if _, ok := cache.Get(3); ok {
		t.Fatal("rejected key must not be stored")
	}

	// Updating an existing key must not consume new capacity.
	if err := cache.Set(1, "updated"); err != nil {
		t.Fatalf("update in full cache failed: %v", err)
	}

	if value, ok := cache.Get(1); !ok || value != "updated" {
		t.Fatalf("expected updated value, got (%q, %t)", value, ok)
	}

	if value, ok := cache.Get(2); !ok || value != "two" {
		t.Fatalf("existing value was lost: (%q, %t)", value, ok)
	}
}

func TestCacheInvalidConfiguration(t *testing.T) {
	t.Run("invalid capacity", func(t *testing.T) {
		for _, capacity := range []int{-10, -1, 0} {
			cache, err := New[string, int](capacity, 0)

			if !errors.Is(err, ErrInvalidCapacity) {
				t.Errorf("capacity=%d: expected ErrInvalidCapacity, got %v", capacity, err)
			}

			if cache != nil {
				cache.Close()
				t.Errorf("capacity=%d: expected nil cache", capacity)
			}
		}
	})

	t.Run("invalid default TTL", func(t *testing.T) {
		cache, err := New[string, int](10, -time.Second)

		if !errors.Is(err, ErrInvalidTTL) {
			t.Fatalf("expected ErrInvalidTTL, got %v", err)
		}

		if cache != nil {
			cache.Close()
			t.Fatal("expected nil cache")
		}
	})

	t.Run("invalid custom TTL", func(t *testing.T) {
		cache := newCache[string, int](t, 2, 0)

		err := cache.SetWithTTL("key", 42, -time.Second)
		if !errors.Is(err, ErrInvalidTTL) {
			t.Fatalf("expected ErrInvalidTTL, got %v", err)
		}

		if _, ok := cache.Get("key"); ok {
			t.Fatal("invalid SetWithTTL must not store a value")
		}
	})
}

func TestCacheGenericTypes(t *testing.T) {
	t.Run("integer keys and boolean values", func(t *testing.T) {
		cache := newCache[int, bool](t, 2, 0)

		if err := cache.Set(1, true); err != nil {
			t.Fatal(err)
		}
		if err := cache.Set(2, false); err != nil {
			t.Fatal(err)
		}

		for key, expected := range map[int]bool{1: true, 2: false} {
			value, ok := cache.Get(key)
			if !ok || value != expected {
				t.Errorf("key=%d: got (%t, %t), want (%t, true)",
					key, value, ok, expected)
			}
		}
	})

	t.Run("struct keys and struct values", func(t *testing.T) {
		type Key struct {
			Region string
			ID     int
		}

		type User struct {
			Name   string
			Active bool
		}

		cache := newCache[Key, User](t, 2, 0)

		key := Key{Region: "eu", ID: 42}
		expected := User{Name: "Alice", Active: true}

		if err := cache.Set(key, expected); err != nil {
			t.Fatal(err)
		}

		value, ok := cache.Get(key)
		if !ok || value != expected {
			t.Fatalf("got (%+v, %t), want (%+v, true)", value, ok, expected)
		}
	})

	t.Run("mixed value types", func(t *testing.T) {
		cache := newCache[string, any](t, 3, 0)

		values := map[string]any{
			"string": "hello",
			"int":    42,
			"bool":   true,
		}

		for key, value := range values {
			if err := cache.Set(key, value); err != nil {
				t.Fatal(err)
			}
		}

		if value, ok := cache.Get("string"); !ok || value != "hello" {
			t.Errorf("string value: got (%v, %t)", value, ok)
		}
		if value, ok := cache.Get("int"); !ok || value != 42 {
			t.Errorf("int value: got (%v, %t)", value, ok)
		}
		if value, ok := cache.Get("bool"); !ok || value != true {
			t.Errorf("bool value: got (%v, %t)", value, ok)
		}
	})
}

func TestCacheDefaultTTL(t *testing.T) {
	cache := newCache[string, string](t, 2, shortTTL)

	if err := cache.Set("key", "value"); err != nil {
		t.Fatal(err)
	}

	if value, ok := cache.Get("key"); !ok || value != "value" {
		t.Fatalf("value must be available immediately: (%q, %t)", value, ok)
	}

	waitUntilMissing(t, cache, "key", 2*time.Second)
}

func TestCacheCustomTTL(t *testing.T) {
	cache := newCache[string, int](t, 3, longTTL)

	if err := cache.SetWithTTL("short", 10, shortTTL); err != nil {
		t.Fatal(err)
	}

	if err := cache.Set("default", 20); err != nil {
		t.Fatal(err)
	}

	waitUntilMissing(t, cache, "short", 2*time.Second)

	if value, ok := cache.Get("default"); !ok || value != 20 {
		t.Fatalf("default TTL entry should still exist: (%d, %t)", value, ok)
	}
}

func TestCacheZeroTTL(t *testing.T) {
	t.Run("default zero TTL", func(t *testing.T) {
		cache := newCache[string, int](t, 1, 0)

		if err := cache.Set("key", 42); err != nil {
			t.Fatal(err)
		}

		time.Sleep(2 * shortTTL)

		if value, ok := cache.Get("key"); !ok || value != 42 {
			t.Fatalf("permanent entry expired: (%d, %t)", value, ok)
		}
	})

	t.Run("custom zero TTL overrides default", func(t *testing.T) {
		cache := newCache[string, int](t, 1, shortTTL)

		if err := cache.SetWithTTL("key", 42, 0); err != nil {
			t.Fatal(err)
		}

		time.Sleep(2 * shortTTL)

		if value, ok := cache.Get("key"); !ok || value != 42 {
			t.Fatalf("permanent entry expired: (%d, %t)", value, ok)
		}
	})
}

func TestCacheTTLRefresh(t *testing.T) {
	cache := newCache[string, int](t, 1, 0)

	if err := cache.SetWithTTL("key", 1, shortTTL); err != nil {
		t.Fatal(err)
	}

	// Refresh before the original TTL expires.
	time.Sleep(shortTTL / 2)

	if err := cache.SetWithTTL("key", 2, longTTL); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	// Wait past the original expiration.
	time.Sleep(shortTTL)

	if value, ok := cache.Get("key"); !ok || value != 2 {
		t.Fatalf("refreshed entry should survive: (%d, %t)", value, ok)
	}
}

func TestCacheExpiredEntryFreesCapacity(t *testing.T) {
	cache := newCache[string, int](t, 1, 0)

	if err := cache.SetWithTTL("old", 1, shortTTL); err != nil {
		t.Fatal(err)
	}

	waitUntilMissing(t, cache, "old", 2*time.Second)

	if err := cache.Set("new", 2); err != nil {
		t.Fatalf("expired entry should free capacity: %v", err)
	}

	if value, ok := cache.Get("new"); !ok || value != 2 {
		t.Fatalf("new entry missing: (%d, %t)", value, ok)
	}
}

func TestCacheExpiredEntriesFreeCapacityWithoutGet(t *testing.T) {
	cache := newCache[string, int](t, 1, 0)

	if err := cache.SetWithTTL("old", 1, shortTTL); err != nil {
		t.Fatal(err)
	}

	// No Get("old") call: expiration must be respected
	// when attempting to insert a new entry.
	time.Sleep(2 * shortTTL)

	if err := cache.Set("new", 2); err != nil {
		t.Fatalf("expired entry should not block insertion: %v", err)
	}

	if value, ok := cache.Get("new"); !ok || value != 2 {
		t.Fatalf("new entry missing: (%d, %t)", value, ok)
	}
}

func TestCacheFlushAll(t *testing.T) {
	cache := newCache[string, int](t, 2, 0)

	if err := cache.Set("one", 1); err != nil {
		t.Fatal(err)
	}
	if err := cache.Set("two", 2); err != nil {
		t.Fatal(err)
	}

	cache.FlushAll()

	for _, key := range []string{"one", "two"} {
		if _, ok := cache.Get(key); ok {
			t.Errorf("key %q still exists after FlushAll", key)
		}
	}

	if err := cache.Set("three", 3); err != nil {
		t.Fatalf("cache should accept writes after FlushAll: %v", err)
	}
}

func TestCacheFlushAllWithTTL(t *testing.T) {
	cache := newCache[string, int](t, 2, longTTL)

	if err := cache.Set("default", 1); err != nil {
		t.Fatal(err)
	}
	if err := cache.SetWithTTL("custom", 2, shortTTL); err != nil {
		t.Fatal(err)
	}

	cache.FlushAll()

	if _, ok := cache.Get("default"); ok {
		t.Fatal("default TTL entry survived FlushAll")
	}
	if _, ok := cache.Get("custom"); ok {
		t.Fatal("custom TTL entry survived FlushAll")
	}

	if err := cache.Set("new", 3); err != nil {
		t.Fatalf("cache should be reusable after FlushAll: %v", err)
	}
}

func TestCacheCloseIdempotent(t *testing.T) {
	cache, err := New[string, int](10, shortTTL)
	if err != nil {
		t.Fatal(err)
	}

	cache.Close()
	cache.Close()
	cache.Close()
}

func TestCacheConcurrentAccess(t *testing.T) {
	const (
		goroutines = 50
		operations = 100
	)

	cache := newCache[int, int](t, goroutines*operations, 0)

	var wg sync.WaitGroup
	start := make(chan struct{})

	for id := range goroutines {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			<-start

			for j := range operations {
				key := id*operations + j

				if err := cache.Set(key, key*10); err != nil {
					t.Errorf("Set(%d): %v", key, err)
					return
				}

				if value, ok := cache.Get(key); !ok || value != key*10 {
					t.Errorf("Get(%d): got (%d, %t)", key, value, ok)
					return
				}
			}
		}(id)
	}

	close(start)
	wg.Wait()

	for key := range goroutines * operations {
		expected := key * 10

		value, ok := cache.Get(key)
		if !ok || value != expected {
			t.Errorf("key=%d: got (%d, %t), want (%d, true)",
				key, value, ok, expected)
		}
	}
}

func TestCacheConcurrentCapacity(t *testing.T) {
	const (
		capacity   = 100
		goroutines = 500
	)

	cache := newCache[int, int](t, capacity, 0)

	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan error, goroutines)

	for id := range goroutines {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			<-start
			results <- cache.Set(id, id)
		}(id)
	}

	close(start)
	wg.Wait()
	close(results)

	successes := 0

	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrFullCapacity):
			// Expected when the cache is full.
		default:
			t.Errorf("unexpected Set error: %v", err)
		}
	}

	if successes != capacity {
		t.Fatalf("expected %d successful inserts, got %d",
			capacity, successes)
	}
}

func TestCacheConcurrentTTLRefresh(t *testing.T) {
	cache := newCache[string, int](t, 1, 0)

	if err := cache.SetWithTTL("key", 0, shortTTL); err != nil {
		t.Fatal(err)
	}

	// Start after the initial entry has expired.
	time.Sleep(2 * shortTTL)

	const goroutines = 20

	var wg sync.WaitGroup
	start := make(chan struct{})

	for id := range goroutines {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			<-start

			if id%2 == 0 {
				cache.Get("key")
			} else {
				if err := cache.SetWithTTL("key", id, longTTL); err != nil {
					t.Errorf("refresh failed: %v", err)
				}
			}
		}(id)
	}

	close(start)
	wg.Wait()

	// At least one successful refresh must remain visible.
	if _, ok := cache.Get("key"); !ok {
		t.Fatal("refreshed entry was unexpectedly removed")
	}
}

func TestCacheConcurrentFlushAll(t *testing.T) {
	cache := newCache[int, int](t, 100, 0)

	const goroutines = 20
	const operations = 100

	var wg sync.WaitGroup

	for id := range goroutines {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for j := range operations {
				key := id*operations + j

				if j%10 == 0 {
					cache.FlushAll()
				} else {
					// Capacity errors are expected under contention.
					err := cache.Set(key, j)
					if err != nil && !errors.Is(err, ErrFullCapacity) {
						t.Errorf("unexpected Set error: %v", err)
					}
				}

				cache.Get(key)
			}
		}(id)
	}

	wg.Wait()

	// The cache must remain usable after concurrent operations.
	cache.FlushAll()

	if err := cache.Set(999, 42); err != nil {
		t.Fatalf("cache unusable after concurrent FlushAll: %v", err)
	}

	if value, ok := cache.Get(999); !ok || value != 42 {
		t.Fatalf("unexpected final value: (%d, %t)", value, ok)
	}
}

func TestCacheInterfaceCompatibility(t *testing.T) {
	// Compile-time check that the concrete implementation
	// satisfies the public interface.
	var _ Cache[string, int] = (*InMemoryCache[string, int])(nil)

	cache := newCache[string, int](t, 1, 0)

	if err := cache.SetWithTTL("key", 42, 0); err != nil {
		t.Fatal(err)
	}

	value, ok := cache.Get("key")
	if !ok || value != 42 {
		t.Fatal(fmt.Sprintf("unexpected result: (%d, %t)", value, ok))
	}
}

func TestJanitorRemovesExpiredEntries(t *testing.T) {
	if testing.Short() {
    t.Skip("skipping long-running janitor test")
}
    cache, err := New[string, int](1, 30*time.Millisecond)
    if err != nil {
        t.Fatal(err)
    }
    defer cache.Close()

    if err := cache.Set("expired", 42); err != nil {
        t.Fatal(err)
    }

    deadline := time.Now().Add(65 * time.Second)

    for {
        cache.mu.RLock()
        _, exists := cache.data["expired"]
        cache.mu.RUnlock()

        if !exists {
            return
        }

        if time.Now().After(deadline) {
            t.Fatal("janitor did not remove expired entry")
        }

        time.Sleep(100 * time.Millisecond)
    }
}
