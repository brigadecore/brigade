package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery/testing" // nolint: lll
	"github.com/brigadecore/brigade/sdk/v2/meta"
	metaTesting "github.com/brigadecore/brigade/sdk/v2/meta/testing"
	"github.com/stretchr/testify/require"
)

func TestProjectMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, Project{}, "Project")
}

func TestProjectListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, ProjectList{}, "ProjectList")
}

func TestNewProjectsClient(t *testing.T) {
	client := NewProjectsClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	)
	require.IsType(t, &projectsClient{}, client)
	rmTesting.RequireBaseClient(t, client.(*projectsClient).BaseClient)
	require.NotNil(t, client.(*projectsClient).authzClient)
	require.Equal(
		t,
		client.(*projectsClient).authzClient,
		client.Authz(),
	)
	require.NotNil(t, client.(*projectsClient).secretsClient)
	require.Equal(t, client.(*projectsClient).secretsClient, client.Secrets())
}

func TestProjectsClientCreate(t *testing.T) {
	testProject := Project{
		ObjectMeta: meta.ObjectMeta{
			ID: "bluebook",
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/projects", r.URL.Path)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				project := Project{}
				err = json.Unmarshal(bodyBytes, &project)
				require.NoError(t, err)
				require.Equal(t, testProject, project)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(server.URL, rmTesting.TestAPIToken, nil)
	project, err := client.Create(
		context.Background(),
		testProject,
	)
	require.NoError(t, err)
	require.Equal(t, testProject, project)
}

func TestProjectsClientCreateFromBytes(t *testing.T) {
	testProject := Project{
		ObjectMeta: meta.ObjectMeta{
			ID: "bluebook",
		},
	}
	testProjectBytes, err := json.Marshal(testProject)
	require.NoError(t, err)
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/projects", r.URL.Path)
				var bodyBytes []byte
				bodyBytes, err = ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, testProjectBytes, bodyBytes)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(server.URL, rmTesting.TestAPIToken, nil)
	project, err := client.CreateFromBytes(context.Background(), testProjectBytes)
	require.NoError(t, err)
	require.Equal(t, testProject, project)
}

func TestProjectsClientList(t *testing.T) {
	testProjects := ProjectList{
		Items: []Project{
			{
				ObjectMeta: meta.ObjectMeta{
					ID: "bluebook",
				},
			},
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/v2/projects", r.URL.Path)
				bodyBytes, err := json.Marshal(testProjects)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(server.URL, rmTesting.TestAPIToken, nil)
	projects, err := client.List(context.Background(), nil, nil)
	require.NoError(t, err)
	require.Equal(t, testProjects, projects)
}

func TestProjectsClientGet(t *testing.T) {
	testProject := Project{
		ObjectMeta: meta.ObjectMeta{
			ID: "bluebook",
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s", testProject.ID),
					r.URL.Path,
				)
				bodyBytes, err := json.Marshal(testProject)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(server.URL, rmTesting.TestAPIToken, nil)
	project, err := client.Get(context.Background(), testProject.ID)
	require.NoError(t, err)
	require.Equal(t, testProject, project)
}

func TestProjectsClientUpdate(t *testing.T) {
	testProject := Project{
		ObjectMeta: meta.ObjectMeta{
			ID: "bluebook",
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s", testProject.ID),
					r.URL.Path,
				)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				project := Project{}
				err = json.Unmarshal(bodyBytes, &project)
				require.NoError(t, err)
				require.Equal(t, testProject, project)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(server.URL, rmTesting.TestAPIToken, nil)
	project, err := client.Update(context.Background(), testProject)
	require.NoError(t, err)
	require.Equal(t, testProject, project)
}

func TestProjectsClientUpdateFromBytes(t *testing.T) {
	testProject := Project{
		ObjectMeta: meta.ObjectMeta{
			ID: "bluebook",
		},
	}
	testProjectBytes, err := json.Marshal(testProject)
	require.NoError(t, err)
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s", testProject.ID),
					r.URL.Path,
				)
				var bodyBytes []byte
				bodyBytes, err = ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, testProjectBytes, bodyBytes)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(server.URL, rmTesting.TestAPIToken, nil)
	project, err := client.UpdateFromBytes(
		context.Background(),
		testProject.ID,
		testProjectBytes,
	)
	require.NoError(t, err)
	require.Equal(t, testProject, project)
}

func TestProjectsClientDelete(t *testing.T) {
	const testProjectID = "bluebook"
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s", testProjectID),
					r.URL.Path,
				)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Delete(context.Background(), testProjectID)
	require.NoError(t, err)
}
