package authz

// RoleAssignment represents the assignment of a Role to a principal such as a
// User or ServiceAccount.
type RoleAssignment struct {
	// Role assigns a Role to the specified principal.
	Role Role `json:"role" bson:"role"`
	// Principal specifies the principal to whom the Role is assigned.
	Principal PrincipalReference `json:"principal" bson:"principal"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific RoleName.
	Scope string `json:"scope,omitempty" bson:"scope,omitempty"`
}

// Matches determines if this RoleAssignment matches the role and scope
// arguments.
func (r RoleAssignment) Matches(role Role, scope string) bool {
	return r.Role.Name == role.Name &&
		(r.Scope == scope || r.Scope == RoleScopeGlobal)
}
