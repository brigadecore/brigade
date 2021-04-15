package core

import (
	libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"
)

const (
	// Core-specific, system-level roles...

	// RoleNameEventCreator is the name of a system-level Role that enables
	// principals to create Events for all Projects-- provided the Events have a
	// specific value in the Source field. This is useful for Event gateways,
	// which should be able to create Events for all Projects, but should NOT be
	// able to impersonate other gateways.
	RoleNameEventCreator libAuthz.RoleName = "EVENT_CREATOR"

	// RoleNameProjectCreator is the name of a system-level Role that enables
	// principals to create new Projects.
	RoleNameProjectCreator libAuthz.RoleName = "PROJECT_CREATOR"

	// Core-specific, project-level roles...

	// RoleNameProjectAdmin is the name of a project-level Role that enables a
	// principal to manage a specific Project.
	RoleNameProjectAdmin libAuthz.RoleName = "ADMIN"

	// RoleNameProjectDeveloper is the name of a project-level Role that enables a
	// principal to update a specific project.
	RoleNameProjectDeveloper libAuthz.RoleName = "DEVELOPER"

	// RoleNameProjectUser is the name of project-level Role that enables a
	// principal to create and manage Events for a specific Project.
	RoleNameProjectUser libAuthz.RoleName = "USER"
)

// Core-specific, system-level roles...

// RoleEventCreator returns a system-level Role that enables principals to
// create Events for all Projects.
func RoleEventCreator() libAuthz.Role {
	return libAuthz.Role{
		Name: RoleNameEventCreator,
	}
}

// RoleProjectCreator returns a system-level Role that enables principals to
// create new Projects.
func RoleProjectCreator() libAuthz.Role {
	return libAuthz.Role{
		Name: RoleNameProjectCreator,
	}
}

// Core-specific, project-level roles...

// RoleProjectAdmin returns a ProjectRole that enables a principal to manage a
// Project.
func RoleProjectAdmin() ProjectRole {
	return ProjectRole{
		Name: RoleNameProjectAdmin,
	}
}

// RoleProjectDeveloper returns a ProjectRole that enables a principal to update
// a Project.
func RoleProjectDeveloper() ProjectRole {
	return ProjectRole{
		Name: RoleNameProjectDeveloper,
	}
}

// RoleProjectUser returns a ProjectRole that enables a principal to create and
// manage Events for a Project.
func RoleProjectUser() ProjectRole {
	return ProjectRole{
		Name: RoleNameProjectUser,
	}
}
