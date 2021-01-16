package authx

import (
	"context"
)

// RoleType is a type whose values can be used to disambiguate one type of Role
// from another. This allows, for instance, system-level Roles to be
// differentiated from project-level Roles.
type RoleType string

// RoleTypeSystem represents a system-level Role.
const RoleTypeSystem RoleType = "SYSTEM"

// RoleName is a type whose value maps to a well-defined Brigade Role.
type RoleName string

// Role represents a set of permissions, with domain-specific meaning, held by a
// principal, such as a User or ServiceAccount via a RoleAssignment.
type Role struct {
	// Type indicates the Role's type, for instance, system-level or
	// project-level.
	Type RoleType `json:"type,omitempty" bson:"type,omitempty"`
	// Name is the name of a Role and has domain-specific meaning.
	Name RoleName `json:"name,omitempty" bson:"name,omitempty"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific RoleName.
	Scope string `json:"scope,omitempty" bson:"scope,omitempty"`
}

// RoleAssignment represents the assignment of a Role to a principal such as a
// User or ServiceAccount.
type RoleAssignment struct {
	// Role assigns a Role to the specified principal.
	Role Role `json:"role" bson:"role"`
	// Principal specifies the principal to whom the Role is assigned.
	Principal PrincipalReference `json:"principal" bson:"principal"`
}

// RoleAssignmentsStore is an interface for components that implement
// RoleAssignment persistence concerns.
type RoleAssignmentsStore interface {
	// Grant the role specified by the RoleAssignment to the principal specified
	// by the RoleAssignment.
	Grant(ctx context.Context, roleAssignment RoleAssignment) error
	// Revoke the role specified by the RoleAssignment for the principal specified
	// by the RoleAssignment.
	Revoke(ctx context.Context, roleAssignment RoleAssignment) error
}
