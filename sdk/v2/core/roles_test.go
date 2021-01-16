package core

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

func TestNewProjectRolesClient(t *testing.T) {
	client := NewProjectRoleAssignmentsClient(testAPIAddress, testAPIToken, nil)
	require.IsType(t, &projectRoleAssignmentsClient{}, client)
	requireBaseClient(t, client.(*projectRoleAssignmentsClient).BaseClient)
}

func TestProjectRoleAssignmentsClientGrant(t *testing.T) {
	testRoleAssignment := authx.RoleAssignment{
		Role: authx.Role{
			Type:  RoleTypeProject,
			Name:  authx.RoleName("ceo"),
			Scope: "bluebook",
		},
		Principal: authx.PrincipalReference{
			Type: authx.PrincipalTypeUser,
			ID:   "tony@starkindustries.com",
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/project-role-assignments", r.URL.Path)
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
	client := NewProjectRoleAssignmentsClient(server.URL, testAPIToken, nil)
	err := client.Grant(context.Background(), testRoleAssignment)
	require.NoError(t, err)
}

func TestProjectRoleAssignmentsClientRevoke(t *testing.T) {
	testRoleAssignment := authx.RoleAssignment{
		Role: authx.Role{
			Type:  RoleTypeProject,
			Name:  authx.RoleName("ceo"),
			Scope: "bluebook",
		},
		Principal: authx.PrincipalReference{
			Type: authx.PrincipalTypeUser,
			ID:   "tony@starkindustries.com",
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, "/v2/project-role-assignments", r.URL.Path)
				require.Equal(
					t,
					testRoleAssignment.Role.Name,
					authx.RoleName(r.URL.Query().Get("roleName")),
				)
				require.Equal(
					t,
					testRoleAssignment.Role.Scope,
					r.URL.Query().Get("projectID"),
				)
				require.Equal(
					t,
					testRoleAssignment.Principal.Type,
					authx.PrincipalType(r.URL.Query().Get("principalType")),
				)
				require.Equal(
					t,
					testRoleAssignment.Principal.ID,
					r.URL.Query().Get("principalID"),
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewProjectRoleAssignmentsClient(server.URL, testAPIToken, nil)
	err := client.Revoke(context.Background(), testRoleAssignment)
	require.NoError(t, err)
}
