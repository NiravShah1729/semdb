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