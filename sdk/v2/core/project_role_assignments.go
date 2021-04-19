package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// ProjectRoleAssignment represents the assignment of a ProjectRole to a
// principal such as a User or ServiceAccount.
type ProjectRoleAssignment struct {
	// Role assigns a Role to the specified principal.
	Role libAuthz.Role `json:"role"`
	// Principal specifies the principal to whom the Role is assigned.
	Principal libAuthz.PrincipalReference `json:"principal"`
}

// MarshalJSON amends ProjectRoleAssignment instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (p ProjectRoleAssignment) MarshalJSON() ([]byte, error) {
	type Alias ProjectRoleAssignment
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ProjectRoleAssignment",
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectRoleAssignmentsClient is the specialized client for managing
// ProjectRoleAssignments with the Brigade API.
type ProjectRoleAssignmentsClient interface {
	// Grant grants the project-level Role specified by the ProjectRoleAssignment
	// to the principal also specified by the ProjectRoleAssignment.
	Grant(
		ctx context.Context,
		projectID string,
		projectRoleAssignment ProjectRoleAssignment,
	) error
	// Revoke revokes the project-level Role specified by the
	// ProjectRoleAssignment for the principal also specified by the
	// ProjectRoleAssignment.
	Revoke(
		ctx context.Context,
		projectID string,
		projectRoleAssignment ProjectRoleAssignment,
	) error
}

type projectRoleAssignmentsClient struct {
	*rm.BaseClient
}

// NewProjectRoleAssignmentsClient returns a specialized client for managing
// project-level RoleAssignments.
func NewProjectRoleAssignmentsClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) ProjectRoleAssignmentsClient {
	return &projectRoleAssignmentsClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (p *projectRoleAssignmentsClient) Grant(
	ctx context.Context,
	projectID string,
	projectRoleAssignment ProjectRoleAssignment,
) error {
	return p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        fmt.Sprintf("v2/projects/%s/role-assignments", projectID),
			ReqBodyObj:  projectRoleAssignment,
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectRoleAssignmentsClient) Revoke(
	ctx context.Context,
	projectID string,
	projectRoleAssignment ProjectRoleAssignment,
) error {
	queryParams := map[string]string{
		"role":          string(projectRoleAssignment.Role),
		"principalType": string(projectRoleAssignment.Principal.Type),
		"principalID":   projectRoleAssignment.Principal.ID,
	}
	return p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/projects/%s/role-assignments", projectID),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}
