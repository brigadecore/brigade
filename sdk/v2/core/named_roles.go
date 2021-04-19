package core

import (
	libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"
)

const (
	// Core-specific, system-level roles...

	// RoleEventCreator is the name of a system-level Role that enables principals
	// to create Events for all Projects-- provided the Events have a specific
	// value in the Source field. This is useful for Event gateways, which should
	// be able to create Events for all Projects, but should NOT be able to
	// impersonate other gateways.
	RoleEventCreator libAuthz.Role = "EVENT_CREATOR"

	// RoleProjectCreator is the name of a system-level Role that enables
	// principals to create new Projects.
	RoleProjectCreator libAuthz.Role = "PROJECT_CREATOR"

	// Core-specific, project-level roles...

	// RoleProjectAdmin is the name of a project-level Role that enables a
	// principal to manage a specific Project.
	RoleProjectAdmin libAuthz.Role = "PROJECT_ADMIN"

	// RoleProjectDeveloper is the name of a project-level Role that enables a
	// principal to update a specific project.
	RoleProjectDeveloper libAuthz.Role = "PROJECT_DEVELOPER"

	// RoleProjectUser is the name of project-level Role that enables a principal
	// to create and manage Events for a specific Project.
	RoleProjectUser libAuthz.Role = "PROJECT_USER"
)
