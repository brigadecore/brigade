package authn

import (
	"context"
	"testing"

	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestUserMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, User{}, "User")
}

func TestUserListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, UserList{}, "UserList")
}

func TestNewUsersService(t *testing.T) {
	usersStore := &MockUsersStore{}
	sessionsStore := &mockSessionsStore{}
	svc := NewUsersService(libAuthz.AlwaysAuthorize, usersStore, sessionsStore)
	require.NotNil(t, svc.(*usersService).authorize)
	require.Same(t, usersStore, svc.(*usersService).usersStore)
	require.Same(t, sessionsStore, svc.(*usersService).sessionsStore)
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
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting users from store",
			service: &usersService{
				authorize: libAuthz.AlwaysAuthorize,
				usersStore: &MockUsersStore{
					ListFn: func(context.Context, meta.ListOptions) (UserList, error) {
						return UserList{}, errors.New("error listing users")
					},
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
				authorize: libAuthz.AlwaysAuthorize,
				usersStore: &MockUsersStore{
					ListFn: func(context.Context, meta.ListOptions) (UserList, error) {
						return UserList{}, nil
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
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(_ User, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "with error from store",
			service: &usersService{
				authorize: libAuthz.AlwaysAuthorize,
				usersStore: &MockUsersStore{
					GetFn: func(context.Context, string) (User, error) {
						return User{}, &meta.ErrNotFound{}
					},
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
				authorize: libAuthz.AlwaysAuthorize,
				usersStore: &MockUsersStore{
					GetFn: func(context.Context, string) (User, error) {
						return testUser, nil
					},
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
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error updating user in store",
			service: &usersService{
				authorize: libAuthz.AlwaysAuthorize,
				usersStore: &MockUsersStore{
					LockFn: func(context.Context, string) error {
						return errors.New("store error")
					},
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
				authorize: libAuthz.AlwaysAuthorize,
				usersStore: &MockUsersStore{
					LockFn: func(context.Context, string) error {
						return nil
					},
				},
				sessionsStore: &mockSessionsStore{
					DeleteByUserFn: func(c context.Context, s string) error {
						return errors.New("store error")
					},
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
				authorize: libAuthz.AlwaysAuthorize,
				usersStore: &MockUsersStore{
					LockFn: func(context.Context, string) error {
						return nil
					},
				},
				sessionsStore: &mockSessionsStore{
					DeleteByUserFn: func(c context.Context, s string) error {
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
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error updating user in store",
			service: &usersService{
				authorize: libAuthz.AlwaysAuthorize,
				usersStore: &MockUsersStore{
					UnlockFn: func(context.Context, string) error {
						return errors.New("store error")
					},
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
				authorize: libAuthz.AlwaysAuthorize,
				usersStore: &MockUsersStore{
					UnlockFn: func(context.Context, string) error {
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
			err := testCase.service.Unlock(
				context.Background(),
				"tony@starkindustries.com",
			)
			testCase.assertions(err)
		})
	}
}
