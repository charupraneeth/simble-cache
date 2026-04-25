package main

import (
	"fmt"
	"sync"
)

// cache
// we'll have a map
// you can add items to the map
// you can remove items from the map
// multiple people can read from the map
// only one would be able to update it

// we'll have a ttl set on each item
// every minute we have routine what checks ttl for each item and remove it

type LRUCache struct {
	capacity   int64
	linkedList LinkedList
	store      map[string]*Node
	mu         sync.Mutex
}

type Node struct {
	value int64
	next  *Node
	prev  *Node
	key   string
}

type LinkedList struct {
	head *Node
	tail *Node
}

func NewLRUCache(capacity int64) *LRUCache {
	lruCache := &LRUCache{}
	lruCache.linkedList = LinkedList{}
	lruCache.store = make(map[string]*Node)

	lruCache.capacity = capacity

	return lruCache
}

func (cache *LRUCache) Get(key string) int64 {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	node, ok := cache.store[key]

	if !ok {
		return -1
	}

	cache.linkedList.MoveToTail(node)
	return node.value

}

func (cache *LRUCache) Set(key string, val int64) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	node := cache.store[key]
	if node != nil {
		node.value = val
		cache.linkedList.MoveToTail(node)
	} else {
		if len(cache.store)+1 > int(cache.capacity) {
			delete(cache.store, cache.linkedList.head.key)
			cache.linkedList.Remove(cache.linkedList.head)
		}
		newNode := cache.linkedList.Append(val, key)
		cache.store[key] = newNode
	}
}

func (list *LinkedList) Append(val int64, key string) *Node {
	newNode := &Node{value: val, key: key}

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

func (list *LinkedList) MoveToTail(node *Node) {
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

func (list *LinkedList) PrintAll() {
	temp := list.head

	for temp != nil {
		fmt.Println(temp.value, temp.key)

		temp = temp.next
	}
}

func (list *LinkedList) Reverse() {
	if list.head == nil {
		return
	}

	list.tail = list.head

	var prev *Node
	curr := list.head

	for curr != nil {
		next := curr.next
		curr.next = prev
		prev = curr
		curr = next
	}

	list.head = prev
}

func (list *LinkedList) HasCycle() bool {
	slow := list.head
	fast := list.head

	for fast != nil && fast.next != nil {

		slow = slow.next

		fast = fast.next.next

		if slow == fast {
			return true
		}
	}

	return false
}

func (list *LinkedList) Remove(n *Node) {

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

func main() {
	cache := NewLRUCache(50) // slightly larger capacity for the test

	// A WaitGroup waits for a collection of goroutines to finish
	var wg sync.WaitGroup

	fmt.Println("Starting stress test...")

	// Spawn 1,000 goroutines hitting the cache at the same time
	for i := 0; i < 1000; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", workerID%10) // Only 10 unique keys, forcing lots of evictions and overwrites

			// Concurrently write
			cache.Set(key, int64(workerID))

			// Concurrently read
			cache.Get(key)

		}(i)
	}

	// Block until all 1,000 goroutines are done
	wg.Wait()

	fmt.Println("Stress test complete! Cache state:")
	cache.linkedList.PrintAll()
}
