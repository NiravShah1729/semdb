package store
import (
	"sync"
	"math"
	"fmt"
	"time"
)
type SemanticEntry struct {
	Key       string
	Text      string
	Vector    []float32
	Value     string
	expiresAt time.Time
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

func (s *SemanticStore) Add(key, text, value string, vector []float32, ttl time.Duration){
	s.mu.Lock()
	defer s.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	s.entries = append(s.entries, SemanticEntry{
		Key:       key,
		Text:      text,
		Vector:    vector,
		Value:     value,
		expiresAt: expiresAt,
	})
}

func (s *SemanticStore) Search(vector []float32, threshold float32) (string, bool){
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.entries) == 0{
		return "",false
	}

	var maxi float32
	maxi = 0

	var temp SemanticEntry
	now := time.Now()
	for _, val := range s.entries {
		if !val.expiresAt.IsZero() && now.After(val.expiresAt) {
			continue
		}
		similarity := cosineSimilarity(vector, val.Vector)
		fmt.Printf("Comparing '%s' with query. Similarity: %f\n", val.Text, similarity)
		if similarity > maxi {
			maxi = similarity
			temp = val
		}
	}

	if maxi >= threshold {
		return temp.Value, true
	} else {
		return "", false
	}
}

func (s *SemanticStore) DeleteExpiredKeys() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	active := make([]SemanticEntry, 0, len(s.entries))
	for _, entry := range s.entries {
		if entry.expiresAt.IsZero() || now.Before(entry.expiresAt) {
			active = append(active, entry)
		}
	}
	s.entries = active
}