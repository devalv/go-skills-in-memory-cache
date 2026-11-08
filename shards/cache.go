package shards

import (
	"errors"
	"hash/maphash"
	"sync"
)

var ErrInvalidShardCount = errors.New("cache shard count must be greater than 0")

type Cache[K comparable, V any] interface {
	Set(k K, v V) error
	Get(k K) (V, bool)
	FlushAll()
}

type shard[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

// InMemoryCache must be created with New and must not be copied after use.
type InMemoryCache[K comparable, V any] struct {
	shards []shard[K, V]
	seed   maphash.Seed
}

// New creates an unbounded cache with shardCount independently locked maps.
func New[K comparable, V any](shardCount int) (*InMemoryCache[K, V], error) {
	if shardCount < 1 {
		return nil, ErrInvalidShardCount
	}

	c := &InMemoryCache[K, V]{
		shards: make([]shard[K, V], shardCount),
		seed:   maphash.MakeSeed(),
	}
	for i := range c.shards {
		c.shards[i].data = make(map[K]V)
	}

	return c, nil
}

func (c *InMemoryCache[K, V]) shardFor(k K) *shard[K, V] {
	index := maphash.Comparable(c.seed, k) % uint64(len(c.shards))

	return &c.shards[index]
}

// Set stores or updates a value and returns nil.
func (c *InMemoryCache[K, V]) Set(k K, v V) error {
	s := c.shardFor(k)
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[k] = v

	return nil
}

func (c *InMemoryCache[K, V]) Get(k K) (V, bool) {
	s := c.shardFor(k)
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[k]

	return v, ok
}

// FlushAll clears the cache atomically with respect to Set and Get.
func (c *InMemoryCache[K, V]) FlushAll() {
	// Every flush locks shards in the same order to avoid deadlocks.
	for i := range c.shards {
		c.shards[i].mu.Lock()
	}
	for i := range c.shards {
		c.shards[i].data = make(map[K]V)
	}
	for i := len(c.shards) - 1; i >= 0; i-- {
		c.shards[i].mu.Unlock()
	}
}
