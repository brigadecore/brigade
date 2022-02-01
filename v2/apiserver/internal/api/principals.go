package api

import (
	"context"
	"encoding/json"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
)

// PrincipalReferenceKind represents the canonical PrincipalReference kind
// string
const PrincipalReferenceKind = "PrincipalReference"

// PrincipalType is a type whose values can be used to disambiguate one type of
// principal from another. For instance, when assigning a Role to a principal
// via a RoleAssignment, a PrincipalType field is used to indicate whether the
// value of the PrincipalID field reflects a User ID or a ServiceAccount ID.
type PrincipalType string

const (
	// PrincipalTypeServiceAccount represents a principal that is authenticated as
	// the root user.
	PrincipalTypeRoot PrincipalType = "ROOT"
	// PrincipalTypeServiceAccount represents a principal that is a
	// ServiceAccount.
	PrincipalTypeServiceAccount PrincipalType = "SERVICE_ACCOUNT"
	// PrincipalTypeUser represents a principal that is a User.
	PrincipalTypeUser PrincipalType = "USER"
)

// PrincipalReference is a reference to any sort of security principal (human
// user, service account, etc.)
type PrincipalReference struct {
	// Type qualifies what kind of principal is referenced by the ID field-- for
	// instance, a User or a ServiceAccount.
	Type PrincipalType `json:"type,omitempty"`
	// ID references a principal. The Type qualifies what type of principal that
	// is-- for instance, a User or a ServiceAccount.
	ID string `json:"id,omitempty"`
}

// MarshalJSON amends PrincipalReference instances with type metadata.
func (p PrincipalReference) MarshalJSON() ([]byte, error) {
	type Alias PrincipalReference
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       PrincipalReferenceKind,
			},
			Alias: (Alias)(p),
		},
	)
}

type principalContextKey struct{}

// ContextWithPrincipal returns a context.Context that has been augmented with
// the provided principal.
func ContextWithPrincipal(
	ctx context.Context,
	principal interface{},
) context.Context {
	return context.WithValue(
		ctx,
		principalContextKey{},
		principal,
	)
}

// PrincipalFromContext extracts a principal from the provided context.Context
// and returns it.
func PrincipalFromContext(ctx context.Context) interface{} {
	return ctx.Value(principalContextKey{})
}

// PrincipalsService is a generalized interface for retrieving information about
// a User or Service account when it is not known which of the two one is
// dealing with. It's decoupled from underlying technology choices (e.g. data
// store) to keep business logic reusable and consistent while the underlying
// tech stack remains free to change.
type PrincipalsService interface {
	// WhoAmI returns a reference to the current authenticated principal.
	WhoAmI(context.Context) (PrincipalReference, error)
}

// principalsService is an implementation of the PrincipalsService interface.
type principalsService struct {
	authorize AuthorizeFn
}

// NewUsersService returns a generalized interface for retrieving principal
// information.
func NewPrincipalsService(authorizeFn AuthorizeFn) PrincipalsService {
	return &principalsService{
		authorize: authorizeFn,
	}
}

func (p *principalsService) WhoAmI(
	ctx context.Context,
) (PrincipalReference, error) {
	ref := PrincipalReference{}
	switch principal := PrincipalFromContext(ctx).(type) {
	case *RootPrincipal:
		ref.Type = PrincipalTypeRoot
		ref.ID = "root"
	case *ServiceAccount:
		ref.Type = PrincipalTypeServiceAccount
		ref.ID = principal.ID
	case *User:
		ref.Type = PrincipalTypeUser
		ref.ID = principal.ID
	default: // What kind of principal is this??? This shouldn't happen.
		return ref, &meta.ErrAuthorization{}
	}
	return ref, nil
}
