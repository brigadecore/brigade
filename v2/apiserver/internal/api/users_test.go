package api

import (
	"context"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestUserMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, User{}, "User")
}

func TestNewUsersService(t *testing.T) {
	usersStore := &mockUsersStore{}
	sessionsStore := &mockSessionsStore{}
	roleAssignmentsStore := &mockRoleAssignmentsStore{}
	projectRoleAssignmentsStore := &mockProjectRoleAssignmentsStore{}
	svc, ok := NewUsersService(
		alwaysAuthorize,
		usersStore,
		sessionsStore,
		roleAssignmentsStore,
		projectRoleAssignmentsStore,
		UsersServiceConfig{},
	).(*usersService)
	require.True(t, ok)
	require.NotNil(t, svc.authorize)
	require.Same(t, usersStore, svc.usersStore)
	require.Same(t, sessionsStore, svc.sessionsStore)
	require.Same(t, roleAssignmentsStore, svc.roleAssignmentsStore)
	require.Same(t, projectRoleAssignmentsStore, svc.projectRoleAssignmentsStore)
}

func TestUserServiceList(t *testing.T) {
	testCases := []struct {
		name       string
		service    UsersService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &usersService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "user management functions unavailable",
			service: &usersService{
				authorize: alwaysAuthorize,
				config: UsersServiceConfig{
					ThirdPartyAuthEnabled: false,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotSupported{}, err)
			},
		},
		{
			name: "error getting users from store",
			service: &usersService{
				authorize: alwaysAuthorize,
				usersStore: &mockUsersStore{
					ListFn: func(context.Context, meta.ListOptions) (UserList, error) {
						return UserList{}, errors.New("error listing users")
					},
				},
				config: UsersServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error listing users")
				require.Contains(t, err.Error(), "error retrieving users from store")
			},
		},
		{
			name: "success",
			service: &usersService{
				authorize: alwaysAuthorize,
				usersStore: &mockUsersStore{
					ListFn: func(context.Context, meta.ListOptions) (UserList, error) {
						return UserList{}, nil
					},
				},
				config: UsersServiceConfig{
					ThirdPartyAuthEnabled: true,
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

func TestUsersServiceGet(t *testing.T) {
	testUser := User{
		ObjectMeta: meta.ObjectMeta{
			ID: "tony@starkindustries.com",
		},
	}
	testCases := []struct {
		name       string
		service    UsersService
		assertions func(user User, err error)
	}{
		{
			name: "unauthorized",
			service: &usersService{
				authorize: neverAuthorize,
			},
			assertions: func(_ User, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "user management functions unavailable",
			service: &usersService{
				authorize: alwaysAuthorize,
				config: UsersServiceConfig{
					ThirdPartyAuthEnabled: false,
				},
			},
			assertions: func(_ User, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotSupported{}, err)
			},
		},
		{
			name: "with error from store",
			service: &usersService{
				authorize: alwaysAuthorize,
				usersStore: &mockUsersStore{
					GetFn: func(context.Context, string) (User, error) {
						return User{}, &meta.ErrNotFound{}
					},
				},
				config: UsersServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(user User, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, errors.Cause(err))
			},
		},
		{
			name: "success",
			service: &usersService{
				authorize: alwaysAuthorize,
				usersStore: &mockUsersStore{
					GetFn: func(context.Context, string) (User, error) {
						return testUser, nil
					},
				},
				config: UsersServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(user User, err error) {
				require.NoError(t, err)
				require.Equal(t, testUser, user)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			user, err := testCase.service.Get(context.Background(), testUser.ID)
			testCase.assertions(user, err)
		})
	}
}

func TestUsersServiceLock(t *testing.T) {
	testCases := []struct {
		name       string
		service    UsersService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &usersService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "user management functions unavailable",
			service: &usersService{
				authorize: alwaysAuthorize,
				config: UsersServiceConfig{
					ThirdPartyAuthEnabled: false,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotSupported{}, err)
			},
		},
		{
			name: "error updating user in store",
			service: &usersService{
				authorize: alwaysAuthorize,
				usersStore: &mockUsersStore{
					LockFn: func(context.Context, string) error {
						return errors.New("store error")
					},
				},
				config: UsersServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "store error")
				require.Contains(t, err.Error(), "error locking user")
			},
		},
		{
			name: "error deleting user sessions from store",
			service: &usersService{
				authorize: alwaysAuthorize,
				usersStore: &mockUsersStore{
					LockFn: func(context.Context, string) error {
						return nil
					},
				},
				sessionsStore: &mockSessionsStore{
					DeleteByUserFn: func(c context.Context, s string) error {
						return errors.New("store error")
					},
				},
				config: UsersServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "store error")
				require.Contains(t, err.Error(), "error deleting user")
				require.Contains(t, err.Error(), "sessions from store")
			},
		},
		{
			name: "success",
			service: &usersService{
				authorize: alwaysAuthorize,
				usersStore: &mockUsersStore{
					LockFn: func(context.Context, string) error {
						return nil
					},
				},
				sessionsStore: &mockSessionsStore{
					DeleteByUserFn: func(c context.Context, s string) error {
						return nil
					},
				},
				config: UsersServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err :=
				testCase.service.Lock(context.Background(), "tony@starkindustries.com")
			testCase.assertions(err)
		})
	}
}

func TestUsersServiceUnlock(t *testing.T) {
	testCases := []struct {
		name       string
		service    UsersService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &usersService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "user management functions unavailable",
			service: &usersService{
				authorize: alwaysAuthorize,
				config: UsersServiceConfig{
					ThirdPartyAuthEnabled: false,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotSupported{}, err)
			},
		},
		{
			name: "error updating user in store",
			service: &usersService{
				authorize: alwaysAuthorize,
				usersStore: &mockUsersStore{
					UnlockFn: func(context.Context, string) error {
						return errors.New("store error")
					},
				},
				config: UsersServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "store error")
				require.Contains(t, err.Error(), "error unlocking user")
			},
		},
		{
			name: "success",
			service: &usersService{
				authorize: alwaysAuthorize,
				usersStore: &mockUsersStore{
					UnlockFn: func(context.Context, string) error {
						return nil
					},
				},
				config: UsersServiceConfig{
					ThirdPartyAuthEnabled: true,
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.service.Unlock(
				context.Background(),
				"tony@starkindustries.com",
			)
			testCase.assertions(err)
		})
	}
}

func TestUsersServiceDelete(t *testing.T) {
	testCases := []struct {
		name       string
		service    UsersService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &usersService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error deleting role assignments",
			service: &usersService{
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
				require.Contains(t, err.Error(), "error deleting user")
				require.Contains(t, err.Error(), "role assignments")
			},
		},
		{
			name: "error deleting project role assignments",
			service: &usersService{
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
				require.Contains(t, err.Error(), "error deleting user")
				require.Contains(t, err.Error(), "project role assignments")
			},
		},
		{
			name: "error deleting user",
			service: &usersService{
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
				usersStore: &mockUsersStore{
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
					"error deleting user",
				)
			},
		},
		{
			name: "error deleting user sessions",
			service: &usersService{
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
				usersStore: &mockUsersStore{
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				sessionsStore: &mockSessionsStore{
					DeleteByUserFn: func(context.Context, string) error {
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
					"error deleting user",
				)
			},
		},
		{
			name: "success",
			service: &usersService{
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
				usersStore: &mockUsersStore{
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				sessionsStore: &mockSessionsStore{
					DeleteByUserFn: func(context.Context, string) error {
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
				testCase.service.Delete(
					context.Background(),
					"tony@starkindustries.com",
				),
			)
		})
	}
}

type mockUsersStore struct {
	CreateFn func(context.Context, User) error
	ListFn   func(context.Context, meta.ListOptions) (UserList, error)
	GetFn    func(context.Context, string) (User, error)
	LockFn   func(context.Context, string) error
	UnlockFn func(context.Context, string) error
	DeleteFn func(context.Context, string) error
}

func (m *mockUsersStore) Create(ctx context.Context, user User) error {
	return m.CreateFn(ctx, user)
}

func (m *mockUsersStore) List(
	ctx context.Context,
	opts meta.ListOptions,
) (UserList, error) {
	return m.ListFn(ctx, opts)
}

func (m *mockUsersStore) Get(ctx context.Context, id string) (User, error) {
	return m.GetFn(ctx, id)
}

func (m *mockUsersStore) Lock(ctx context.Context, id string) error {
	return m.LockFn(ctx, id)
}

func (m *mockUsersStore) Unlock(ctx context.Context, id string) error {
	return m.UnlockFn(ctx, id)
}

func (m *mockUsersStore) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}
