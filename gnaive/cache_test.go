package gnaive

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
)

func TestCache(t *testing.T) {
	cache, err := New[string, string](1)
	if err != nil {
		t.Fatalf("cache creating should be ok: %v", err)
	}

	if value, ok := cache.Get("missing"); ok || value != "" {
		t.Fatalf("missing key: got (%q, %t)", value, ok)
	}

	if err := cache.Set("key", "0"); err != nil {
		t.Fatalf("set should succeed: %v", err)
	}

	if value, ok := cache.Get("key"); !ok || value != "0" {
		t.Fatalf("stored zero: got (%q, %t)", value, ok)
	}

	if err := cache.Set("key", "42"); err != nil {
		t.Fatalf("updating existing key should succeed: %v", err)
	}

	if value, ok := cache.Get("key"); !ok || value != "42" {
		t.Fatalf("updated key: got (%q, %t)", value, ok)
	}

	err = cache.Set("cap-error", "87")
	if !errors.Is(err, ErrFullCapacity) {
		t.Fatalf("expected ErrFullCapacity, got %v", err)
	}

	cache.Delete("key")
	cache.Delete("missing")

	if _, ok := cache.Get("key"); ok {
		t.Fatal("deleted key still present")
	}

	if err := cache.Set("new-key", "100"); err != nil {
		t.Fatalf("set after deletion should succeed: %v", err)
	}
}

func TestCacheInvalidCapacity(t *testing.T) {
	for _, capacity := range []int{-100, -1, 0} {
		t.Run(fmt.Sprintf("capacity=%d", capacity), func(t *testing.T) {
			cache, err := New[string, int](capacity)

			if !errors.Is(err, ErrInvalidCapacity) {
				t.Fatalf("expected ErrInvalidCapacity, got %v", err)
			}

			if cache != nil {
				t.Fatal("expected nil cache for invalid capacity")
			}
		})
	}
}

func TestCacheGenericTypes(t *testing.T) {
	t.Run("int keys and int values", func(t *testing.T) {
		cache, err := New[int, int](2)
		if err != nil {
			t.Fatal(err)
		}

		if err := cache.Set(1, 100); err != nil {
			t.Fatal(err)
		}

		if err := cache.Set(2, 0); err != nil {
			t.Fatal(err)
		}

		if value, ok := cache.Get(1); !ok || value != 100 {
			t.Fatalf("expected (100, true), got (%d, %t)", value, ok)
		}

		if value, ok := cache.Get(2); !ok || value != 0 {
			t.Fatalf("expected (0, true), got (%d, %t)", value, ok)
		}

		if value, ok := cache.Get(3); ok || value != 0 {
			t.Fatalf("expected (0, false), got (%d, %t)", value, ok)
		}
	})

	t.Run("string keys and struct values", func(t *testing.T) {
		type User struct {
			ID   int
			Name string
		}

		cache, err := New[string, User](2)
		if err != nil {
			t.Fatal(err)
		}

		expected := User{ID: 1, Name: "Alice"}

		if err := cache.Set("user:1", expected); err != nil {
			t.Fatal(err)
		}

		actual, ok := cache.Get("user:1")
		if !ok || actual != expected {
			t.Fatalf("expected (%+v, true), got (%+v, %t)", expected, actual, ok)
		}
	})

	t.Run("struct keys and slice values", func(t *testing.T) {
		type Key struct {
			Region string
			ID     int
		}

		cache, err := New[Key, []int](2)
		if err != nil {
			t.Fatal(err)
		}

		key := Key{Region: "eu", ID: 42}
		expected := []int{10, 20, 30}

		if err := cache.Set(key, expected); err != nil {
			t.Fatal(err)
		}

		actual, ok := cache.Get(key)
		if !ok {
			t.Fatal("expected key to exist")
		}

		if len(actual) != len(expected) {
			t.Fatalf("expected length %d, got %d", len(expected), len(actual))
		}

		for i := range expected {
			if actual[i] != expected[i] {
				t.Fatalf("index %d: expected %d, got %d", i, expected[i], actual[i])
			}
		}
	})

	t.Run("int keys and pointer values", func(t *testing.T) {
		type Item struct {
			Value string
		}

		cache, err := New[int, *Item](1)
		if err != nil {
			t.Fatal(err)
		}

		expected := &Item{Value: "hello"}

		if err := cache.Set(42, expected); err != nil {
			t.Fatal(err)
		}

		actual, ok := cache.Get(42)
		if !ok || actual != expected {
			t.Fatalf("expected (%p, true), got (%p, %t)", expected, actual, ok)
		}
	})

	t.Run("bool keys and bool values", func(t *testing.T) {
		cache, err := New[bool, bool](2)
		if err != nil {
			t.Fatal(err)
		}

		if err := cache.Set(true, false); err != nil {
			t.Fatal(err)
		}

		value, ok := cache.Get(true)
		if !ok || value != false {
			t.Fatalf("expected (false, true), got (%t, %t)", value, ok)
		}
	})
}

func TestCacheCapacity(t *testing.T) {
	cache, err := New[int, string](2)
	if err != nil {
		t.Fatal(err)
	}

	if err := cache.Set(1, "one"); err != nil {
		t.Fatal(err)
	}

	if err := cache.Set(2, "two"); err != nil {
		t.Fatal(err)
	}

	if err := cache.Set(3, "three"); !errors.Is(err, ErrFullCapacity) {
		t.Fatalf("expected ErrFullCapacity, got %v", err)
	}

	if _, ok := cache.Get(3); ok {
		t.Fatal("rejected key should not exist")
	}

	// Updating an existing key must work even when cache is full.
	if err := cache.Set(1, "updated"); err != nil {
		t.Fatalf("updating existing key should succeed: %v", err)
	}

	if value, ok := cache.Get(1); !ok || value != "updated" {
		t.Fatalf("expected (updated, true), got (%q, %t)", value, ok)
	}

	cache.Delete(2)

	// Deleting a key should free capacity.
	if err := cache.Set(3, "three"); err != nil {
		t.Fatalf("set after deletion should succeed: %v", err)
	}

	if value, ok := cache.Get(3); !ok || value != "three" {
		t.Fatalf("expected (three, true), got (%q, %t)", value, ok)
	}
}

func TestThreadSafe(t *testing.T) {
	const (
		goroutines = 100
		operations = 1000
	)

	cache, err := New[int, string](operations)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := range goroutines {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			<-start

			for j := range operations {
				key := j
				value := strconv.Itoa(id)

				if err := cache.Set(key, value); err != nil {
					t.Errorf("Set(%d): %v", key, err)
					return
				}

				cache.Get(key)

				if j%2 == 0 {
					cache.Delete(key)
				}
			}
		}(i)
	}

	close(start)
	wg.Wait()
}

func TestThreadSafeUniqueKeys(t *testing.T) {
	const (
		goroutines = 50
		operations = 100
	)

	total := goroutines * operations

	cache, err := New[int, int](total)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	for i := range goroutines {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for j := range operations {
				key := id*operations + j

				if err := cache.Set(key, key*10); err != nil {
					t.Errorf("Set(%d): %v", key, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all writes after goroutines finish.
	for key := range total {
		expected := key * 10

		value, ok := cache.Get(key)
		if !ok || value != expected {
			t.Errorf(
				"key %d: expected (%d, true), got (%d, %t)",
				key, expected, value, ok,
			)
		}
	}
}

func TestThreadSafeCapacity(t *testing.T) {
	const (
		capacity   = 100
		goroutines = 1000
	)

	cache, err := New[int, int](capacity)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	var successful sync.WaitGroup

	// Count successful writes using a channel.
	results := make(chan bool, goroutines)

	for i := range goroutines {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			err := cache.Set(id, id)

			switch {
			case err == nil:
				results <- true
			case errors.Is(err, ErrFullCapacity):
				results <- false
			default:
				t.Errorf("unexpected Set error: %v", err)
				results <- false
			}
		}(i)
	}

	wg.Wait()
	close(results)

	_ = successful

	count := 0
	for ok := range results {
		if ok {
			count++
		}
	}

	if count != capacity {
		t.Fatalf("expected %d successful writes, got %d", capacity, count)
	}
}
