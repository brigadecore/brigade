package api

import (
	"context"
	"encoding/json"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

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
	// Set sets the value of a new Secret or updates the value of an existing
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
	authorize        AuthorizeFn
	projectAuthorize ProjectAuthorizeFn
	projectsStore    ProjectsStore
	secretsStore     SecretsStore
}

// NewSecretsService returns a specialized interface for managing Secrets.
func NewSecretsService(
	authorizeFn AuthorizeFn,
	projectAuthorize ProjectAuthorizeFn,
	projectsStore ProjectsStore,
	secretsStore SecretsStore,
) SecretsService {
	return &secretsService{
		authorize:        authorizeFn,
		projectAuthorize: projectAuthorize,
		projectsStore:    projectsStore,
		secretsStore:     secretsStore,
	}
}

func (s *secretsService) List(
	ctx context.Context,
	projectID string,
	opts meta.ListOptions,
) (SecretList, error) {
	if err := s.authorize(ctx, RoleReader, ""); err != nil {
		return SecretList{}, err
	}

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
			"error getting secrets for project %q from store",
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
	if err := s.projectAuthorize(ctx, projectID, RoleProjectAdmin); err != nil {
		return err
	}

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
			"error setting secret %q for project %q in store",
			secret.Key,
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
	if err := s.projectAuthorize(ctx, projectID, RoleProjectAdmin); err != nil {
		return err
	}

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
			"error unsetting secret %q for project %q in store",
			key,
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
