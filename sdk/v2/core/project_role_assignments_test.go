package core

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brigadecore/brigade/sdk/v2/authn"
	rmTesting "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery/testing" // nolint: lll
	libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"
	"github.com/stretchr/testify/require"
)

func TestNewProjectRoleAssignmentsClient(t *testing.T) {
	client := NewProjectRoleAssignmentsClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	)
	require.IsType(t, &projectRoleAssignmentsClient{}, client)
	rmTesting.RequireBaseClient(
		t,
		client.(*projectRoleAssignmentsClient).BaseClient,
	)
}

func TestProjectRoleAssignmentsClientGrant(t *testing.T) {
	testProjectRoleAssignment := ProjectRoleAssignment{
		Role: ProjectRole{
			Name:      libAuthz.RoleName("ceo"),
			ProjectID: "bluebook",
		},
		Principal: authn.PrincipalReference{
			Type: authn.PrincipalTypeUser,
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
				projectRoleAssignment := ProjectRoleAssignment{}
				err = json.Unmarshal(bodyBytes, &projectRoleAssignment)
				require.NoError(t, err)
				require.Equal(t, testProjectRoleAssignment, projectRoleAssignment)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewProjectRoleAssignmentsClient(
		server.URL,
		rmTesting.TestAPIToken,
		nil,
	)
	err := client.Grant(context.Background(), testProjectRoleAssignment)
	require.NoError(t, err)
}

func TestProjectRoleAssignmentsClientRevoke(t *testing.T) {
	testProjectRoleAssignment := ProjectRoleAssignment{
		Role: ProjectRole{
			Name:      libAuthz.RoleName("ceo"),
			ProjectID: "bluebook",
		},
		Principal: authn.PrincipalReference{
			Type: authn.PrincipalTypeUser,
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
					testProjectRoleAssignment.Role.Name,
					libAuthz.RoleName(r.URL.Query().Get("roleName")),
				)
				require.Equal(
					t,
					testProjectRoleAssignment.Role.ProjectID,
					r.URL.Query().Get("projectID"),
				)
				require.Equal(
					t,
					testProjectRoleAssignment.Principal.Type,
					authn.PrincipalType(r.URL.Query().Get("principalType")),
				)
				require.Equal(
					t,
					testProjectRoleAssignment.Principal.ID,
					r.URL.Query().Get("principalID"),
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewProjectRoleAssignmentsClient(
		server.URL,
		rmTesting.TestAPIToken,
		nil,
	)
	err := client.Revoke(context.Background(), testProjectRoleAssignment)
	require.NoError(t, err)
}
