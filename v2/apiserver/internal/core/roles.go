package core

import "github.com/brigadecore/brigade/v2/apiserver/internal/authx"

// RoleTypeProject represents a project-level Role.
const RoleTypeProject authx.RoleType = "PROJECT"

const (
	// RoleNameProjectAdmin is the name of a project-level Role that enables
	// principals to manage all aspects of a given Project, including the
	// Project's secrets.
	RoleNameProjectAdmin authx.RoleName = "PROJECT_ADMIN"
	// RoleNameProjectDeveloper is the name of a project-level Role that enables
	// principals to update Projects. This Role does NOT enable event creation
	// or secret management.
	RoleNameProjectDeveloper authx.RoleName = "PROJECT_DEVELOPER"
	// RoleNameProjectUser is the name of a project-level Role that enables
	// principals to create and manage Events for a Project.
	RoleNameProjectUser authx.RoleName = "PROJECT_USER"
)

// RoleProjectAdmin returns a Role that enables a principal to manage a Project
// having an ID field whose value matches that of the Scope field. If the value
// of the Scope field is RoleScopeGlobal ("*"), then the Role is unbounded and
// enables a principal to manage all Projects.
func RoleProjectAdmin(projectID string) authx.Role {
	return authx.Role{
		Type:  RoleTypeProject,
		Name:  RoleNameProjectAdmin,
		Scope: projectID,
	}
}

// RoleProjectDeveloper returns a Role that enables a principal to read and
// update a Project having an ID field whose value matches that of the Scope
// field. If the value of the Scope field is RoleScopeGlobal ("*"), then the
// Role is unbounded and enables a principal to read and update all Projects.
func RoleProjectDeveloper(projectID string) authx.Role {
	return authx.Role{
		Type:  RoleTypeProject,
		Name:  RoleNameProjectDeveloper,
		Scope: projectID,
	}
}

// RoleProjectUser returns a Role that enables a principal read, create, and
// manage Events for a Project having an ID field whose value matches the Scope
// field. If the value of the Scope field is RoleScopeGlobal ("*"), then the
// Role is unbounded and enables a principal to read, create, and manage Events
// for all Projects.
func RoleProjectUser(projectID string) authx.Role {
	return authx.Role{
		Type:  RoleTypeProject,
		Name:  RoleNameProjectUser,
		Scope: projectID,
	}
}
