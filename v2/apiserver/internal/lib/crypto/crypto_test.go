package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
