package core

import (
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/system"
)

// Core-specific, system-level roles...

// RoleEventCreator returns a system-level Role that enables principals to
// create Events for all Projects.
func RoleEventCreator() libAuthz.Role {
	return libAuthz.Role{
		Type: system.RoleTypeSystem,
		Name: "EVENT_CREATOR",
	}
}

// RoleProjectCreator returns a system-level Role that enables principals to
// create new Projects.
func RoleProjectCreator() libAuthz.Role {
	return libAuthz.Role{
		Type: system.RoleTypeSystem,
		Name: "PROJECT_CREATOR",
	}
}

// Core-specific, project-level roles...

// RoleProjectAdmin returns a project-level Role that enables a principal to
// manage a Project.
func RoleProjectAdmin() libAuthz.Role {
	return libAuthz.Role{
		Type: RoleTypeProject,
		Name: "ADMIN",
	}
}

// RoleProjectDeveloper returns a project-level Role that enables a principal to
// update a Project.
func RoleProjectDeveloper() libAuthz.Role {
	return libAuthz.Role{
		Type: RoleTypeProject,
		Name: "DEVELOPER",
	}
}

// RoleProjectUser returns a project-level Role that enables a principal to
// create and manage Events for a Project.
func RoleProjectUser() libAuthz.Role {
	return libAuthz.Role{
		Type: RoleTypeProject,
		Name: "USER",
	}
}

// Special core-specific roles...
//
// These are reserved for use by system components and are NOT assignable to
// Users and ServiceAccounts.

// RoleObserver returns a system-level Role that enables principals to update
// Worker and Job status based on observation of the underlying workload
// execution substrate. This Role exists exclusively for use by Brigade's
// Observer component.
func RoleObserver() libAuthz.Role {
	return libAuthz.Role{
		Type: system.RoleTypeSystem,
		Name: "OBSERVER",
	}
}

// RoleScheduler returns a system-level Role that enables principals to initiate
// execution of a Worker or Job on the underlying workload execution substrate.
// This Role exists exclusively for use by Brigade's Scheduler component.
func RoleScheduler() libAuthz.Role {
	return libAuthz.Role{
		Type: system.RoleTypeSystem,
		Name: "SCHEDULER",
	}
}

// RoleWorker returns an event-level Role that enables principals to create new
// Jobs, monitor the status of those Jobs, and access their logs. This Role is
// exclusively for the use of Brigade Workers.
func RoleWorker() libAuthz.Role {
	return libAuthz.Role{
		Type: system.RoleTypeSystem,
		Name: "WORKER",
	}
}
