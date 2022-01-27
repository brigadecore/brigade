package sdk

const (
	// Core-specific, project-level roles...

	// RoleProjectAdmin is the name of a project-level Role that enables a
	// principal to manage a specific Project.
	RoleProjectAdmin Role = "PROJECT_ADMIN"

	// RoleProjectDeveloper is the name of a project-level Role that enables a
	// principal to update a specific project.
	RoleProjectDeveloper Role = "PROJECT_DEVELOPER"

	// RoleProjectUser is the name of project-level Role that enables a principal
	// to create and manage Events for a specific Project.
	RoleProjectUser Role = "PROJECT_USER"
)
