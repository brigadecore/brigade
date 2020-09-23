package system

import (
	"context"
	"net/http"

	"github.com/brigadecore/brigade/sdk/v2/authx"
	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// RoleTypeSystem represents a system-level Role.
const RoleTypeSystem authx.RoleType = "SYSTEM"

const (
	// RoleNameAdmin is the name of a system-level Role that enables principals to
	// manage Users, ServiceAccounts, and system-level permissions for Users and
	// ServiceAccounts.
	RoleNameAdmin authx.RoleName = "ADMIN"
	// RoleNameEventCreator is the name of a system-level Role that enables
	// principals to create Events for all Projects.
	RoleNameEventCreator authx.RoleName = "EVENT_CREATOR"
	// RoleNameProjectCreator is the name of a system-level Role that enables
	// principals to create new Projects.
	RoleNameProjectCreator authx.RoleName = "PROJECT_CREATOR"
	// RoleNameReader is the name of a system-level Role that enables global read
	// access.
	RoleNameReader authx.RoleName = "READER"
)

// RolesClient is the specialized client for managing system-level
// RoleAssignments with the Brigade API.
type RolesClient interface {

	// TODO: This needs a function for listing available system roles

	// TODO: This needs a function for listing system role assignments

	// Grant grants the system-level Role specified by the RoleAssignment to the
	// principal also specified by the RoleAssignment.
	Grant(context.Context, authx.RoleAssignment) error

	// Revoke revokes the system-level Role specified by the RoleAssignment for
	// principal also specified by the RoleAssignment.
	Revoke(context.Context, authx.RoleAssignment) error
}

type rolesClient struct {
	*rm.BaseClient
}

// NewRolesClient returns a specialized client for managing System
// Roles.
func NewRolesClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) RolesClient {
	return &rolesClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (r *rolesClient) Grant(
	ctx context.Context,
	roleAssignment authx.RoleAssignment,
) error {
	return r.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/system/role-assignments",
			AuthHeaders: r.BearerTokenAuthHeaders(),
			ReqBodyObj:  roleAssignment,
			SuccessCode: http.StatusOK,
		},
	)
}

func (r *rolesClient) Revoke(
	ctx context.Context,
	roleAssignment authx.RoleAssignment,
) error {
	queryParams := map[string]string{
		"role":          string(roleAssignment.Role),
		"principalType": string(roleAssignment.PrincipalType),
		"principalID":   roleAssignment.PrincipalID,
	}
	return r.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/system/role-assignments",
			AuthHeaders: r.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}
