package store

import (
	"hash/fnv"
)

const numShards = 16

// hashKey calculates a 32-bit FNV-1a hash for a given string key.
func hashKey(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

// getShard returns the Shard responsible for the specified key.
func (s *Store) getShard(key string) *Shard {
	index := hashKey(key) % numShards
	return &s.shards[index]
}
