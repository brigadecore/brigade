package authn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery/testing" // nolint: lll
	metaTesting "github.com/brigadecore/brigade/sdk/v2/meta/testing"
	"github.com/stretchr/testify/require"
)

func TestThirdPartyAuthDetailsMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		ThirdPartyAuthDetails{},
		"ThirdPartyAuthDetails",
	)
}

func TestNewSessionsClient(t *testing.T) {
	client := NewSessionsClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	)
	require.IsType(t, &sessionsClient{}, client)
	rmTesting.RequireBaseClient(t, client.(*sessionsClient).BaseClient)
}

func TestSessionsClientCreateRootSession(t *testing.T) {
	const testRootPassword = "foobar"
	testSessionToken := Token{
		Value: "opensesame",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/sessions", r.URL.Path)
				require.Contains(t, r.Header.Get("Authorization"), "Basic")
				bodyBytes, err := json.Marshal(testSessionToken)
				require.NoError(t, err)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewSessionsClient(server.URL, rmTesting.TestAPIToken, nil)
	token, err := client.CreateRootSession(context.Background(), testRootPassword)
	require.NoError(t, err)
	require.Equal(t, testSessionToken, token)
}

func TestSessionsClientCreateUserSession(t *testing.T) {
	testThirdPartyAuthOpts := &ThirdPartyAuthOptions{
		SuccessURL: "https://example.com/success",
	}
	testThirdPartyAuthDetails := ThirdPartyAuthDetails{
		Token: "opensesame",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/sessions", r.URL.Path)
				require.Empty(t, r.Header.Get("Authorization"))
				require.Equal(
					t,
					testThirdPartyAuthOpts.SuccessURL,
					r.URL.Query().Get("successURL"),
				)
				bodyBytes, err := json.Marshal(testThirdPartyAuthDetails)
				require.NoError(t, err)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewSessionsClient(server.URL, rmTesting.TestAPIToken, nil)
	thirdPartyAuthDetails, err :=
		client.CreateUserSession(context.Background(), testThirdPartyAuthOpts)
	require.NoError(t, err)
	require.Equal(t, testThirdPartyAuthDetails, thirdPartyAuthDetails)
}

func TestSessionsClientDelete(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, "/v2/session", r.URL.Path)
				require.Contains(t, r.Header.Get("Authorization"), "Bearer")
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewSessionsClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Delete(context.Background())
	require.NoError(t, err)
}
