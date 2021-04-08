package authz

// RoleName is a type whose value maps to a well-defined Brigade Role.
type RoleName string

// RoleScopeGlobal represents an unbounded scope.
const RoleScopeGlobal = "*"

// Role represents a set of permissions.
type Role struct {
	// Name is the name of a Role and has domain-specific meaning.
	Name RoleName `json:"name,omitempty" bson:"name,omitempty"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific RoleName.
	Scope string `json:"scope,omitempty" bson:"scope,omitempty"`
}

// Matches determines if this Role matches the requiredRole argument. This Role
// is a match for the required one if the Name fields have the same values in
// both AND if the value of this Role's Scope field is either the same as that
// of the required Role's Scope field OR is unbounded ("*").
//
// Note that order is important. A.Matches(B) does not guarantee B.Matches(A).
func (r Role) Matches(requiredRole Role) bool {
	return r.Name == requiredRole.Name &&
		(r.Scope == requiredRole.Scope || r.Scope == RoleScopeGlobal)
}
