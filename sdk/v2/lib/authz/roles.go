package authz

// RoleType is a type whose values can be used to disambiguate one type of Role
// from another.
type RoleType string

// RoleName is a type whose value maps to a well-defined Brigade Role.
type RoleName string

// RoleScopeGlobal represents an unbounded scope.
const RoleScopeGlobal = "*"

// Role represents a set of permissions, with domain-specific meaning, held by a
// principal, such as a User or ServiceAccount via a RoleAssignment.
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
