package authx

import (
	"context"
	"encoding/json"
	"net/http"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
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
	Type RoleType `json:"type,omitempty"`
	// Name is the name of a Role and has domain-specific meaning.
	Name RoleName `json:"name,omitempty"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific RoleName.
	Scope string `json:"scope,omitempty"`
}

// RoleAssignment represents the assignment of a Role to a principal such as a
// User or ServiceAccount.
type RoleAssignment struct {
	// Role assigns a Role to the specified principal.
	Role Role `json:"role"`
	// Principal specifies the principal to whom the Role is assigned.
	Principal PrincipalReference `json:"principal"`
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
				Kind:       "RoleAssignment",
			},
			Alias: (Alias)(r),
		},
	)
}

// RoleAssignmentsClient is the specialized client for managing RoleAssignments
// with the Brigade API.
type RoleAssignmentsClient interface {
	// Grant grants the system-level Role specified by the RoleAssignment to the
	// principal also specified by the RoleAssignment.
	Grant(context.Context, RoleAssignment) error
	// Revoke revokes the system-level Role specified by the RoleAssignment for
	// principal also specified by the RoleAssignment.
	Revoke(context.Context, RoleAssignment) error
}

type roleAssignmentsClient struct {
	*rm.BaseClient
}

// NewRoleAssignmentsClient returns a specialized client for managing
// RoleAssignments.
func NewRoleAssignmentsClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) RoleAssignmentsClient {
	return &roleAssignmentsClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (r *roleAssignmentsClient) Grant(
	ctx context.Context,
	roleAssignment RoleAssignment,
) error {
	return r.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/role-assignments",
			AuthHeaders: r.BearerTokenAuthHeaders(),
			ReqBodyObj:  roleAssignment,
			SuccessCode: http.StatusOK,
		},
	)
}

func (r *roleAssignmentsClient) Revoke(
	ctx context.Context,
	roleAssignment RoleAssignment,
) error {
	queryParams := map[string]string{
		"roleType":      string(roleAssignment.Role.Type),
		"roleName":      string(roleAssignment.Role.Name),
		"principalType": string(roleAssignment.Principal.Type),
		"principalID":   roleAssignment.Principal.ID,
	}
	if roleAssignment.Role.Scope != "" {
		queryParams["roleScope"] = string(roleAssignment.Role.Scope)
	}
	return r.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/role-assignments",
			AuthHeaders: r.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}
