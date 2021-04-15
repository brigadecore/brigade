package core

import libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"

// ProjectRoleScopeGlobal represents an unbounded project scope.
const ProjectRoleScopeGlobal = "*"

// ProjectRole represents a set of project-level permissions.
type ProjectRole struct {
	// Name is the name of a ProjectRole and has domain-specific meaning.
	Name libAuthz.RoleName `json:"name,omitempty"`
}
