# go-skills-in-memory-cache

```plain
go-skills-in-memory-cache/
├── go.mod
├── README.md
├── Makefile
├── naive/
│   ├── cache.go
│   └── cache_test.go
├── gnaive/
│   ├── cache.go
│   └── cache_test.go
├── ttl/
│   ├── cache.go
│   └── cache_test.go
├── examples/
│   └── naive/
│       └── main.go
├── examples/
│   └── gnaive/
│       └── main.go
└── examples/
    └── ttl/
        └── main.go
```


## naive/cache.go
Потокобезопасный, но, только string/string

```go
type Cache[K comparable, V any] interface {
	Set(k K, v V) error
	Get(k K) (V, bool)
	Delete(k K)
}
```

## gnaive/cache.go
Потокобезопасный с использованием дженериков

```go
type Cache[K comparable, V any] interface {
	Set(k K, v V) error
	Get(k K) (V, bool)
	Delete(k K)
}
```

## ttl/cache.go
Потокобезопасный с использованием дженериков и TTL

```go
type Cache[K comparable, V any] interface {
	Set(k K, v V) error
	SetWithTTL(k K, v V, ttl time.Duration) error
	Get(k K) (V, bool)
	FlushAll()
	Close()
}
```
