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
├── examples/
│   └── naive/
│       └── main.go
└── examples/
    └── naive/
        └── main.go
```


## naive/cache.go
Потокобезопасный, но, только string/string

## gnaive/cache.go
Потокобезопасный с использованием дженериков
