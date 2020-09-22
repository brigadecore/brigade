package authx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/stretchr/testify/require"
)

func TestUserMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, User{}, "User")
}

func TestUserListMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, UserList{}, "UserList")
}

func TestNewUsersClient(t *testing.T) {
	client := NewUsersClient(testAPIAddress, testAPIToken, nil)
	require.IsType(t, &usersClient{}, client)
	requireBaseClient(t, client.(*usersClient).BaseClient)
}

func TestUsersClientList(t *testing.T) {
	testUsers := UserList{
		Items: []User{
			{
				ObjectMeta: meta.ObjectMeta{
					ID: "tony@starkindustries.com",
				},
			},
			{
				ObjectMeta: meta.ObjectMeta{
					ID: "pepper@starkindustries.com",
				},
			},
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/v2/users", r.URL.Path)
				bodyBytes, err := json.Marshal(testUsers)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewUsersClient(server.URL, testAPIToken, nil)
	users, err := client.List(context.Background(), nil, nil)
	require.NoError(t, err)
	require.Equal(t, testUsers, users)
}

func TestUsersClientGet(t *testing.T) {
	testUser := User{
		ObjectMeta: meta.ObjectMeta{
			ID: "tony@starkindustries.com",
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/users/%s", testUser.ID),
					r.URL.Path,
				)
				bodyBytes, err := json.Marshal(testUser)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewUsersClient(server.URL, testAPIToken, nil)
	user, err := client.Get(context.Background(), testUser.ID)
	require.NoError(t, err)
	require.Equal(t, testUser, user)
}

func TestUsersClientLock(t *testing.T) {
	const testUserID = "tony@starkindustries.com"
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/users/%s/lock", testUserID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewUsersClient(server.URL, testAPIToken, nil)
	err := client.Lock(context.Background(), testUserID)
	require.NoError(t, err)
}

func TestUsersClientUnlock(t *testing.T) {
	const testUserID = "tony@starkindustries.com"
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/users/%s/lock", testUserID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewUsersClient(server.URL, testAPIToken, nil)
	err := client.Unlock(context.Background(), testUserID)
	require.NoError(t, err)
}
