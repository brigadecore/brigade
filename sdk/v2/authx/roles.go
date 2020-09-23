package authx

import (
	"encoding/json"

	"github.com/brigadecore/brigade/sdk/v2/meta"
)

// RoleName is a type whose value maps to a well-defined Brigade Role.
type RoleName string

// RoleType is a type whose values can be used to differentiate one type of Role
// from another. This allows (for instance) system-level Roles to be
// differentiated from project-level Roles.
type RoleType string

// Role represents a set of permissions, with domain-specific meaning, held by a
// principal, such as a User or ServiceAccount.
type Role struct {
	// Type indicates the Role's type, for instance, system-level or
	// project-level.
	Type RoleType `json:"type,omitempty"`
	// Name is the name of a Role and has domain-specific meaning.
	Name RoleName `json:"name,omitempty"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific RoleName.
	Scope string `json:"scope,omitempty"`
}

// RoleAssignment represents the assignment of a Role to a principal.
type RoleAssignment struct {
	// Role specifies a Role.
	Role RoleName `json:"role,omitempty"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific RoleName.
	Scope string `json:"scope,omitempty"`
	// PrincipalType qualifies what kind of principal is referenced by the
	// PrincipalID field.
	PrincipalType PrincipalType `json:"principalType,omitempty"`
	// PrincipalID references a principal. The PrincipalType qualifies what type
	// of principal that is-- for instance, a User or a ServiceAccount.
	PrincipalID string `json:"principalID,omitempty"`
}

// MarshalJSON amends RoleAssignment instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (r RoleAssignment) MarshalJSON() ([]byte, error) {
	type Alias RoleAssignment
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "RoleAssignment",
			},
			Alias: (Alias)(r),
		},
	)
}
