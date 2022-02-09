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
	client, ok := NewProjectRoleAssignmentsClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	).(*projectRoleAssignmentsClient)
	require.True(t, ok)
	rmTesting.RequireBaseClient(t, client.BaseClient)
}

func TestProjectRoleAssignmentsClientGrant(t *testing.T) {
	const testProjectID = "bluebook"
	testProjectRoleAssignment := ProjectRoleAssignment{
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
	const testUserID = "tony@starkindustries.com"
	const testRole = Role("ceo")
	testProjectRoleAssignments := ProjectRoleAssignmentList{
		Items: []ProjectRoleAssignment{
			{
				ProjectID: testProjectID,
				Principal: PrincipalReference{
					Type: PrincipalTypeUser,
					ID:   testUserID,
				},
				Role: testRole,
			},
		},
	}
	testCases := []struct {
		name       string
		selector   ProjectRoleAssignmentsSelector
		assertions func(*testing.T, *http.Request)
	}{
		{
			name: "with project ID specified",
			selector: ProjectRoleAssignmentsSelector{
				ProjectID: testProjectID,
				Principal: &PrincipalReference{
					Type: PrincipalTypeUser,
					ID:   testUserID,
				},
				Role: testRole,
			},
			assertions: func(t *testing.T, r *http.Request) {
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s/role-assignments", testProjectID),
					r.URL.Path,
				)
			},
		},
		{
			name: "without project ID specified",
			selector: ProjectRoleAssignmentsSelector{
				Principal: &PrincipalReference{
					Type: PrincipalTypeUser,
					ID:   testUserID,
				},
				Role: testRole,
			},
			assertions: func(t *testing.T, r *http.Request) {
				require.Equal(
					t,
					"/v2/project-role-assignments",
					r.URL.Path,
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						defer r.Body.Close()
						require.Equal(t, http.MethodGet, r.Method)
						require.Equal(
							t,
							testCase.selector.Principal.Type,
							PrincipalType(r.URL.Query().Get("principalType")),
						)
						require.Equal(
							t,
							testCase.selector.Principal.ID,
							r.URL.Query().Get("principalID"),
						)
						require.Equal(
							t,
							testCase.selector.Role,
							Role(r.URL.Query().Get("role")),
						)
						testCase.assertions(t, r)
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
				&testCase.selector,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, testProjectRoleAssignments, projectRoleAssignments)
		})
	}
}

func TestProjectRoleAssignmentsClientRevoke(t *testing.T) {
	const testProjectID = "bluebook"
	testProjectRoleAssignment := ProjectRoleAssignment{
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
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s/role-assignments", testProjectID),
					r.URL.Path,
				)
				require.Equal(
					t,
					testProjectRoleAssignment.Role,
					Role(r.URL.Query().Get("role")),
				)
				require.Equal(
					t,
					testProjectRoleAssignment.Principal.Type,
					PrincipalType(r.URL.Query().Get("principalType")),
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
