package system

import "github.com/brigadecore/brigade/v2/apiserver/internal/authx"

// RoleTypeSystem represents a system-level Role.
const RoleTypeSystem authx.RoleType = "SYSTEM"

const (
	// RoleNameAdmin is the name of a system-level Role that enables principals to
	// manage Users, ServiceAccounts, and system-level permissions for Users and
	// ServiceAccounts.
	RoleNameAdmin authx.RoleName = "ADMIN"
	// RoleNameEventCreator is the name of a system-level Role that enables
	// principals to create Events for all Projects. This Role is useful for
	// ServiceAccounts used for gateways.
	RoleNameEventCreator authx.RoleName = "EVENT_CREATOR"
	// RoleNameProjectCreator is the name of a system-level Role that enables
	// principals to create new Projects.
	RoleNameProjectCreator authx.RoleName = "PROJECT_CREATOR"
	// RoleNameReader is the name of a system-level Role that enables global read
	// access.
	RoleNameReader authx.RoleName = "READER"

	// Special roles
	//
	// These are reserved for use by system components and are NOT assignable to
	// Users and ServiceAccounts.

	// RoleNameObserver is the name of a system-level Role that enables principals
	// to updates Worker and Job status based on observation of the underlying
	// workload execution substrate. This Role exists exclusively for use by
	// Brigade's Observer component.
	RoleNameObserver authx.RoleName = "OBSERVER"
	// RoleNameScheduler is the name of a system-level Role that enables
	// principals to initiate execution of a Worker or Job on the underlying
	// workload execution substrate. This Role exists exclusively for use by
	// Brigade's Scheduler component.
	RoleNameScheduler authx.RoleName = "SCHEDULER"
	// RoleNameWorker is the name of an event-level Role that enables principals
	// to create new Jobs. This Role is exclusively for the use of Brigade
	// Workers.
	RoleNameWorker authx.RoleName = "WORKER"
)

// RoleAdmin returns a Role that enables a principal to manage Users,
// ServiceAccounts, and globally scoped permissions for Users and
// ServiceAccounts.
func RoleAdmin() authx.Role {
	return authx.Role{
		Type: RoleTypeSystem,
		Name: RoleNameAdmin,
	}
}

// RoleEventCreator returns a Role that enables a principal to create new Events
// having a Source field whose value matches that of the Scope field. This Role
// is useful for ServiceAccounts used for gateways.
func RoleEventCreator(eventSource string) authx.Role {
	return authx.Role{
		Type:  RoleTypeSystem,
		Name:  RoleNameEventCreator,
		Scope: eventSource,
	}
}

// RoleProjectCreator returns a Role that enables a principal to create new
// Projects.
func RoleProjectCreator() authx.Role {
	return authx.Role{
		Type: RoleTypeSystem,
		Name: RoleNameProjectCreator,
	}
}

// RoleReader returns a Role that enables a principal to list and read Projects,
// Events, Users, and Service Accounts.
func RoleReader() authx.Role {
	return authx.Role{
		Type: RoleTypeSystem,
		Name: RoleNameReader,
	}
}

// Special roles
//
// These are reserved for use by system components and are NOT assignable to
// Users and ServiceAccounts.

// RoleObserver returns a Role that enables a principal to update Worker and Job
// statuses based on observations of the underlying workload execution
// substrate. This Role is exclusively for the use of the Observer component.
func RoleObserver() authx.Role {
	return authx.Role{
		Type: RoleTypeSystem,
		Name: RoleNameObserver,
	}
}

// RoleScheduler returns a Role that enables a principal to initiate execution
// of Workers and Jobs on the underlying workload execution substrate. This Role
// is exclusively for the use of the Scheduler component.
func RoleScheduler() authx.Role {
	return authx.Role{
		Type: RoleTypeSystem,
		Name: RoleNameScheduler,
	}
}

// RoleWorker returns a Role that enables a principal to create Jobs for the
// Event whose ID matches the Scope. This Role is exclusively for the use of
// Workers.
func RoleWorker(eventID string) authx.Role {
	return authx.Role{
		Type:  RoleTypeSystem,
		Name:  RoleNameWorker,
		Scope: eventID,
	}
}
