package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewAuthnClient(t *testing.T) {
	client :=
		NewAuthnClient(rmTesting.TestAPIAddress, rmTesting.TestAPIToken, nil)
	require.IsType(t, &authnClient{}, client)
	require.NotNil(t, client.(*authnClient).serviceAccountsClient)
	require.Equal(
		t,
		client.(*authnClient).serviceAccountsClient,
		client.ServiceAccounts(),
	)
	require.NotNil(t, client.(*authnClient).sessionsClient)
	require.Equal(t, client.(*authnClient).sessionsClient, client.Sessions())
	require.NotNil(t, client.(*authnClient).usersClient)
	require.Equal(t, client.(*authnClient).usersClient, client.Users())
}

func TestAuthnClientWhoAmI(t *testing.T) {
	testRef := PrincipalReference{
		Type: PrincipalTypeServiceAccount,
		ID:   "friday",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					"/v2/whoami",
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
				bodyBytes, err := json.Marshal(testRef)
				require.NoError(t, err)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewAuthnClient(server.URL, rmTesting.TestAPIToken, nil)
	ref, err := client.WhoAmI(context.Background())
	require.NoError(t, err)
	require.Equal(t, testRef, ref)
}
