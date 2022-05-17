package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
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
				Kind:       ProjectKind,
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectUpdateOptions represents useful, optional settings for updating a
// Project. It currently has no fields, but exists to preserve the possibility
// of future expansion without having to change client function signatures.
type ProjectUpdateOptions struct {
	// CreateIfNotFound when set to true will cause a non-existing Project to be
	// created instead of updated.
	CreateIfNotFound bool
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
	Qualifiers Qualifiers `json:"qualifiers,omitempty" bson:"qualifiers,omitempty"` // nolint: lll
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
	Labels map[string]string `json:"labels,omitempty" bson:"labels,omitempty"`
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
	List(context.Context, meta.ListOptions) (meta.List[Project], error)
	// Get retrieves a single Project specified by its identifier. If the
	// specified Project does not exist, implementations MUST return a
	// *meta.ErrNotFound error.
	Get(context.Context, string) (Project, error)
	// Update updates an existing Project. If the specified Project does not
	// exist, implementations MUST return a *meta.ErrNotFound error.
	// Implementations may assume the Project passed to this function has been
	// pre-validated.
	Update(context.Context, Project, ProjectUpdateOptions) error
	// Delete deletes a single Project specified by its identifier. If the
	// specified Project does not exist, implementations MUST return a
	// *meta.ErrNotFound error.
	Delete(context.Context, string) error
}

type projectsService struct {
	authorize                   AuthorizeFn
	projectAuthorize            ProjectAuthorizeFn
	projectsStore               ProjectsStore
	eventsStore                 EventsStore
	logsStore                   CoolLogsStore
	projectRoleAssignmentsStore ProjectRoleAssignmentsStore
	substrate                   Substrate
}

// NewProjectsService returns a specialized interface for managing Projects.
func NewProjectsService(
	authorizeFn AuthorizeFn,
	projectAuthorize ProjectAuthorizeFn,
	projectsStore ProjectsStore,
	eventsStore EventsStore,
	logsStore CoolLogsStore,
	projectRoleAssignmentsStore ProjectRoleAssignmentsStore,
	substrate Substrate,
) ProjectsService {
	return &projectsService{
		authorize:                   authorizeFn,
		projectAuthorize:            projectAuthorize,
		projectsStore:               projectsStore,
		eventsStore:                 eventsStore,
		logsStore:                   logsStore,
		projectRoleAssignmentsStore: projectRoleAssignmentsStore,
		substrate:                   substrate,
	}
}

func (p *projectsService) Create(
	ctx context.Context,
	project Project,
) (Project, error) {
	if err := p.authorize(ctx, RoleProjectCreator, ""); err != nil {
		return project, err
	}

	now := time.Now().UTC()
	project.Created = &now

	_, err := p.projectsStore.Get(ctx, project.ID)
	if err != nil {
		if _, ok := err.(*meta.ErrNotFound); !ok {
			return project, errors.Wrapf(
				err,
				"error checking for project %q existence",
				project.ID,
			)
		}
	} else {
		return project, &meta.ErrConflict{
			Type: ProjectKind,
			ID:   project.ID,
			Reason: fmt.Sprintf(
				"Project %q already exists.",
				project.ID,
			),
		}
	}

	// Add substrate-specific details BEFORE we persist.
	project, err = p.substrate.CreateProject(ctx, project)
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
	principal := PrincipalFromContext(ctx)

	var principalRef PrincipalReference
	switch prin := principal.(type) {
	case *User:
		principalRef = PrincipalReference{
			Type: PrincipalTypeUser,
			ID:   prin.ID,
		}
	case *ServiceAccount:
		principalRef = PrincipalReference{
			Type: PrincipalTypeServiceAccount,
			ID:   prin.ID,
		}
	default:
		return project, nil
	}

	if err = p.projectRoleAssignmentsStore.Grant(
		ctx,
		ProjectRoleAssignment{
			ProjectID: project.ID,
			Role:      RoleProjectAdmin,
			Principal: principalRef,
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
	if err = p.projectRoleAssignmentsStore.Grant(
		ctx,
		ProjectRoleAssignment{
			ProjectID: project.ID,
			Role:      RoleProjectDeveloper,
			Principal: principalRef,
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
	if err = p.projectRoleAssignmentsStore.Grant(
		ctx,
		ProjectRoleAssignment{
			ProjectID: project.ID,
			Role:      RoleProjectUser,
			Principal: principalRef,
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
) (meta.List[Project], error) {
	if err := p.authorize(ctx, RoleReader, ""); err != nil {
		return meta.List[Project]{}, err
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
	if err := p.authorize(ctx, RoleReader, ""); err != nil {
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

func (p *projectsService) Update(
	ctx context.Context,
	project Project,
	opts ProjectUpdateOptions,
) error {
	if err := p.authorize(ctx, RoleReader, ""); err != nil {
		return err
	}

	if _, err := p.projectsStore.Get(ctx, project.ID); err != nil {
		_, isErrNotFound := errors.Cause(err).(*meta.ErrNotFound)
		if !isErrNotFound || !opts.CreateIfNotFound {
			return errors.Wrapf(err, "error retrieving project %q from store",
				project.ID)
		}

		_, err = p.Create(ctx, project)
		return err
	}

	if err :=
		p.projectAuthorize(ctx, project.ID, RoleProjectDeveloper); err != nil {
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
	if err := p.authorize(ctx, RoleReader, ""); err != nil {
		return err
	}

	project, err := p.projectsStore.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving project %q from store", id)
	}

	if err := p.projectAuthorize(ctx, id, RoleProjectAdmin); err != nil {
		return err
	}

	// Delete all events associated with this project
	if err := p.eventsStore.DeleteByProjectID(ctx, id); err != nil {
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
	if err :=
		p.projectRoleAssignmentsStore.RevokeByProjectID(ctx, id); err != nil {
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
	) (meta.List[Project], error)
	ListSubscribers(
		ctx context.Context,
		event Event,
	) (meta.List[Project], error)
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
