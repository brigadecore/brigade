package core

import (
	"context"
	"encoding/json"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// SecretList is an ordered and pageable list of Secrets.
type SecretList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Secrets.
	Items []Secret `json:"items,omitempty"`
}

// All of the following functions are implemented on the SecretsList type to
// facilitate sorting by each constituent Secret's Key field.

// Len returns the cardinality of the underlying Items field.
func (s SecretList) Len() int {
	return len(s.Items)
}

// Swap swaps Secret i and Secret j in the underlying Items field.
func (s SecretList) Swap(i, j int) {
	s.Items[i], s.Items[j] = s.Items[j], s.Items[i]
}

// Less returns true when Secret i's Key field is less than Secret j's Key field
// in the underlying Items field (when compared lexically).
func (s SecretList) Less(i, j int) bool {
	return s.Items[i].Key < s.Items[j].Key
}

// MarshalJSON amends SecretList instances with type metadata.
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

// Secret represents Project-level sensitive information.
type Secret struct {
	// Key is a key by which the secret can referred.
	Key string `json:"key,omitempty"`
	// Value is the sensitive information. This is a write-only field.
	Value string `json:"value,omitempty"`
}

// MarshalJSON amends Secret instances with type metadata.
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

// SecretsService is the specialized interface for managing Secrets. It's
// decoupled from underlying technology choices (e.g. data store, message bus,
// etc.) to keep business logic reusable and consistent while the underlying
// tech stack remains free to change.
type SecretsService interface {
	// List returns a SecretList whose Items (Secrets) contain Keys only and
	// not Values (all Value fields are empty). i.e. Once a secret is set, end
	// clients are unable to retrieve values.
	List(
		ctx context.Context,
		projectID string,
		opts meta.ListOptions,
	) (SecretList, error)
	// Set set the value of a new Secret or updates the value of an existing
	// Secret. If the specified Project does not exist, implementations MUST
	// return a *meta.ErrNotFound error. If the specified Key does not exist, it
	// is created. If the specified Key does exist, its corresponding Value is
	// overwritten.
	Set(
		ctx context.Context,
		projectID string,
		secret Secret,
	) error
	// Unset clears the value of an existing Secret. If the specified Project does
	// not exist, implementations MUST return a *meta.ErrNotFound error. If the
	// specified Key does not exist, no error is returned.
	Unset(ctx context.Context, projectID string, key string) error
}

type secretsService struct {
	projectsStore ProjectsStore
	secretsStore  SecretsStore
}

// NewSecretsService returns a specialized interface for managing Secrets.
func NewSecretsService(
	projectsStore ProjectsStore,
	secretsStore SecretsStore,
) SecretsService {
	return &secretsService{
		projectsStore: projectsStore,
		secretsStore:  secretsStore,
	}
}

func (s *secretsService) List(
	ctx context.Context,
	projectID string,
	opts meta.ListOptions,
) (SecretList, error) {
	secrets := SecretList{}
	project, err := s.projectsStore.Get(ctx, projectID)
	if err != nil {
		return secrets, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	if opts.Limit == 0 {
		opts.Limit = 20
	}
	if secrets, err =
		s.secretsStore.List(ctx, project, opts); err != nil {
		return secrets, errors.Wrapf(
			err,
			"error getting worker secrets for project %q from store",
			projectID,
		)
	}
	return secrets, nil
}

func (s *secretsService) Set(
	ctx context.Context,
	projectID string,
	secret Secret,
) error {
	project, err := s.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	if err := s.secretsStore.Set(ctx, project, secret); err != nil {
		return errors.Wrapf(
			err,
			"error setting secret for project %q worker in store",
			projectID,
		)
	}
	return nil
}

func (s *secretsService) Unset(
	ctx context.Context,
	projectID string,
	key string,
) error {
	project, err := s.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	if err :=
		s.secretsStore.Unset(ctx, project, key); err != nil {
		return errors.Wrapf(
			err,
			"error unsetting secrets for project %q worker in store",
			projectID,
		)
	}
	return nil
}

// SecretsStore is an interface for components that implement Secret persistence
// concerns.
type SecretsStore interface {
	// List returns a SecretList, with its Items (Secrets) ordered lexically by
	// Key.
	List(ctx context.Context,
		project Project,
		opts meta.ListOptions,
	) (SecretList, error)
	// Set adds or updates the provided Secret associated with the specified
	// Project.
	Set(ctx context.Context, project Project, secret Secret) error
	// Unset clears (deletes) the Secret (identified by its Key) associated with
	// the specified Project.
	Unset(ctx context.Context, project Project, key string) error
}
