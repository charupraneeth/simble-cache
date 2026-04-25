package cache

import (
	"sync"
	"testing"
	"time"
)

// --- Get / Set ---

func TestSetAndGet(t *testing.T) {
	c := NewLRUCache[string](5)
	c.Set("hello", "world", 0)

	val, ok := c.Get("hello")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if val != "world" {
		t.Fatalf("expected 'world', got %q", val)
	}
}

func TestGetMissingKey(t *testing.T) {
	c := NewLRUCache[int](5)

	_, ok := c.Get("missing")
	if ok {
		t.Fatal("expected missing key to return false")
	}
}

func TestSetUpdatesExistingKey(t *testing.T) {
	c := NewLRUCache[string](5)
	c.Set("key", "first", 0)
	c.Set("key", "second", 0)

	val, ok := c.Get("key")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if val != "second" {
		t.Fatalf("expected 'second', got %q", val)
	}
}

// --- Delete ---

func TestDelete(t *testing.T) {
	c := NewLRUCache[string](5)
	c.Set("key", "value", 0)
	c.Delete("key")

	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestDeleteMissingKeyIsNoop(t *testing.T) {
	c := NewLRUCache[string](5)
	// Should not panic
	c.Delete("nonexistent")
}

// --- LRU Eviction ---

func TestLRUEviction(t *testing.T) {
	c := NewLRUCache[int](3)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Set("c", 3, 0)

	// Access "a" to mark it as recently used
	c.Get("a")

	// Adding "d" should evict the LRU item, which is now "b"
	c.Set("d", 4, 0)

	_, ok := c.Get("b")
	if ok {
		t.Fatal("expected 'b' to be evicted as LRU")
	}

	// "a", "c", and "d" should still be present
	for _, key := range []string{"a", "c", "d"} {
		if _, ok := c.Get(key); !ok {
			t.Fatalf("expected key %q to still exist", key)
		}
	}
}

func TestLRUEvictionOrder(t *testing.T) {
	c := NewLRUCache[int](2)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)

	// "a" is now LRU. Adding "c" should evict "a".
	c.Set("c", 3, 0)

	if _, ok := c.Get("a"); ok {
		t.Fatal("expected 'a' to be evicted")
	}
	if _, ok := c.Get("b"); !ok {
		t.Fatal("expected 'b' to still exist")
	}
	if _, ok := c.Get("c"); !ok {
		t.Fatal("expected 'c' to still exist")
	}
}

// --- TTL ---

func TestTTLExpiry(t *testing.T) {
	c := NewLRUCache[string](5)

	// Set with a 50ms TTL
	c.Set("short", "lived", 50)

	// Should exist immediately
	if _, ok := c.Get("short"); !ok {
		t.Fatal("expected key to exist before expiry")
	}

	// Wait for it to expire
	time.Sleep(100 * time.Millisecond)

	// Should be gone now
	if _, ok := c.Get("short"); ok {
		t.Fatal("expected key to be expired")
	}
}

func TestNoTTLDoesNotExpire(t *testing.T) {
	c := NewLRUCache[string](5)
	c.Set("persistent", "value", 0)

	time.Sleep(50 * time.Millisecond)

	if _, ok := c.Get("persistent"); !ok {
		t.Fatal("expected key with no TTL to persist")
	}
}

// --- Generics ---

func TestGenericStruct(t *testing.T) {
	type User struct{ Name string }

	c := NewLRUCache[User](5)
	c.Set("user:1", User{Name: "Charu"}, 0)

	user, ok := c.Get("user:1")
	if !ok {
		t.Fatal("expected user to exist")
	}
	if user.Name != "Charu" {
		t.Fatalf("expected 'Charu', got %q", user.Name)
	}
}

// --- Concurrency ---

func TestConcurrentAccess(t *testing.T) {
	c := NewLRUCache[int](50)

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := "key"
			c.Set(key, id, -1)
			c.Get(key)
		}(i)
	}
	wg.Wait()
}
