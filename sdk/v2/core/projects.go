package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/meta"
)

// ProjectsSelector represents useful filter criteria when selecting multiple
// Projects for API group operations like list. It currently has no fields, but
// exists to preserve the possibility of future expansion without having to
// change client function signatures.
type ProjectsSelector struct{}

// ProjectList is an ordered and pageable list of ProjectS.
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
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty"`
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
				Kind:       "Project",
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectSpec is the technical component of a Project. It pairs
// EventSubscriptions with a prototypical WorkerSpec that is used as a template
// for creating new Workers.
type ProjectSpec struct {
	// EventSubscription defines a set of trigger conditions under which a new
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
	// Source specifies the origin of an Event (e.g. a gateway).
	Source string `json:"source,omitempty"`
	// Types enumerates specific Events of interest from the specified source.
	// This is useful in narrowing a subscription when a source also emits many
	// events that are NOT of interest.
	Types []string `json:"types,omitempty"`
	// Labels enumerates specific key/value pairs with which Events of interest
	// must be labeled. An event must have ALL of these labels to match this
	// subscription.
	Labels Labels `json:"labels,omitempty"`
}

// KubernetesConfig represents Kubernetes-specific configuration. This is used
// primarily at the Project level, but is also denormalized onto Events so that
// Event handling doesn't required a Project lookup to obtain
// Kubernetes-specific configuration.
type KubernetesConfig struct {
	// Namespace is the dedicated Kubernetes namespace for the Project. This is
	// NOT specified by clients when creating a new Project. The namespace is
	// created by / assigned by the system. This detail is a necessity to prevent
	// clients from naming existing namespaces in an attempt to hijack them.
	Namespace string `json:"namespace,omitempty"`
}

// ProjectsClient is the specialized client for managing Projects with the
// Brigade API.
type ProjectsClient interface {
	// Create creates a new Project.
	Create(context.Context, Project) (Project, error)
	// CreateFromBytes creates a new Project using raw (unprocessed by the client)
	// bytes, presumably originating from a file. This is the preferred way to
	// create Projects defined by an end user since server-side validation will
	// then be applied directly to the Project definition as the user has written
	// it (i.e. WITHOUT any normalization or corrections the client may have made
	// when unmarshaling the original data or when marshaling the outbound
	// request).
	CreateFromBytes(context.Context, []byte) (Project, error)
	// List returns a ProjectList, with its Items (Projects) ordered
	// alphabetically by Project ID.
	List(context.Context, ProjectsSelector, meta.ListOptions) (ProjectList, error)
	// Get retrieves a single Project specified by its identifier.
	Get(context.Context, string) (Project, error)
	// Update updates an existing Project.
	Update(context.Context, Project) (Project, error)
	// UpdateFromBytes updates an existing Project using raw (unprocessed by the
	// client) bytes, presumably originating from a file. This is the preferred
	// way to update Projects defined by an end user since server-side validation
	// will then be applied directly to the Project definition as the user has
	// written it (i.e. WITHOUT any normalization or corrections the client may
	// have made when unmarshaling the original data or when marshaling the
	// outbound request).
	UpdateFromBytes(context.Context, string, []byte) (Project, error)
	// Delete deletes a single Project specified by its identifier.
	Delete(context.Context, string) error

	// Roles returns a specialized client for Project Role management.
	Roles() ProjectRolesClient

	// Secrets returns a specialized client for Secret management.
	Secrets() SecretsClient
}

type projectsClient struct {
	*restmachinery.BaseClient
	// rolesClient is a specialized client for Project Role management.
	rolesClient ProjectRolesClient
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
		BaseClient:    restmachinery.NewBaseClient(apiAddress, apiToken, opts),
		rolesClient:   NewProjectRolesClient(apiAddress, apiToken, opts),
		secretsClient: NewSecretsClient(apiAddress, apiToken, opts),
	}
}

func (p *projectsClient) Create(
	ctx context.Context,
	project Project,
) (Project, error) {
	createdProject := Project{}
	return createdProject, p.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/projects",
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  project,
			SuccessCode: http.StatusCreated,
			RespObj:     &createdProject,
		},
	)
}

func (p *projectsClient) CreateFromBytes(
	ctx context.Context,
	projectBytes []byte,
) (Project, error) {
	createdProject := Project{}
	return createdProject, p.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/projects",
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  projectBytes,
			SuccessCode: http.StatusCreated,
			RespObj:     &createdProject,
		},
	)
}

func (p *projectsClient) List(
	ctx context.Context,
	_ ProjectsSelector,
	opts meta.ListOptions,
) (ProjectList, error) {
	projects := ProjectList{}
	return projects, p.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/projects",
			AuthHeaders: p.BearerTokenAuthHeaders(),
			QueryParams: p.AppendListQueryParams(nil, opts),
			SuccessCode: http.StatusOK,
			RespObj:     &projects,
		},
	)
}

func (p *projectsClient) Get(
	ctx context.Context,
	id string,
) (Project, error) {
	project := Project{}
	return project, p.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/projects/%s", id),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &project,
		},
	)
}

func (p *projectsClient) Update(
	ctx context.Context,
	project Project,
) (Project, error) {
	updatedProject := Project{}
	return updatedProject, p.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/projects/%s", project.ID),
			AuthHeaders: p.BearerTokenAuthHeaders(),
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
) (Project, error) {
	updatedProject := Project{}
	return updatedProject, p.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/projects/%s", projectID),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  projectBytes,
			SuccessCode: http.StatusOK,
			RespObj:     &updatedProject,
		},
	)
}

func (p *projectsClient) Delete(ctx context.Context, id string) error {
	return p.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/projects/%s", id),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsClient) Roles() ProjectRolesClient {
	return p.rolesClient
}

func (p *projectsClient) Secrets() SecretsClient {
	return p.secretsClient
}
