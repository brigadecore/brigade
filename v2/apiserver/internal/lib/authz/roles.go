package authz

// RoleName is a type whose value maps to a well-defined Brigade Role.
type RoleName string

// RoleScopeGlobal represents an unbounded scope.
const RoleScopeGlobal = "*"

// Role represents a set of permissions.
type Role struct {
	// Name is the name of a Role and has domain-specific meaning.
	Name RoleName `json:"name,omitempty" bson:"name,omitempty"`
}
