package store

import (
	"sync"
	"time"
)

// Shard represents a single partition of the store with its own lock and maps.
type Shard struct {
	mu     sync.RWMutex
	data   map[string]entry
	hashes map[string]map[string]string
}

// Store represents the sharded in-memory key-value database.
type Store struct {
	shards [numShards]Shard
	Semantic *SemanticStore
}

// entry holds a string value and an optional expiration timestamp.
type entry struct {
	value     string
	expiresAt time.Time
}

// NewStore initializes and returns a Store with all 16 shards allocated.
func NewStore() *Store {
	s := &Store{
		Semantic: NewSemanticStore(),
	}
	for i := 0; i < numShards; i++ {
		s.shards[i].data = make(map[string]entry)
		s.shards[i].hashes = make(map[string]map[string]string)
	}
	return s
}

// Get retrieves the value associated with a key if it exists and has not expired.
func (s *Store) Get(key string) (string, bool) {
	shard := s.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()

	val, ok := shard.data[key]
	if ok {
		if !val.expiresAt.IsZero() && time.Now().After(val.expiresAt) {
			return "", false
		}
		return val.value, true
	}
	return "", false
}

// Exists returns the count of given keys that currently exist and are unexpired.
func (s *Store) Exists(keys ...string) int {
	count := 0
	for _, key := range keys {
		shard := s.getShard(key)
		shard.mu.RLock()
		val, ok := shard.data[key]
		if ok {
			if !val.expiresAt.IsZero() && time.Now().After(val.expiresAt) {
				shard.mu.RUnlock()
				continue
			}
			count++
		}
		shard.mu.RUnlock()
	}
	return count
}

// Set stores a key-value pair with an optional expiration TTL.
func (s *Store) Set(key string, val string, ttl time.Duration) {
	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	shard.data[key] = entry{value: val, expiresAt: expiresAt}
}

// Del removes specified keys from their respective shards and returns deleted count.
func (s *Store) Del(keys ...string) int {
	count := 0
	for _, key := range keys {
		shard := s.getShard(key)
		shard.mu.Lock()
		if _, ok := shard.data[key]; ok {
			delete(shard.data, key)
			count++
		}
		shard.mu.Unlock()
	}
	return count
}

// Expire sets a TTL duration on an existing key, returning false if non-existent or expired.
func (s *Store) Expire(key string, ttl time.Duration) bool {
	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	val, ok := shard.data[key]
	if !ok {
		return false
	}
	if !val.expiresAt.IsZero() && time.Now().After(val.expiresAt) {
		delete(shard.data, key)
		return false
	}

	val.expiresAt = time.Now().Add(ttl)
	shard.data[key] = val
	return true
}

// TTL returns remaining seconds until key expiration (-1 if no TTL, -2 if missing/expired).
func (s *Store) TTL(key string) int64 {
	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	val, ok := shard.data[key]
	if !ok || (!val.expiresAt.IsZero() && time.Now().After(val.expiresAt)) {
		return int64(-2)
	}

	if val.expiresAt.IsZero() {
		return int64(-1)
	}

	return int64(time.Until(val.expiresAt).Seconds())
}

// deleteExpiredKeys iterates over all shards to remove expired entries safely.
func (s *Store) deleteExpiredKeys() {
	now := time.Now()
	for i := 0; i < numShards; i++ {
		shard := &s.shards[i]
		shard.mu.Lock()
		for key, item := range shard.data {
			if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
				delete(shard.data, key)
			}
		}
		shard.mu.Unlock()
	}
	s.Semantic.DeleteExpiredKeys()
}

// StartActiveEviction launches a background goroutine to periodically clean up expired keys.
func (s *Store) StartActiveEviction() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			s.deleteExpiredKeys()
		}
	}()
}

// Hash Commands

// HSet sets field in the hash stored at key; returns 1 if new field was created, 0 if updated.
func (s *Store) HSet(key, field, val string) int {
	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	hash, ok := shard.hashes[key]
	if !ok {
		hash = make(map[string]string)
		shard.hashes[key] = hash
	}

	_, exists := hash[field]
	hash[field] = val
	if exists {
		return 0
	}
	return 1
}

// HGet retrieves the value associated with field in the hash stored at key.
func (s *Store) HGet(key, field string) (string, bool) {
	shard := s.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()

	val, ok := shard.hashes[key]
	if ok {
		i, ok := val[field]
		if ok {
			return i, true
		}
	}
	return "", false
}

// HDel removes specified field from the hash stored at key; returns 1 if deleted, 0 if missing.
func (s *Store) HDel(key, field string) int {
	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if hash, ok := shard.hashes[key]; ok {
		if _, exists := hash[field]; exists {
			delete(hash, field)
			return 1
		}
	}
	return 0
}

// HGetAll returns a copy of all fields and values stored in the hash at key.
func (s *Store) HGetAll(key string) map[string]string {
	shard := s.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()

	val, ok := shard.hashes[key]
	result := make(map[string]string)
	if !ok {
		return result
	}
	for k, v := range val {
		result[k] = v
	}
	return result
}