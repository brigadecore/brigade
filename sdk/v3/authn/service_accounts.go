package authn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	rm "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	"github.com/brigadecore/brigade/sdk/v3/restmachinery"
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

// ServiceAccountCreateOptions represents useful, optional settings for the
// creation of new ServiceAccounts. It currently has no fields, but exists to
// preserve the possibility of future expansion without having to change client
// function signatures.
type ServiceAccountCreateOptions struct{}

// ServiceAccountGetOptions represents useful, optional criteria for the
// retrieval of a ServiceAccount. It currently has no fields, but exists to
// preserve the possibility of future expansion without having to change client
// function signatures.
type ServiceAccountGetOptions struct{}

// ServiceAccountLockOptions represents useful, optional settings for the
// locking of a ServiceAccount. It currently has no fields, but exists to
// preserve the possibility of future expansion without having to change client
// function signatures.
type ServiceAccountLockOptions struct{}

// ServiceAccountUnlockOptions represents useful, optional settings for the
// unlocking of a ServiceAccount. It currently has no fields, but exists to
// preserve the possibility of future expansion without having to change client
// function signatures.
type ServiceAccountUnlockOptions struct{}

// ServiceAccountDeleteOptions represents useful, optional settings for the
// deletion of a ServiceAccount. It currently has no fields, but exists to
// preserve the possibility of future expansion without having to change client
// function signatures.
type ServiceAccountDeleteOptions struct{}

// ServiceAccountsSelector represents useful filter criteria when selecting
// multiple ServiceAccounts for API group operations like list. It currently has
// no fields, but exists to preserve the possibility of future expansion without
// having to change client function signatures.
type ServiceAccountsSelector struct{}

// ServiceAccountsClient is the specialized client for managing ServiceAccounts
// with the Brigade API.
type ServiceAccountsClient interface {
	// Create creates a new ServiceAccount.
	Create(
		context.Context,
		ServiceAccount,
		*ServiceAccountCreateOptions,
	) (Token, error)
	// List returns a ServiceAccountList.
	List(
		context.Context,
		*ServiceAccountsSelector,
		*meta.ListOptions,
	) (ServiceAccountList, error)
	// Get retrieves a single ServiceAccount specified by its identifier.
	Get(
		context.Context,
		string,
		*ServiceAccountGetOptions,
	) (ServiceAccount, error)

	// Lock revokes system access for a single ServiceAccount specified by its
	// identifier.
	Lock(context.Context, string, *ServiceAccountLockOptions) error
	// Unlock restores system access for a single ServiceAccount (after presumably
	// having been revoked) specified by its identifier. It returns a new Token.
	Unlock(context.Context, string, *ServiceAccountUnlockOptions) (Token, error)

	// Delete deletes a single ServiceAccount specified by its identifier.
	Delete(context.Context, string, *ServiceAccountDeleteOptions) error
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
	_ *ServiceAccountCreateOptions,
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
	_ *ServiceAccountGetOptions,
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

func (s *serviceAccountsClient) Lock(
	ctx context.Context,
	id string,
	_ *ServiceAccountLockOptions,
) error {
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
	_ *ServiceAccountUnlockOptions,
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

func (s *serviceAccountsClient) Delete(
	ctx context.Context,
	id string,
	_ *ServiceAccountDeleteOptions,
) error {
	return s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/service-accounts/%s", id),
			SuccessCode: http.StatusOK,
		},
	)
}
