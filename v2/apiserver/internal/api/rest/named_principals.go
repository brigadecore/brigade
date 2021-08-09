package rest

import "github.com/brigadecore/brigade/v2/apiserver/internal/api"

var (
	// Root is a singleton that represents Brigade's "root" user.
	root = &rootPrincipal{}
	// Scheduler is a singleton that represents Brigade's scheduler component.
	scheduler = &schedulerPrincipal{}
	// Observer is a singleton that represents Brigade's observer component.
	observer = &observerPrincipal{}
)

// rootPrincipal is an implementation of the Principal interface for the "root"
// user.
type rootPrincipal struct{}

func (r *rootPrincipal) RoleAssignments() []api.RoleAssignment {
	return []api.RoleAssignment{
		{Role: api.RoleAdmin},
		{Role: api.RoleReader},
		{
			Role:  api.RoleEventCreator,
			Scope: api.RoleScopeGlobal,
		},
		{Role: api.RoleProjectCreator},
	}
}

func (r *rootPrincipal) ProjectRoleAssignments() []api.ProjectRoleAssignment {
	return []api.ProjectRoleAssignment{
		{
			ProjectID: api.ProjectRoleScopeGlobal,
			Role:      api.RoleProjectAdmin,
		},
		{
			ProjectID: api.ProjectRoleScopeGlobal,
			Role:      api.RoleProjectDeveloper,
		},
		{
			ProjectID: api.ProjectRoleScopeGlobal,
			Role:      api.RoleProjectUser,
		},
	}
}

// schedulerPrincipal is an implementation of the Principal interface that
// represents the scheduler component, which is a special class of user because,
// although it cannot do much, it has the UNIQUE ability to launch Workers and
// Jobs.
type schedulerPrincipal struct{}

func (s *schedulerPrincipal) RoleAssignments() []api.RoleAssignment {
	return []api.RoleAssignment{
		{Role: api.RoleReader},
		{Role: api.RoleScheduler},
	}
}

// observerPrincipal is an implementation of the Principal interface that
// represents the observer component, which is a special class of user because,
// although it cannot do much, it has the UNIQUE ability to update Worker and
// Job statuses.
type observerPrincipal struct{}

func (o *observerPrincipal) RoleAssignments() []api.RoleAssignment {
	return []api.RoleAssignment{
		{Role: api.RoleReader},
		{Role: api.RoleObserver},
	}
}

// workerPrincipal is an implementation of the Principal interface that
// represents an Event's Worker, which is a special class of user because,
// although it cannot do much, it has the UNIQUE ability to create new Jobs.
type workerPrincipal struct {
	eventID string
}

func (w *workerPrincipal) RoleAssignments() []api.RoleAssignment {
	return []api.RoleAssignment{
		{Role: api.RoleReader},
		{
			Role:  api.RoleWorker,
			Scope: w.eventID,
		},
	}
}

// worker returns an Principal that represents the specified Event's Worker.
func worker(eventID string) *workerPrincipal {
	return &workerPrincipal{
		eventID: eventID,
	}
}
