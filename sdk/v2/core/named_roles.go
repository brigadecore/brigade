package core

import (
	libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"
)

const (
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
