package authz

import (
	"context"
	"net/http"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

const (
	// PrincipalTypeServiceAccount represents a principal that is a
	// ServiceAccount.
	PrincipalTypeServiceAccount libAuthz.PrincipalType = "SERVICE_ACCOUNT"
	// PrincipalTypeUser represents a principal that is a User.
	PrincipalTypeUser libAuthz.PrincipalType = "USER"
)

// RoleAssignmentsClient is the specialized client for managing RoleAssignments
// with the Brigade API.
type RoleAssignmentsClient interface {
	// Grant grants the system-level Role specified by the RoleAssignment to the
	// principal also specified by the RoleAssignment.
	Grant(context.Context, libAuthz.RoleAssignment) error
	// Revoke revokes the system-level Role specified by the RoleAssignment for
	// the principal also specified by the RoleAssignment.
	Revoke(context.Context, libAuthz.RoleAssignment) error
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
	roleAssignment libAuthz.RoleAssignment,
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
	roleAssignment libAuthz.RoleAssignment,
) error {
	queryParams := map[string]string{
		"roleType":      string(roleAssignment.Role.Type),
		"role":          string(roleAssignment.Role.Name),
		"principalType": string(roleAssignment.Principal.Type),
		"principalID":   roleAssignment.Principal.ID,
	}
	if roleAssignment.Scope != "" {
		queryParams["scope"] = string(roleAssignment.Scope)
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
