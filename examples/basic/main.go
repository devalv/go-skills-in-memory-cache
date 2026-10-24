package main

import (
	"fmt"
	"runtime"

	"github.com/devalv/go-skills-in-memory-cache/naive"
)

func main() {
	const entries = 100_000
	const valueSize = 1024

	cache := naive.New[int, []byte]()
	printMemory("Before filling")
	for i := range entries {
		cache.Set(i, make([]byte, valueSize))
	}
	printMemory("After filling")
	value, ok := cache.Get(0)
	fmt.Printf("Get(0): found=%t, size=%d bytes\n", ok, len(value))

	for i := range entries {
		cache.Delete(i)
	}
	printMemory("After deleting")
	// Keep the cache alive so measurements include its remaining map storage.
	runtime.KeepAlive(cache)
}

func printMemory(label string) {
	runtime.GC()
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	fmt.Printf("%s: heap allocated=%.2f MiB, heap objects=%d\n",
		label, float64(stats.HeapAlloc)/(1024*1024), stats.HeapObjects)
}
