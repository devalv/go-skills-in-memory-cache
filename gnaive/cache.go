package gnaive

import (
	"errors"
	"sync"
)

var (
	ErrInvalidCapacity = errors.New("cache capacity must be non-negative")
	ErrFullCapacity    = errors.New("cache capacity is full")
)

type Cache[K comparable, V any] interface {
	Set(k K, v V) error
    Get(k K) (V, bool)
    Delete(k K)
}

type InMemoryCache[K comparable, V any] struct {
	data map[K]V
	cap  int
	mu   sync.RWMutex
}

func New[K comparable, V any](capacity int) (*InMemoryCache[K, V], error) {
	if capacity < 1 {
		return nil, ErrInvalidCapacity
	}

	return &InMemoryCache[K, V]{
		data: make(map[K]V),
		cap:  capacity,
	}, nil
}

func (c *InMemoryCache[K, V]) Set(k K, v V) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, present := c.data[k]
	if !present && len(c.data) >= c.cap {
		return ErrFullCapacity
	}
	c.data[k] = v

	return nil
}

func (c *InMemoryCache[K, V]) Get(k K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.data[k]

	return data, ok
}

func (c *InMemoryCache[K, V]) Delete(k K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, k)
}
