package rand

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSeeded(t *testing.T) {
	rand := NewSeeded()
	require.NotNil(t, rand)
	require.IsType(t, &seeded{}, rand)
	require.NotNil(t, rand.(*seeded).seededRand)
	require.NotNil(t, rand.(*seeded).mut)
}

func TestIntn(t *testing.T) {
	// We have no visibility into the underlying *mathrand.Rand, so we'll test
	// that it is indeed seeded by comparing the results of Intn ro results from
	// unseeded Rand. We use a very large n to reduce the liklihood of a
	// coincidental match.
	const n = 2147483647
	seededRand := NewSeeded()
	require.NotEqual(t, rand.Intn(n), seededRand.Intn(n))
}
