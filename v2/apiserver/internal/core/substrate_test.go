package core

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubstrateWorkerCountMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, &SubstrateWorkerCount{}, "SubstrateWorkerCount")
}

func TestSubstrateJobCountMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, &SubstrateJobCount{}, "SubstrateJobCount")
}

func TestNewSubstrateService(t *testing.T) {
	substrate := &mockSubstrate{}
	svc := NewSubstrateService(substrate)
	require.Same(t, substrate, svc.(*substrateService).substrate)
}

func TestSubstrateServiceCountRunningWorkers(t *testing.T) {
	const testCount = 5
	testCases := []struct {
		name       string
		service    SubstrateService
		assertions func(SubstrateWorkerCount, error)
	}{
		{
			name: "error counting workers in substrate",
			service: &substrateService{
				substrate: &mockSubstrate{
					CountRunningWorkersFn: func(
						context.Context,
					) (SubstrateWorkerCount, error) {
						return SubstrateWorkerCount{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(_ SubstrateWorkerCount, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(
					t,
					err.Error(),
					"error counting running workers on substrate",
				)
			},
		},
		{
			name: "success",
			service: &substrateService{
				substrate: &mockSubstrate{
					CountRunningWorkersFn: func(
						context.Context,
					) (SubstrateWorkerCount, error) {
						return SubstrateWorkerCount{
							Count: testCount,
						}, nil
					},
				},
			},
			assertions: func(count SubstrateWorkerCount, err error) {
				require.NoError(t, err)
				require.Equal(t, testCount, count.Count)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			count, err := testCase.service.CountRunningWorkers(context.Background())
			testCase.assertions(count, err)
		})
	}
}

func TestSubstrateServiceCountRunningJobs(t *testing.T) {
	const testCount = 5
	testCases := []struct {
		name       string
		service    SubstrateService
		assertions func(SubstrateJobCount, error)
	}{
		{
			name: "error counting jobs in substrate",
			service: &substrateService{
				substrate: &mockSubstrate{
					CountRunningJobsFn: func(
						context.Context,
					) (SubstrateJobCount, error) {
						return SubstrateJobCount{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(_ SubstrateJobCount, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(
					t,
					err.Error(),
					"error counting running jobs on substrate",
				)
			},
		},
		{
			name: "success",
			service: &substrateService{
				substrate: &mockSubstrate{
					CountRunningJobsFn: func(
						context.Context,
					) (SubstrateJobCount, error) {
						return SubstrateJobCount{
							Count: testCount,
						}, nil
					},
				},
			},
			assertions: func(count SubstrateJobCount, err error) {
				require.NoError(t, err)
				require.Equal(t, testCount, count.Count)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			count, err := testCase.service.CountRunningJobs(context.Background())
			testCase.assertions(count, err)
		})
	}
}
