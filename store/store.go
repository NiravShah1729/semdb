package store

import (
	"sync"
	"time"
)

type Store struct {
	mu sync.RWMutex
	data map[string]entry
}

type entry struct {
	value string
	expiresAt time.Time
}

func NewStore() *Store{
	return &Store{
		data: make(map[string]entry),
	}
}

func (s *Store) Get(key string) (string,bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()	

	val,ok := s.data[key]
	if ok {
		if !val.expiresAt.IsZero() && time.Now().After(val.expiresAt) {
			return "",false
		}else{
			return val.value,true
		}
	}
	return "",false

}
func (s *Store) Exists(keys ...string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _,i := range keys {
		val,ok := s.data[i]
		if ok {
			if !val.expiresAt.IsZero() && time.Now().After(val.expiresAt) {
				continue
			}else{
				count++
			}
		}
	}
	return count
}
func (s *Store) Set(key string,val string,ttl time.Duration){
	s.mu.Lock()
	defer s.mu.Unlock()
	var expiresAt time.Time
	if ttl > 0{
		expiresAt = time.Now().Add(ttl)
	}
	s.data[key] = entry{val,expiresAt}
}

func (s *Store) Del(keys ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for _,i := range keys {
		_,ok := s.data[i]
		if ok {
			delete(s.data,i)
			count++
		}
	}
	return count
}

func (s *Store) Expire(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	val,ok := s.data[key]

	if !ok {
		return false
	}
	// Treat key as non-existent if it already expired
    if !val.expiresAt.IsZero() && time.Now().After(val.expiresAt) {
        delete(s.data, key) // Cleanup expired key
        return false
    }

    val.expiresAt = time.Now().Add(ttl)
    s.data[key] = val // Write back to map
    return true
}

func (s *Store) TTL(key string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	val,ok := s.data[key]

	if !ok || (!val.expiresAt.IsZero() && time.Now().After(val.expiresAt)) {
    	return int64(-2)
	}
	
	if val.expiresAt.IsZero() {
		return int64(-1)
	}

	return int64(time.Until(val.expiresAt).Seconds())
}

func (s *Store) deleteExpiredKeys(){
	s.mu.Lock()
	defer s.mu.Unlock()
	for val,i := range s.data {
		if !i.expiresAt.IsZero() && time.Now().After(i.expiresAt) {
			delete(s.data,val)
		}
	} 
}

func (s *Store) StartActiveEviction(){
	go func () {
		ticker := time.NewTicker(10*time.Second)
		defer ticker.Stop()
		for range ticker.C {
			
			s.deleteExpiredKeys()
			
		}
	}()
}