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

const (
	// ProjectRoleAssignmentKind represents the canonical ProjectRoleAssignment
	// kind string
	ProjectRoleAssignmentKind = "ProjectRoleAssignment"

	// ProjectRoleAssignmentListKind represents the canonical
	// ProjectRoleAssignmentList kind string
	ProjectRoleAssignmentListKind = "ProjectRoleAssignmentList"
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
				Kind:       ProjectRoleAssignmentKind,
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectRoleAssignmentList is an ordered and pageable list of
// ProjectRoleAssignments.
type ProjectRoleAssignmentList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of ProjectRoleAssignments.
	Items []ProjectRoleAssignment `json:"items,omitempty"`
}

// MarshalJSON amends ProjectRoleAssignmentList instances with type metadata so
// that clients do not need to be concerned with the tedium of doing so.
func (p ProjectRoleAssignmentList) MarshalJSON() ([]byte, error) {
	type Alias ProjectRoleAssignmentList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       ProjectRoleAssignmentListKind,
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectRoleAssignmentsSelector represents useful filter criteria when
// selecting multiple ProjectRoleAssignments for API group operations like list.
type ProjectRoleAssignmentsSelector struct {
	// Principal specifies that only ProjectRoleAssignments for the specified
	// Principal should be selected.
	Principal *libAuthz.PrincipalReference
	// Role specified that only ProjectRoleAssignments for the specified Role
	// should be selected.
	Role libAuthz.Role
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
	// List returns a ProjectRoleAssignmentsList, with its Items
	// (ProjectRoleAssignments) ordered by principal type, principalID, project,
	// and role. Criteria for which ProjectRoleAssignments should be retrieved can
	// be specified using the ProjectRoleAssignmentsSelector parameter.
	List(
		ctx context.Context,
		projectID string,
		selector *ProjectRoleAssignmentsSelector,
		opts *meta.ListOptions,
	) (ProjectRoleAssignmentList, error)
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

func (p *projectRoleAssignmentsClient) List(
	ctx context.Context,
	projectID string,
	selector *ProjectRoleAssignmentsSelector,
	opts *meta.ListOptions,
) (ProjectRoleAssignmentList, error) {
	queryParams := map[string]string{}
	if selector != nil {
		if selector.Principal != nil {
			queryParams["principalType"] = string(selector.Principal.Type)
			queryParams["principalID"] = selector.Principal.ID
		}
		if selector.Role != "" {
			queryParams["role"] = string(selector.Role)
		}
	}
	projectRoleAssignments := ProjectRoleAssignmentList{}
	return projectRoleAssignments, p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/projects/%s/role-assignments", projectID),
			QueryParams: p.AppendListQueryParams(queryParams, opts),
			SuccessCode: http.StatusOK,
			RespObj:     &projectRoleAssignments,
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
