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
├── lru/
│   ├── cache.go
│   └── cache_test.go
├── shards/
│   ├── cache.go
│   └── cache_test.go
└── examples/
    ├── naive/
    │   └── main.go
    ├── gnaive/
    │   └── main.go
    ├── ttl/
    │   └── main.go
    ├── lru/
    │   └── main.go
    └── shards/
        └── main.go
```


## naive/cache.go
Потокобезопасный, но, только string/string

```go
func New(capacity int) (*InMemoryCache, error) {
```

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
func New[K comparable, V any](capacity int) (*InMemoryCache[K, V], error) {
```

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
func New[K comparable, V any](
	capacity int,
	ttl time.Duration,
) (*InMemoryCache[K, V], error) {
```

```go
type Cache[K comparable, V any] interface {
	Set(k K, v V) error
	SetWithTTL(k K, v V, ttl time.Duration) error
	Get(k K) (V, bool)
	FlushAll()
	Close()
}
```

## lru/cache.go
Потокобезопасный кеш с дженериками и вытеснением наименее недавно использованного элемента (LRU), без TTL.
Ограничение по cap

```go
func New[K comparable, V any](capacity int) (*InMemoryCache[K, V], error) {
```


```go
type Cache[K comparable, V any] interface {
	Set(k K, v V) error
	Get(k K) (V, bool)
	FlushAll()
}
```

## shards/cache.go
Потокобезопасный кеш с дженериками и шардированием ключей, без TTL, LRU и ограничения количества записей.

```go
func New[K comparable, V any](shardCount int) (*InMemoryCache[K, V], error)

type Cache[K comparable, V any] interface {
    Set(k K, v V) error
    Get(k K) (V, bool)
    FlushAll()
}
```

Шард выбирается через стандартную `hash/maphash.Comparable` с неизменным seed кеша.
Ключи сохраняются целиком: совпадение хешей не приводит к перезаписи разных ключей.
Каждый шард содержит свою map и `sync.RWMutex`: `Set` блокирует только свой шард,
а несколько `Get` одного шарда могут выполняться одновременно.
`Set` возвращает `nil`. `FlushAll` блокирует все шарды в фиксированном порядке
и атомарно очищает кеш относительно `Set` и `Get`.
Обращения к одному ключу или разным ключам одного шарда используют общую блокировку;
сходство значений не влияет на выбор шарда.
