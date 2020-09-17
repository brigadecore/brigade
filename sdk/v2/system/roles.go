package system

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/brigadecore/brigade/sdk/v2/authx"
	"github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
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
	*restmachinery.BaseClient
}

// NewRolesClient returns a specialized client for managing System
// Roles.
func NewRolesClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) RolesClient {
	return &rolesClient{
		BaseClient: &restmachinery.BaseClient{
			APIAddress: apiAddress,
			APIToken:   apiToken,
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: allowInsecure,
					},
				},
			},
		},
	}
}

func (r *rolesClient) Grant(
	ctx context.Context,
	roleAssignment authx.RoleAssignment,
) error {
	return r.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
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
		restmachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/system/role-assignments",
			AuthHeaders: r.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}
