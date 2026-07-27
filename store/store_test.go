package store

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestShardedStoreBasic(t *testing.T) {
	s := NewStore()

	// Test SET & GET
	s.Set("key1", "val1", 0)
	val, ok := s.Get("key1")
	if !ok || val != "val1" {
		t.Errorf("Expected val1, got %v (ok=%v)", val, ok)
	}

	// Test GET non-existent
	_, ok = s.Get("key2")
	if ok {
		t.Errorf("Expected false for non-existent key")
	}

	// Test EXISTS
	s.Set("key2", "val2", 0)
	count := s.Exists("key1", "key2", "key3")
	if count != 2 {
		t.Errorf("Expected 2 existing keys, got %d", count)
	}

	// Test DEL
	delCount := s.Del("key1", "key3")
	if delCount != 1 {
		t.Errorf("Expected 1 deleted key, got %d", delCount)
	}

	_, ok = s.Get("key1")
	if ok {
		t.Errorf("Expected key1 to be deleted")
	}
}

func TestShardedStoreTTL(t *testing.T) {
	s := NewStore()

	// Test TTL on non-existent key
	if ttl := s.TTL("nonexistent"); ttl != -2 {
		t.Errorf("Expected -2 for non-existent key TTL, got %d", ttl)
	}

	// Test key without expiration
	s.Set("perm", "val", 0)
	if ttl := s.TTL("perm"); ttl != -1 {
		t.Errorf("Expected -1 for key without expiration, got %d", ttl)
	}

	// Test EXPIRE & TTL
	s.Set("temp", "val", 0)
	ok := s.Expire("temp", 2*time.Second)
	if !ok {
		t.Errorf("Expected Expire to return true")
	}

	ttl := s.TTL("temp")
	if ttl <= 0 || ttl > 2 {
		t.Errorf("Expected TTL around 2s, got %d", ttl)
	}

	// Expire non-existent key
	if ok := s.Expire("nonexistent", 5*time.Second); ok {
		t.Errorf("Expected Expire to return false for missing key")
	}
}

func TestShardedStoreHashes(t *testing.T) {
	s := NewStore()

	// HSET
	res := s.HSet("myhash", "field1", "val1")
	if res != 1 {
		t.Errorf("Expected HSet to return 1 for new field, got %d", res)
	}

	res = s.HSet("myhash", "field1", "val1_updated")
	if res != 0 {
		t.Errorf("Expected HSet to return 0 for existing field, got %d", res)
	}

	// HGET
	val, ok := s.HGet("myhash", "field1")
	if !ok || val != "val1_updated" {
		t.Errorf("Expected val1_updated, got %v (ok=%v)", val, ok)
	}

	_, ok = s.HGet("myhash", "field2")
	if ok {
		t.Errorf("Expected false for non-existent field")
	}

	// HGETALL
	s.HSet("myhash", "field2", "val2")
	all := s.HGetAll("myhash")
	if len(all) != 2 || all["field1"] != "val1_updated" || all["field2"] != "val2" {
		t.Errorf("Unexpected HGetAll output: %v", all)
	}

	// HDEL
	delRes := s.HDel("myhash", "field1")
	if delRes != 1 {
		t.Errorf("Expected HDel to return 1, got %d", delRes)
	}

	_, ok = s.HGet("myhash", "field1")
	if ok {
		t.Errorf("Expected field1 to be deleted from hash")
	}
}

func TestShardedStoreConcurrency(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	workers := 20
	opsPerWorker := 100

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < opsPerWorker; j++ {
				key := fmt.Sprintf("key_%d_%d", workerID, j)
				val := fmt.Sprintf("val_%d", j)
				s.Set(key, val, 0)
				s.Get(key)
				s.HSet("shared_hash", key, val)
				s.HGet("shared_hash", key)
			}
		}(i)
	}

	wg.Wait()
}
