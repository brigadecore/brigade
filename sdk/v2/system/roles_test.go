package system

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brigadecore/brigade/sdk/v2/authx"
	"github.com/stretchr/testify/require"
)

func TestNewRolesClient(t *testing.T) {
	client := NewRolesClient(testAPIAddress, testAPIToken, nil)
	require.IsType(t, &rolesClient{}, client)
	requireBaseClient(t, client.(*rolesClient).BaseClient)
}

func TestRolesClientGrant(t *testing.T) {
	testRoleAssignment := authx.RoleAssignment{
		Role:          authx.RoleName("ceo"),
		PrincipalType: authx.PrincipalTypeUser,
		PrincipalID:   "tony@starkindustries.com",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/system/role-assignments", r.URL.Path)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				roleAssignment := authx.RoleAssignment{}
				err = json.Unmarshal(bodyBytes, &roleAssignment)
				require.NoError(t, err)
				require.Equal(t, testRoleAssignment, roleAssignment)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewRolesClient(server.URL, testAPIToken, nil)
	err := client.Grant(context.Background(), testRoleAssignment)
	require.NoError(t, err)
}

func TestRolesClientRevoke(t *testing.T) {
	testRoleAssignment := authx.RoleAssignment{
		Role:          authx.RoleName("ceo"),
		PrincipalType: authx.PrincipalTypeUser,
		PrincipalID:   "tony@starkindustries.com",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, "/v2/system/role-assignments", r.URL.Path)
				require.Equal(
					t,
					testRoleAssignment.Role,
					authx.RoleName(r.URL.Query().Get("role")),
				)
				require.Equal(
					t,
					testRoleAssignment.PrincipalType,
					authx.PrincipalType(r.URL.Query().Get("principalType")),
				)
				require.Equal(
					t,
					testRoleAssignment.PrincipalID,
					r.URL.Query().Get("principalID"),
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewRolesClient(server.URL, testAPIToken, nil)
	err := client.Revoke(context.Background(), testRoleAssignment)
	require.NoError(t, err)
}
