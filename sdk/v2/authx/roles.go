package authx

import (
	"encoding/json"

	"github.com/brigadecore/brigade/sdk/v2/meta"
)

// RoleName is a type whose value maps to a well-defined Brigade Role.
type RoleName string

const (
	// RoleNameAdmin is the name of a system-level Role that enables principals to
	// manage Users, ServiceAccounts, and system-level permissions for Users and
	// ServiceAccounts.
	RoleNameAdmin RoleName = "ADMIN"
	// RoleNameEventCreator is the name of a system-level Role that enables
	// principals to create Events for all Projects.
	RoleNameEventCreator RoleName = "EVENT_CREATOR"
	// RoleNameProjectAdmin is the name of a project-level Role that enables
	// principals to manage all aspects of a given Project, including the
	// Project's secrets.
	RoleNameProjectAdmin RoleName = "PROJECT_ADMIN"
	// RoleNameProjectCreator is the name of a system-level Role that enables
	// principals to create new Projects.
	RoleNameProjectCreator RoleName = "PROJECT_CREATOR"
	// RoleNameProjectDeveloper is the name of a project-level Role that enables
	// principals to update Projects. This Role does NOT enable event creation
	// or secret management.
	RoleNameProjectDeveloper RoleName = "PROJECT_DEVELOPER"
	// RoleNameProjectUser is the name of a project-level Role that enables
	// principals to create and manage Events for a Project.
	RoleNameProjectUser RoleName = "PROJECT_USER"
	// RoleNameReader is the name of a system-level Role that enables global read
	// access.
	RoleNameReader RoleName = "READER"
)

// RoleType is a type whose values can be used to disambiguate one type of Role
// from another. This allows, for instance, system-level Roles to be
// differentiated from project-level Roles.
type RoleType string

const (
	// RoleTypeProject represents a project-level Role.
	RoleTypeProject RoleType = "PROJECT"
	// RoleTypeSystem represents a system-level Role.
	RoleTypeSystem RoleType = "SYSTEM"
)

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
