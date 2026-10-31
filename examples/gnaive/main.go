package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/devalv/go-skills-in-memory-cache/gnaive"
)

func main() {
	const entries = 100_000
	const valueSize = 1024

	cache, err := gnaive.New[any, any](entries)
	if err != nil {
		fmt.Printf("failed to create cache: %v\n", err)
		os.Exit(1)
	}

	printMemory("Before filling")

	for i := range entries {
		var key any
		var value any

		switch i % 3 {
		case 0:
			key = i
			value = strings.Repeat("x", valueSize)
		case 1:
			key = strconv.Itoa(i)
			value = i
		case 2:
			key = fmt.Sprintf("bool:%d", i)
			value = i%2 == 0
		}

		if err := cache.Set(key, value); err != nil {
			fmt.Printf("failed to set value: %v\n", err)
			os.Exit(1)
		}
	}

	printMemory("After filling")

	// Check different key and value types.
	for _, key := range []any{0, "1", true, false} {
		value, ok := cache.Get(key)
		fmt.Printf(
			"Get(%v): found=%t, type=%T, value=%v\n",
			key, ok, value, value,
		)
	}

	// Delete using the same keys as during insertion.
	for i := range entries {
		var key any

		switch i % 3 {
		case 0:
			key = i
		case 1:
			key = strconv.Itoa(i)
		case 2:
			key = i%2 == 0
		}

		cache.Delete(key)
	}

	printMemory("After deleting")

	runtime.KeepAlive(cache)
}

func printMemory(label string) {
	runtime.GC()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	fmt.Printf(
		"%s: heap allocated=%.2f MiB, heap objects=%d\n",
		label,
		float64(stats.HeapAlloc)/(1024*1024),
		stats.HeapObjects,
	)
}
