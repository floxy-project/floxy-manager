package simplecache

import (
	"sync"
	"testing"
	"time"
)

func TestCache_GetSet(t *testing.T) {
	cache := New[string, string]()

	// Test setting and getting a value
	cache.Set("key1", "value1", time.Minute)

	value, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if value != "value1" {
		t.Errorf("Expected 'value1', got '%s'", value)
	}

	// Test getting non-existent key
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("Expected not to find nonexistent key")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := New[string, int]()

	cache.Set("key1", 42, time.Minute)

	// Verify key exists
	value, found := cache.Get("key1")
	if !found || value != 42 {
		t.Error("Expected to find key1 with value 42")
	}

	// Delete key
	cache.Delete("key1")

	// Verify key is gone
	_, found = cache.Get("key1")
	if found {
		t.Error("Expected key1 to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := New[string, string]()

	cache.Set("key1", "value1", time.Minute)
	cache.Set("key2", "value2", time.Minute)

	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}

	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be cleared")
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := New[string, string]()

	// Set value with very short TTL
	cache.Set("key1", "value1", 10*time.Millisecond)

	// Should be available immediately
	value, found := cache.Get("key1")
	if !found || value != "value1" {
		t.Error("Expected to find key1 immediately")
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should be expired now
	_, found = cache.Get("key1")
	if found {
		t.Error("Expected key1 to be expired")
	}
}

func TestCache_TypeConversion(t *testing.T) {
	cache := New[string, any]()

	// Test different types
	cache.Set("string", "hello", time.Minute)
	cache.Set("int", 42, time.Minute)
	cache.Set("bool", true, time.Minute)
	cache.Set("float", 3.14, time.Minute)

	// Test string
	if value, found := cache.Get("string"); !found || value != "hello" {
		t.Error("String value test failed")
	}

	// Test int
	if value, found := cache.Get("int"); !found || value != 42 {
		t.Error("Int value test failed")
	}

	// Test bool
	if value, found := cache.Get("bool"); !found || value != true {
		t.Error("Bool value test failed")
	}

	// Test float
	if value, found := cache.Get("float"); !found || value != 3.14 {
		t.Error("Float value test failed")
	}
}

func TestCache_Concurrency(t *testing.T) {
	cache := New[int, int]()

	const numGoroutines = 100
	const numOperations = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := goroutineID*numOperations + j
				cache.Set(key, key*2, time.Minute)
			}
		}(i)
	}

	wg.Wait()

	// Verify some values
	for i := 0; i < 10; i++ {
		value, found := cache.Get(i)
		if !found {
			t.Errorf("Expected to find key %d", i)
		}
		if value != i*2 {
			t.Errorf("Expected value %d for key %d, got %d", i*2, i, value)
		}
	}
}

func TestCache_Cleanup(t *testing.T) {
	cache := New[string, string]()

	// Set some values with different TTLs
	cache.Set("short", "value1", 10*time.Millisecond)
	cache.Set("long", "value2", time.Minute)

	// Verify both exist
	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}

	// Wait for short TTL to expire
	time.Sleep(20 * time.Millisecond)

	// Run cleanup
	cache.Cleanup()

	// Short should be gone, long should remain
	if cache.Size() != 1 {
		t.Errorf("Expected size 1 after cleanup, got %d", cache.Size())
	}

	_, found := cache.Get("short")
	if found {
		t.Error("Expected 'short' to be cleaned up")
	}

	value, found := cache.Get("long")
	if !found || value != "value2" {
		t.Error("Expected 'long' to still exist")
	}
}

func TestCache_StartCleanup(t *testing.T) {
	cache := New[string, string]()

	// Start cleanup goroutine
	cache.StartCleanup(10 * time.Millisecond)

	// Set a value with short TTL
	cache.Set("test", "value", 20*time.Millisecond)

	// Verify it exists
	_, found := cache.Get("test")
	if !found {
		t.Error("Expected 'test' to exist initially")
	}

	// Wait for cleanup to run
	time.Sleep(50 * time.Millisecond)

	// Should be cleaned up by now
	_, found = cache.Get("test")
	if found {
		t.Error("Expected 'test' to be cleaned up by background goroutine")
	}
}

func TestCache_Size(t *testing.T) {
	cache := New[string, string]()

	// Initially empty
	if cache.Size() != 0 {
		t.Errorf("Expected initial size 0, got %d", cache.Size())
	}

	// Add some items
	cache.Set("key1", "value1", time.Minute)
	cache.Set("key2", "value2", time.Minute)

	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}

	// Add expired item (should still count in size until cleanup)
	cache.Set("key3", "value3", 1*time.Millisecond)
	time.Sleep(2 * time.Millisecond)

	if cache.Size() != 3 {
		t.Errorf("Expected size 3 (including expired), got %d", cache.Size())
	}

	// Cleanup should remove expired items
	cache.Cleanup()
	if cache.Size() != 2 {
		t.Errorf("Expected size 2 after cleanup, got %d", cache.Size())
	}
}

func TestCache_ZeroValue(t *testing.T) {
	cache := New[string, int]()

	// Test getting zero value for non-existent key
	value, found := cache.Get("nonexistent")
	if found {
		t.Error("Expected not to find nonexistent key")
	}
	if value != 0 {
		t.Errorf("Expected zero value for int, got %d", value)
	}
}

func TestCache_Overwrite(t *testing.T) {
	cache := New[string, string]()

	// Set initial value
	cache.Set("key1", "value1", time.Minute)

	value, found := cache.Get("key1")
	if !found || value != "value1" {
		t.Error("Expected initial value")
	}

	// Overwrite with new value
	cache.Set("key1", "value2", time.Minute)

	value, found = cache.Get("key1")
	if !found || value != "value2" {
		t.Error("Expected overwritten value")
	}
}

func TestCache_ComplexKey(t *testing.T) {
	cache := New[struct{ A, B int }, string]()

	key := struct{ A, B int }{A: 1, B: 2}
	cache.Set(key, "value", time.Minute)

	value, found := cache.Get(key)
	if !found || value != "value" {
		t.Error("Expected to find complex key")
	}

	// Test with different struct instance but same values
	key2 := struct{ A, B int }{A: 1, B: 2}
	value, found = cache.Get(key2)
	if !found || value != "value" {
		t.Error("Expected to find complex key with different instance")
	}
}

func BenchmarkCache_Set(b *testing.B) {
	cache := New[string, string]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", "value", time.Minute)
	}
}

func BenchmarkCache_Get(b *testing.B) {
	cache := New[string, string]()
	cache.Set("key", "value", time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

func BenchmarkCache_GetMiss(b *testing.B) {
	cache := New[string, string]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("nonexistent")
	}
}

func BenchmarkCache_Concurrent(b *testing.B) {
	cache := New[int, int]()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Set(i, i*2, time.Minute)
			cache.Get(i)
			i++
		}
	})
}
