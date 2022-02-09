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

func TestSecretMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, Secret{}, "Secret")
}

func TestSecretListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, SecretList{}, "SecretList")
}

func TestNewSecretsClient(t *testing.T) {
	client, ok := NewSecretsClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	).(*secretsClient)
	require.True(t, ok)
	rmTesting.RequireBaseClient(t, client.BaseClient)
}

func TestSecretsClientList(t *testing.T) {
	const testProjectID = "bluebook"
	testSecrets := SecretList{
		Items: []Secret{
			{
				Key:   "soylentgreen",
				Value: "people",
			},
			{
				Key:   "whodunit",
				Value: "thebutler",
			},
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s/secrets", testProjectID),
					r.URL.Path,
				)
				bodyBytes, err := json.Marshal(testSecrets)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewSecretsClient(server.URL, rmTesting.TestAPIToken, nil)
	secrets, err := client.List(context.Background(), testProjectID, nil)
	require.NoError(t, err)
	require.Equal(t, testSecrets, secrets)
}

func TestSecretsClientSet(t *testing.T) {
	const testProjectID = "bluebook"
	testSecret := Secret{
		Key:   "soylentgreen",
		Value: "people",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf(
						"/v2/projects/%s/secrets/%s",
						testProjectID,
						testSecret.Key,
					),
					r.URL.Path,
				)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				secret := Secret{}
				err = json.Unmarshal(bodyBytes, &secret)
				require.NoError(t, err)
				require.Equal(t, testSecret, secret)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewSecretsClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Set(context.Background(), testProjectID, testSecret, nil)
	require.NoError(t, err)
}

func TestSecretsClientUnset(t *testing.T) {
	const testProjectID = "bluebook"
	const testSecretKey = "soylentgreen"
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(
					t,
					fmt.Sprintf(
						"/v2/projects/%s/secrets/%s",
						testProjectID,
						testSecretKey,
					),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewSecretsClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Unset(context.Background(), testProjectID, testSecretKey, nil)
	require.NoError(t, err)
}
