package authz

// RoleType is a type whose values can be used to disambiguate one type of Role
// from another.
type RoleType string

// RoleName is a type whose value maps to a well-defined Brigade Role.
type RoleName string

// RoleScopeGlobal represents an unbounded scope.
const RoleScopeGlobal = "*"

// Role represents a set of permissions.
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

// Matches determines if this Role matches the requiredRole argument. This Role
// is a match for the required one if the Type and Name fields have the same
// values in both AND if the value of this Role's Scope field is either the same
// as that of the required Role's Scope field OR is unbounded ("*").
//
// Note that order is important. A.Matches(B) does not guarantee B.Matches(A).
func (r Role) Matches(requiredRole Role) bool {
	return r.Type == requiredRole.Type &&
		r.Name == requiredRole.Name &&
		(r.Scope == requiredRole.Scope || r.Scope == RoleScopeGlobal)
}
