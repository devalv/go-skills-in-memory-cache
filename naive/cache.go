package naive

import (
	"errors"
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
	_, present := c.data[k]

	if !present && len(c.data) >= c.cap {
		return ErrFullCapacity
	}
	c.data[k] = v

	return nil
}

func (c *InMemoryCache) Get(k string) (string, bool) {
	data, ok := c.data[k]

	return data, ok
}

func (c *InMemoryCache) Delete(k string) {
	delete(c.data, k)
}

// TODO: аргумент с lock/mu
// TODO: тесты с лок и без лок
// TODO: лок и анлок - сравнить скорость обычного и рв мутекс
