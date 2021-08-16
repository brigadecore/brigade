package authn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// ServiceAccount represents a non-human Brigade user, such as an Event
// gateway.
type ServiceAccount struct {
	// ObjectMeta encapsulates ServiceAccount metadata.
	meta.ObjectMeta `json:"metadata"`
	// Description is a natural language description of the ServiceAccount's
	// purpose.
	Description string `json:"description,omitempty"`
	// Locked indicates when the ServiceAccount has been locked out of the system
	// by an administrator. If this field's value is nil, the ServiceAccount is
	// not locked.
	Locked *time.Time `json:"locked,omitempty"`
}

// MarshalJSON amends ServiceAccount instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (s ServiceAccount) MarshalJSON() ([]byte, error) {
	type Alias ServiceAccount
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ServiceAccount",
			},
			Alias: (Alias)(s),
		},
	)
}

// ServiceAccountList is an ordered and pageable list of ServiceAccounts.
type ServiceAccountList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of ServiceAccounts.
	Items []ServiceAccount `json:"items,omitempty"`
}

// MarshalJSON amends ServiceAccountList instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (s ServiceAccountList) MarshalJSON() ([]byte, error) {
	type Alias ServiceAccountList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ServiceAccountList",
			},
			Alias: (Alias)(s),
		},
	)
}

// ServiceAccountsSelector represents useful filter criteria when selecting
// multiple ServiceAccounts for API group operations like list. It currently has
// no fields, but exists to preserve the possibility of future expansion without
// having to change client function signatures.
type ServiceAccountsSelector struct{}

// ServiceAccountsClient is the specialized client for managing ServiceAccounts
// with the Brigade API.
type ServiceAccountsClient interface {
	// Create creates a new ServiceAccount.
	Create(context.Context, ServiceAccount) (Token, error)
	// List returns a ServiceAccountList.
	List(
		context.Context,
		*ServiceAccountsSelector,
		*meta.ListOptions,
	) (ServiceAccountList, error)
	// Get retrieves a single ServiceAccount specified by its identifier.
	Get(context.Context, string) (ServiceAccount, error)

	// Lock revokes system access for a single ServiceAccount specified by its
	// identifier.
	Lock(context.Context, string) error
	// Unlock restores system access for a single ServiceAccount (after presumably
	// having been revoked) specified by its identifier. It returns a new Token.
	Unlock(context.Context, string) (Token, error)

	// Delete deletes a single ServiceAccount specified by its identifier.
	Delete(context.Context, string) error
}

type serviceAccountsClient struct {
	*rm.BaseClient
}

// NewServiceAccountsClient returns a specialized client for managing
// ServiceAccounts.
func NewServiceAccountsClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) ServiceAccountsClient {
	return &serviceAccountsClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (s *serviceAccountsClient) Create(
	ctx context.Context,
	serviceAccount ServiceAccount,
) (Token, error) {
	token := Token{}
	return token, s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/service-accounts",
			ReqBodyObj:  serviceAccount,
			SuccessCode: http.StatusCreated,
			RespObj:     &token,
		},
	)
}

func (s *serviceAccountsClient) List(
	ctx context.Context,
	_ *ServiceAccountsSelector,
	opts *meta.ListOptions,
) (ServiceAccountList, error) {
	serviceAccounts := ServiceAccountList{}
	return serviceAccounts, s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/service-accounts",
			QueryParams: s.AppendListQueryParams(nil, opts),
			SuccessCode: http.StatusOK,
			RespObj:     &serviceAccounts,
		},
	)
}

func (s *serviceAccountsClient) Get(
	ctx context.Context,
	id string,
) (ServiceAccount, error) {
	serviceAccount := ServiceAccount{}
	return serviceAccount, s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/service-accounts/%s", id),
			SuccessCode: http.StatusOK,
			RespObj:     &serviceAccount,
		},
	)
}

func (s *serviceAccountsClient) Lock(ctx context.Context, id string) error {
	return s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/service-accounts/%s/lock", id),
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *serviceAccountsClient) Unlock(
	ctx context.Context,
	id string,
) (Token, error) {
	token := Token{}
	return token, s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/service-accounts/%s/lock", id),
			SuccessCode: http.StatusOK,
			RespObj:     &token,
		},
	)
}

func (s *serviceAccountsClient) Delete(ctx context.Context, id string) error {
	return s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/service-accounts/%s", id),
			SuccessCode: http.StatusOK,
		},
	)
}
