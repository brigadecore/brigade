package core

import libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"

// ProjectRoleScopeGlobal represents an unbounded project scope.
const ProjectRoleScopeGlobal = "*"

// ProjectRole represents a set of project-level permissions.
type ProjectRole struct {
	// Name is the name of a Role and has domain-specific meaning.
	Name libAuthz.RoleName `json:"name,omitempty" bson:"name,omitempty"`
}
