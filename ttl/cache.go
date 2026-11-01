package ttl

import (
	"errors"
	"sync"
	"time"
)

const defaultCleanupInterval = time.Minute

var (
	ErrInvalidCapacity = errors.New("cache capacity must be greater than 0")
	ErrFullCapacity    = errors.New("cache capacity is full")
	ErrInvalidTTL      = errors.New("cache ttl must be non-negative")
)

type entry[V any] struct {
	value     V
	expiresAt time.Time // zero value means no expiration
}

func (e entry[V]) expired(now time.Time) bool {
	return !e.expiresAt.IsZero() && !now.Before(e.expiresAt)
}

type Cache[K comparable, V any] interface {
	Set(k K, v V) error
	SetWithTTL(k K, v V, ttl time.Duration) error
	Get(k K) (V, bool)
	FlushAll()
	Close()
}

type InMemoryCache[K comparable, V any] struct {
	data      map[K]entry[V]
	cap       int
	mu        sync.RWMutex
	ttl       time.Duration
	janitor   *time.Ticker
	stop      chan struct{}
	done      chan struct{}
	closeOnce sync.Once
}

func New[K comparable, V any](
	capacity int,
	ttl time.Duration,
) (*InMemoryCache[K, V], error) {
	if capacity < 1 {
		return nil, ErrInvalidCapacity
	}
	if ttl < 0 {
		return nil, ErrInvalidTTL
	}

	c := &InMemoryCache[K, V]{
		data:    make(map[K]entry[V]),
		cap:     capacity,
		ttl:     ttl,
		janitor: time.NewTicker(defaultCleanupInterval),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}

	go c.cleanupLoop()

	return c, nil
}

// Set stores a value under key, using the cache's default TTL.
func (c *InMemoryCache[K, V]) Set(k K, v V) error {
	return c.SetWithTTL(k, v, c.ttl)
}

// SetWithTTL stores a value with a custom TTL.
// A TTL of zero means no expiration.
func (c *InMemoryCache[K, V]) SetWithTTL(
	k K,
	v V,
	ttl time.Duration,
) error {
	if ttl < 0 {
		return ErrInvalidTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	_, present := c.data[k]

	if !present && len(c.data) >= c.cap {
		// The lock is already held.
		c.deleteExpiredLocked(now)

		// Check capacity again after removing expired entries.
		if len(c.data) >= c.cap {
			return ErrFullCapacity
		}
	}

	e := entry[V]{value: v}
	if ttl > 0 {
		e.expiresAt = now.Add(ttl)
	}

	c.data[k] = e

	return nil
}

func (c *InMemoryCache[K, V]) Get(k K) (V, bool) {
	var zero V

	c.mu.RLock()

	e, ok := c.data[k]
	if !ok {
		c.mu.RUnlock()

		return zero, false
	}

	if !e.expired(time.Now()) {
		c.mu.RUnlock()

		return e.value, true
	}

	c.mu.RUnlock()

	// The entry was expired when first read.
	// Acquire a write lock and check it again.
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok = c.data[k]
	if !ok {
		return zero, false
	}

	// Another goroutine may have refreshed the entry.
	if !e.expired(time.Now()) {
		return e.value, true
	}

	delete(c.data, k)

	return zero, false
}

func (c *InMemoryCache[K, V]) FlushAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[K]entry[V])
}

// Close stops the background cleanup goroutine.
// Calling Close multiple times is safe.
func (c *InMemoryCache[K, V]) Close() {
	c.closeOnce.Do(func() {
		c.janitor.Stop()
		close(c.stop)
	})

	// Wait until the cleanup goroutine has exited.
	<-c.done
}

func (c *InMemoryCache[K, V]) cleanupLoop() {
	defer close(c.done)

	for {
		select {
		case <-c.janitor.C:
			c.deleteExpired()
		case <-c.stop:
			return
		}
	}
}

// deleteExpired acquires the write lock.
// It is intended for use by the background janitor.
func (c *InMemoryCache[K, V]) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deleteExpiredLocked(time.Now())
}

// deleteExpiredLocked removes all expired entries.
// The caller MUST already hold c.mu.Lock().
func (c *InMemoryCache[K, V]) deleteExpiredLocked(now time.Time) {
	for k, e := range c.data {
		if e.expired(now) {
			delete(c.data, k)
		}
	}
}
