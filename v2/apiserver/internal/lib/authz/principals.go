package authz

// PrincipalType is a type whose values can be used to disambiguate one type of
// principal from another. For instance, when assigning a Role to a principal
// via a RoleAssignment, a PrincipalType field is used to indicate whether the
// value of the PrincipalID field reflects a User ID or a ServiceAccount ID.
type PrincipalType string

// PrincipalReference is a reference to any sort of security principal (human
// user, service account, etc.)
type PrincipalReference struct {
	// Type qualifies what kind of principal is referenced by the ID field-- for
	// instance, a User or a ServiceAccount.
	Type PrincipalType `json:"type,omitempty"`
	// ID references a principal. The Type qualifies what type of principal that
	// is-- for instance, a User or a ServiceAccount.
	ID string `json:"id,omitempty"`
}
