package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	testCases := []struct {
		name string
		salt string
	}{
		{
			name: "without salt",
			salt: "",
		},
		{
			name: "with salt",
			salt: "n pepa",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			const testInput = "shoop"
			hash := Hash(testCase.salt, testInput)
			assert.NotEqual(t, testInput, hash)
			// This is how long a sha256 sum should be
			assert.Equal(t, 64, len(hash))
		})
	}
}

func TestNewToken(t *testing.T) {
	testCases := []struct {
		name         string
		requestedLen int
		expectedLen  int
	}{
		{
			name:         "requested length < 64",
			requestedLen: 63,
			expectedLen:  64,
		},
		{
			name:         "requested length >= 64",
			requestedLen: 64,
			expectedLen:  64,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			token := NewToken(testCase.requestedLen)
			require.Equal(t, testCase.expectedLen, len(token))
		})
	}
}
