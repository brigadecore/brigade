package authn

import (
	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/system"
)

var (
	// Root is a singleton that represents Brigade's "root" user.
	root = &rootPrincipal{}
	// Scheduler is a singleton that represents Brigade's scheduler component.
	scheduler = &schedulerPrincipal{}
	// Observer is a singleton that represents Brigade's observer component.
	observer = &observerPrincipal{}
)

// rootPrincipal is an implementation of the libAuthn.Principal interface for
// the "root" user.
type rootPrincipal struct{}

func (r *rootPrincipal) RoleAssignments() []libAuthz.RoleAssignment {
	return []libAuthz.RoleAssignment{
		{Role: system.RoleAdmin()},
		{Role: system.RoleReader()},
		{
			Role:  core.RoleEventCreator(),
			Scope: libAuthz.RoleScopeGlobal,
		},
		{
			Role:  core.RoleProjectAdmin(),
			Scope: libAuthz.RoleScopeGlobal,
		},
		{Role: core.RoleProjectCreator()},
		{
			Role:  core.RoleProjectDeveloper(),
			Scope: libAuthz.RoleScopeGlobal,
		},
		{
			Role:  core.RoleProjectUser(),
			Scope: libAuthz.RoleScopeGlobal,
		},
	}
}

// schedulerPrincipal is an implementation of the libAuthn.Principal interface
// that represents the scheduler component, which is a special class of user
// because, although it cannot do much, it has the UNIQUE ability to launch
// Workers and Jobs.
type schedulerPrincipal struct{}

func (s *schedulerPrincipal) RoleAssignments() []libAuthz.RoleAssignment {
	return []libAuthz.RoleAssignment{
		{Role: system.RoleReader()},
		{Role: core.RoleScheduler()},
	}
}

// observerPrincipal is an implementation of the libAuthn.Principal interface
// that represents the observer component, which is a special class of user
// because, although it cannot do much, it has the UNIQUE ability to update
// Worker and Job statuses.
type observerPrincipal struct{}

func (o *observerPrincipal) RoleAssignments() []libAuthz.RoleAssignment {
	return []libAuthz.RoleAssignment{
		{Role: system.RoleReader()},
		{Role: core.RoleObserver()},
	}
}

// workerPrincipal is an implementation of the libAuthn.Principal interface that
// represents an Event's Worker, which is a special class of user because,
// although it cannot do much, it has the UNIQUE ability to create new Jobs.
type workerPrincipal struct {
	eventID string
}

func (w *workerPrincipal) RoleAssignments() []libAuthz.RoleAssignment {
	return []libAuthz.RoleAssignment{
		{Role: system.RoleReader()},
		{
			Role:  core.RoleWorker(),
			Scope: w.eventID,
		},
	}
}

// worker returns an libAuthn.Principal that represents the specified Event's
// Worker.
func worker(eventID string) *workerPrincipal {
	return &workerPrincipal{
		eventID: eventID,
	}
}
