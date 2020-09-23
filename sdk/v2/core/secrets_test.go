package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecretMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, Secret{}, "Secret")
}

func TestSecretListMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, SecretList{}, "SecretList")
}

func TestNewSecretsClient(t *testing.T) {
	client := NewSecretsClient(testAPIAddress, testAPIToken, nil)
	require.IsType(t, &secretsClient{}, client)
	requireBaseClient(t, client.(*secretsClient).BaseClient)
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
	client := NewSecretsClient(server.URL, testAPIToken, nil)
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
	client := NewSecretsClient(server.URL, testAPIToken, nil)
	err := client.Set(context.Background(), testProjectID, testSecret)
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
	client := NewSecretsClient(server.URL, testAPIToken, nil)
	err := client.Unset(context.Background(), testProjectID, testSecretKey)
	require.NoError(t, err)
}
