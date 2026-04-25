// Package cache provides a thread-safe, in-memory LRU (Least Recently Used) cache
// with optional per-key TTL (Time To Live) expiration.
//
// The cache uses a doubly linked list paired with a hash map to guarantee O(1)
// time complexity for both Get and Set operations. A background goroutine
// automatically evicts expired keys once per minute.
//
// Example usage:
//
//	cache := cache.NewLRUCache(100)
//	cache.Set("username", 42, 60000) // expires in 60 seconds
//	val := cache.Get("username")     // returns 42
package cache

import (
	"fmt"
	"sync"
	"time"
)

// LRUCache is a thread-safe, in-memory cache with LRU eviction and per-key TTL support.
// Use NewLRUCache to create an instance.
type LRUCache struct {
	capacity   int
	linkedList linkedList
	store      map[string]*node
	mu         sync.Mutex
}

type node struct {
	value int64
	next  *node
	prev  *node
	key   string
	ttl   time.Time
}

type linkedList struct {
	head *node
	tail *node
}

// NewLRUCache creates and returns a new LRUCache with the specified maximum capacity.
// Once the cache is full, the least recently used item is evicted to make room.
// A background goroutine is automatically started to sweep expired keys every minute.
func NewLRUCache(capacity int) *LRUCache {
	lruCache := &LRUCache{}
	lruCache.store = make(map[string]*node)
	lruCache.capacity = capacity

	go lruCache.startCleanupRoutine()
	return lruCache
}

// Get retrieves the value associated with the given key.
// If the key does not exist or has expired, it returns -1.
// Accessing a key marks it as the most recently used item.
func (cache *LRUCache) Get(key string) int64 {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	node, ok := cache.store[key]
	if !ok {
		return -1
	}

	if !node.ttl.IsZero() && time.Now().After(node.ttl) {
		cache.deleteLocked(key)
		return -1
	}

	cache.linkedList.MoveToTail(node)
	return node.value
}

// Set inserts or updates a key-value pair in the cache.
// ttlInMs is the time-to-live for the key in milliseconds.
// Pass 0 to store the key without an expiration.
// Pass a negative value to use the default TTL of 1 year.
// If the cache is at capacity, the least recently used item is evicted.
func (cache *LRUCache) Set(key string, val int64, ttlInMs int64) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if ttlInMs < 0 {
		ttlInMs = 1000 * 60 * 60 * 24 * 365 // set default to 1 year
	}

	node := cache.store[key]
	if node != nil {
		node.value = val
		if ttlInMs > 0 {
			node.ttl = time.Now().Add(time.Duration(ttlInMs) * time.Millisecond)
		}
		cache.linkedList.MoveToTail(node)
	} else {
		if len(cache.store)+1 > int(cache.capacity) {
			delete(cache.store, cache.linkedList.head.key)
			cache.linkedList.Remove(cache.linkedList.head)
		}
		var actualTTL time.Time
		if ttlInMs > 0 {
			actualTTL = time.Now().Add(time.Duration(ttlInMs) * time.Millisecond)
		}
		newNode := cache.linkedList.Append(val, key, actualTTL)
		cache.store[key] = newNode
	}
}

func (cache *LRUCache) deleteLocked(key string) {
	node := cache.store[key]

	if node == nil {
		return
	}

	cache.linkedList.Remove(node)
	delete(cache.store, key)
}

// Delete removes the entry with the given key from the cache.
// It is a no-op if the key does not exist.
func (cache *LRUCache) Delete(key string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.deleteLocked(key)
}

// startCleanupRoutine runs a blocking loop that sweeps the cache every minute,
// removing any entries whose TTL has expired. It is intended to be run in a goroutine.
func (cache *LRUCache) startCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)

	for {
		<-ticker.C
		cache.mu.Lock()
		curr := cache.linkedList.head

		for curr != nil {
			next := curr.next
			if !curr.ttl.IsZero() && time.Now().After(curr.ttl) {
				cache.deleteLocked(curr.key)
			}
			curr = next
		}
		cache.mu.Unlock()
	}
}

func (list *linkedList) Append(val int64, key string, ttl time.Time) *node {
	newNode := &node{value: val, key: key, ttl: ttl}

	// Handle the empty list scenario
	if list.tail == nil {
		list.head = newNode
		list.tail = newNode
		return newNode
	}

	newNode.prev = list.tail
	list.tail.next = newNode
	list.tail = newNode

	return newNode
}

func (list *linkedList) MoveToTail(node *node) {
	// 1. If it's already the tail, do nothing!
	if node == list.tail {
		return
	}

	// 2. Unlink it safely using the method you already built
	list.Remove(node)

	// 3. Now attach it to the tail
	node.prev = list.tail
	node.next = nil

	if list.tail != nil {
		list.tail.next = node
	}
	list.tail = node

	// (Optional edge case check: if list was empty, which shouldn't happen here, but good practice)
	if list.head == nil {
		list.head = node
	}
}

func (list *linkedList) PrintAll() {
	temp := list.head

	for temp != nil {
		fmt.Println(temp.value, temp.key)

		temp = temp.next
	}
}

func (list *linkedList) Remove(n *node) {

	if n == nil {
		return
	}

	if n.prev != nil {
		n.prev.next = n.next
	} else {
		// Removing the head
		list.head = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	} else {
		// Removing the tail
		list.tail = n.prev
	}

	n.prev = nil
	n.next = nil
}
