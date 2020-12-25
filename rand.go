package retry

import (
	"math/rand"
	"sync"
)

// Randomizer is the rand interface
//go:generate mockgen -destination=mocks/mock_randomizer.go -package=mocks . Randomizer
type Randomizer interface {
	Int63n(int64) int64
}

// Rand is math.Rand wrapper
// with concurency protection.
//
type Rand struct {
	mu   sync.RWMutex
	rand *rand.Rand
}

// NewRand initialize new Rand.
//
func NewRand(seed int64) *Rand {
	randSource := rand.NewSource(seed)
	r := rand.New(randSource)

	return &Rand{
		rand: r,
	}
}

// Int63n calls rand.Int63n with lock protection.
//
func (r *Rand) Int63n(n int64) int64 {
	r.mu.Lock()
	result := r.rand.Int63n(n)
	r.mu.Unlock()
	return result
}
