package shards

import (
	"errors"
	"math"
	"sync"
	"testing"
)

var _ Cache[string, int] = (*InMemoryCache[string, int])(nil)

func TestInvalidShardCount(t *testing.T) {
	for _, count := range []int{-1, 0} {
		c, err := New[int, int](count)
		if c != nil || !errors.Is(err, ErrInvalidShardCount) {
			t.Fatalf("New(%d) = (%v, %v)", count, c, err)
		}
	}
}

func TestCache(t *testing.T) {
	for _, count := range []int{1, 3, 16} {
		c, err := New[int, int](count)
		if err != nil {
			t.Fatal(err)
		}
		if value, ok := c.Get(1); ok || value != 0 {
			t.Fatalf("missing key = (%d, %t)", value, ok)
		}
		for i := range 100 {
			if err := c.Set(i, i); err != nil {
				t.Fatal(err)
			}
		}
		if err := c.Set(1, 0); err != nil {
			t.Fatal(err)
		}
		for i := range 100 {
			want := i
			if i == 1 {
				want = 0
			}
			if value, ok := c.Get(i); !ok || value != want {
				t.Fatalf("Get(%d) = (%d, %t), want %d", i, value, ok, want)
			}
		}
		c.FlushAll()
		c.FlushAll()
		for i := range 100 {
			if _, ok := c.Get(i); ok {
				t.Fatalf("key %d survived FlushAll", i)
			}
		}
		if err := c.Set(1, 42); err != nil {
			t.Fatal(err)
		}
		if value, ok := c.Get(1); !ok || value != 42 {
			t.Fatal("cache cannot be reused after FlushAll")
		}
	}
}

func TestGenericKeys(t *testing.T) {
	type key struct {
		Region string
		ID     int
	}
	c, err := New[any, []int](16)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []any{1, "1", key{"eu", 1}, float64(0)} {
		if err := c.Set(k, []int{42}); err != nil {
			t.Fatal(err)
		}
	}
	for _, k := range []any{1, "1", key{"eu", 1}, math.Copysign(0, -1)} {
		if value, ok := c.Get(k); !ok || len(value) != 1 || value[0] != 42 {
			t.Fatalf("Get(%v) = (%v, %t)", k, value, ok)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	c, err := New[int, int](16)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for worker := range 16 {
		wg.Go(func() {
			for i := range 100 {
				k := worker*100 + i
				if err := c.Set(k, k); err != nil {
					t.Error(err)
					return
				}
				if value, ok := c.Get(k); !ok || value != k {
					t.Errorf("Get(%d) = (%d, %t)", k, value, ok)
				}
			}
		})
	}
	wg.Wait()
	for range 4 {
		wg.Go(func() {
			for i := range 100 {
				if err := c.Set(i, i); err != nil {
					t.Error(err)
					return
				}
				c.Get(i)
				c.FlushAll()
			}
		})
	}
	wg.Wait()
	c.FlushAll()
	for i := range 1600 {
		if _, ok := c.Get(i); ok {
			t.Fatalf("key %d survived final flush", i)
		}
	}
}
