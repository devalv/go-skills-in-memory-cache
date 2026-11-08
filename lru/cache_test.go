package lru

import (
	"errors"
	"sync"
	"testing"
)

var _ Cache[string, int] = (*InMemoryCache[string, int])(nil)

func TestInvalidCapacity(t *testing.T) {
	for _, capacity := range []int{-1, 0} {
		c, err := New[int, int](capacity)
		if c != nil || !errors.Is(err, ErrInvalidCapacity) {
			t.Fatalf("New(%d) = (%v, %v)", capacity, c, err)
		}
	}
}

func TestEviction(t *testing.T) {
	for _, operation := range []string{"get", "update"} {
		t.Run(operation, func(t *testing.T) {
			c, err := New[string, int](2)
			if err != nil {
				t.Fatal(err)
			}
			if err := c.Set("a", 1); err != nil {
				t.Fatal(err)
			}
			if err := c.Set("b", 2); err != nil {
				t.Fatal(err)
			}
			if operation == "get" {
				c.Get("a")
			} else {
				if err := c.Set("a", 0); err != nil {
					t.Fatal(err)
				}
			}
			c.Get("missing")
			if err := c.Set("c", 3); err != nil {
				t.Fatal(err)
			}
			if _, ok := c.Get("b"); ok {
				t.Fatal("least recently used key was not evicted")
			}
			want := 1
			if operation == "update" {
				want = 0
			}
			if value, ok := c.Get("a"); !ok || value != want {
				t.Fatalf("Get(a) = (%d, %t), want (%d, true)", value, ok, want)
			}
			c.FlushAll()
			if value, ok := c.Get("a"); ok || value != 0 {
				t.Fatalf("flushed key = (%d, %t)", value, ok)
			}
			if _, ok := c.Get("c"); ok {
				t.Fatal("FlushAll did not remove all keys")
			}
			if len(c.data) != 0 || c.order.Len() != 0 {
				t.Fatal("FlushAll did not clear map and list")
			}
			c.FlushAll()
			if err := c.Set("c", 3); err != nil {
				t.Fatal(err)
			}
			if err := c.Set("d", 4); err != nil {
				t.Fatal(err)
			}
			if value, ok := c.Get("c"); !ok || value != 3 {
				t.Fatal("insertion after flush evicted an entry unnecessarily")
			}
			if err := c.Set("e", 5); err != nil {
				t.Fatal(err)
			}
			if _, ok := c.Get("d"); ok {
				t.Fatal("eviction after flush did not remove least recently used key")
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	c, err := New[int, int](1)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			for j := range 100 {
				if err := c.Set(i, j); err != nil {
					t.Error(err)
					return
				}
				c.Get(i)
				c.FlushAll()
			}
		})
	}
	wg.Wait()
	if len(c.data) != c.order.Len() || len(c.data) > 1 {
		t.Fatal("map and list are inconsistent")
	}
}
