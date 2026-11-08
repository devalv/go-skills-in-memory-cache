package lru

import (
	"container/list"
	"errors"
	"sync"
)

var (
	ErrInvalidCapacity = errors.New("cache capacity must be greater than 0")
	ErrInvalidType     = errors.New("unexpected list entry type")
)

type entry[K comparable, V any] struct {
	key   K
	value V
}

type Cache[K comparable, V any] interface {
	Set(k K, v V) error
	Get(k K) (V, bool)
	FlushAll()
}

// InMemoryCache must be created with New and must not be copied after use.
type InMemoryCache[K comparable, V any] struct {
	mu    sync.Mutex
	data  map[K]*list.Element
	order list.List // Most recently used entries are at the front.
	cap   int
}

// New creates a cache holding at most capacity entries.
func New[K comparable, V any](capacity int) (*InMemoryCache[K, V], error) {
	if capacity < 1 {
		return nil, ErrInvalidCapacity
	}

	return &InMemoryCache[K, V]{
		data: make(map[K]*list.Element),
		cap:  capacity,
	}, nil
}

// Set stores a value, evicting the least recently used entry when full.
// It returns nil for both insertion and update, including eviction.
func (c *InMemoryCache[K, V]) Set(k K, v V) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.data[k]; ok {
		element.Value = entry[K, V]{key: k, value: v}
		c.order.MoveToFront(element)

		return nil
	}

	c.data[k] = c.order.PushFront(entry[K, V]{key: k, value: v})
	if len(c.data) > c.cap {
		oldest := c.order.Back()
		item, ok := oldest.Value.(entry[K, V])
		if !ok {
			panic(ErrInvalidType)
		}
		delete(c.data, item.key)
		c.order.Remove(oldest)
	}

	return nil
}

// Get returns a value and marks it as recently used when found.
func (c *InMemoryCache[K, V]) Get(k K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.data[k]; ok {
		item, ok := element.Value.(entry[K, V])
		if !ok {
			panic(ErrInvalidType)
		}
		c.order.MoveToFront(element)

		return item.value, true
	}
	var zero V

	return zero, false
}

func (c *InMemoryCache[K, V]) FlushAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[K]*list.Element)
	c.order.Init()
}
