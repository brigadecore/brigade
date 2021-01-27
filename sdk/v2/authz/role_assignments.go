package authz

import (
	"context"
	"encoding/json"
	"net/http"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// PrincipalType is a type whose values can be used to disambiguate one type of
// principal from another. For instance, when assigning a Role to a principal
// via a RoleAssignment, a PrincipalType field is used to indicate whether the
// value of the PrincipalID field reflects a User ID or a ServiceAccount ID.
type PrincipalType string

const (
	// PrincipalTypeServiceAccount represents a principal that is a
	// ServiceAccount.
	PrincipalTypeServiceAccount PrincipalType = "SERVICE_ACCOUNT"
	// PrincipalTypeUser represents a principal that is a User.
	PrincipalTypeUser PrincipalType = "USER"
)

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

// RoleAssignment represents the assignment of a Role to a principal such as a
// User or ServiceAccount.
type RoleAssignment struct {
	// Role assigns a Role to the specified principal.
	Role libAuthz.Role `json:"role"`
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
	// the principal also specified by the RoleAssignment.
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
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}
