package authn

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// OIDCAuthDetails encapsulates all information required for a client
// authenticating by means of OpenID Connect to complete the authentication
// process using a third-party OIDC identity provider.
type OIDCAuthDetails struct {
	// AuthURL is a URL that can be requested in a user's web browser to complete
	// authentication via a third-party OIDC identity provider.
	AuthURL string `json:"authURL,omitempty"`
	// Token is an opaque bearer token issued by Brigade to correlate a User with
	// a Session. It remains unactivated (useless) until the OIDC authentication
	// workflow is successfully completed. Clients may expect that that the token
	// expires (at an interval determined by a system administrator) and, for
	// simplicity, is NOT refreshable. When the token has expired,
	// re-authentication is required.
	Token string `json:"token,omitempty"`
}

// MarshalJSON amends OIDCAuthDetails instances with type metadata so
// that clients do not need to be concerned with the tedium of doing so.
func (o OIDCAuthDetails) MarshalJSON() ([]byte, error) {
	type Alias OIDCAuthDetails
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "OIDCAuthDetails",
			},
			Alias: (Alias)(o),
		},
	)
}

// SessionsClient is the specialized client for managing Brigade API Sessions.
type SessionsClient interface {
	// CreateRootSession creates a Session for the root user (if enabled by the
	// system administrator) and returns a Token with a short expiry period
	// (determined by a system administrator). In contrast to most other
	// operations exposed by the Brigade API, a valid token is not required to
	// invoke this.
	CreateRootSession(ctx context.Context, password string) (Token, error)
	// CreateUserSession creates a new User Session and initiates an OpenID
	// Connect authentication workflow. It returns an OIDCAuthDetails containing
	// all information required to continue the authentication process with a
	// third-party OIDC identity provider.
	CreateUserSession(context.Context) (OIDCAuthDetails, error)
	// Delete deletes the Session identified by the token in use by this client.
	Delete(context.Context) error
}

type sessionsClient struct {
	*rm.BaseClient
}

// NewSessionsClient returns a specialized client for managing Brigade API
// Sessions.
func NewSessionsClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) SessionsClient {
	return &sessionsClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (s *sessionsClient) CreateRootSession(
	ctx context.Context,
	password string,
) (Token, error) {
	includeAuthHeader := false
	token := Token{}
	return token, s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:            http.MethodPost,
			Path:              "v2/sessions",
			IncludeAuthHeader: &includeAuthHeader,
			Headers: map[string]string{
				"Authorization": fmt.Sprintf(
					"Basic %s",
					base64.StdEncoding.EncodeToString(
						[]byte(fmt.Sprintf("root:%s", password)),
					),
				),
			},
			QueryParams: map[string]string{
				"root": "true",
			},
			SuccessCode: http.StatusCreated,
			RespObj:     &token,
		},
	)
}

func (s *sessionsClient) CreateUserSession(
	ctx context.Context,
) (OIDCAuthDetails, error) {
	includeAuthHeader := false
	oidcAuthDetails := OIDCAuthDetails{}
	return oidcAuthDetails, s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:            http.MethodPost,
			Path:              "v2/sessions",
			IncludeAuthHeader: &includeAuthHeader,
			SuccessCode:       http.StatusCreated,
			RespObj:           &oidcAuthDetails,
		},
	)
}

func (s *sessionsClient) Delete(ctx context.Context) error {
	return s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/session",
			SuccessCode: http.StatusOK,
		},
	)
}
