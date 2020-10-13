package authx

import (
	"context"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestUserMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, User{}, "User")
}

func TestNewUsersService(t *testing.T) {
	store := &mockUsersStore{}
	svc := NewUsersService(store)
	require.Same(t, store, svc.(*usersService).store)
}

func TestUsersServiceGet(t *testing.T) {
	testUser := User{
		ObjectMeta: meta.ObjectMeta{
			ID: "tony@starkindustries.com",
		},
	}
	testCases := []struct {
		name       string
		store      UsersStore
		assertions func(user User, err error)
	}{
		{
			name: "with error from store",
			store: &mockUsersStore{
				GetFn: func(context.Context, string) (User, error) {
					return User{}, &meta.ErrNotFound{}
				},
			},
			assertions: func(user User, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, errors.Cause(err))
			},
		},
		{
			name: "success",
			store: &mockUsersStore{
				GetFn: func(context.Context, string) (User, error) {
					return testUser, nil
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
			service := &usersService{
				store: testCase.store,
			}
			user, err := service.Get(context.Background(), testUser.ID)
			testCase.assertions(user, err)
		})
	}
}

type mockUsersStore struct {
	CreateFn func(context.Context, User) error
	GetFn    func(context.Context, string) (User, error)
}

func (m *mockUsersStore) Create(ctx context.Context, user User) error {
	return m.CreateFn(ctx, user)
}

func (m *mockUsersStore) Get(ctx context.Context, id string) (User, error) {
	return m.GetFn(ctx, id)
}
