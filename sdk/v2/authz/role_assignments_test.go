package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery/testing" // nolint: lll
	libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"
	metaTesting "github.com/brigadecore/brigade/sdk/v2/meta/testing"
	"github.com/stretchr/testify/require"
)

func TestRoleAssignmentListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		RoleAssignmentList{},
		RoleAssignmentListKind,
	)
}

func TestNewRoleAssignmentsClient(t *testing.T) {
	client := NewRoleAssignmentsClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	)
	require.IsType(t, &roleAssignmentsClient{}, client)
	rmTesting.RequireBaseClient(t, client.(*roleAssignmentsClient).BaseClient)
}

func TestRoleAssignmentsClientGrant(t *testing.T) {
	testRoleAssignment := libAuthz.RoleAssignment{
		Role: libAuthz.Role("ceo"),
		Principal: libAuthz.PrincipalReference{
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
				roleAssignment := libAuthz.RoleAssignment{}
				err = json.Unmarshal(bodyBytes, &roleAssignment)
				require.NoError(t, err)
				require.Equal(t, testRoleAssignment, roleAssignment)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewRoleAssignmentsClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Grant(context.Background(), testRoleAssignment)
	require.NoError(t, err)
}

func TestRoleAssignmentsClientList(t *testing.T) {
	testRoleAssignments := RoleAssignmentList{
		Items: []libAuthz.RoleAssignment{
			{
				Principal: libAuthz.PrincipalReference{
					Type: PrincipalTypeUser,
					ID:   "tony@starkindustries.com",
				},
				Role:  libAuthz.Role("ceo"),
				Scope: "corporate",
			},
		},
	}
	testRoleAssignmentsSelector := RoleAssignmentsSelector{
		Principal: &libAuthz.PrincipalReference{
			Type: PrincipalTypeUser,
			ID:   "tony@starkindustries.com",
		},
		Role: libAuthz.Role("ceo"),
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
					libAuthz.PrincipalType(r.URL.Query().Get("principalType")),
				)
				require.Equal(
					t,
					testRoleAssignmentsSelector.Principal.ID,
					r.URL.Query().Get("principalID"),
				)
				require.Equal(
					t,
					testRoleAssignmentsSelector.Role,
					libAuthz.Role(r.URL.Query().Get("role")),
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
	testRoleAssignment := libAuthz.RoleAssignment{
		Role: libAuthz.Role("ceo"),
		Principal: libAuthz.PrincipalReference{
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
					libAuthz.Role(r.URL.Query().Get("role")),
				)
				require.Equal(
					t,
					testRoleAssignment.Principal.Type,
					libAuthz.PrincipalType(r.URL.Query().Get("principalType")),
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
	err := client.Revoke(context.Background(), testRoleAssignment)
	require.NoError(t, err)
}
