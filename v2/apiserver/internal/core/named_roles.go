package core

import "github.com/brigadecore/brigade/v2/apiserver/internal/authx"

// Core-specific, system-level roles...

// RoleEventCreator returns a system-level Role that enables principals to
// create Events for all Projects-- provided the Events have a value in the
// Source field that matches the value in this Role's Scope field. This is
// useful for Event gateways, which should be able to create Events for all
// Projects, but should NOT be able to impersonate other gateways.
func RoleEventCreator(eventSource string) authx.Role {
	return authx.Role{
		Type:  authx.RoleTypeSystem,
		Name:  "EVENT_CREATOR",
		Scope: eventSource,
	}
}

// RoleProjectCreator returns a system-level Role that enables principals to
// create new Projects.
func RoleProjectCreator() authx.Role {
	return authx.Role{
		Type: authx.RoleTypeSystem,
		Name: "PROJECT_CREATOR",
	}
}

// Core-specific, project-level roles...

// RoleProjectAdmin returns a project-level Role that enables a principal to
// manage the Project whose ID matches the value of the Scope field. If the
// value of the Scope field is RoleScopeGlobal ("*"), then the Role is unbounded
// and enables a principal to manage all Projects.
func RoleProjectAdmin(projectID string) authx.Role {
	return authx.Role{
		Type:  RoleTypeProject,
		Name:  "ADMIN",
		Scope: projectID,
	}
}

// RoleProjectDeveloper returns a project-level Role that enables a principal to
// update the Project whose ID matches the value of the Scope field. If the
// value of the Scope field is RoleScopeGlobal ("*"), then the Role is unbounded
// and enables a principal to update all Projects.
func RoleProjectDeveloper(projectID string) authx.Role {
	return authx.Role{
		Type:  RoleTypeProject,
		Name:  "DEVELOPER",
		Scope: projectID,
	}
}

// RoleProjectUser returns a project-level Role that enables a principal to
// create and manage Events for the Project whose ID matches the value of the
// Scope field. If the value of the Scope field is RoleScopeGlobal ("*"), then
// the Role is unbounded and enables a principal to create and manage Events for
// all Projects.
func RoleProjectUser(projectID string) authx.Role {
	return authx.Role{
		Type:  RoleTypeProject,
		Name:  "USER",
		Scope: projectID,
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
func RoleObserver() authx.Role {
	return authx.Role{
		Type: authx.RoleTypeSystem,
		Name: "OBSERVER",
	}
}

// RoleScheduler returns a system-level Role that enables principals to initiate
// execution of a Worker or Job on the underlying workload execution substrate.
// This Role exists exclusively for use by Brigade's Scheduler component.
func RoleScheduler() authx.Role {
	return authx.Role{
		Type: authx.RoleTypeSystem,
		Name: "SCHEDULER",
	}
}

// RoleWorker returns an event-level Role that enables principals to create new
// Jobs, monitor the status of those Jobs, and access their logs. This Role is
// exclusively for the use of Brigade Workers.
func RoleWorker(eventID string) authx.Role {
	return authx.Role{
		Type:  authx.RoleTypeSystem,
		Name:  "WORKER",
		Scope: eventID,
	}
}
