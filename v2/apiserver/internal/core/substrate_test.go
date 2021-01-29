package core

import (
	"context"
	"errors"
	"testing"

	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestSubstrateWorkerCountMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		&SubstrateWorkerCount{},
		"SubstrateWorkerCount",
	)
}

func TestSubstrateJobCountMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		&SubstrateJobCount{},
		"SubstrateJobCount",
	)
}

func TestNewSubstrateService(t *testing.T) {
	substrate := &mockSubstrate{}
	svc := NewSubstrateService(libAuthz.AlwaysAuthorize, substrate)
	require.NotNil(t, svc.(*substrateService).authorize)
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
			name: "unauthorized",
			service: &substrateService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(_ SubstrateWorkerCount, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error counting workers in substrate",
			service: &substrateService{
				authorize: libAuthz.AlwaysAuthorize,
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
				authorize: libAuthz.AlwaysAuthorize,
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
			name: "unauthorized",
			service: &substrateService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(_ SubstrateJobCount, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error counting jobs in substrate",
			service: &substrateService{
				authorize: libAuthz.AlwaysAuthorize,
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
				authorize: libAuthz.AlwaysAuthorize,
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
