package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brigadecore/brigade/sdk/v2/authx"
	"github.com/stretchr/testify/require"
)

func TestNewProjectRolesClient(t *testing.T) {
	client := NewProjectRolesClient(testAPIAddress, testAPIToken, nil)
	require.IsType(t, &projectRolesClient{}, client)
	requireBaseClient(t, client.(*projectRolesClient).BaseClient)
}

func TestProjectRolesClientGrant(t *testing.T) {
	const testProjectID = "bluebook"
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
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s/role-assignments", testProjectID),
					r.URL.Path,
				)
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
	client := NewProjectRolesClient(server.URL, testAPIToken, nil)
	err := client.Grant(
		context.Background(),
		testProjectID,
		testRoleAssignment,
	)
	require.NoError(t, err)
}

func TestProjectRolesClientRevoke(t *testing.T) {
	const testProjectID = "bluebook"
	testRoleAssignment := authx.RoleAssignment{
		Role:          authx.RoleName("ceo"),
		PrincipalType: authx.PrincipalTypeUser,
		PrincipalID:   "tony@starkindustries.com",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s/role-assignments", testProjectID),
					r.URL.Path,
				)
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
	client := NewProjectRolesClient(server.URL, testAPIToken, nil)
	err := client.Revoke(context.Background(), testProjectID, testRoleAssignment)
	require.NoError(t, err)
}
