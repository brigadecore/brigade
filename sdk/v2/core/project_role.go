package core

import libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"

// ProjectRole represents a set of project-level permissions.
type ProjectRole struct {
	// Name is the name of a ProjectRole and has domain-specific meaning.
	Name libAuthz.RoleName `json:"name,omitempty"`
	// ProjectID qualifies the scope of the ProjectRole.
	ProjectID string `json:"projectID,omitempty"`
}
