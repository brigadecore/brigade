package api

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

func TestNewServiceAccountService(t *testing.T) {
	serviceAccountsStore := &mockServiceAccountStore{}
	roleAssignmentsStore := &mockRoleAssignmentsStore{}
	projectRoleAssignmentsStore := &mockProjectRoleAssignmentsStore{}
	svc, ok := NewServiceAccountsService(
		alwaysAuthorize,
		serviceAccountsStore,
		roleAssignmentsStore,
		projectRoleAssignmentsStore,
	).(*serviceAccountsService)
	require.True(t, ok)
	require.NotNil(t, svc.authorize)
	require.Same(t, serviceAccountsStore, svc.serviceAccountsStore)
}

func TestServiceAccountsServiceCreate(t *testing.T) {
	testCases := []struct {
		name       string
		service    ServiceAccountsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &serviceAccountsService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error creating service account in store",
			service: &serviceAccountsService{
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
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
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
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
			name: "unauthorized",
			service: &serviceAccountsService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting service accounts from store",
			service: &serviceAccountsService{
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
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
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
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
			name: "unauthorized",
			service: &serviceAccountsService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting service account from store",
			service: &serviceAccountsService{
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
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
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
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
				serviceAccountsStore: &mockServiceAccountStore{
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
				serviceAccountsStore: &mockServiceAccountStore{
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
			name: "unauthorized",
			service: &serviceAccountsService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error updating service account in store",
			service: &serviceAccountsService{
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
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
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
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
			name: "unauthorized",
			service: &serviceAccountsService{
				authorize: neverAuthorize,
			},
			assertions: func(_ Token, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error updating service account in store",
			service: &serviceAccountsService{
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
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
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
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

func TestServiceAccountsDelete(t *testing.T) {
	testCases := []struct {
		name       string
		service    ServiceAccountsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &serviceAccountsService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error deleting role assignments",
			service: &serviceAccountsService{
				authorize: alwaysAuthorize,
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					RevokeByPrincipalFn: func(context.Context, PrincipalReference) error {
						return errors.New("store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "store error")
				require.Contains(t, err.Error(), "error deleting service account")
				require.Contains(t, err.Error(), "role assignments")
			},
		},
		{
			name: "error deleting project role assignments",
			service: &serviceAccountsService{
				authorize: alwaysAuthorize,
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					RevokeByPrincipalFn: func(context.Context, PrincipalReference) error {
						return nil
					},
				},
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					RevokeByPrincipalFn: func(context.Context, PrincipalReference) error {
						return errors.New("store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "store error")
				require.Contains(t, err.Error(), "error deleting service account")
				require.Contains(t, err.Error(), "project role assignments")
			},
		},
		{
			name: "error deleting service account",
			service: &serviceAccountsService{
				authorize: alwaysAuthorize,
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					RevokeByPrincipalFn: func(context.Context, PrincipalReference) error {
						return nil
					},
				},
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					RevokeByPrincipalFn: func(context.Context, PrincipalReference) error {
						return nil
					},
				},
				serviceAccountsStore: &mockServiceAccountStore{
					DeleteFn: func(context.Context, string) error {
						return errors.New("store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "store error")
				require.Contains(
					t,
					err.Error(),
					"error deleting service account",
				)
			},
		},
		{
			name: "success",
			service: &serviceAccountsService{
				authorize: alwaysAuthorize,
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					RevokeByPrincipalFn: func(context.Context, PrincipalReference) error {
						return nil
					},
				},
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					RevokeByPrincipalFn: func(context.Context, PrincipalReference) error {
						return nil
					},
				},
				serviceAccountsStore: &mockServiceAccountStore{
					DeleteFn: func(context.Context, string) error {
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
			testCase.assertions(
				testCase.service.Delete(context.Background(), "jarvis"),
			)
		})
	}
}

type mockServiceAccountStore struct {
	CreateFn func(context.Context, ServiceAccount) error
	ListFn   func(
		context.Context,
		meta.ListOptions,
	) (ServiceAccountList, error)
	GetFn              func(context.Context, string) (ServiceAccount, error)
	GetByHashedTokenFn func(context.Context, string) (ServiceAccount, error)
	LockFn             func(context.Context, string) error
	UnlockFn           func(
		ctx context.Context,
		id string,
		newHashedToken string,
	) error
	DeleteFn func(context.Context, string) error
}

func (m *mockServiceAccountStore) Create(
	ctx context.Context,
	serviceAccount ServiceAccount,
) error {
	return m.CreateFn(ctx, serviceAccount)
}

func (m *mockServiceAccountStore) List(
	ctx context.Context,
	opts meta.ListOptions,
) (ServiceAccountList, error) {
	return m.ListFn(ctx, opts)
}

func (m *mockServiceAccountStore) Get(
	ctx context.Context,
	id string,
) (ServiceAccount, error) {
	return m.GetFn(ctx, id)
}

func (m *mockServiceAccountStore) GetByHashedToken(
	ctx context.Context,
	token string,
) (ServiceAccount, error) {
	return m.GetByHashedTokenFn(ctx, token)
}

func (m *mockServiceAccountStore) Lock(ctx context.Context, id string) error {
	return m.LockFn(ctx, id)
}

func (m *mockServiceAccountStore) Unlock(
	ctx context.Context,
	id string,
	newHashedToken string,
) error {
	return m.UnlockFn(ctx, id, newHashedToken)
}

func (m *mockServiceAccountStore) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}
