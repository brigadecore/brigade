package authn

import (
	"context"
	"encoding/json"
	"time"

	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/brigadecore/brigade/v2/apiserver/internal/system"
	"github.com/pkg/errors"
)

// ServiceAccountKind represents the canonical Service Account kind string
const ServiceAccountKind = "ServiceAccount"

// ServiceAccount represents a non-human Brigade user, such as an Event
// gateway.
type ServiceAccount struct {
	// ObjectMeta encapsulates ServiceAccount metadata.
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	// Description is a natural language description of the ServiceAccount's
	// purpose.
	Description string `json:"description" bson:"description"`
	// HashedToken is a secure, one-way hash of the ServiceAccount's token.
	HashedToken string `json:"-" bson:"hashedToken"`
	// Locked indicates when the ServiceAccount has been locked out of the system
	// by an administrator. If this field's value is nil, the ServiceAccount is
	// not locked.
	Locked *time.Time `json:"locked,omitempty" bson:"locked"`
}

// MarshalJSON amends ServiceAccount instances with type metadata.
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

// MarshalJSON amends ServiceAccountList instances with type metadata.
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

// ServiceAccountsService is the specialized interface for managing
// ServiceAccounts. It's decoupled from underlying technology choices (e.g. data
// store) to keep business logic reusable and consistent while the underlying
// tech stack remains free to change.
type ServiceAccountsService interface {
	// Create creates a new ServiceAccount. If a ServiceAccount having the same ID
	// already exists, implementations MUST return a *meta.ErrConflict error.
	Create(context.Context, ServiceAccount) (Token, error)
	// List retrieves a ServiceAccountList.
	List(context.Context, meta.ListOptions) (ServiceAccountList, error)
	// Get retrieves a single ServiceAccount specified by its identifier. If the
	// specified ServiceAccount does not exist, implementations MUST return a
	// *meta.ErrNotFound error.
	Get(context.Context, string) (ServiceAccount, error)
	// GetByToken retrieves a single ServiceAccount specified by token. If no
	// such ServiceAccount exists, implementations MUST return a *meta.ErrNotFound
	// error.
	GetByToken(context.Context, string) (ServiceAccount, error)

	// Lock revokes system access for a single ServiceAccount specified by its
	// identifier. If the specified ServiceAccount does not exist, implementations
	// MUST return a *meta.ErrNotFound error.
	Lock(context.Context, string) error
	// Unlock restores system access for a single ServiceAccount (after presumably
	// having been revoked) specified by its identifier. It returns a new Token.
	// If the specified ServiceAccount does not exist, implementations MUST return
	// a *meta.ErrNotFound error.
	Unlock(context.Context, string) (Token, error)
}

type serviceAccountsService struct {
	authorize libAuthz.AuthorizeFn
	store     ServiceAccountsStore
}

// NewServiceAccountsService returns a specialized interface for managing
// ServiceAccounts.
func NewServiceAccountsService(
	authorizeFn libAuthz.AuthorizeFn,
	store ServiceAccountsStore,
) ServiceAccountsService {
	return &serviceAccountsService{
		authorize: authorizeFn,
		store:     store,
	}
}

func (s *serviceAccountsService) Create(
	ctx context.Context,
	serviceAccount ServiceAccount,
) (Token, error) {
	token := Token{}

	if err := s.authorize(ctx, system.RoleAdmin(), ""); err != nil {
		return token, err
	}

	token.Value = crypto.NewToken(256)
	now := time.Now().UTC()
	serviceAccount.Created = &now
	serviceAccount.HashedToken = crypto.Hash("", token.Value)
	if err := s.store.Create(ctx, serviceAccount); err != nil {
		return token, errors.Wrapf(
			err,
			"error storing new service account %q",
			serviceAccount.ID,
		)
	}
	return token, nil
}

func (s *serviceAccountsService) List(
	ctx context.Context,
	opts meta.ListOptions,
) (ServiceAccountList, error) {
	if err := s.authorize(ctx, system.RoleReader(), ""); err != nil {
		return ServiceAccountList{}, err
	}

	if opts.Limit == 0 {
		opts.Limit = 20
	}
	serviceAccounts, err := s.store.List(ctx, opts)
	if err != nil {
		return serviceAccounts,
			errors.Wrap(err, "error retrieving service accounts from store")
	}
	return serviceAccounts, nil
}

func (s *serviceAccountsService) Get(
	ctx context.Context,
	id string,
) (ServiceAccount, error) {
	if err := s.authorize(ctx, system.RoleReader(), ""); err != nil {
		return ServiceAccount{}, err
	}

	serviceAccount, err := s.store.Get(ctx, id)
	if err != nil {
		return serviceAccount, errors.Wrapf(
			err,
			"error retrieving service account %q from store",
			id,
		)
	}
	return serviceAccount, nil
}

func (s *serviceAccountsService) GetByToken(
	ctx context.Context,
	token string,
) (ServiceAccount, error) {
	// No authz requirements here because this is is never invoked at the explicit
	// request of an end user; rather it is invoked only by the system itself.

	serviceAccount, err := s.store.GetByHashedToken(
		ctx,
		crypto.Hash("", token),
	)
	if err != nil {
		return serviceAccount, errors.Wrap(
			err,
			"error retrieving service account from store by token",
		)
	}
	return serviceAccount, nil
}

func (s *serviceAccountsService) Lock(ctx context.Context, id string) error {
	if err := s.authorize(ctx, system.RoleAdmin(), ""); err != nil {
		return err
	}

	if err := s.store.Lock(ctx, id); err != nil {
		return errors.Wrapf(
			err,
			"error locking service account %q in the store",
			id,
		)
	}
	return nil
}

func (s *serviceAccountsService) Unlock(
	ctx context.Context,
	id string,
) (Token, error) {
	if err := s.authorize(ctx, system.RoleAdmin(), ""); err != nil {
		return Token{}, err
	}

	newToken := Token{
		Value: crypto.NewToken(256),
	}
	if err := s.store.Unlock(
		ctx,
		id,
		crypto.Hash("", newToken.Value),
	); err != nil {
		return newToken, errors.Wrapf(
			err,
			"error unlocking service account %q in the store",
			id,
		)
	}
	return newToken, nil
}

// ServiceAccountsStore is an interface for components that implement
// ServiceAccount persistence concerns.
type ServiceAccountsStore interface {
	// Create persists a new ServiceAccount in the underlying data store. If a
	// ServiceAccount having the same ID already exists, implementations MUST
	// return a *meta.ErrConflict error.
	Create(context.Context, ServiceAccount) error
	// List retrieves a ServiceAccountList from the underlying data store, with
	// its Items (ServiceAccounts) ordered by ID.
	List(context.Context, meta.ListOptions) (ServiceAccountList, error)
	// Get retrieves a single ServiceAccount from the underlying data store. If
	// the specified ServiceAccount does not exist, implementations MUST return
	// a *meta.ErrNotFound error.
	Get(context.Context, string) (ServiceAccount, error)
	// GetByHashedToken retrieves a single ServiceAccount having the provided
	// hashed token from the underlying data store. If no such ServiceAccount
	// exists, implementations MUST return a *meta.ErrNotFound error.
	GetByHashedToken(context.Context, string) (ServiceAccount, error)

	// Lock updates the specified ServiceAccount in the underlying data store to
	// reflect that it has been locked out of the system. If the specified
	// ServiceAccount does not exist, implementations MUST return a
	// *meta.ErrNotFound error.
	Lock(context.Context, string) error
	// Unlock updates the specified ServiceAccount in the underlying data store to
	// reflect that it's system access (after presumably having been revoked) has
	// been restored. A hashed token must be provided as a replacement for the
	// existing token. If the specified ServiceAccount does not exist,
	// implementations MUST return a *meta.ErrNotFound error.
	Unlock(ctx context.Context, id string, newHashedToken string) error
}
