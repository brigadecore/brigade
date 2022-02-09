package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	metaTesting "github.com/brigadecore/brigade/sdk/v3/meta/testing"
	"github.com/stretchr/testify/require"
)

func TestRoleAssignmentMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, RoleAssignment{}, RoleAssignmentKind)
}

func TestRoleAssignmentListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		RoleAssignmentList{},
		RoleAssignmentListKind,
	)
}

func TestNewRoleAssignmentsClient(t *testing.T) {
	client, ok := NewRoleAssignmentsClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	).(*roleAssignmentsClient)
	require.True(t, ok)
	rmTesting.RequireBaseClient(t, client.BaseClient)
}

func TestRoleAssignmentsClientGrant(t *testing.T) {
	testRoleAssignment := RoleAssignment{
		Role: Role("ceo"),
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
	client := NewRoleAssignmentsClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Grant(context.Background(), testRoleAssignment, nil)
	require.NoError(t, err)
}

func TestRoleAssignmentsClientList(t *testing.T) {
	testRoleAssignments := RoleAssignmentList{
		Items: []RoleAssignment{
			{
				Principal: PrincipalReference{
					Type: PrincipalTypeUser,
					ID:   "tony@starkindustries.com",
				},
				Role:  Role("ceo"),
				Scope: "corporate",
			},
		},
	}
	testRoleAssignmentsSelector := RoleAssignmentsSelector{
		Principal: &PrincipalReference{
			Type: PrincipalTypeUser,
			ID:   "tony@starkindustries.com",
		},
		Role: Role("ceo"),
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/v2/role-assignments", r.URL.Path)
				require.Equal(
					t,
					testRoleAssignmentsSelector.Principal.Type,
					PrincipalType(r.URL.Query().Get("principalType")),
				)
				require.Equal(
					t,
					testRoleAssignmentsSelector.Principal.ID,
					r.URL.Query().Get("principalID"),
				)
				require.Equal(
					t,
					testRoleAssignmentsSelector.Role,
					Role(r.URL.Query().Get("role")),
				)
				bodyBytes, err := json.Marshal(testRoleAssignments)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewRoleAssignmentsClient(server.URL, rmTesting.TestAPIToken, nil)
	roleAssignments, err := client.List(
		context.Background(),
		&testRoleAssignmentsSelector,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, testRoleAssignments, roleAssignments)
}

func TestRoleAssignmentsClientRevoke(t *testing.T) {
	testRoleAssignment := RoleAssignment{
		Role: Role("ceo"),
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
					testRoleAssignment.Role,
					Role(r.URL.Query().Get("role")),
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
	client := NewRoleAssignmentsClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Revoke(context.Background(), testRoleAssignment, nil)
	require.NoError(t, err)
}
