package authx

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/meta"
)

// OIDCAuthDetails encapsulates all information required for a client
// authenticating by means of OpenID Connect to complete the authentication
// process using a third-party OIDC identity provider.
type OIDCAuthDetails struct {
	// OAuth2State is an opaque token issued by Brigade that must be sent to the
	// OIDC identity provider as a query parameter when the OIDC authentication
	// workflow continues (in the user's web browser). The OIDC identity provider
	// includes this token when it issues a callback to the Brigade API server
	// after successful authentication. This permits the Brigade API server to
	// correlate the successful authentication by the OIDC identity provider with
	// an existing, but as-yet-unactivated token. This proof of authentication
	// allows Brigade to activate the token and associate it with the User that
	// the OIDC identity provider indicates has successfully completed
	// authentication.
	OAuth2State string `json:"oauth2State,omitempty"`
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
	// CreateRootSession creates a Session for the root user (if enabled by th
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
	*restmachinery.BaseClient
}

// NewSessionsClient returns a specialized client for managing Brigade API
// Sessions.
func NewSessionsClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) SessionsClient {
	return &sessionsClient{
		BaseClient: restmachinery.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (s *sessionsClient) CreateRootSession(
	ctx context.Context,
	password string,
) (Token, error) {
	token := Token{}
	return token, s.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/sessions",
			AuthHeaders: s.BasicAuthHeaders("root", password),
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
	oidcAuthDetails := OIDCAuthDetails{}
	return oidcAuthDetails, s.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/sessions",
			SuccessCode: http.StatusCreated,
			RespObj:     &oidcAuthDetails,
		},
	)
}

func (s *sessionsClient) Delete(ctx context.Context) error {
	return s.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/session",
			AuthHeaders: s.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}
