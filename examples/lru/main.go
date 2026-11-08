package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/devalv/go-skills-in-memory-cache/lru"
)

func main() {
	const entries = 100_000
	const valueSize = 1024

	implementation, err := lru.New[int, []byte](entries)
	if err != nil {
		fmt.Printf("failed to create cache: %v\n", err)
		os.Exit(1)
	}
	var cache lru.Cache[int, []byte] = implementation

	printMemory("Before filling")
	for i := range entries {
		if err := cache.Set(i, make([]byte, valueSize)); err != nil {
			fmt.Printf("failed to set value: %v\n", err)
			os.Exit(1)
		}
	}
	printMemory("After filling")

	// Accessing key 0 protects it from the next eviction.
	cache.Get(0)
	if err := cache.Set(entries, make([]byte, valueSize)); err != nil {
		fmt.Printf("failed to set value: %v\n", err)
		os.Exit(1)
	}
	for _, key := range []int{0, 1, entries} {
		value, ok := cache.Get(key)
		fmt.Printf("Get(%d): found=%t, size=%d bytes\n", key, ok, len(value))
	}

	cache.FlushAll()
	printMemory("After FlushAll")
	runtime.KeepAlive(cache)
}

func printMemory(label string) {
	runtime.GC()
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	fmt.Printf("%s: heap allocated=%.2f MiB, heap objects=%d\n",
		label, float64(stats.HeapAlloc)/(1024*1024), stats.HeapObjects)
}
