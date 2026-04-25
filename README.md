# simble-cache

[![Tests](https://github.com/charupraneeth/simble-cache/actions/workflows/test.yml/badge.svg)](https://github.com/charupraneeth/simble-cache/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/charupraneeth/simble-cache.svg)](https://pkg.go.dev/github.com/charupraneeth/simble-cache)
[![Go Report Card](https://goreportcard.com/badge/github.com/charupraneeth/simble-cache)](https://goreportcard.com/report/github.com/charupraneeth/simble-cache)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A thread-safe, generic, in-process LRU cache for Go with optional per-key TTL expiration.

## Features

- **Generic** — store any type (`string`, `int`, `struct`, etc.) with full compile-time type safety
- **O(1) Get & Set** — backed by a hash map + doubly linked list
- **LRU Eviction** — automatically evicts the least recently used item when the cache is full
- **TTL Expiration** — per-key time-to-live, lazily expired on `Get` and swept by a background goroutine every minute
- **Thread-safe** — safe for concurrent use from multiple goroutines

## Installation

```bash
go get github.com/charupraneeth/simble-cache
```

> Requires Go 1.21+

## Quick Start

```go
package main

import (
    "fmt"
    cache "github.com/charupraneeth/simble-cache"
)

func main() {
    // Create a cache of strings with a capacity of 100 items
    c := cache.NewLRUCache[string](100)

    // Set a key with a 60-second TTL
    c.Set("greeting", "hello", 60_000)

    // Get returns (value, ok) — just like a map lookup
    if val, ok := c.Get("greeting"); ok {
        fmt.Println(val) // hello
    }

    // Delete a key explicitly
    c.Delete("greeting")
}
```

### Storing Custom Structs

```go
type User struct {
    ID   int
    Name string
}

userCache := cache.NewLRUCache[User](50)
userCache.Set("user:1", User{ID: 1, Name: "Charu"}, -1) // no expiry

user, ok := userCache.Get("user:1")
// user is of type User — no type assertions needed!
```

## API

### `NewLRUCache[V any](capacity int) *LRUCache[V]`

Creates a new LRU cache that stores values of type `V`, with the given maximum capacity. A background goroutine is automatically started to sweep expired entries every minute.

### `Set(key string, val V, ttlInMs int64)`

Inserts or updates a key-value pair.

| `ttlInMs` | Behavior |
|---|---|
| `0` | No expiration |
| `> 0` | Expires after the given duration in milliseconds |
| `< 0` | Default TTL of 1 year |

If the cache is at capacity, the least recently used item is evicted before inserting.

### `Get(key string) (V, bool)`

Retrieves the value for a key. Returns the zero value of `V` and `false` if the key is missing or has expired. Accessing a key marks it as the most recently used item.

### `Delete(key string)`

Removes a key from the cache. No-op if the key does not exist.

## License

[MIT](LICENSE)
