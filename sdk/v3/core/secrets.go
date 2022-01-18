package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	rm "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	"github.com/brigadecore/brigade/sdk/v3/restmachinery"
)

// Secret represents Project-level sensitive information.
type Secret struct {
	// Key is a key by which the secret can referred.
	Key string `json:"key,omitempty"`
	// Value is the sensitive information. This is a write-only field.
	Value string `json:"value,omitempty"`
}

// MarshalJSON amends Secret instances with type metadata so that clients do not
// need to be concerned with the tedium of doing so.
func (s Secret) MarshalJSON() ([]byte, error) {
	type Alias Secret
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "Secret",
			},
			Alias: (Alias)(s),
		},
	)
}

// SecretList is an ordered and pageable list of Secrets.
type SecretList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Secrets.
	Items []Secret `json:"items,omitempty"`
}

// MarshalJSON amends SecretList instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (s SecretList) MarshalJSON() ([]byte, error) {
	type Alias SecretList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "SecretList",
			},
			Alias: (Alias)(s),
		},
	)
}

// SecretSetOptions represents useful, optional settings for assigning a new
// value to a Project Secret. It currently has no fields, but exists to preserve
// the possibility of future expansion without having to change client function
// signatures.
type SecretSetOptions struct{}

// SecretUnsetOptions represents useful, optional settings for un-assigning a
// value from a Project Secret. It currently has no fields, but exists to
// preserve the possibility of future expansion without having to change client
// function signatures.
type SecretUnsetOptions struct{}

// SecretsClient is the specialized client for managing Secrets with the
// Brigade API.
type SecretsClient interface {
	// List returns a SecretList whose Items (Secrets) contain Keys only and not
	// Values (all Value fields are empty). i.e. Once a secret is set, end clients
	// are unable to retrieve values.
	List(
		ctx context.Context,
		projectID string,
		opts *meta.ListOptions,
	) (SecretList, error)
	// Set sets the value of a new Secret or updates the value of an existing
	// Secret. If the specified Key does not exist, it is created. If the
	// specified Key does exist, its corresponding Value is overwritten.
	Set(
		ctx context.Context,
		projectID string,
		secret Secret,
		opts *SecretSetOptions,
	) error
	// Unset clears the value of an existing Secret. If the specified Key does not
	// exist, no error is returned.
	Unset(
		ctx context.Context,
		projectID string,
		key string,
		opts *SecretUnsetOptions,
	) error
}

type secretsClient struct {
	*rm.BaseClient
}

// NewSecretsClient returns a specialized client for managing
// Secrets.
func NewSecretsClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) SecretsClient {
	return &secretsClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (s *secretsClient) List(
	ctx context.Context,
	projectID string,
	opts *meta.ListOptions,
) (SecretList, error) {
	secrets := SecretList{}
	return secrets, s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/projects/%s/secrets", projectID),
			QueryParams: s.AppendListQueryParams(nil, opts),
			SuccessCode: http.StatusOK,
			RespObj:     &secrets,
		},
	)
}

func (s *secretsClient) Set(
	ctx context.Context,
	projectID string,
	secret Secret,
	_ *SecretSetOptions,
) error {
	return s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method: http.MethodPut,
			Path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				secret.Key,
			),
			ReqBodyObj:  secret,
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *secretsClient) Unset(
	ctx context.Context,
	projectID string,
	key string,
	_ *SecretUnsetOptions,
) error {
	return s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method: http.MethodDelete,
			Path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				key,
			),
			SuccessCode: http.StatusOK,
		},
	)
}
