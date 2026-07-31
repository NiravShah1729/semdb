package store
import (
	"sync"
	"math"
)
type SemanticEntry struct {
	Key string
	Text string
	Vector []float32
	Value string
}

type SemanticStore struct {
	mu sync.RWMutex
	entries []SemanticEntry
}

func NewSemanticStore() *SemanticStore{
	return &SemanticStore{
		entries: make([]SemanticEntry, 0),
	}
}

func cosineSimilarity(a, b []float32) float32 {
    var dot, normA, normB float32
    for i := range a {
        dot += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    return dot / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

func (s *SemanticStore) Add(key, text,value string , vector []float32){
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = append(s.entries, SemanticEntry{
		Key: key,
		Text: text,
		Vector: vector,
		Value: value,
	})
}

func (s *SemanticStore) Search(vector []float32,threshold float32) (string, bool){
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.entries) == 0{
		return "",false
	}

	var maxi float32
	maxi = 0

	var temp SemanticEntry
	for _,val := range s.entries {
		similarity := cosineSimilarity(vector,val.Vector)
		if similarity > maxi {
			maxi = similarity
			temp = val
		}
	}

	if maxi >= threshold {
		return temp.Value,true
	}else{
		return "",false
	}
}