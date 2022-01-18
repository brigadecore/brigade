package authz

import (
	"encoding/json"

	"github.com/brigadecore/brigade/sdk/v3/meta"
)

// RoleAssignmentKind represents the canonical RoleAssignment kind string
const RoleAssignmentKind = "RoleAssignment"

// RoleAssignment represents the assignment of a Role to a principal such as a
// User or ServiceAccount.
type RoleAssignment struct {
	// Role assigns a Role to the specified principal.
	Role Role `json:"role"`
	// Principal specifies the principal to whom the Role is assigned.
	Principal PrincipalReference `json:"principal"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific Role.
	Scope string `json:"scope,omitempty"`
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
				Kind:       RoleAssignmentKind,
			},
			Alias: (Alias)(r),
		},
	)
}
