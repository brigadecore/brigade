package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/authz"
	libAuthn "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authn"
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/brigadecore/brigade/v2/apiserver/internal/system"
	"github.com/pkg/errors"
)

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
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	// Description is a natural language description of the Project.
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	// Spec is an instance of a ProjectSpec that pairs EventSubscriptions with
	// a WorkerTemplate.
	Spec ProjectSpec `json:"spec" bson:"spec"`
	// Kubernetes contains Kubernetes-specific details of the Project's
	// environment. These details are populated by Brigade so that sufficiently
	// authorized Kubernetes users may obtain the information needed to directly
	// modify a Project's environment to facilitate certain advanced use cases.
	Kubernetes *KubernetesDetails `json:"kubernetes,omitempty" bson:"kubernetes,omitempty"` // nolint: lll
}

// MarshalJSON amends Project instances with type metadata.
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

// ProjectList is an ordered and pageable list of Projects.
type ProjectList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Projects.
	Items []Project `json:"items"`
}

// MarshalJSON amends ProjectList instances with type metadata.
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

// ProjectSpec is the technical component of a Project. It pairs
// EventSubscriptions with a prototypical WorkerSpec that is used as a template
// for creating new Workers.
type ProjectSpec struct {
	// EventSubscription defines a set of trigger conditions under which a new
	// Worker should be created.
	EventSubscriptions []EventSubscription `json:"eventSubscriptions,omitempty" bson:"eventSubscriptions,omitempty"` // nolint: lll
	// WorkerTemplate is a prototypical WorkerSpec.
	WorkerTemplate WorkerSpec `json:"workerTemplate" bson:"workerTemplate"`
}

// EventSubscription defines a set of Events of interest. ProjectSpecs utilize
// these in defining the Events that should trigger the execution of a new
// Worker. An Event matches a subscription if it meets ALL of the specified
// criteria.
type EventSubscription struct {
	// Source specifies the origin of an Event (e.g. a gateway). This is a
	// required field.
	Source string `json:"source,omitempty" bson:"source,omitempty"`
	// Types enumerates specific Events of interest from the specified Source.
	// This is useful in narrowing a subscription when a Source also emits many
	// Event types that are NOT of interest. This is a required field. The value
	// "*" may be utilized to denote that ALL events originating from the
	// specified Source are of interest.
	Types []string `json:"types,omitempty" bson:"types,omitempty"`
	// Labels defines an EXACT set of key/value pairs with which Events of
	// interest must be labeled. An Event having additional labels not included in
	// the subscription does NOT match that subscription. Likewise a subscription
	// having additional labels not included in the Event does NOT match that
	// Event. This strict requirement prevents accidental subscriptions. For
	// instance, consider an Event gateway brokering events from GitHub. If Events
	// (per that gateway's own documentation) were labeled `repo=<repository
	// name>`, no would-be subscriber to Events from that gateway will succeed
	// unless their subscription includes that matching label. This effectively
	// PREVENTS a scenario where a subscriber who has forgotten to apply
	// applicable labels accidentally subscribes to ALL events from the GitHub
	// gateway, regardless of the repository of origin.
	Labels Labels `json:"labels,omitempty" bson:"labels,omitempty"`
}

// KubernetesDetails represents Kubernetes-specific configuration.
type KubernetesDetails struct {
	// Namespace is the dedicated Kubernetes namespace for the Project. This is
	// NOT specified by clients when creating a new Project. The namespace is
	// created by / assigned by the system. This detail is a necessity to prevent
	// clients from naming existing namespaces in an attempt to hijack them.
	Namespace string `json:"namespace,omitempty" bson:"namespace,omitempty"`
}

// ProjectsService is the specialized interface for managing Projects. It's
// decoupled from underlying technology choices (e.g. data store, message bus,
// etc.) to keep business logic reusable and consistent while the underlying
// tech stack remains free to change.
type ProjectsService interface {
	// Create creates a new Project. If a Project with the specified identifier
	// already exists, implementations MUST return a *meta.ErrConflict error.
	// Implementations may assume the Project passed to this function has been
	// pre-validated.
	Create(context.Context, Project) (Project, error)
	// List returns a ProjectList, with its Items (Projects) ordered
	// alphabetically by Project ID.
	List(context.Context, meta.ListOptions) (ProjectList, error)
	// Get retrieves a single Project specified by its identifier. If the
	// specified Project does not exist, implementations MUST return a
	// *meta.ErrNotFound error.
	Get(context.Context, string) (Project, error)
	// Update updates an existing Project. If the specified Project does not
	// exist, implementations MUST return a *meta.ErrNotFound error.
	// Implementations may assume the Project passed to this function has been
	// pre-validated.
	Update(context.Context, Project) error
	// Delete deletes a single Project specified by its identifier. If the
	// specified Project does not exist, implementations MUST return a
	// *meta.ErrNotFound error.
	Delete(context.Context, string) error
}

type projectsService struct {
	authorize            libAuthz.AuthorizeFn
	projectsStore        ProjectsStore
	eventsStore          EventsStore
	logsStore            CoolLogsStore
	roleAssignmentsStore authz.RoleAssignmentsStore
	substrate            Substrate
}

// NewProjectsService returns a specialized interface for managing Projects.
func NewProjectsService(
	authorizeFn libAuthz.AuthorizeFn,
	projectsStore ProjectsStore,
	eventsStore EventsStore,
	logsStore CoolLogsStore,
	roleAssignmentsStore authz.RoleAssignmentsStore,
	substrate Substrate,
) ProjectsService {
	return &projectsService{
		authorize:            authorizeFn,
		projectsStore:        projectsStore,
		eventsStore:          eventsStore,
		logsStore:            logsStore,
		roleAssignmentsStore: roleAssignmentsStore,
		substrate:            substrate,
	}
}

func (p *projectsService) Create(
	ctx context.Context,
	project Project,
) (Project, error) {
	if err := p.authorize(ctx, RoleProjectCreator()); err != nil {
		return project, err
	}

	now := time.Now().UTC()
	project.Created = &now

	// Add substrate-specific details BEFORE we persist.
	project, err := p.substrate.CreateProject(ctx, project)
	if err != nil {
		return project, errors.Wrapf(
			err,
			"error creating project %q on the substrate",
			project.ID,
		)
	}

	if err = p.projectsStore.Create(ctx, project); err != nil {
		return project,
			errors.Wrapf(err, "error storing new project %q", project.ID)
	}

	// Make the current user an admin, developer, and user of the project
	principal := libAuthn.PrincipalFromContext(ctx)

	var principalRef authz.PrincipalReference
	switch prin := principal.(type) {
	case *authn.User:
		principalRef = authz.PrincipalReference{
			Type: authz.PrincipalTypeUser,
			ID:   prin.ID,
		}
	case *authn.ServiceAccount:
		principalRef = authz.PrincipalReference{
			Type: authz.PrincipalTypeServiceAccount,
			ID:   prin.ID,
		}
	default:
		return project, nil
	}

	if err = p.roleAssignmentsStore.Grant(
		ctx,
		authz.RoleAssignment{
			Principal: principalRef,
			Role:      RoleProjectAdmin(project.ID),
		},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error making %s %q ADMIN of new project %q",
			principalRef.Type,
			principalRef.ID,
			project.ID,
		)
	}
	if err = p.roleAssignmentsStore.Grant(
		ctx,
		authz.RoleAssignment{
			Principal: principalRef,
			Role:      RoleProjectDeveloper(project.ID),
		},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error making %s %q DEVELOPER of new project %q",
			principalRef.Type,
			principalRef.ID,
			project.ID,
		)
	}
	if err = p.roleAssignmentsStore.Grant(
		ctx,
		authz.RoleAssignment{
			Principal: principalRef,
			Role:      RoleProjectUser(project.ID),
		},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error making %s %q USER of new project %q",
			principalRef.Type,
			principalRef.ID,
			project.ID,
		)
	}

	return project, nil
}

func (p *projectsService) List(
	ctx context.Context,
	opts meta.ListOptions,
) (ProjectList, error) {
	if err := p.authorize(ctx, system.RoleReader()); err != nil {
		return ProjectList{}, err
	}

	if opts.Limit == 0 {
		opts.Limit = 20
	}
	projects, err := p.projectsStore.List(ctx, opts)
	if err != nil {
		return projects, errors.Wrap(err, "error retrieving projects from store")
	}
	return projects, nil
}

func (p *projectsService) Get(
	ctx context.Context,
	id string,
) (Project, error) {
	if err := p.authorize(ctx, system.RoleReader()); err != nil {
		return Project{}, err
	}

	project, err := p.projectsStore.Get(ctx, id)
	if err != nil {
		return project, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			id,
		)
	}
	return project, nil
}

func (p *projectsService) Update(ctx context.Context, project Project) error {
	if err :=
		p.authorize(ctx, RoleProjectDeveloper(project.ID)); err != nil {
		return err
	}

	if err := p.projectsStore.Update(ctx, project); err != nil {
		return errors.Wrapf(
			err,
			"error updating project %q in store",
			project.ID,
		)
	}
	return nil
}

func (p *projectsService) Delete(ctx context.Context, id string) error {
	if err := p.authorize(ctx, RoleProjectAdmin(id)); err != nil {
		return err
	}

	project, err := p.projectsStore.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving project %q from store", id)
	}

	// Delete all events associated with this project
	if _, _, err := p.eventsStore.DeleteMany(
		ctx,
		EventsSelector{ProjectID: id, WorkerPhases: WorkerPhasesAll()},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting all events associated with project %q",
			id,
		)
	}

	// Delete all logs associated with this project
	if err := p.logsStore.DeleteProjectLogs(ctx, id); err != nil {
		return errors.Wrapf(
			err,
			"error deleting all logs associated with project %q",
			id,
		)
	}

	// Delete all role assignments associated with this project. If we didn't do
	// this and someone, in the future, created a new project with the same name,
	// that new project would begin life with some existing principals having
	// permissions they ought not have.
	if err := p.roleAssignmentsStore.RevokeMany(
		ctx,
		authz.RoleAssignment{
			Role: libAuthz.Role{
				Type:  RoleTypeProject,
				Scope: id,
			},
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error revoking all role assignments associated with project %q",
			id,
		)
	}

	// Delete the project itself
	if err := p.projectsStore.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error removing project %q from store", id)
	}
	if err := p.substrate.DeleteProject(ctx, project); err != nil {
		return errors.Wrapf(
			err,
			"error deleting project %q from substrate",
			id,
		)
	}
	return nil
}

// ProjectsStore is an interface for components that implement Project
// persistence concerns.
type ProjectsStore interface {
	// Create stores the provided Project. Implementations MUST return a
	// *meta.ErrConflict error if a Project having the indicated identifier
	// already exists.
	Create(context.Context, Project) error
	// List returns a ProjectList, with its Items (Projects) ordered
	// alphabetically by Project ID.
	List(
		context.Context,
		meta.ListOptions,
	) (ProjectList, error)
	ListSubscribers(
		ctx context.Context,
		event Event,
	) (ProjectList, error)
	// Get returns a Project having the indicated ID. If no such Project exists,
	// implementations MUST return a *meta.ErrNotFound error.
	Get(context.Context, string) (Project, error)
	// Update updates the provided Project in storage, presuming that Project
	// already exists in storage. If no Project having the indicated ID already
	// exists, implementations MUST return a *meta.ErrNotFound error.
	// Implementations MUST apply updates ONLY to the Description and Spec fields,
	// as only these fields are intended to be mutable. Implementations MUST
	// ignore changes to all other fields when updating.
	Update(context.Context, Project) error
	// Delete deletes the specified Project. If no Project having the given
	// identifier is found, implementations MUST return a *meta.ErrNotFound error.
	Delete(context.Context, string) error
}
