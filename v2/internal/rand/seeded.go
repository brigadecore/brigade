package rand

import (
	mathrand "math/rand"
	"sync"
	"time"
)

// Seeded is an interface for seeded, concurrency-safe random number generators.
type Seeded interface {
	// Intn returns, as an int, a non-negative pseudo-random number in [0,n). It
	// panics if n <= 0.
	Intn(max int) int
	// Float64 returns, as a float64, a pseudo-random number in [0.0,1.0).
	Float64() float64
}

type seeded struct {
	seededRand *mathrand.Rand
	mut        *sync.Mutex
}

// NewSeeded returns a seeded, concurrency-safe random number generator.
func NewSeeded() Seeded {
	rnd := mathrand.New(mathrand.NewSource(time.Now().UTC().UnixNano()))
	return &seeded{
		seededRand: rnd,
		mut:        &sync.Mutex{},
	}
}

func (s *seeded) Intn(max int) int {
	s.mut.Lock()
	defer s.mut.Unlock()
	return s.seededRand.Intn(max)
}

func (s *seeded) Float64() float64 {
	s.mut.Lock()
	defer s.mut.Unlock()
	return s.seededRand.Float64()
}
