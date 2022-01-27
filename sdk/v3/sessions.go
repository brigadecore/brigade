package sdk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	rm "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	"github.com/brigadecore/brigade/sdk/v3/restmachinery"
)

// UserSessionCreateOptions encapsulates user-specified options when creating a
// new Session that will authenticate using a third-party identity provider.
type UserSessionCreateOptions struct {
	// SuccessURL indicates where users should be redirected to after successful
	// completion of a third-party authentication workflow. If this is left
	// unspecified, users will be redirected to a default success page.
	SuccessURL string
}

// ThirdPartyAuthDetails encapsulates all information required for a client
// authenticating by means of a third-party identity provider to complete the
// authentication workflow.
type ThirdPartyAuthDetails struct {
	// AuthURL is a URL that can be requested in a user's web browser to complete
	// authentication via a third-party identity provider.
	AuthURL string `json:"authURL,omitempty"`
	// Token is an opaque bearer token issued by Brigade to correlate a User with
	// a Session. It remains unactivated (useless) until the authentication
	// workflow is successfully completed. Clients may expect that that the token
	// expires (at an interval determined by a system administrator) and, for
	// simplicity, is NOT refreshable. When the token has expired,
	// re-authentication is required.
	Token string `json:"token,omitempty"`
}

// MarshalJSON amends ThirdPartyAuthDetails instances with type metadata so
// that clients do not need to be concerned with the tedium of doing so.
func (t ThirdPartyAuthDetails) MarshalJSON() ([]byte, error) {
	type Alias ThirdPartyAuthDetails
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ThirdPartyAuthDetails",
			},
			Alias: (Alias)(t),
		},
	)
}

// RootSessionCreateOptions represents useful, optional settings for the
// creation of a root Session. It currently has no fields, but exists to
// preserve the possibility of future expansion without having to change client
// function signatures.
type RootSessionCreateOptions struct{}

// SessionDeleteOptions represents useful, optional settings for the deletion of
// a Session. It currently has no fields, but exists to preserve the possibility
// of future expansion without having to change client function signatures.
type SessionDeleteOptions struct{}

// SessionsClient is the specialized client for managing Brigade API Sessions.
type SessionsClient interface {
	// CreateRootSession creates a Session for the root user (if enabled by the
	// system administrator) and returns a Token with a short expiry period
	// (determined by a system administrator). In contrast to most other
	// operations exposed by the Brigade API, a valid token is not required to
	// invoke this.
	CreateRootSession(
		ctx context.Context,
		password string,
		opts *RootSessionCreateOptions,
	) (Token, error)
	// CreateUserSession creates a new User Session and initiates an
	// authentication workflow with a third-party identity provider. It returns
	// ThirdPartyAuthDetails containing all information required to continue the
	// authentication process.
	CreateUserSession(
		context.Context,
		*UserSessionCreateOptions,
	) (ThirdPartyAuthDetails, error)
	// Delete deletes the Session identified by the token in use by this client.
	Delete(context.Context, *SessionDeleteOptions) error
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
	_ *RootSessionCreateOptions,
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
	opts *UserSessionCreateOptions,
) (ThirdPartyAuthDetails, error) {
	includeAuthHeader := false
	thirdPartyAuthDetails := ThirdPartyAuthDetails{}
	queryParams := map[string]string{}
	if opts != nil {
		if opts.SuccessURL != "" {
			queryParams["successURL"] = opts.SuccessURL
		}
	}
	return thirdPartyAuthDetails, s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:            http.MethodPost,
			Path:              "v2/sessions",
			IncludeAuthHeader: &includeAuthHeader,
			SuccessCode:       http.StatusCreated,
			RespObj:           &thirdPartyAuthDetails,
			QueryParams:       queryParams,
		},
	)
}

func (s *sessionsClient) Delete(
	ctx context.Context,
	_ *SessionDeleteOptions,
) error {
	return s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/session",
			SuccessCode: http.StatusOK,
		},
	)
}
