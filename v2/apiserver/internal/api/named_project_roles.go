package api

const (
	// RoleProjectAdmin represents a project-level Role that enables a principal
	// to manage a Project.
	RoleProjectAdmin Role = "PROJECT_ADMIN"

	// RoleProjectDeveloper represents a project-level Role that enables a
	// principal to update a Project.
	RoleProjectDeveloper Role = "PROJECT_DEVELOPER"

	// RoleProjectUser represents a project-level Role that enables a principal to
	// create and manage Events for a Project.
	RoleProjectUser Role = "PROJECT_USER"
)
