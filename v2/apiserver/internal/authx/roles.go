package authx

import "context"

// RoleType is a type whose values can be used to disambiguate one type of Role
// from another. This allows, for instance, system-level Roles to be
// differentiated from project-level Roles.
type RoleType string

const (
	// RoleTypeEvent represents an event-level Role. In practice, these are only
	// used by Workers who have no business reading or writing anything related to
	// any Event other than the one they were spawned to handle.
	RoleTypeEvent RoleType = "EVENT"
	// RoleTypeProject represents a project-level Role.
	RoleTypeProject RoleType = "PROJECT"
	// RoleTypeSystem represents a system-level Role.
	RoleTypeSystem RoleType = "SYSTEM"
)

// RoleName is a type whose value maps to a well-defined Brigade Role.
type RoleName string

// RoleScopeGlobal represents an unbounded scope.
const RoleScopeGlobal = "*"

// Role represents a set of permissions, with domain-specific meaning, held by a
// principal, such as a User or ServiceAccount via a RoleAssignment.
type Role struct {
	// Type indicates the Role's type, for instance, system-level or
	// project-level.
	Type RoleType `json:"type" bson:"type"`
	// Name is the name of a Role and has domain-specific meaning.
	Name RoleName `json:"name" bson:"name"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific RoleName.
	Scope string `json:"scope" bson:"scope"`
}

// RoleAssignment represents the assignment of a Role to a principal such as a
// User or ServiceAccount.
type RoleAssignment struct {
	// Note the explicit lack of RoleType field. Project-level RoleAssignments
	// and system-level RoleAssignment are handled by different endpoints. The
	// endpoints provide that context.

	// Role specifies a Role.
	Role RoleName `json:"role"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific RoleName.
	Scope string `json:"scope"`
	// PrincipalType qualifies what kind of principal is referenced by the
	// PrincipalID field-- for instance, a User or a ServiceAccount.
	PrincipalType PrincipalType `json:"principalType"`
	// PrincipalID references a principal. The PrincipalType qualifies what type
	// of principal that is-- for instance, a User or a ServiceAccount.
	PrincipalID string `json:"principalID"`
}

// RolesStore is an interface for components that implement Role persistence
// concerns.
type RolesStore interface {
	Grant(
		ctx context.Context,
		principalType PrincipalType,
		principalID string,
		roles ...Role,
	) error
	Revoke(
		ctx context.Context,
		principalType PrincipalType,
		principalID string,
		roles ...Role,
	) error
}
