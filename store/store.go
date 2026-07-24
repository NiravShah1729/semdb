package store

import "sync"

type Store struct {
	mu sync.RWMutex
	data map[string]string
}

func NewStore() *Store{
	return &Store{
		data: make(map[string]string),
	}
}

func (s *Store) Get(key string) (string,bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()	

	val,ok := s.data[key]
	if ok {
		return val,true
	}
	return "",false

}
func (s *Store) Exists(keys ...string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _,i := range keys {
		_,ok := s.data[i]
		if ok {
			count++
		}
	}
	return count
}
func (s *Store) Set(key string,val string){
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = val
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

