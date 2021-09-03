package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveEnvVars(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func()
		input          string
		expectedResult string
	}{
		{
			name:           "no substitutions",
			input:          "foobar",
			expectedResult: "foobar",
		},
		{
			name:           "failed substitution",
			input:          "${SUB}bar",
			expectedResult: "bar",
		},
		{
			name: "one substitution",
			setup: func() {
				t.Setenv("SUB", "foo")
			},
			input:          "${SUB}bar",
			expectedResult: "foobar",
		},
		{
			name: "multiple substitutions",
			setup: func() {
				t.Setenv("SUB1", "foo")
				t.Setenv("SUB2", "bar")
			},
			input:          "${SUB1}${SUB2}",
			expectedResult: "foobar",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup()
			}
			require.Equal(t, testCase.expectedResult, resolveEnvVars(testCase.input))
		})
	}
}
