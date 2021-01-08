package authx

const (
	// RoleNameAdmin is the name of a system-level Role that enables principals to
	// manage Users, ServiceAccounts, and system-level permissions for Users and
	// ServiceAccounts.
	RoleNameAdmin RoleName = "ADMIN"
	// RoleNameEventCreator is the name of a system-level Role that enables
	// principals to create Events for all Projects. This is useful for Event
	// gateways.
	RoleNameEventCreator RoleName = "EVENT_CREATOR"
	// RoleNameProjectAdmin is the name of a project-level Role that enables
	// principals to manage all aspects of a given Project, including the
	// Project's secrets, and project-level permissions for Users and
	// ServiceAccounts.
	RoleNameProjectAdmin RoleName = "PROJECT_ADMIN"
	// RoleNameProjectCreator is the name of a system-level Role that enables
	// principals to create new Projects.
	RoleNameProjectCreator RoleName = "PROJECT_CREATOR"
	// RoleNameProjectDeveloper is the name of a project-level Role that enables
	// principals to update Projects. This Role does NOT enable event creation,
	// secret management, or management of project-level permissions for Users and
	// ServiceAccounts.
	RoleNameProjectDeveloper RoleName = "PROJECT_DEVELOPER"
	// RoleNameProjectUser is the name of a project-level Role that enables
	// principals to create and manage Events for a Project.
	RoleNameProjectUser RoleName = "PROJECT_USER"
	// RoleNameReader is the name of a system-level Role that enables global read
	// access.
	RoleNameReader RoleName = "READER"

	// Special roles
	//
	// These are reserved for use by system components and are NOT assignable to
	// Users and ServiceAccounts.

	// RoleNameObserver is the name of a system-level Role that enables principals
	// to updates Worker and Job status based on observation of the underlying
	// workload execution substrate. This Role exists exclusively for use by
	// Brigade's Observer component.
	RoleNameObserver RoleName = "OBSERVER"
	// RoleNameScheduler is the name of a system-level Role that enables
	// principals to initiate execution of a Worker or Job on the underlying
	// workload execution substrate. This Role exists exclusively for use by
	// Brigade's Scheduler component.
	RoleNameScheduler RoleName = "SCHEDULER"
	// RoleNameWorker is the name of an event-level Role that enables principals
	// to create new Jobs , monitor the status of those Jobs, and access their
	// logs. This Role is exclusively for the use of Brigade Workers.
	RoleNameWorker RoleName = "WORKER"
)

// RoleAdmin returns a system-level Role that enables principals to manage
// Users, ServiceAccounts, and system-level permissions for Users and
// ServiceAccounts.
func RoleAdmin() Role {
	return Role{
		Type: RoleTypeSystem,
		Name: RoleNameAdmin,
	}
}

// RoleEventCreator returns a system-level Role that enables principals to
// create Events for all Projects-- provided the Events have the specified
// source. This is useful for Event gateways, which should be able to create
// Events for all Projects, but should NOT be able to impersonate other
// gateways.
func RoleEventCreator(eventSource string) Role {
	return Role{
		Type:  RoleTypeSystem,
		Name:  RoleNameEventCreator,
		Scope: eventSource,
	}
}

// RoleProjectAdmin returns a project-level Role that enables principals to
// manage all aspects of a given Project, including the Project's secrets, and
// project-level permissions for Users and ServiceAccounts.
func RoleProjectAdmin(projectID string) Role {
	return Role{
		Type:  RoleTypeProject,
		Name:  RoleNameProjectAdmin,
		Scope: projectID,
	}
}

// RoleProjectCreator returns a system-level Role that enables principals to
// create new Projects.
func RoleProjectCreator() Role {
	return Role{
		Type: RoleTypeSystem,
		Name: RoleNameProjectCreator,
	}
}

// RoleProjectDeveloper returns a project-level Role that enables principals to
// update the given Project. This Role does NOT enable event creation, secret
// management, or management of project-level permissions for Users and
// ServiceAccounts.
func RoleProjectDeveloper(projectID string) Role {
	return Role{
		Type:  RoleTypeProject,
		Name:  RoleNameProjectDeveloper,
		Scope: projectID,
	}
}

// RoleProjectUser is the name of a project-level Role that enables principals
// to create and manage Events for the specified Project.
func RoleProjectUser(projectID string) Role {
	return Role{
		Type:  RoleTypeProject,
		Name:  RoleNameProjectUser,
		Scope: projectID,
	}
}

// RoleReader returns a system-level Role that enables global read access.
func RoleReader() Role {
	return Role{
		Type: RoleTypeSystem,
		Name: RoleNameReader,
	}
}

// Special roles
//
// These are reserved for use by system components and are NOT assignable to
// Users and ServiceAccounts.

// RoleObserver returns a system-level Role that enables principals to updates
// Worker and Job status based on observation of the underlying workload
// execution substrate. This Role exists exclusively for use by Brigade's
// Observer component.
func RoleObserver() Role {
	return Role{
		Type: RoleTypeSystem,
		Name: RoleNameObserver,
	}
}

// RoleScheduler returns a system-level Role that enables principals to initiate
// execution of a Worker or Job on the underlying workload execution substrate.
// This Role exists exclusively for use by Brigade's Scheduler component.
func RoleScheduler() Role {
	return Role{
		Type: RoleTypeSystem,
		Name: RoleNameScheduler,
	}
}

// RoleWorker returns an event-level Role that enables principals to create new
// Jobs, monitor the status of those Jobs, and access their logs. This Role is
// exclusively for the use of Brigade Workers.
func RoleWorker(eventID string) Role {
	return Role{
		Type:  RoleTypeEvent,
		Name:  RoleNameWorker,
		Scope: eventID,
	}
}
