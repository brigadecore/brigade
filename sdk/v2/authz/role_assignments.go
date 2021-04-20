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

const (
	// RoleAssignmentListKind represents the canonical RoleAssignmentList kind
	// string
	RoleAssignmentListKind = "RoleAssignmentList"

	// PrincipalTypeServiceAccount represents a principal that is a
	// ServiceAccount.
	PrincipalTypeServiceAccount libAuthz.PrincipalType = "SERVICE_ACCOUNT"
	// PrincipalTypeUser represents a principal that is a User.
	PrincipalTypeUser libAuthz.PrincipalType = "USER"
)

// RoleAssignmentList is an ordered and pageable list of RoleAssignments.
type RoleAssignmentList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of RoleAssignments.
	Items []libAuthz.RoleAssignment `json:"items,omitempty"`
}

// MarshalJSON amends RoleAssignmentList instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (r RoleAssignmentList) MarshalJSON() ([]byte, error) {
	type Alias RoleAssignmentList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       RoleAssignmentListKind,
			},
			Alias: (Alias)(r),
		},
	)
}

// RoleAssignmentsSelector represents useful filter criteria when selecting
// multiple RoleAssignments for API group operations like list.
type RoleAssignmentsSelector struct {
	// Principal specifies that only RoleAssignments for the specified Principal
	// should be selected.
	Principal *libAuthz.PrincipalReference
	// TODO: Document this
	Role libAuthz.Role
}

// RoleAssignmentsClient is the specialized client for managing RoleAssignments
// with the Brigade API.
type RoleAssignmentsClient interface {
	// Grant grants the system-level Role specified by the RoleAssignment to the
	// principal also specified by the RoleAssignment.
	Grant(context.Context, libAuthz.RoleAssignment) error
	// List returns a RoleAssignmentsList, with its Items (RoleAssignments)
	// ordered by principal type, principalID, role, and scope. Criteria for which
	// RoleAssignments should be retrieved can be specified using the
	// RoleAssignmentsSelector parameter.
	List(
		context.Context,
		*RoleAssignmentsSelector,
		*meta.ListOptions,
	) (RoleAssignmentList, error)
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

func (r *roleAssignmentsClient) List(
	ctx context.Context,
	selector *RoleAssignmentsSelector,
	opts *meta.ListOptions,
) (RoleAssignmentList, error) {
	queryParams := map[string]string{}
	if selector != nil {
		if selector.Principal != nil {
			queryParams["principalType"] = string(selector.Principal.Type)
			queryParams["principalID"] = string(selector.Principal.ID)
		}
		if selector.Role != "" {
			queryParams["role"] = string(selector.Role)
		}
	}
	roleAssignments := RoleAssignmentList{}
	return roleAssignments, r.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/role-assignments",
			QueryParams: r.AppendListQueryParams(queryParams, opts),
			SuccessCode: http.StatusOK,
			RespObj:     &roleAssignments,
		},
	)
}

func (r *roleAssignmentsClient) Revoke(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
) error {
	queryParams := map[string]string{
		"role":          string(roleAssignment.Role),
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
