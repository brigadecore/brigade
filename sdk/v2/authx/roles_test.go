package authx

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoleAssignmentMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, RoleAssignment{}, "RoleAssignment")
}

func TestNewRoleAssignmentsClient(t *testing.T) {
	client := NewRoleAssignmentsClient(testAPIAddress, testAPIToken, nil)
	require.IsType(t, &roleAssignmentsClient{}, client)
	requireBaseClient(t, client.(*roleAssignmentsClient).BaseClient)
}

func TestRoleAssignmentsClientGrant(t *testing.T) {
	testRoleAssignment := RoleAssignment{
		Role: Role{
			Type: RoleTypeSystem,
			Name: RoleName("ceo"),
		},
		Principal: PrincipalReference{
			Type: PrincipalTypeUser,
			ID:   "tony@starkindustries.com",
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/role-assignments", r.URL.Path)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				roleAssignment := RoleAssignment{}
				err = json.Unmarshal(bodyBytes, &roleAssignment)
				require.NoError(t, err)
				require.Equal(t, testRoleAssignment, roleAssignment)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewRoleAssignmentsClient(server.URL, testAPIToken, nil)
	err := client.Grant(context.Background(), testRoleAssignment)
	require.NoError(t, err)
}

func TestRoleAssignmentsClientRevoke(t *testing.T) {
	testRoleAssignment := RoleAssignment{
		Role: Role{
			Type: RoleTypeSystem,
			Name: RoleName("ceo"),
		},
		Principal: PrincipalReference{
			Type: PrincipalTypeUser,
			ID:   "tony@starkindustries.com",
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, "/v2/role-assignments", r.URL.Path)
				require.Equal(
					t,
					testRoleAssignment.Role.Type,
					RoleType(r.URL.Query().Get("roleType")),
				)
				require.Equal(
					t,
					testRoleAssignment.Role.Name,
					RoleName(r.URL.Query().Get("roleName")),
				)
				require.Equal(
					t,
					testRoleAssignment.Principal.Type,
					PrincipalType(r.URL.Query().Get("principalType")),
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
	client := NewRoleAssignmentsClient(server.URL, testAPIToken, nil)
	err := client.Revoke(context.Background(), testRoleAssignment)
	require.NoError(t, err)
}
