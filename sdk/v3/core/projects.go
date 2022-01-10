package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	rm "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	"github.com/brigadecore/brigade/sdk/v3/restmachinery"
)

// ProjectKind represents the canonical Project kind string
const ProjectKind = "Project"

// Project is Brigade's fundamental configuration, management, and isolation
// construct.
// - Configuration: Users define Projects to pair EventSubscriptions with
//   template WorkerSpecs.
// - Management: Project administrators govern Project access by granting and
//   revoking project-level Roles to/from principals (such as Users or
//   ServiceAccounts)
// - Isolation: All workloads (Workers and Jobs) spawned to handle a given
//   Project's Events are isolated from other Projects' workloads in the
//   underlying workload execution substrate.
type Project struct {
	// ObjectMeta contains Project metadata.
	meta.ObjectMeta `json:"metadata"`
	// Description is a natural language description of the Project.
	Description string `json:"description,omitempty"`
	// Spec is an instance of a ProjectSpec that pairs EventSubscriptions with
	// a WorkerTemplate.
	Spec ProjectSpec `json:"spec"`
	// Kubernetes contains Kubernetes-specific details of the Project's
	// environment. These details are populated by Brigade so that sufficiently
	// authorized Kubernetes users may obtain the information needed to directly
	// modify a Project's environment to facilitate certain advanced use cases.
	// Clients MUST leave the value of this field nil when using the API to create
	// or update a Project.
	Kubernetes *KubernetesDetails `json:"kubernetes,omitempty"`
}

// MarshalJSON amends Project instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (p Project) MarshalJSON() ([]byte, error) {
	type Alias Project
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       ProjectKind,
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectList is an ordered and pageable list of Projects.
type ProjectList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Projects.
	Items []Project `json:"items,omitempty"`
}

// MarshalJSON amends ProjectList instances with type metadata so that clients
// do not need to be concerned with the tedium of doing so.
func (p ProjectList) MarshalJSON() ([]byte, error) {
	type Alias ProjectList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ProjectList",
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectsSelector represents useful filter criteria when selecting multiple
// Projects for API group operations like list. It currently has no fields, but
// exists to preserve the possibility of future expansion without having to
// change client function signatures.
type ProjectsSelector struct{}

// ProjectSpec is the technical component of a Project. It pairs
// EventSubscriptions with a prototypical WorkerSpec that is used as a template
// for creating new Workers.
type ProjectSpec struct {
	// EventSubscriptions defines a set of trigger conditions under which a new
	// Worker should be created.
	EventSubscriptions []EventSubscription `json:"eventSubscriptions,omitempty"`
	// WorkerTemplate is a prototypical WorkerSpec.
	WorkerTemplate WorkerSpec `json:"workerTemplate"`
}

// EventSubscription defines a set of Events of interest. ProjectSpecs utilize
// these in defining the Events that should trigger the execution of a new
// Worker. An Event matches a subscription if it meets ALL of the specified
// criteria.
type EventSubscription struct {
	// Source specifies the origin of an Event (e.g. a gateway). This is a
	// required field.
	Source string `json:"source,omitempty"`
	// Types enumerates specific Events of interest from the specified Source.
	// This is useful in narrowing a subscription when a Source also emits many
	// Event types that are NOT of interest. This is a required field. The value
	// "*" may be utilized to denote that ALL events originating from the
	// specified Source are of interest.
	Types []string `json:"types,omitempty"`
	// Qualifiers specifies an EXACT set of key/value pairs with which an Event
	// MUST also be qualified for a Project to be considered subscribed. To
	// demonstrate the usefulness of this, consider any event from a hypothetical
	// GitHub gateway. If, by design, that gateway does not intend for any Project
	// to subscribe to ALL Events (i.e. regardless of which repository they
	// originated from), then that gateway can QUALIFY Events it emits into
	// Brigade's event bus with repo=<repository name>. Projects wishing to
	// subscribe to Events from the GitHub gateway MUST include that Qualifier in
	// their EventSubscription. Note that the Qualifiers field's "MUST match"
	// subscription semantics differ from the Labels field's "MAY match"
	// subscription semantics.
	Qualifiers map[string]string `json:"qualifiers,omitempty"`
	// Labels optionally specifies filter criteria as key/value pairs with which
	// an Event MUST also be labeled for a Project to be considered subscribed. If
	// the Event has ADDITIONAL labels, not mentioned by this EventSubscription,
	// these do not preclude a match. To demonstrate the usefulness of this,
	// consider any event from a hypothetical Slack gateway. If, by design, that
	// gateway intends for Projects to select between subscribing to ALL Events or
	// ONLY events originating from a specific channel, then that gateway can
	// LABEL Events it emits into Brigade's event bus with channel=<channel name>.
	// Projects wishing to subscribe to ALL Events from the Slack gateway MAY omit
	// that Label from their EventSubscription, while Projects wishing to
	// subscribe to only Events originating from a specific channel MAY include
	// that Label in their EventSubscription. Note that the Labels field's "MAY
	// match" subscription semantics differ from the Qualifiers field's "MUST
	// match" subscription semantics.
	Labels map[string]string `json:"labels,omitempty"`
}

// KubernetesDetails represents Kubernetes-specific configuration.
type KubernetesDetails struct {
	// Namespace is the dedicated Kubernetes namespace for the Project. This is
	// NOT specified by clients when creating a new Project. The namespace is
	// created by / assigned by the system. This detail is a necessity to prevent
	// clients from naming existing namespaces in an attempt to hijack them.
	Namespace string `json:"namespace,omitempty"`
}

// ProjectCreateOptions represents useful, optional settings for creating a new
// Project. It currently has no fields, but exists to preserve the possibility
// of future expansion without having to change client function signatures.
type ProjectCreateOptions struct{}

// ProjectGetOptions represents useful, optional criteria for retrieving a
// Project. It currently has no fields, but exists to preserve the possibility
// of future expansion without having to change client function signatures.
type ProjectGetOptions struct{}

// ProjectUpdateOptions represents useful, optional settings for updating a
// Project. It currently has no fields, but exists to preserve the possibility
// of future expansion without having to change client function signatures.
type ProjectUpdateOptions struct{}

// ProjectDeleteOptions represents useful, optional settings for deleting a
// Project. It currently has no fields, but exists to preserve the possibility
// of future expansion without having to change client function signatures.
type ProjectDeleteOptions struct{}

// ProjectsClient is the specialized client for managing Projects with the
// Brigade API.
type ProjectsClient interface {
	// Create creates a new Project.
	Create(context.Context, Project, *ProjectCreateOptions) (Project, error)
	// CreateFromBytes creates a new Project using raw (unprocessed by the client)
	// bytes, presumably originating from a file. This is the preferred way to
	// create Projects defined by an end user since server-side validation will
	// then be applied directly to the Project definition as the user has written
	// it (i.e. WITHOUT any normalization or corrections the client may have made
	// when unmarshaling the original data or when marshaling the outbound
	// request).
	CreateFromBytes(
		context.Context,
		[]byte,
		*ProjectCreateOptions,
	) (Project, error)
	// List returns a ProjectList, with its Items (Projects) ordered
	// alphabetically by Project ID.
	List(
		context.Context,
		*ProjectsSelector,
		*meta.ListOptions,
	) (ProjectList, error)
	// Get retrieves a single Project specified by its identifier.
	Get(context.Context, string, *ProjectGetOptions) (Project, error)
	// Update updates an existing Project.
	Update(context.Context, Project, *ProjectUpdateOptions) (Project, error)
	// UpdateFromBytes updates an existing Project using raw (unprocessed by the
	// client) bytes, presumably originating from a file. This is the preferred
	// way to update Projects defined by an end user since server-side validation
	// will then be applied directly to the Project definition as the user has
	// written it (i.e. WITHOUT any normalization or corrections the client may
	// have made when unmarshaling the original data or when marshaling the
	// outbound request).
	UpdateFromBytes(
		context.Context,
		string,
		[]byte,
		*ProjectUpdateOptions,
	) (Project, error)
	// Delete deletes a single Project specified by its identifier.
	Delete(context.Context, string, *ProjectDeleteOptions) error

	// Authz returns a specialized client for managing project-level authorization
	// concerns.
	Authz() AuthzClient

	// Secrets returns a specialized client for Secret management.
	Secrets() SecretsClient
}

type projectsClient struct {
	*rm.BaseClient
	// authzClient is a specialized client for managing project-level
	// authorization concerns.
	authzClient AuthzClient
	// secretsClient is a specialized client for Secret management.
	secretsClient SecretsClient
}

// NewProjectsClient returns a specialized client for managing Projects.
func NewProjectsClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) ProjectsClient {
	return &projectsClient{
		BaseClient:    rm.NewBaseClient(apiAddress, apiToken, opts),
		authzClient:   NewAuthzClient(apiAddress, apiToken, opts),
		secretsClient: NewSecretsClient(apiAddress, apiToken, opts),
	}
}

func (p *projectsClient) Create(
	ctx context.Context,
	project Project,
	_ *ProjectCreateOptions,
) (Project, error) {
	createdProject := Project{}
	return createdProject, p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/projects",
			ReqBodyObj:  project,
			SuccessCode: http.StatusCreated,
			RespObj:     &createdProject,
		},
	)
}

func (p *projectsClient) CreateFromBytes(
	ctx context.Context,
	projectBytes []byte,
	_ *ProjectCreateOptions,
) (Project, error) {
	createdProject := Project{}
	return createdProject, p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/projects",
			ReqBodyObj:  projectBytes,
			SuccessCode: http.StatusCreated,
			RespObj:     &createdProject,
		},
	)
}

func (p *projectsClient) List(
	ctx context.Context,
	_ *ProjectsSelector,
	opts *meta.ListOptions,
) (ProjectList, error) {
	projects := ProjectList{}
	return projects, p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/projects",
			QueryParams: p.AppendListQueryParams(nil, opts),
			SuccessCode: http.StatusOK,
			RespObj:     &projects,
		},
	)
}

func (p *projectsClient) Get(
	ctx context.Context,
	id string,
	_ *ProjectGetOptions,
) (Project, error) {
	project := Project{}
	return project, p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/projects/%s", id),
			SuccessCode: http.StatusOK,
			RespObj:     &project,
		},
	)
}

func (p *projectsClient) Update(
	ctx context.Context,
	project Project,
	_ *ProjectUpdateOptions,
) (Project, error) {
	updatedProject := Project{}
	return updatedProject, p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/projects/%s", project.ID),
			ReqBodyObj:  project,
			SuccessCode: http.StatusOK,
			RespObj:     &updatedProject,
		},
	)
}

func (p *projectsClient) UpdateFromBytes(
	ctx context.Context,
	projectID string,
	projectBytes []byte,
	_ *ProjectUpdateOptions,
) (Project, error) {
	updatedProject := Project{}
	return updatedProject, p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/projects/%s", projectID),
			ReqBodyObj:  projectBytes,
			SuccessCode: http.StatusOK,
			RespObj:     &updatedProject,
		},
	)
}

func (p *projectsClient) Delete(
	ctx context.Context,
	id string,
	_ *ProjectDeleteOptions,
) error {
	return p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/projects/%s", id),
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsClient) Authz() AuthzClient {
	return p.authzClient
}

func (p *projectsClient) Secrets() SecretsClient {
	return p.secretsClient
}
