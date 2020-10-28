package retries

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestManagedRetries(t *testing.T) {
	const testMaxAttempts = 10
	const testProcess = "foo"
	// Override the function that returns the backoff intervals so that we
	// can preempt long waits.
	jitteredExpBackoffFn = func(int, time.Duration) time.Duration {
		return time.Millisecond
	}
	testCases := []struct {
		name       string
		fn         func() (bool, error)
		assertions func(err error)
	}{
		{
			name: "unrecoverable error",
			fn: func() (bool, error) {
				return false, errors.New("don't bother") // A unrecoverable failure
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Equal(t, err.Error(), "don't bother")
			},
		},
		{
			name: "fn handled error internally, but requests retry",
			fn: func() (bool, error) {
				return true, nil // Always retry
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t, err.Error(),
					fmt.Sprintf(
						"failed %d attempt(s) to %s",
						testMaxAttempts,
						testProcess,
					),
				)
			},
		},
		{
			name: "fn returns error",
			fn: func() (bool, error) {
				return true, errors.New("something went wrong") // Always retry
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t, err.Error(),
					fmt.Sprintf(
						"failed %d attempt(s) to %s",
						testMaxAttempts,
						testProcess,
					),
				)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "fn succeeds",
			fn: func() (bool, error) {
				return false, nil // All good
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := ManageRetries(
				context.Background(),
				testProcess,
				testMaxAttempts,
				time.Minute,
				testCase.fn,
			)
			testCase.assertions(err)
		})
	}
}

func TestJitteredExpBackoff(t *testing.T) {
	testCases := []struct {
		failureCount int
		cap          time.Duration
		expectedMin  time.Duration
		expectedMax  time.Duration
	}{
		{
			failureCount: 1,
			cap:          time.Minute,
			expectedMin:  time.Second,
			expectedMax:  2 * time.Second,
		},
		{
			failureCount: 2,
			cap:          time.Minute,
			expectedMin:  2 * time.Second,
			expectedMax:  4 * time.Second,
		},
		{
			failureCount: 3,
			cap:          time.Minute,
			expectedMin:  4 * time.Second,
			expectedMax:  8 * time.Second,
		},
		{
			failureCount: 4,
			cap:          time.Minute,
			expectedMin:  8 * time.Second,
			expectedMax:  16 * time.Second,
		},
		{
			failureCount: 5,
			cap:          time.Minute,
			expectedMin:  16 * time.Second,
			expectedMax:  32 * time.Second,
		},
		{
			failureCount: 6,
			cap:          time.Minute,
			expectedMin:  30 * time.Second,
			expectedMax:  time.Minute,
		},
		{
			failureCount: 7,
			cap:          time.Minute,
			expectedMin:  30 * time.Second,
			expectedMax:  time.Minute,
		},
		{
			failureCount: 8,
			cap:          time.Minute,
			expectedMin:  30 * time.Second,
			expectedMax:  time.Minute,
		},
	}
	for _, testCase := range testCases {
		t.Run(strconv.Itoa(testCase.failureCount), func(t *testing.T) {
			delay1 := jitteredExpBackoff(testCase.failureCount, testCase.cap)

			require.Less(t, testCase.expectedMin.Seconds(), delay1.Seconds())
			require.Less(t, delay1.Seconds(), testCase.expectedMax.Seconds())

			// Make sure the jitter works
			delay2 := jitteredExpBackoff(testCase.failureCount, testCase.cap)
			require.NotEqual(t, delay1, delay2)
		})
	}
}
