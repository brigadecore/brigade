package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brigadecore/brigade/sdk/v3/authz"
	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	libAuthz "github.com/brigadecore/brigade/sdk/v3/lib/authz"
	metaTesting "github.com/brigadecore/brigade/sdk/v3/meta/testing"
	"github.com/stretchr/testify/require"
)

func TestProjectRoleAssignmentMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		ProjectRoleAssignment{},
		ProjectRoleAssignmentKind,
	)
}

func TestProjectRoleAssignmentListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		ProjectRoleAssignmentList{},
		ProjectRoleAssignmentListKind,
	)
}

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
	const testProjectID = "bluebook"
	testProjectRoleAssignment := ProjectRoleAssignment{
		Role: libAuthz.Role("ceo"),
		Principal: libAuthz.PrincipalReference{
			Type: authz.PrincipalTypeUser,
			ID:   "tony@starkindustries.com",
		},
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
	err :=
		client.Grant(
			context.Background(),
			testProjectID,
			testProjectRoleAssignment,
			nil,
		)
	require.NoError(t, err)
}

func TestProjectRoleAssignmentsClientList(t *testing.T) {
	const testProjectID = "bluebook"
	testProjectRoleAssignments := ProjectRoleAssignmentList{
		Items: []ProjectRoleAssignment{
			{
				Principal: libAuthz.PrincipalReference{
					Type: authz.PrincipalTypeUser,
					ID:   "tony@starkindustries.com",
				},
				Role: libAuthz.Role("ceo"),
			},
		},
	}
	testProjectRoleAssignmentsSelector := ProjectRoleAssignmentsSelector{
		Principal: &libAuthz.PrincipalReference{
			Type: authz.PrincipalTypeUser,
			ID:   "tony@starkindustries.com",
		},
		Role: libAuthz.Role("ceo"),
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s/role-assignments", testProjectID),
					r.URL.Path,
				)
				require.Equal(
					t,
					testProjectRoleAssignmentsSelector.Principal.Type,
					libAuthz.PrincipalType(r.URL.Query().Get("principalType")),
				)
				require.Equal(
					t,
					testProjectRoleAssignmentsSelector.Principal.ID,
					r.URL.Query().Get("principalID"),
				)
				require.Equal(
					t,
					testProjectRoleAssignmentsSelector.Role,
					libAuthz.Role(r.URL.Query().Get("role")),
				)
				bodyBytes, err := json.Marshal(testProjectRoleAssignments)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewProjectRoleAssignmentsClient(
		server.URL,
		rmTesting.TestAPIToken,
		nil,
	)
	projectRoleAssignments, err := client.List(
		context.Background(),
		testProjectID,
		&testProjectRoleAssignmentsSelector,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, testProjectRoleAssignments, projectRoleAssignments)
}

func TestProjectRoleAssignmentsClientRevoke(t *testing.T) {
	const testProjectID = "bluebook"
	testProjectRoleAssignment := ProjectRoleAssignment{
		Role: libAuthz.Role("ceo"),
		Principal: libAuthz.PrincipalReference{
			Type: authz.PrincipalTypeUser,
			ID:   "tony@starkindustries.com",
		},
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
					testProjectRoleAssignment.Role,
					libAuthz.Role(r.URL.Query().Get("role")),
				)
				require.Equal(
					t,
					testProjectRoleAssignment.Principal.Type,
					libAuthz.PrincipalType(r.URL.Query().Get("principalType")),
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
	err := client.Revoke(
		context.Background(),
		testProjectID,
		testProjectRoleAssignment,
		nil,
	)
	require.NoError(t, err)
}
