package os

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetEnvVar(t *testing.T) {
	const testDefaultVal = "default"
	testCases := []struct {
		name       string
		setup      func()
		assertions func()
	}{
		{
			name: "env var exists",
			setup: func() {
				err := os.Setenv("FOO1", "foo")
				require.NoError(t, err)
			},
			assertions: func() {
				require.Equal(t, "foo", GetEnvVar("FOO1", testDefaultVal))
			},
		},
		{
			name: "env var does not exist",
			assertions: func() {
				require.Equal(t, testDefaultVal, GetEnvVar("FOO2", testDefaultVal))
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup()
			}
			testCase.assertions()
		})
	}
}

func TestGetRequiredEnvVar(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func()
	}{
		{
			name: "env var exists",
			setup: func() {
				err := os.Setenv("BAR1", "bar")
				require.NoError(t, err)
			},
			assertions: func() {
				val, err := GetRequiredEnvVar("BAR1")
				require.NoError(t, err)
				require.Equal(t, "bar", val)
			},
		},
		{
			name: "env var does not exist",
			assertions: func() {
				_, err := GetRequiredEnvVar("BAR2")
				require.Error(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup()
			}
			testCase.assertions()
		})
	}
}

func TestGetStringSliceFromEnvVar(t *testing.T) {
	testDefaultVal := []string{}
	testCases := []struct {
		name       string
		setup      func()
		assertions func()
	}{
		{
			name: "env var exists",
			setup: func() {
				err := os.Setenv("SLICE1", "foo,bar")
				require.NoError(t, err)
			},
			assertions: func() {
				require.Equal(
					t,
					[]string{"foo", "bar"},
					GetStringSliceFromEnvVar("SLICE1", testDefaultVal),
				)
			},
		},
		{
			name: "env var does not exist",
			assertions: func() {
				require.Equal(
					t,
					testDefaultVal,
					GetStringSliceFromEnvVar("SLICE2", testDefaultVal),
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup()
			}
			testCase.assertions()
		})
	}
}

func TestGetIntFromEnvVar(t *testing.T) {
	const testDefaultVal = 1
	testCases := []struct {
		name       string
		setup      func()
		assertions func()
	}{
		{
			name: "env var exists",
			setup: func() {
				err := os.Setenv("BAT1", "42")
				require.NoError(t, err)
			},
			assertions: func() {
				val, err := GetIntFromEnvVar("BAT1", testDefaultVal)
				require.NoError(t, err)
				require.Equal(t, 42, val)
			},
		},
		{
			name: "env var does not exist",
			assertions: func() {
				val, err := GetIntFromEnvVar("BAT2", testDefaultVal)
				require.NoError(t, err)
				require.Equal(t, testDefaultVal, val)
			},
		},
		{
			name: "env var value not parsable as int",
			setup: func() {
				err := os.Setenv("BAT3", "life, the universe, and everything")
				require.NoError(t, err)
			},
			assertions: func() {
				_, err := GetIntFromEnvVar("BAT3", testDefaultVal)
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as an int")
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup()
			}
			testCase.assertions()
		})
	}
}

func TestGetBoolFromEnvVar(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func()
	}{
		{
			name: "env var exists",
			setup: func() {
				err := os.Setenv("BAZ1", "true")
				require.NoError(t, err)
			},
			assertions: func() {
				val, err := GetBoolFromEnvVar("BAZ1", false)
				require.NoError(t, err)
				require.Equal(t, true, val)
			},
		},
		{
			name: "env var does not exist",
			assertions: func() {
				val, err := GetBoolFromEnvVar("BAZ2", true)
				require.NoError(t, err)
				require.True(t, val)
			},
		},
		{
			name: "env var value not parsable as int",
			setup: func() {
				err := os.Setenv("BAZ3", "not really")
				require.NoError(t, err)
			},
			assertions: func() {
				_, err := GetBoolFromEnvVar("BAZ3", false)
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a bool")
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup()
			}
			testCase.assertions()
		})
	}
}

func TestGetDurationFromEnvVar(t *testing.T) {
	const testDefaultVal = time.Minute
	testCases := []struct {
		name       string
		setup      func()
		assertions func()
	}{
		{
			name: "env var exists",
			setup: func() {
				err := os.Setenv("BAZ1", "20s")
				require.NoError(t, err)
			},
			assertions: func() {
				val, err := GetDurationFromEnvVar("BAZ1", testDefaultVal)
				require.NoError(t, err)
				require.Equal(t, 20*time.Second, val)
			},
		},
		{
			name: "env var does not exist",
			assertions: func() {
				val, err := GetDurationFromEnvVar("BAZ2", testDefaultVal)
				require.NoError(t, err)
				require.Equal(t, testDefaultVal, val)
			},
		},
		{
			name: "env var value not parsable as duration",
			setup: func() {
				err := os.Setenv("BAZ3", "life, the universe, and everything")
				require.NoError(t, err)
			},
			assertions: func() {
				_, err := GetDurationFromEnvVar("BAZ3", testDefaultVal)
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a duration")
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup()
			}
			testCase.assertions()
		})
	}
}
