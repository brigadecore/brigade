package core

import (
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
)

const (
	// Core-specific, system-level roles...

	// RoleEventCreator represents a system-level Role that enables principals to
	// create Events for all Projects.
	RoleEventCreator libAuthz.Role = "EVENT_CREATOR"

	// RoleProjectCreator represents a system-level Role that enables principals
	// to create new Projects.
	RoleProjectCreator libAuthz.Role = "PROJECT_CREATOR"

	// Core-specific, ProjectRoles...

	// RoleProjectAdmin represents a project-level Role that enables a principal
	// to manage a Project.
	RoleProjectAdmin libAuthz.Role = "PROJECT_ADMIN"

	// RoleProjectDeveloper represents a project-level Role that enables a
	// principal to update a Project.
	RoleProjectDeveloper libAuthz.Role = "PROJECT_DEVELOPER"

	// RoleProjectUser represents a project-level Role that enables a principal to
	// create and manage Events for a Project.
	RoleProjectUser libAuthz.Role = "PROJECT_USER"

	// Special core-specific roles...
	//
	// These are reserved for use by system components and are NOT assignable to
	// Users and ServiceAccounts.

	// RoleObserver represents a system-level Role that enables principals to
	// update Worker and Job status based on observation of the underlying
	// workload execution substrate. This Role exists exclusively for use by
	// Brigade's Observer component.
	RoleObserver libAuthz.Role = "OBSERVER"

	// RoleScheduler represents a system-level Role that enables principals to
	// initiate execution of a Worker or Job on the underlying workload execution
	// substrate. This Role exists exclusively for use by Brigade's Scheduler
	// component.
	RoleScheduler libAuthz.Role = "SCHEDULER"

	// RoleWorker represents an event-level Role that enables principals to create
	// new Jobs, monitor the status of those Jobs, and access their logs. This
	// Role is exclusively for the use of Brigade Workers.
	RoleWorker libAuthz.Role = "WORKER"
)
