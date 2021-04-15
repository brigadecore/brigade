package authz

// RoleName is a type whose value maps to a well-defined Brigade Role.
type RoleName string

// RoleScopeGlobal represents an unbounded scope.
const RoleScopeGlobal = "*"

// Role represents a set of permissions, with domain-specific meaning, held by a
// principal, such as a User or ServiceAccount via a RoleAssignment.
type Role struct {
	// Name is the name of a Role and has domain-specific meaning.
	Name RoleName `json:"name,omitempty"`
}
