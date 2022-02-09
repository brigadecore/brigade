package sdk

// nolint: lll
import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	metaTesting "github.com/brigadecore/brigade/sdk/v3/meta/testing"
	"github.com/stretchr/testify/require"
)

func TestServiceAccountMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, ServiceAccount{}, "ServiceAccount")
}

func TestServiceAccountListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		ServiceAccountList{},
		"ServiceAccountList",
	)
}

func TestNewServiceAccountsClient(t *testing.T) {
	client, ok := NewServiceAccountsClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	).(*serviceAccountsClient)
	require.True(t, ok)
	rmTesting.RequireBaseClient(t, client.BaseClient)
}

func TestServiceAccountsClientCreate(t *testing.T) {
	testServiceAccount := ServiceAccount{
		ObjectMeta: meta.ObjectMeta{
			ID: "jarvis",
		},
	}
	testServiceAccountToken := Token{
		Value: "opensesame",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/service-accounts", r.URL.Path)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				serviceAccount := ServiceAccount{}
				err = json.Unmarshal(bodyBytes, &serviceAccount)
				require.NoError(t, err)
				require.Equal(t, testServiceAccount, serviceAccount)
				bodyBytes, err = json.Marshal(testServiceAccountToken)
				require.NoError(t, err)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewServiceAccountsClient(server.URL, rmTesting.TestAPIToken, nil)
	token, err := client.Create(
		context.Background(),
		testServiceAccount,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, testServiceAccountToken, token)
}

func TestServiceAccountsClientList(t *testing.T) {
	testServiceAccounts := ServiceAccountList{
		Items: []ServiceAccount{
			{
				ObjectMeta: meta.ObjectMeta{
					ID: "jarvis",
				},
			},
			{
				ObjectMeta: meta.ObjectMeta{
					ID: "friday",
				},
			},
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/v2/service-accounts", r.URL.Path)
				bodyBytes, err := json.Marshal(testServiceAccounts)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewServiceAccountsClient(server.URL, rmTesting.TestAPIToken, nil)
	serviceAccounts, err := client.List(context.Background(), nil, nil)
	require.NoError(t, err)
	require.Equal(t, testServiceAccounts, serviceAccounts)
}

func TestServiceAccountsClientGet(t *testing.T) {
	testServiceAccount := ServiceAccount{
		ObjectMeta: meta.ObjectMeta{
			ID: "jarvis",
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/service-accounts/%s", testServiceAccount.ID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
				bodyBytes, err := json.Marshal(testServiceAccount)
				require.NoError(t, err)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewServiceAccountsClient(server.URL, rmTesting.TestAPIToken, nil)
	serviceAccount, err :=
		client.Get(context.Background(), testServiceAccount.ID, nil)
	require.NoError(t, err)
	require.Equal(t, testServiceAccount, serviceAccount)
}

func TestServiceAccountsClientLock(t *testing.T) {
	const testServiceAccountID = "jarvis"
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/service-accounts/%s/lock", testServiceAccountID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewServiceAccountsClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Lock(context.Background(), testServiceAccountID, nil)
	require.NoError(t, err)
}

func TestServiceAccountsClientUnlock(t *testing.T) {
	const testServiceAccountID = "jarvis"
	testServiceAccountToken := Token{
		Value: "opensesame",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/service-accounts/%s/lock", testServiceAccountID),
					r.URL.Path,
				)
				bodyBytes, err := json.Marshal(testServiceAccountToken)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewServiceAccountsClient(server.URL, rmTesting.TestAPIToken, nil)
	token, err := client.Unlock(context.Background(), testServiceAccountID, nil)
	require.NoError(t, err)
	require.Equal(t, testServiceAccountToken, token)
}

func TestServiceAccountsClientDelete(t *testing.T) {
	const testServiceAccountID = "jarvis"
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/service-accounts/%s", testServiceAccountID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewServiceAccountsClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Delete(context.Background(), testServiceAccountID, nil)
	require.NoError(t, err)
}
