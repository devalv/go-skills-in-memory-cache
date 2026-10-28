package naive

import (
	"errors"
	"sync"
)

var (
	ErrInvalidCapacity = errors.New("cache capacity must be non-negative")
	ErrFullCapacity    = errors.New("cache capacity is full")
)

type Cache interface {
	Set(k, v string)
	Get(k string) (string, bool)
}

type InMemoryCache struct {
	data map[string]string
	cap  int
	mu   sync.RWMutex
}

func New(capacity int) (*InMemoryCache, error) {
	if capacity < 1 {
		return nil, ErrInvalidCapacity
	}

	return &InMemoryCache{
		data: make(map[string]string),
		cap:  capacity,
	}, nil
}

func (c *InMemoryCache) Set(k, v string) error {
	// TODO: Cache[K comparable, V any] -> gnaive/cache
	c.mu.Lock()
	defer c.mu.Unlock()

	_, present := c.data[k]
	if !present && len(c.data) >= c.cap {
		return ErrFullCapacity
	}
	c.data[k] = v

	return nil
}

func (c *InMemoryCache) Get(k string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.data[k]

	return data, ok
}

func (c *InMemoryCache) Delete(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, k)
}
