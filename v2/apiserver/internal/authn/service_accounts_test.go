package authn

import (
	"context"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestServiceAccountMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, ServiceAccount{}, "ServiceAccount")
}

func TestServiceAccountListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		ServiceAccountList{},
		"ServiceAccountList",
	)
}

func TestNewServiceAccountService(t *testing.T) {
	store := &MockServiceAccountStore{}
	svc := NewServiceAccountsService(store)
	require.Same(t, store, svc.(*serviceAccountsService).store)
}

func TestServiceAccountsServiceCreate(t *testing.T) {
	testCases := []struct {
		name       string
		service    ServiceAccountsService
		assertions func(error)
	}{
		{
			name: "error creating service account in store",
			service: &serviceAccountsService{
				store: &MockServiceAccountStore{
					CreateFn: func(context.Context, ServiceAccount) error {
						return errors.New("store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "store error")
				require.Contains(t, err.Error(), "error storing new service account")
			},
		},
		{
			name: "success",
			service: &serviceAccountsService{
				store: &MockServiceAccountStore{
					CreateFn: func(context.Context, ServiceAccount) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := testCase.service.Create(context.Background(), ServiceAccount{})
			testCase.assertions(err)
		})
	}
}

func TestServiceAccountsServiceList(t *testing.T) {
	testCases := []struct {
		name       string
		service    ServiceAccountsService
		assertions func(error)
	}{
		{
			name: "error getting service accounts from store",
			service: &serviceAccountsService{
				store: &MockServiceAccountStore{
					ListFn: func(
						context.Context,
						meta.ListOptions,
					) (ServiceAccountList, error) {
						return ServiceAccountList{},
							errors.New("error listing service accounts")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error listing service accounts")
				require.Contains(
					t,
					err.Error(),
					"error retrieving service accounts from store",
				)
			},
		},
		{
			name: "success",
			service: &serviceAccountsService{
				store: &MockServiceAccountStore{
					ListFn: func(
						context.Context,
						meta.ListOptions,
					) (ServiceAccountList, error) {
						return ServiceAccountList{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err :=
				testCase.service.List(context.Background(), meta.ListOptions{})
			testCase.assertions(err)
		})
	}
}

func TestServiceAccountsServiceGet(t *testing.T) {
	testCases := []struct {
		name       string
		service    ServiceAccountsService
		assertions func(error)
	}{
		{
			name: "error getting service account from store",
			service: &serviceAccountsService{
				store: &MockServiceAccountStore{
					GetFn: func(context.Context, string) (ServiceAccount, error) {
						return ServiceAccount{}, errors.New("error getting service account")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error getting service account")
				require.Contains(t, err.Error(), "error retrieving service account")
			},
		},
		{
			name: "success",
			service: &serviceAccountsService{
				store: &MockServiceAccountStore{
					GetFn: func(context.Context, string) (ServiceAccount, error) {
						return ServiceAccount{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err :=
				testCase.service.Get(context.Background(), "jarvis")
			testCase.assertions(err)
		})
	}
}

func TestServiceAccountsServiceGetByToken(t *testing.T) {
	const testToken = "abcdefghijklmnopqrstuvwxyz"
	testCases := []struct {
		name       string
		service    ServiceAccountsService
		assertions func(error)
	}{
		{
			name: "error getting service account from store",
			service: &serviceAccountsService{
				store: &MockServiceAccountStore{
					GetByHashedTokenFn: func(
						context.Context,
						string,
					) (ServiceAccount, error) {
						return ServiceAccount{}, errors.New("error getting service account")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error getting service account")
				require.Contains(
					t, err.Error(),
					"error retrieving service account from store by token",
				)
			},
		},
		{
			name: "success",
			service: &serviceAccountsService{
				store: &MockServiceAccountStore{
					GetByHashedTokenFn: func(
						context.Context,
						string,
					) (ServiceAccount, error) {
						return ServiceAccount{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err :=
				testCase.service.GetByToken(context.Background(), testToken)
			testCase.assertions(err)
		})
	}
}

func TestServiceAccountsLock(t *testing.T) {
	testCases := []struct {
		name       string
		service    ServiceAccountsService
		assertions func(error)
	}{
		{
			name: "error updating service account in store",
			service: &serviceAccountsService{
				store: &MockServiceAccountStore{
					LockFn: func(context.Context, string) error {
						return errors.New("store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "store error")
				require.Contains(t, err.Error(), "error locking service account")
			},
		},
		{
			name: "success",
			service: &serviceAccountsService{
				store: &MockServiceAccountStore{
					LockFn: func(context.Context, string) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.service.Lock(context.Background(), "jarvis")
			testCase.assertions(err)
		})
	}
}

func TestServiceAccountsUnlock(t *testing.T) {
	testCases := []struct {
		name       string
		service    ServiceAccountsService
		assertions func(token Token, err error)
	}{
		{
			name: "error updating service account in store",
			service: &serviceAccountsService{
				store: &MockServiceAccountStore{
					UnlockFn: func(context.Context, string, string) error {
						return errors.New("store error")
					},
				},
			},
			assertions: func(_ Token, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "store error")
				require.Contains(t, err.Error(), "error unlocking service account")
			},
		},
		{
			name: "success",
			service: &serviceAccountsService{
				store: &MockServiceAccountStore{
					UnlockFn: func(context.Context, string, string) error {
						return nil
					},
				},
			},
			assertions: func(token Token, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, token.Value)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			token, err := testCase.service.Unlock(context.Background(), "jarvis")
			testCase.assertions(token, err)
		})
	}
}
